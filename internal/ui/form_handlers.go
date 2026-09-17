package ui

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/integrations"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var errEmptyLogin = errors.New("user not found")

func (app *ReactAppWrapper) findUserForLogin(idOrEmail string) (*model.User, error) {
	idOrEmail = strings.TrimSpace(idOrEmail)
	if idOrEmail == "" {
		return nil, errEmptyLogin
	}
	if user, err := app.userStorer.GetUser(idOrEmail); err == nil && user != nil {
		return user, nil
	}
	users, err := app.userStorer.GetUsers()
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		if u == nil {
			continue
		}
		if strings.EqualFold(u.ID, idOrEmail) || strings.EqualFold(u.Email, idOrEmail) {
			return u, nil
		}
	}
	return nil, errEmptyLogin
}

func (app *ReactAppWrapper) formLogin(c *gin.Context) {
	email := strings.TrimSpace(c.PostForm("email"))
	password := c.PostForm("password")
	next := c.Query("next")
	if next == "" {
		next = c.PostForm("next")
	}
	if next == "" || !strings.HasPrefix(next, "/") {
		next = "/"
	}

	if app.cfg.CreateFirstUser {
		user, err := model.NewUser(email, password)
		if err == nil {
			user.IsAdmin = true
			if err := app.userStorer.RegisterUser(user); err == nil {
				app.cfg.CreateFirstUser = false
			}
		}
	}

	user, err := app.findUserForLogin(email)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login?error="+urlQuery("Invalid email or password")+"&email="+urlQuery(email))
		return
	}
	ok, err := user.CheckPassword(password)
	if err != nil || !ok {
		c.Redirect(http.StatusSeeOther, "/login?error="+urlQuery("Invalid email or password")+"&email="+urlQuery(email))
		return
	}
	tokenString, expiresAfter, err := app.issueWebTokenForUser(user, uuid.NewString())
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login?error="+urlQuery("Login failed"))
		return
	}
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(cookieName, tokenString, int(expiresAfter.Seconds()), "/", "", app.cfg.HTTPSCookie, true)
	c.Redirect(http.StatusSeeOther, next)
}

func urlQuery(s string) string {
	return url.QueryEscape(s)
}

func (app *ReactAppWrapper) formLogout(c *gin.Context) {
	c.SetCookie(cookieName, "/", -1, "", "", false, true)
	c.Redirect(http.StatusSeeOther, "/login")
}

func (app *ReactAppWrapper) formProfileTheme(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	user := app.getModelUser(u.ID)
	if user == nil {
		redirectFlash(c, "/profile", "error", "User not found")
		return
	}
	themeID := c.PostForm("themeId")
	if themeID == "" {
		themeID = "default"
	}
	if _, _, err := app.themes.getXML(themeID); err != nil {
		redirectFlash(c, "/profile", "error", "Unknown theme")
		return
	}
	overrides := map[string]string{}
	for k, vals := range c.Request.PostForm {
		if strings.HasPrefix(k, "override-") && len(vals) > 0 && vals[0] != "" {
			overrides[strings.TrimPrefix(k, "override-")] = vals[0]
		}
	}
	user.ThemeID = themeID
	user.ThemeColorOverrides = overrides
	if err := app.userStorer.UpdateUser(user); err != nil {
		redirectFlash(c, "/profile", "error", "Failed to save theme")
		return
	}
	redirectFlash(c, "/profile", "success", "Theme deployed")
}

func (app *ReactAppWrapper) formProfilePassword(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	user := app.getModelUser(u.ID)
	if user == nil {
		redirectFlash(c, "/profile", "error", "User not found")
		return
	}
	cur := c.PostForm("currentPassword")
	neu := c.PostForm("newPassword")
	ok, err := user.CheckPassword(cur)
	if err != nil || !ok {
		redirectFlash(c, "/profile", "error", "Current password incorrect")
		return
	}
	if neu == "" {
		redirectFlash(c, "/profile", "error", "New password required")
		return
	}
	user.SetPassword(neu)
	if err := app.userStorer.UpdateUser(user); err != nil {
		redirectFlash(c, "/profile", "error", "Failed to update password")
		return
	}
	redirectFlash(c, "/profile", "success", "Password updated")
}

