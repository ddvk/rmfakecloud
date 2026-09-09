package ui

import (
	"bytes"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type pageUser struct {
	ID    string
	Email string
	Admin bool
}

func (app *ReactAppWrapper) optionalUser(c *gin.Context) *pageUser {
	token, err := c.Cookie(cookieName)
	if err != nil || token == "" {
		return nil
	}
	claims := &WebUserClaims{}
	if err := common.ClaimsFromToken(claims, token, app.cfg.JWTSecretKey); err != nil {
		return nil
	}
	admin := false
	for _, r := range claims.Roles {
		if r == AdminRole {
			admin = true
			break
		}
	}
	return &pageUser{ID: claims.UserID, Email: claims.Email, Admin: admin}
}

func (app *ReactAppWrapper) requirePageUser(c *gin.Context) *pageUser {
	u := app.optionalUser(c)
	if u == nil {
		next := url.QueryEscape(c.Request.URL.RequestURI())
		c.Redirect(http.StatusFound, "/login?next="+next)
		c.Abort()
		return nil
	}
	// Populate gin context for backends that expect auth middleware keys
	app.hydrateAuthContext(c, u)
	return u
}

func (app *ReactAppWrapper) hydrateAuthContext(c *gin.Context, u *pageUser) {
	token, _ := c.Cookie(cookieName)
	claims := &WebUserClaims{}
	_ = common.ClaimsFromToken(claims, token, app.cfg.JWTSecretKey)
	uid := common.SanitizeUid(claims.UserID)
	c.Set(userIDContextKey, uid)
	c.Set(browserIDContextKey, claims.BrowserID)
	c.Set(backendVersionKey, common.Sync10)
	for _, s := range strings.Fields(claims.Scopes) {
		if s == isSync15Key {
			c.Set(backendVersionKey, common.Sync15)
		}
	}
	if u.Admin {
		c.Set(AdminRole, true)
	}
}

func (app *ReactAppWrapper) loadUserTheme(c *gin.Context, u *pageUser) (themeXML []byte, css string, chrome string, formFactor string) {
	formFactor = clientFormFactor(c.Request)
	themeID := "default"
	var overrides map[string]string
	if u != nil {
		user, err := app.userStorer.GetUser(u.ID)
		if err == nil && user != nil {
			if user.ThemeID != "" {
				themeID = user.ThemeID
			}
			overrides = user.ThemeColorOverrides
		}
	}
	xmlBytes, _, err := app.themes.getXML(themeID)
	if err != nil {
		xmlBytes, _, _ = app.themes.getXML("default")
	}
	xmlBytes = applyColorOverridesXML(xmlBytes, overrides)
	css, err = themeToCSS(xmlBytes, formFactor)
	if err != nil {
		log.Warn("theme css: ", err)
		css = ""
	}
	chrome = chromeStyleFromTheme(xmlBytes, formFactor)
	return xmlBytes, css, chrome, formFactor
}

func defaultNav(path string, admin bool) []navItem {
	items := []navItem{
		{ID: "documents", Href: "/documents", Label: "Documents", Icon: "folder"},
		{ID: "integrations", Href: "/integrations", Label: "Integrations", Icon: "puzzle"},
		{ID: "connect", Href: "/connect", Label: "Connect", Icon: "link"},
		{ID: "screenshare", Href: "/screenshare", Label: "Screen share", Icon: "display"},
		{ID: "admin", Href: "/admin", Label: "Admin", Icon: "gear", AdminOnly: true},
		{ID: "help", Href: "/help", Label: "Help", Icon: "book"},
		{ID: "profile", Href: "/profile", Label: "Profile", Icon: "person"},
	}
	for i := range items {
		if items[i].Href == path || (items[i].Href != "/" && strings.HasPrefix(path, items[i].Href)) {
			items[i].Active = true
		}
		if path == "/" && items[i].ID == "documents" {
			// home is not documents
			items[i].Active = false
		}
	}
	_ = admin
	return items
}

type navItem struct {
	ID, Href, Label, Icon string
	AdminOnly, Active     bool
}

func (app *ReactAppWrapper) renderPage(c *gin.Context, pageXML []byte) {
	setFormFactorHeaders(c)
	html, ok, err := app.transformPageXML(pageXML)
	if err != nil {
		log.Warn("server xslt failed, using browser bootstrap: ", err)
	}
	if ok && len(html) > 0 {
		c.Data(http.StatusOK, "text/html; charset=utf-8", html)
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", bootstrapXSLTHTML(pageXML))
}

func (app *ReactAppWrapper) flashFromQuery(c *gin.Context) (ftype, msg string) {
	ftype = c.Query("flash")
	msg = c.Query("msg")
	return
}

func redirectFlash(c *gin.Context, path, ftype, msg string) {
	u := path
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	c.Redirect(http.StatusSeeOther, u+sep+"flash="+url.QueryEscape(ftype)+"&msg="+url.QueryEscape(msg))
}

func xmlCDATA(s string) string {
	return "<![CDATA[" + strings.ReplaceAll(s, "]]>", "]]]]><![CDATA[>") + "]]>"
}

func writePageOpen(b *bytes.Buffer, kind, title, path, chrome, css string, u *pageUser, flashType, flashMsg string, nav []navItem) {
	fmt.Fprintf(b, `<?xml version="1.0" encoding="UTF-8"?>`)
	fmt.Fprintf(b, `<page kind="%s" title="%s" path="%s" chrome-style="%s">`,
		xmlAttr(kind), xmlAttr(title), xmlAttr(path), xmlAttr(chrome))
	if u != nil {
		admin := "false"
		if u.Admin {
			admin = "true"
		}
		fmt.Fprintf(b, `<user id="%s" email="%s" admin="%s"/>`, xmlAttr(u.ID), xmlAttr(u.Email), admin)
	}
	if flashMsg != "" {
		if flashType == "" {
			flashType = "info"
		}
		fmt.Fprintf(b, `<flash type="%s">%s</flash>`, xmlAttr(flashType), esc(flashMsg))
	}
	if css != "" {
		fmt.Fprintf(b, `<theme-css>%s</theme-css>`, xmlCDATA(css))
	}
	b.WriteString(`<nav brand-position="start" user-menu="end">`)
	for _, it := range nav {
		adminOnly := ""
		if it.AdminOnly {
			adminOnly = ` admin-only="true"`
		}
		active := ""
		if it.Active {
			active = ` active="true"`
		}
		fmt.Fprintf(b, `<item id="%s" href="%s" label="%s" icon="%s"%s%s/>`,
			xmlAttr(it.ID), xmlAttr(it.Href), xmlAttr(it.Label), xmlAttr(it.Icon), adminOnly, active)
	}
	b.WriteString(`</nav>`)
}

func writePageClose(b *bytes.Buffer) {
	b.WriteString(`</page>`)
}

func esc(s string) string {
	return html.EscapeString(s)
}

func xmlAttr(s string) string {
	return html.EscapeString(s)
}

func (app *ReactAppWrapper) getModelUser(id string) *model.User {
	u, err := app.userStorer.GetUser(id)
	if err != nil {
		return nil
	}
	return u
}
