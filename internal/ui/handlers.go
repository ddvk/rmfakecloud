package ui

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	userIDContextKey    = "userID"
	browserIDContextKey = "browserID"
	isSync15Key         = "sync15"
	docIDParam          = "docid"
	uiLogger            = "[ui] "
	useridParam         = "userid"
	cookieName          = ".Authrmfakecloud"
)

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
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Registrations are closed"})
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
	expires := time.Now().Add(expiresAfter).Unix()
	claims := &WebUserClaims{
		UserID:    user.ID,
		BrowserID: uuid.NewString(),
		Email:     user.Email,
		Scopes:    scopes,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expires,
			Issuer:    "rmFake WEB",
			Audience:  WebUsage,
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

	uid := c.GetString(userIDContextKey)

	if user.ID != uid {
		log.Error("Trying to change password for a different user.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cant do that"})
		return
	}

	ok, err := user.CheckPassword(req.CurrentPassword)
	if !ok {
		if err != nil {
			log.Error(err)
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid email or password"})
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
	uid := c.GetString(userIDContextKey)

	user, err := app.userStorer.GetUser(uid)
	if err != nil {
		log.Error("Unable to find user: ", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	code, err := app.codeConnector.NewCode(user.ID)
	if err != nil {
		log.Error("Unable to generate new device code: ", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to generate new code"})
		return
	}

	c.JSON(http.StatusOK, code)
}

func getBackend(c *gin.Context) backend {
	blah, ok := c.Get("backend")
	if !ok {
		panic("not there")
	}
	return blah.(backend)

}
func (app *ReactAppWrapper) listDocuments(c *gin.Context) {
	uid := c.GetString(userIDContextKey)

	var tree *viewmodel.DocumentTree

	backend := getBackend(c)
	tree, err := backend.GetDocumentTree(uid)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, tree)
}
func (app *ReactAppWrapper) getDocument(c *gin.Context) {
	uid := c.GetString(userIDContextKey)
	docid := common.ParamS(docIDParam, c)
	log.Info("exporting ", docid)
	backend := getBackend(c)
	reader, err := backend.Export(uid, docid, "pdf", 0)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	defer reader.Close()
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

func (app *ReactAppWrapper) getDocumentMetadata(c *gin.Context) {
	uid := c.GetString(userIDContextKey)
	docid := common.ParamS(docIDParam, c)
	metadata, err := app.metadataStore.GetMetadata(uid, docid)

	if err != nil {
		badReq(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, metadata)
}

func (app *ReactAppWrapper) updateDocument(c *gin.Context) {
	upd := viewmodel.UpdateDoc{}
	if err := c.ShouldBindJSON(&upd); err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}
	uid := c.GetString(userIDContextKey)
	docid := common.ParamS(docIDParam, c)

	metadata, err := app.metadataStore.GetMetadata(uid, docid)

	if err != nil {
		badReq(c, err.Error())
		return
	}

	dirty := false

	// Rename
	if len(upd.Name) > 0 && upd.Name != metadata.VissibleName {
		metadata.VissibleName = upd.Name
		dirty = true
	}

	// Move
	if len(upd.ParentID) > 0 && upd.ParentID != metadata.Parent {
		parent, err := app.metadataStore.GetMetadata(uid, upd.ParentID)

		if err != nil {
			badReq(c, err.Error())
			return
		}

		metadata.Parent = parent.ID
		dirty = true
	}

	if upd.SetParentToRoot && metadata.Parent != "" {
		metadata.Parent = ""
		dirty = true
	}

	if dirty {
		metadata.Version += 1
		if err = app.metadataStore.UpdateMetadata(uid, metadata); err != nil {
			badReq(c, err.Error())
			return
		}

		app.h.Notify(uid, "web", hub.DocumentNotification{
			ID:      metadata.ID,
			Type:    metadata.Type,
			Version: metadata.Version,
			Parent:  metadata.Parent,
			Name:    metadata.VissibleName,
		}, hub.DocAddedEvent)

		log.Info(uiLogger, "document updated: id=", metadata.ID)
	}

	c.Status(http.StatusOK)
}

func (app *ReactAppWrapper) deleteDocument(c *gin.Context) {
	uid := c.GetString(userIDContextKey)
	docId := common.ParamS(docIDParam, c)
	backend := getBackend(c)
	err := backend.DeleteDocument(uid, docId)

	if err != nil {
		log.Error("Delete document error: ", err)
		badReq(c, err.Error())
		return
	}

	c.Status(http.StatusOK)
}
func (app *ReactAppWrapper) createDocument(c *gin.Context) {
	uid := c.GetString(userIDContextKey)
	_ = c.GetBool(isSync15Key)
	log.Info("uploading documents from: ", uid)

	backend := getBackend(c)

	form, err := c.MultipartForm()
	if err != nil {
		log.Error(err)
		badReq(c, "not multiform")
		return
	}
	parentID := ""
	if parent, ok := form.Value["parent"]; ok {
		parentID = parent[0]
	}
	log.Info("Parent: " + parentID)

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

		_, err = backend.CreateDocument(uid, file.Filename, parentID, f)
		if err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}
	backend.Sync(uid)
	c.Status(http.StatusOK)
}

func (app *ReactAppWrapper) getAppUsers(c *gin.Context) {
	// Try to find the user
	users, err := app.userStorer.GetUsers()

	if err != nil {
		log.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to get users."})
		return
	}

	uilist := make([]viewmodel.User, 0)
	for _, u := range users {
		usr := viewmodel.User{
			ID:        u.ID,
			Email:     u.Email,
			Name:      u.Name,
			CreatedAt: u.CreatedAt,
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
	if uid == c.GetString(userIDContextKey) {
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

func (app *ReactAppWrapper) profile(c *gin.Context) {
	uid := c.GetString(userIDContextKey)

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