func (app *ReactAppWrapper) formPasskeyDelete(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	raw, err := model.ParseCredentialIDBase64(c.Param("id"))
	if err != nil {
		redirectFlash(c, "/profile", "error", "Invalid passkey id")
		return
	}
	user := app.getModelUser(u.ID)
	if user == nil {
		redirectFlash(c, "/profile", "error", "User not found")
		return
	}
	if !user.RemoveWebAuthnCredential(raw) {
		redirectFlash(c, "/profile", "error", "Passkey not found")
		return
	}
	if err := app.userStorer.UpdateUser(user); err != nil {
		redirectFlash(c, "/profile", "error", "Failed to delete passkey")
		return
	}
	redirectFlash(c, "/profile", "success", "Passkey deleted")
}

func (app *ReactAppWrapper) formReissueDevice(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	if app.issueDeviceToken == nil {
		redirectFlash(c, "/profile", "error", "Device token signing not configured")
		return
	}
	deviceID := strings.TrimSpace(c.PostForm("deviceId"))
	if deviceID == "" {
		redirectFlash(c, "/profile", "error", "Device id required")
		return
	}
	user, err := app.userStorer.GetUser(u.ID)
	if err != nil || user == nil {
		redirectFlash(c, "/profile", "error", "User not found")
		return
	}
	reg, ok := user.GetRegisteredDevice(deviceID)
	if !ok {
		redirectFlash(c, "/profile", "error", "Device not registered")
		return
	}
	token, err := app.issueDeviceToken(u.ID, deviceID, reg.DeviceDesc)
	if err != nil {
		log.Error(err)
		redirectFlash(c, "/profile", "error", "Could not issue device token")
		return
	}
	user.UpsertRegisteredDevice(deviceID, reg.DeviceDesc, reg.DeviceLink)
	if err := app.userStorer.UpdateUser(user); err != nil {
		log.Warn("reissue device persist: ", err)
	}

	_, css, chrome, _ := app.loadUserTheme(c, u)
	var b bytes.Buffer
	writePageOpen(&b, "device-token", "Device token — rmfakecloud", "/profile", chrome, css, u, "", "", defaultNav("/profile", u.Admin))
	b.WriteString(`<body><device-token`)
	fmt.Fprintf(&b, ` device-id="%s" device-desc="%s" model="%s">`,
		xmlAttr(reg.DeviceID), xmlAttr(reg.DeviceDesc), xmlAttr(deviceModelLabel(reg)))
	fmt.Fprintf(&b, `<token>%s</token>`, esc(token))
	b.WriteString(`</device-token></body>`)
	writePageClose(&b)
	app.renderPage(c, b.Bytes())
}

func deviceModelLabel(d model.RegisteredDevice) string {
	if strings.TrimSpace(d.Model) != "" {
		return d.Model
	}
	if strings.TrimSpace(d.DeviceDesc) != "" {
		return d.DeviceDesc
	}
	return "Unknown"
}

func (app *ReactAppWrapper) formUploadDocument(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	backend := app.getBackend(c)
	parentID := c.PostForm("parent")
	if parentID == "root" {
		parentID = ""
	}
	file, err := c.FormFile("file")
	if err != nil {
		redirectFlash(c, documentsRedirect(parentID), "error", "No file uploaded")
		return
	}
	ext := strings.ToLower(path.Ext(file.Filename))
	if ext == storage.TemplateFileExt {
		if !u.Admin {
			redirectFlash(c, documentsRedirect(parentID), "error", "Only admins can upload templates")
			return
		}
		f, err := file.Open()
		if err != nil {
			redirectFlash(c, "/admin/templates", "error", "Cannot open upload")
			return
		}
		defer f.Close()
		if _, err := backend.CreateDocument(u.ID, file.Filename, "", f); err != nil {
			log.Error(err)
			redirectFlash(c, "/admin/templates", "error", err.Error())
			return
		}
		backend.Sync(u.ID)
		redirectFlash(c, "/admin/templates", "success", "Template uploaded")
		return
	}
	f, err := file.Open()
	if err != nil {
		redirectFlash(c, documentsRedirect(parentID), "error", "Cannot open upload")
		return
	}
	defer f.Close()
	if _, err := backend.CreateDocument(u.ID, file.Filename, parentID, f); err != nil {
		log.Error(err)
		redirectFlash(c, documentsRedirect(parentID), "error", err.Error())
		return
	}
	backend.Sync(u.ID)
	redirectFlash(c, documentsRedirect(parentID), "success", "Uploaded")
}

