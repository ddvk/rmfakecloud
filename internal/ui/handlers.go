package ui

import (
	"fmt"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/integrations"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

const (
	userIDContextKey    = "userID"
	browserIDContextKey = "browserID"
	isSync15Key         = "sync15"
	docIDParam          = "docid"
	intIDParam          = "intid"
	uiLogger            = "[ui] "
	ui10                = " [10] "
	useridParam         = "userid"
	cookieName          = ".Authrmfakecloud"
)

func userID(c *gin.Context) string {
	//TODO: suppress the warning
	//codeql[go/path-injection]
	return c.GetString(userIDContextKey)
}

func (app *ReactAppWrapper) register(c *gin.Context) {

	if !app.cfg.RegistrationOpen {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	client := c.ClientIP()
	log.Info(client)

	if client != "localhost" &&
		client != "::1" &&
		client != "127.0.0.1" {
		c.AbortWithStatusJSON(http.StatusForbidden, viewmodel.NewErrorResponse("Registrations are closed"))
		return
	}

	var form viewmodel.LoginForm
	if err := c.ShouldBindJSON(&form); err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Check this user doesn't already exist
	_, err := app.userStorer.GetUser(form.Email)
	if err == nil {
		badReq(c, "already taken")
		return
	}

	user, err := model.NewUser(form.Email, form.Password)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = app.userStorer.RegisterUser(user)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (app *ReactAppWrapper) login(c *gin.Context) {
	if app.cfg.OIDCConfig != nil {
		if app.cfg.OIDCConfig.Only {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}

	var form viewmodel.LoginForm
	if err := c.ShouldBindJSON(&form); err != nil {
		log.Error(uiLogger, err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	// not really thread safe
	if app.cfg.CreateFirstUser {
		log.Info("Creating an admin user")
		user, err := model.NewUser(form.Email, form.Password)
		if err != nil {
			log.Error("[login]", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		user.IsAdmin = true
		err = app.userStorer.RegisterUser(user)
		if err != nil {
			log.Error("[login] Register ", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		app.cfg.CreateFirstUser = false
	}

	// Try to find the user
	user, err := app.userStorer.GetUser(form.Email)
	if err != nil {
		log.Error(uiLogger, err, " cannot load user, login failed ip: ", c.ClientIP())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if ok, err := user.CheckPassword(form.Password); err != nil || !ok {
		if err != nil {
			log.Error(err)
		} else if !ok {
			log.Warn(uiLogger, "wrong password for: ", form.Email, ", login failed ip: ", c.ClientIP())
		}
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	scopes := ""
	if user.Sync15 {
		scopes = isSync15Key
	}
	expiresAfter := 24 * time.Hour
	expires := time.Now().Add(expiresAfter)
	claims := &WebUserClaims{
		UserID:    user.ID,
		BrowserID: uuid.NewString(),
		Email:     user.Email,
		Scopes:    scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expires),
			Issuer:    "rmFake WEB",
			Audience:  []string{WebUsage},
		},
	}
	if user.IsAdmin {
		claims.Roles = []string{AdminRole}
	} else {
		claims.Roles = []string{"User"}
	}

	tokenString, err := common.SignClaims(claims, app.cfg.JWTSecretKey)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	log.Debug("cookie expires after: ", expiresAfter)
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(cookieName, tokenString, int(expiresAfter.Seconds()), "/", "", app.cfg.HTTPSCookie, true)

	c.String(http.StatusOK, tokenString)
}

func (app *ReactAppWrapper) oidcInfo(c *gin.Context) {
	if app.cfg.OIDCConfig != nil {
		c.JSON(http.StatusOK, viewmodel.OIDCInfo{
			Enabled: true,
			Label:   app.cfg.OIDCConfig.Label,
			Only:    app.cfg.OIDCConfig.Only,
		})
	} else {
		c.JSON(http.StatusOK, viewmodel.OIDCInfo{
			Enabled: false,
			Label:   "",
			Only:    false,
		})
	}
}

func (app *ReactAppWrapper) oidcAuth(c *gin.Context) {
	if app.cfg.OIDCConfig == nil {
		log.Error("OIDC configuration required")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	provider, err := oidc.NewProvider(c, app.cfg.OIDCConfig.ConfigURL)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	app.oidcProvider = provider
	app.oauth2Config = &oauth2.Config{
		ClientID:     app.cfg.OIDCConfig.ClientID,
		ClientSecret: app.cfg.OIDCConfig.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  app.cfg.StorageURL + "/login",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	c.String(http.StatusOK, app.oauth2Config.AuthCodeURL("")) // todo: store and verify OAuth2 states
}

func (app *ReactAppWrapper) oidcCallback(c *gin.Context) {
	if app.cfg.OIDCConfig == nil {
		log.Error("OIDC configuration required")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var req viewmodel.OIDCCallback

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}

	verifier := app.oidcProvider.Verifier(&oidc.Config{ClientID: app.cfg.OIDCConfig.ClientID})

	// Verify state and errors.
	oauth2Token, err := app.oauth2Config.Exchange(c, req.Code)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Parse and verify ID Token payload.
	idToken, err := verifier.Verify(c, rawIDToken)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Extract custom claims
	var oidcClaims struct {
		Email string `json:"email"`
	}
	if err := idToken.Claims(&oidcClaims); err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// not really thread safe
	if app.cfg.CreateFirstUser {
		// Random string
		password := uuid.NewString()

		log.Infof("Creating an admin user (with random password \"%s\")", password)
		user, err := model.NewUser(oidcClaims.Email, password)
		if err != nil {
			log.Error("[login]", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		user.IsAdmin = true
		err = app.userStorer.RegisterUser(user)
		if err != nil {
			log.Error("[login] Register ", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		app.cfg.CreateFirstUser = false
	}

	// Try to find the user
	user, err := app.userStorer.GetUser(oidcClaims.Email)
	if err != nil {
		log.Error(uiLogger, err, " cannot load user, login failed ip: ", c.ClientIP())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	scopes := ""
	if user.Sync15 {
		scopes = isSync15Key
	}
	expiresAfter := 24 * time.Hour
	expires := time.Now().Add(expiresAfter)
	claims := &WebUserClaims{
		UserID:    user.ID,
		BrowserID: uuid.NewString(),
		Email:     user.Email,
		Scopes:    scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expires),
			Issuer:    "rmFake WEB",
			Audience:  []string{WebUsage},
		},
	}
	if user.IsAdmin {
		claims.Roles = []string{AdminRole}
	} else {
		claims.Roles = []string{"User"}
	}

	tokenString, err := common.SignClaims(claims, app.cfg.JWTSecretKey)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	log.Debug("cookie expires after: ", expiresAfter)
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(cookieName, tokenString, int(expiresAfter.Seconds()), "/", "", app.cfg.HTTPSCookie, true)

	c.String(http.StatusOK, tokenString)
}

func (app *ReactAppWrapper) changePassword(c *gin.Context) {
	var req viewmodel.ResetPasswordForm

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}

	user, err := app.userStorer.GetUser(req.UserID)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	uid := userID(c)

	if user.ID != uid {
		log.Error("Trying to change password for a different user.")
		c.AbortWithStatusJSON(http.StatusBadRequest, viewmodel.NewErrorResponse("cant do that"))
		return
	}

	ok, err := user.CheckPassword(req.CurrentPassword)
	if !ok {
		if err != nil {
			log.Error(err)
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, viewmodel.NewErrorResponse("Invalid email or password"))
		return
	}

	if req.NewPassword != "" {
		user.SetPassword(req.NewPassword)
	}

	err = app.userStorer.UpdateUser(user)

	if err != nil {
		log.Error("error updating user", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (app *ReactAppWrapper) newCode(c *gin.Context) {
	uid := userID(c)

	user, err := app.userStorer.GetUser(uid)
	if err != nil {
		log.Error("Unable to find user: ", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, viewmodel.NewErrorResponse(err.Error()))
		return
	}

	code, err := app.codeConnector.NewCode(user.ID)
	if err != nil {
		log.Error("Unable to generate new device code: ", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, viewmodel.NewErrorResponse("Unable to generate new code"))
		return
	}

	c.JSON(http.StatusOK, code)
}

func (app *ReactAppWrapper) getBackend(c *gin.Context) backend {
	s, ok := c.Get(backendVersionKey)
	if !ok {
		panic("key not set")
	}
	backend, ok := app.backends[s.(common.SyncVersion)]
	if !ok {
		panic("backend not found")
	}
	return backend
}

func (app *ReactAppWrapper) listDocuments(c *gin.Context) {
	uid := userID(c)

	var tree *viewmodel.DocumentTree

	backend := app.getBackend(c)
	tree, err := backend.GetDocumentTree(uid)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, tree)
}
func (app *ReactAppWrapper) getDocument(c *gin.Context) {
	uid := userID(c)
	docid := common.ParamS(docIDParam, c)

	exportType := "pdf"
	var exportOption storage.ExportOption = 0

	log.Info("exporting ", docid)
	backend := app.getBackend(c)

	reader, err := backend.Export(uid, docid, exportType, exportOption)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	defer reader.Close()
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

func (app *ReactAppWrapper) getDocumentMetadata(c *gin.Context) {
	uid := userID(c)
	docid := common.ParamS(docIDParam, c)
	// if err != nil {
	// 	log.Error(err)
	// 	c.AbortWithStatus(http.StatusInternalServerError)
	// 	return
	// }
	log.Info(uid, docid)
	c.JSON(http.StatusOK, "TODO")

}

// move rename
func (app *ReactAppWrapper) updateDocument(c *gin.Context) {
	upd := viewmodel.UpdateDoc{}
	if err := c.ShouldBindJSON(&upd); err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}
	backend := app.getBackend(c)
	uid := userID(c)
	log.Info(uiLogger, ui10, "updatedoc")
	err := backend.UpdateDocument(uid, upd.DocumentID, upd.Name, upd.ParentID)
	if err != nil {
		badReq(c, err.Error())
		return
	}

	c.Status(http.StatusOK)
}
func (app *ReactAppWrapper) deleteDocument(c *gin.Context) {
	uid := userID(c)
	docid := c.Param("docid")
	backend := app.getBackend(c)

	err := backend.DeleteDocument(uid, docid)
	if err != nil {
		badReq(c, err.Error())
	}
	c.Status(http.StatusOK)
}

func (app *ReactAppWrapper) createFolder(c *gin.Context) {
	upd := viewmodel.NewFolder{}
	if err := c.ShouldBindJSON(&upd); err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}
	uid := userID(c)

	backend := app.getBackend(c)

	doc, err := backend.CreateFolder(uid, upd.Name, upd.ParentID)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, doc)
}

func (app *ReactAppWrapper) createDocument(c *gin.Context) {
	uid := userID(c)
	log.Info("uploading documents from: ", uid)

	backend := app.getBackend(c)

	form, err := c.MultipartForm()
	if err != nil {
		log.Error(err)
		badReq(c, "not multiform")
		return
	}
	parentID := ""
	if parent, ok := form.Value["parent"]; ok {
		if parent[0] != "root" {
			parentID = parent[0]
		}
	}

	log.Info("Parent: " + parentID)

	docs := []*storage.Document{}
	for _, file := range form.File["file"] {
		f, err := file.Open()
		if err != nil {
			log.Error("[ui] ", err)
			badReq(c, "cant open attachment")
			return
		}

		defer f.Close()
		//do the stuff
		log.Info(uiLogger, fmt.Sprintf("Uploading %s , size: %d", file.Filename, file.Size))

		doc, err := backend.CreateDocument(uid, file.Filename, parentID, f)
		if err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		docs = append(docs, doc)
	}
	backend.Sync(uid)
	c.JSON(http.StatusOK, docs)
}

func (app *ReactAppWrapper) getAppUsers(c *gin.Context) {
	// Try to find the user
	users, err := app.userStorer.GetUsers()

	if err != nil {
		log.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, viewmodel.NewErrorResponse("Unable to get users."))
		return
	}

	uilist := make([]viewmodel.User, 0)
	for _, u := range users {
		usr := viewmodel.User{
			ID:        u.ID,
			Email:     u.Email,
			Name:      u.Name,
			CreatedAt: u.CreatedAt,
			IsAdmin:   u.IsAdmin,
		}
		uilist = append(uilist, usr)
	}
	c.JSON(http.StatusOK, uilist)
}

func (app *ReactAppWrapper) getUser(c *gin.Context) {
	uid := c.Param(useridParam)
	log.Info("Requested: ", uid)

	// Try to find the user
	user, err := app.userStorer.GetUser(uid)
	if err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}

	if user == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, "Invalid user")
		return
	}
	if uid != user.ID && !IsAdmin(c) {
		log.Warn("Only admins can query other users")
		c.AbortWithStatusJSON(http.StatusUnauthorized, "")
		return
	}

	vmUser := &viewmodel.User{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	}
	for _, i := range user.Integrations {
		vmUser.Integrations = append(vmUser.Integrations, i.Name)
	}

	c.JSON(http.StatusOK, vmUser)
}

func (app *ReactAppWrapper) updateUser(c *gin.Context) {
	var req viewmodel.User
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	user, err := app.userStorer.GetUser(req.ID)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if user == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, "Invalid user")
		return
	}
	if req.NewPassword != "" {
		user.SetPassword(req.NewPassword)
	}

	if req.Email != "" {
		user.Email = req.Email
	}

	err = app.userStorer.UpdateUser(user)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusAccepted)
}
func (app *ReactAppWrapper) deleteUser(c *gin.Context) {
	uid := c.Param(useridParam)
	if uid == userID(c) {
		log.Error("can't remove current user ")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	err := app.userStorer.RemoveUser(uid)
	if err != nil {
		log.Error("can't remove ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusAccepted)
}

func (app *ReactAppWrapper) createUser(c *gin.Context) {
	var req viewmodel.NewUser
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}

	user, err := model.NewUser(req.ID, req.NewPassword)

	if err != nil {
		log.Error("can't create ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	user.Email = req.Email

	err = app.userStorer.UpdateUser(user)
	if err != nil {
		log.Error("can't create ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusCreated)
}

func (app *ReactAppWrapper) listIntegrations(c *gin.Context) {
	uid := userID(c)

	user, err := app.userStorer.GetUser(uid)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.JSON(http.StatusOK, user.Integrations)
}

func warnLocalfsEdition(c *gin.Context, int *model.IntegrationConfig) {
	s, err := yaml.Marshal(gin.H{"integrations": []*model.IntegrationConfig{int}})
	if err != nil {
		log.Error("error updating user", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.AbortWithStatusJSON(http.StatusForbidden,
		viewmodel.NewErrorResponse("To avoid security issues with local directory integration, you have to manually edit your .userprofile file:\n\n"+string(s)))
}

func (app *ReactAppWrapper) createIntegration(c *gin.Context) {
	int := model.IntegrationConfig{}
	if err := c.ShouldBindJSON(&int); err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}

	if int.Provider == integrations.LocalfsProvider {
		int.ID = uuid.NewString()
		warnLocalfsEdition(c, &int)
		return
	}

	uid := userID(c)

	user, err := app.userStorer.GetUser(uid)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	int.ID = uuid.NewString()
	user.Integrations = append(user.Integrations, int)

	err = app.userStorer.UpdateUser(user)

	if err != nil {
		log.Error("error updating user", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, int)
}

func (app *ReactAppWrapper) getIntegration(c *gin.Context) {
	uid := userID(c)

	intid := common.ParamS(intIDParam, c)

	user, err := app.userStorer.GetUser(uid)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	for _, integration := range user.Integrations {
		if integration.ID == intid {
			c.JSON(http.StatusOK, integration)
			return
		}
	}

	c.AbortWithStatus(http.StatusNotFound)
}

func (app *ReactAppWrapper) updateIntegration(c *gin.Context) {
	int := model.IntegrationConfig{}
	if err := c.ShouldBindJSON(&int); err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}

	if int.Provider == integrations.LocalfsProvider {
		warnLocalfsEdition(c, &int)
		return
	}

	uid := userID(c)

	intid := common.ParamS(intIDParam, c)

	user, err := app.userStorer.GetUser(uid)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	for idx, integration := range user.Integrations {
		if integration.ID == intid {
			int.ID = integration.ID
			user.Integrations[idx] = int

			err = app.userStorer.UpdateUser(user)

			if err != nil {
				log.Error("error updating user", err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			c.JSON(http.StatusOK, int)
			return
		}
	}

	c.AbortWithStatus(http.StatusNotFound)
}

func (app *ReactAppWrapper) deleteIntegration(c *gin.Context) {
	uid := userID(c)

	intid := common.ParamS(intIDParam, c)

	user, err := app.userStorer.GetUser(uid)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	for idx, integration := range user.Integrations {
		if integration.ID == intid {
			user.Integrations = append(user.Integrations[:idx], user.Integrations[idx+1:]...)

			err = app.userStorer.UpdateUser(user)

			if err != nil {
				log.Error("error updating user", err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			c.Status(http.StatusAccepted)
			return
		}
	}

	c.AbortWithStatus(http.StatusNotFound)
}

func (app *ReactAppWrapper) exploreIntegration(c *gin.Context) {
	uid := userID(c)

	integrationID := common.ParamS(intIDParam, c)

	integrationProvider, err := integrations.GetStorageIntegrationProvider(app.userStorer, uid, integrationID)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	folder := common.ParamS("path", c)
	if folder == "" {
		folder = "root"
	}

	response, err := integrationProvider.List(folder, 2)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (app *ReactAppWrapper) getMetadataIntegration(c *gin.Context) {
	uid := userID(c)

	integrationID := common.ParamS(intIDParam, c)

	integrationProvider, err := integrations.GetStorageIntegrationProvider(app.userStorer, uid, integrationID)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	fileid := common.ParamS("path", c)

	response, err := integrationProvider.GetMetadata(fileid)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (app *ReactAppWrapper) downloadThroughIntegration(c *gin.Context) {
	uid := userID(c)

	integrationID := common.ParamS(intIDParam, c)

	integrationProvider, err := integrations.GetStorageIntegrationProvider(app.userStorer, uid, integrationID)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	fileid := common.ParamS("path", c)

	response, size, err := integrationProvider.Download(fileid)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	defer response.Close()

	c.DataFromReader(http.StatusOK, size, "", response, nil)
}
