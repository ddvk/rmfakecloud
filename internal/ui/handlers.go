package ui

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
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
		log.Error(err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if ok, err := user.CheckPassword(form.Password); err != nil || !ok {
		if err != nil {
			log.Error(err)
		} else if !ok {
			log.Warn(uiLogger, "password mismatch for: ", form.Email)
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
	claims := &common.WebUserClaims{
		UserID:    user.ID,
		BrowserID: uuid.NewString(),
		Email:     user.Email,
		Scopes:    scopes,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expires,
			Issuer:    "rmFake WEB",
			Audience:  common.WebUsage,
		},
	}
	if user.IsAdmin {
		claims.Roles = []string{"Admin"}
	} else {
		claims.Roles = []string{"User"}
	}

	tokenString, err := common.SignClaims(claims, app.cfg.JWTSecretKey)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	httpOnly := true
	secure := app.cfg.IsHTTPS()
	log.Debug("cookie expires after: ", expiresAfter)
	c.SetCookie(cookieName, tokenString, int(expiresAfter.Seconds()), "/", "", secure, httpOnly)

	c.String(http.StatusOK, tokenString)
}

func (app *ReactAppWrapper) resetPassword(c *gin.Context) {

	var form viewmodel.ResetPasswordForm

	if err := c.ShouldBindJSON(&form); err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}

	user, err := app.userStorer.GetUser(form.Email)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	uid := c.GetString(userIDContextKey)

	if user.ID != uid {
		log.Error("Trying to change password for a different user.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cannt do that"})
		return
	}

	ok, err := user.CheckPassword(form.CurrentPassword)
	if !ok {
		if err != nil {
			log.Error(err)
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid email or password"})
		return
	}

	user.SetPassword(form.NewPassword)

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
	docid := c.Param(docIDParam)
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

func (app *ReactAppWrapper) updateDocument(c *gin.Context) {
	upd := viewmodel.UpdateDoc{}
	if err := c.ShouldBindJSON(&upd); err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}
	// uid := c.GetString(userID)
	// docid := c.Param("docid")

	c.Status(http.StatusOK)
}
func (app *ReactAppWrapper) deleteDocument(c *gin.Context) {
	// uid := c.GetString(userID)
	// docid := c.Param("docid")

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
			ID:    u.ID,
			Email: u.Email,
			Name:  u.Name,
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

	c.JSON(http.StatusOK, user)
}

func (app *ReactAppWrapper) updateUser(c *gin.Context) {
	c.Status(http.StatusCreated)
}
