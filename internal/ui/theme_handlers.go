package ui

import (
	"net/http"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/gin-gonic/gin"
)

type profileThemeRequest struct {
	ThemeID              string            `json:"themeId"`
	ThemeColorOverrides  map[string]string `json:"themeColorOverrides"`
}

type saveThemeRequest struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Published bool   `json:"published"`
	XML       string `json:"xml"`
}

func (app *ReactAppWrapper) listThemes(c *gin.Context) {
	admin := IsAdmin(c)
	list, err := app.themes.list(admin)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, list)
}

func isAdminContext(c *gin.Context) bool {
	return IsAdmin(c)
}

func (app *ReactAppWrapper) getThemePublic(c *gin.Context) {
	id := c.Param("id")
	xmlBytes, builtin, err := app.themes.getXML(id)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	root, err := parseThemeRoot(xmlBytes)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	published := strings.EqualFold(root.Published, "true") || root.Published == ""
	if !published && !app.tryIsAdmin(c) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	setFormFactorHeaders(c)
	c.JSON(http.StatusOK, gin.H{
		"id":         id,
		"name":       root.Name,
		"published":  published,
		"builtin":    builtin,
		"xml":        string(xmlBytes),
		"formFactor": clientFormFactor(c.Request),
		"assets": gin.H{
			"cssXsl":    "/ui/api/themes/assets/theme-to-css.xsl",
			"layoutXsl": "/ui/api/themes/assets/theme-to-layout.xsl",
		},
	})
}

func (app *ReactAppWrapper) tryIsAdmin(c *gin.Context) bool {
	token, err := c.Cookie(cookieName)
	if err != nil {
		token, err = common.GetToken(c)
	}
	if err != nil || token == "" {
		return false
	}
	claims := &WebUserClaims{}
	if err := common.ClaimsFromToken(claims, token, app.cfg.JWTSecretKey); err != nil {
		return false
	}
	for _, r := range claims.Roles {
		if r == AdminRole {
			return true
		}
	}
	return false
}

func (app *ReactAppWrapper) updateTheme(c *gin.Context) {
	id := c.Param("id")
	var req saveThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.XML == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if req.ID == "" {
		req.ID = id
	}
	if req.ID != id {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "id mismatch"})
		return
	}
	if err := app.themes.save(req.ID, req.Name, req.Published, []byte(req.XML)); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "id": req.ID})
}

func (app *ReactAppWrapper) getThemeAsset(c *gin.Context) {
	name := c.Param("name")
	b, err := app.themes.asset(name)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	ct := "application/xslt+xml"
	if strings.HasSuffix(name, ".xml") {
		ct = "application/xml"
	}
	c.Data(http.StatusOK, ct, b)
}

func (app *ReactAppWrapper) saveTheme(c *gin.Context) {
	var req saveThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ID == "" || req.XML == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if err := app.themes.save(req.ID, req.Name, req.Published, []byte(req.XML)); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "id": req.ID})
}

func (app *ReactAppWrapper) publishTheme(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Published bool `json:"published"`
	}
	body.Published = true
	_ = c.ShouldBindJSON(&body)
	if err := app.themes.setPublished(id, body.Published); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (app *ReactAppWrapper) deleteTheme(c *gin.Context) {
	id := c.Param("id")
	if err := app.themes.delete(id); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (app *ReactAppWrapper) getProfileTheme(c *gin.Context) {
	uid := userID(c)
	user, err := app.userStorer.GetUser(uid)
	if err != nil || user == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	themeID := user.ThemeID
	if themeID == "" {
		themeID = "default"
	}
	c.JSON(http.StatusOK, gin.H{
		"themeId":             themeID,
		"themeColorOverrides": user.ThemeColorOverrides,
		"formFactor":          clientFormFactor(c.Request),
	})
}

func (app *ReactAppWrapper) putProfileTheme(c *gin.Context) {
	var req profileThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if req.ThemeID == "" {
		req.ThemeID = "default"
	}
	xmlBytes, _, err := app.themes.getXML(req.ThemeID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "unknown theme"})
		return
	}
	root, err := parseThemeRoot(xmlBytes)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	published := strings.EqualFold(root.Published, "true") || root.Published == ""
	if !published && !isAdminContext(c) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "theme is not published"})
		return
	}
	uid := userID(c)
	user, err := app.userStorer.GetUser(uid)
	if err != nil || user == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	user.ThemeID = req.ThemeID
	user.ThemeColorOverrides = sanitizeColorOverrides(req.ThemeColorOverrides)
	if err := app.userStorer.UpdateUser(user); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"themeId":             user.ThemeID,
		"themeColorOverrides": user.ThemeColorOverrides,
	})
}

func sanitizeColorOverrides(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	allowed := map[string]bool{
		"background1": true, "background2": true,
		"foreground1": true, "foreground2": true, "foreground3": true,
		"action": true, "accept": true, "reject": true,
	}
	out := make(map[string]string)
	for k, v := range in {
		if !allowed[k] {
			continue
		}
		v = strings.TrimSpace(v)
		if v == "" || !isSafeCSSColor(v) {
			continue
		}
		out[k] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func isSafeCSSColor(v string) bool {
	if len(v) > 32 {
		return false
	}
	// #rgb, #rrggbb, or simple rgb()/named short
	if strings.HasPrefix(v, "#") {
		for _, ch := range v[1:] {
			if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
				return false
			}
		}
		return len(v) == 4 || len(v) == 7
	}
	return false
}