func (app *ReactAppWrapper) formCreateFolder(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	backend := app.getBackend(c)
	parentID := c.PostForm("parent")
	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		redirectFlash(c, documentsRedirect(parentID), "error", "Folder name required")
		return
	}
	if _, err := backend.CreateFolder(u.ID, name, parentID); err != nil {
		redirectFlash(c, documentsRedirect(parentID), "error", err.Error())
		return
	}
	backend.Sync(u.ID)
	redirectFlash(c, documentsRedirect(parentID), "success", "Folder created")
}

func (app *ReactAppWrapper) formDeleteDocument(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	docid := c.Param("docid")
	parentID := c.PostForm("parent")
	if app.isSyncedTemplateDoc(c, u.ID, docid) {
		if !u.Admin {
			redirectFlash(c, documentsRedirect(parentID), "error", "Only admins can delete templates")
			return
		}
		backend := app.getBackend(c)
		if err := backend.DeleteDocument(u.ID, docid); err != nil {
			redirectFlash(c, "/admin/templates", "error", err.Error())
			return
		}
		backend.Sync(u.ID)
		redirectFlash(c, "/admin/templates", "success", "Deleted")
		return
	}
	backend := app.getBackend(c)
	if err := backend.DeleteDocument(u.ID, docid); err != nil {
		redirectFlash(c, documentsRedirect(parentID), "error", err.Error())
		return
	}
	backend.Sync(u.ID)
	redirectFlash(c, documentsRedirect(parentID), "success", "Deleted")
}

func (app *ReactAppWrapper) formUpdateDocument(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	docid := c.Param("docid")
	name := strings.TrimSpace(c.PostForm("name"))
	parentID := c.PostForm("parent")
	redirectParent := c.PostForm("redirect")
	if redirectParent == "" {
		redirectParent = parentID
	}
	if name == "" {
		redirectFlash(c, documentsRedirect(redirectParent), "error", "Name required")
		return
	}
	if app.isSyncedTemplateDoc(c, u.ID, docid) {
		if !u.Admin {
			redirectFlash(c, documentsRedirect(redirectParent), "error", "Only admins can rename templates")
			return
		}
		backend := app.getBackend(c)
		if err := backend.UpdateDocument(u.ID, docid, name, ""); err != nil {
			redirectFlash(c, "/admin/templates", "error", err.Error())
			return
		}
		backend.Sync(u.ID)
		redirectFlash(c, "/admin/templates", "success", "Updated")
		return
	}
	backend := app.getBackend(c)
	if err := backend.UpdateDocument(u.ID, docid, name, parentID); err != nil {
		redirectFlash(c, documentsRedirect(redirectParent), "error", err.Error())
		return
	}
	backend.Sync(u.ID)
	redirectFlash(c, documentsRedirect(redirectParent), "success", "Updated")
}

func (app *ReactAppWrapper) formPinDocument(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	docid := c.Param("docid")
	redirectParent := c.PostForm("redirect")
	pinnedRaw := strings.ToLower(strings.TrimSpace(c.PostForm("pinned")))
	pinned := pinnedRaw == "1" || pinnedRaw == "true" || pinnedRaw == "on" || pinnedRaw == "yes"
	if app.isSyncedTemplateDoc(c, u.ID, docid) {
		redirectFlash(c, documentsRedirect(redirectParent), "error", "Templates cannot be favorited here")
		return
	}
	backend := app.getBackend(c)
	if err := backend.SetDocumentPinned(u.ID, docid, pinned); err != nil {
		redirectFlash(c, documentsRedirect(redirectParent), "error", err.Error())
		return
	}
	backend.Sync(u.ID)
	msg := "Removed from favorites"
	if pinned {
		msg = "Added to favorites"
	}
	redirectFlash(c, documentsRedirect(redirectParent), "success", msg)
}

func documentsRedirect(parentID string) string {
	if parentID == "" {
		return "/documents"
	}
	return "/documents?folder=" + parentID
}

