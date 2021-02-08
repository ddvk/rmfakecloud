package ui

import (
	"net/http"
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	userID = "userID"
)

func (app *ReactAppWrapper) register(c *gin.Context) {
	if !app.cfg.RegistrationOpen {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Registrations are closed"})
		return
	}

	var form loginForm
	if err := c.ShouldBindJSON(&form); err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}

	// Check this user doesn't already exist
	user, err := app.userStorer.GetUser(form.Email)
	if user != nil {
		badReq(c, "alread taken")
		return
	}

	user, err = model.NewUser(form.Email, form.Password)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	err = app.userStorer.RegisterUser(user)
	if err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)

}
func (app *ReactAppWrapper) login(c *gin.Context) {
	var form loginForm
	if err := c.ShouldBindJSON(&form); err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}

	// Try to find the user
	user, err := app.userStorer.GetUser(form.Email)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if ok, err := user.CheckPassword(form.Password); err != nil || !ok {
		log.Error(err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	claims := &common.WebUserClaims{
		UserId: user.Id,
		Email:  user.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(12 * time.Hour).Unix(),
			Issuer:    "rmFake WEB",
			Audience:  common.WebUsage,
		},
	}
	if user.IsAdmin {
		claims.Roles = []string{"admin"}
	}

	tokenString, err := common.SignClaims(claims, app.cfg.JWTSecretKey)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.SetCookie(".AuthCookie", tokenString, 1, "/", "rmfakecloud", true, true)

	c.JSON(http.StatusOK, gin.H{"auth_token": tokenString, "user": gin.H{}})

}

func (app *ReactAppWrapper) newCode(c *gin.Context) {
	uid := c.GetString(userID)
	if uid == "" {
		log.Error("Unable to find userId in context")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	user, err := app.userStorer.GetUser(uid)
	if err != nil {
		log.Error("Unable to find user: ", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	code, err := app.codeConnector.NewCode(user.Id)
	if err != nil {
		log.Error("Unable to generate new device code: ", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to generate new code"})
		return
	}

	c.JSON(http.StatusOK, code)
}
func (app *ReactAppWrapper) listDocuments(c *gin.Context) {
	tree := DocumentTree{}

	c.JSON(http.StatusOK, tree)
}

func (app *ReactAppWrapper) getAppUsers(c *gin.Context) {
	// Try to find the user
	users, err := app.userStorer.GetUsers()

	if err != nil {
		log.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to get users."})
		return
	}

	uilist := make([]user, 0)
	for _, u := range users {
		usr := user{
			ID:    u.Id,
			Email: u.Email,
			Name:  u.Name,
		}
		uilist = append(uilist, usr)
	}
	c.JSON(http.StatusOK, uilist)
}
func (app *ReactAppWrapper) updateUser(c *gin.Context) {
}
func (app *ReactAppWrapper) getUser(c *gin.Context) {
	uid := c.Param("userid")
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