func (app *ReactAppWrapper) formCreateUser(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil || !u.Admin {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	id := strings.TrimSpace(c.PostForm("userid"))
	email := strings.TrimSpace(c.PostForm("email"))
	pass := c.PostForm("password")
	if pass == "" {
		pass = c.PostForm("newPassword")
	}
	if id == "" || pass == "" {
		redirectFlash(c, "/admin", "error", "User ID and password required")
		return
	}
	user, err := model.NewUser(id, pass)
	if err != nil {
		redirectFlash(c, "/admin", "error", err.Error())
		return
	}
	user.Email = email
	if err := app.userStorer.UpdateUser(user); err != nil {
		redirectFlash(c, "/admin", "error", err.Error())
		return
	}
	redirectFlash(c, "/admin", "success", "User created")
}

func (app *ReactAppWrapper) formDeleteUser(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil || !u.Admin {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	uid := c.Param("userid")
	if uid == u.ID {
		redirectFlash(c, "/admin", "error", "Cannot delete yourself")
		return
	}
	if err := app.userStorer.RemoveUser(uid); err != nil {
		redirectFlash(c, "/admin", "error", err.Error())
		return
	}
	redirectFlash(c, "/admin", "success", "User deleted")
}

func (app *ReactAppWrapper) formCreateIntegration(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	user := app.getModelUser(u.ID)
	if user == nil {
		redirectFlash(c, "/integrations", "error", "User not found")
		return
	}
	intg := model.IntegrationConfig{
		ID:       uuid.NewString(),
		Name:     strings.TrimSpace(c.PostForm("name")),
		Provider: strings.TrimSpace(c.PostForm("provider")),
		Username: c.PostForm("username"),
		Password: c.PostForm("password"),
		Address:  c.PostForm("address"),
		Path:     c.PostForm("path"),
		Endpoint: c.PostForm("endpoint"),
	}
	if intg.Provider == integrations.LocalfsProvider {
		redirectFlash(c, "/integrations", "error", "Local filesystem integrations must be edited in the user profile file on the server")
		return
	}
	user.Integrations = append(user.Integrations, intg)
	if err := app.userStorer.UpdateUser(user); err != nil {
		redirectFlash(c, "/integrations", "error", err.Error())
		return
	}
	redirectFlash(c, "/integrations", "success", "Integration created")
}

func (app *ReactAppWrapper) formDeleteIntegration(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil {
		return
	}
	user := app.getModelUser(u.ID)
	if user == nil {
		redirectFlash(c, "/integrations", "error", "User not found")
		return
	}
	id := c.Param("intid")
	out := user.Integrations[:0]
	for _, i := range user.Integrations {
		if i.ID != id {
			out = append(out, i)
		}
	}
	user.Integrations = out
	if err := app.userStorer.UpdateUser(user); err != nil {
		redirectFlash(c, "/integrations", "error", err.Error())
		return
	}
	redirectFlash(c, "/integrations", "success", "Integration deleted")
}

func (app *ReactAppWrapper) formSaveTheme(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil || !u.Admin {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	id := strings.TrimSpace(c.PostForm("id"))
	name := strings.TrimSpace(c.PostForm("name"))
	xmlBody := c.PostForm("xml")
	published := c.PostForm("published") == "true" || c.PostForm("published") == "on"
	if id == "" || xmlBody == "" {
		redirectFlash(c, "/admin/themes", "error", "ID and XML required")
		return
	}
	if err := app.themes.save(id, name, published, []byte(xmlBody)); err != nil {
		redirectFlash(c, "/admin/themes?id="+id, "error", err.Error())
		return
	}
	redirectFlash(c, "/admin/themes?id="+id, "success", "Theme saved")
}

func (app *ReactAppWrapper) formPublishTheme(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil || !u.Admin {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	id := c.Param("id")
	published := c.PostForm("published") != "false"
	if err := app.themes.setPublished(id, published); err != nil {
		redirectFlash(c, "/admin/themes?id="+id, "error", err.Error())
		return
	}
	redirectFlash(c, "/admin/themes?id="+id, "success", "Publish state updated")
}

func (app *ReactAppWrapper) formDeleteTheme(c *gin.Context) {
	u := app.requirePageUser(c)
	if u == nil || !u.Admin {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	id := c.Param("id")
	if err := app.themes.delete(id); err != nil {
		redirectFlash(c, "/admin/themes", "error", err.Error())
		return
	}
	redirectFlash(c, "/admin/themes", "success", "Theme deleted")
}
