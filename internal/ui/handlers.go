package ui

import (
	"net/http"

	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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
		badReq(c, err.Error())
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
		badReq(c, err.Error())
		return
	}

	if user == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if ok, err := user.CheckPassword(form.Password); err != nil || !ok {
		log.Error(err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, "Invalid email or password")
		return
	}

	token := user.NewAuth0Token("ui", "")

	tokenString, err := token.SignedString(app.cfg.JWTSecretKey)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":       user,
		"auth_token": tokenString,
	})

}

func (app *ReactAppWrapper) newCode(c *gin.Context) {
	uid := c.GetString("userId")
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

	code, err := user.NewUserCode()
	if err != nil {
		log.Error("Unable to generate new device code: ", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to generate new code"})
		return
	}

	app.userStorer.UpdateUser(user)

	c.JSON(http.StatusOK, code)
}
func (app *ReactAppWrapper) listDocuments(c *gin.Context) {
	documentList := DocumentList{
		Documents: []Document{
			{
				ID:       "001",
				Name:     "The Adventures of Huckleberry Finn by Mark Twain",
				ImageUrl: "https://picsum.photos/100/150",
				ParentId: "root",
			},
			{
				ID:       "002",
				Name:     "The Great Gatsby by F. Scott Fizgerald",
				ImageUrl: "https://images-na.ssl-images-amazon.com/images/I/41iers%2BHLSL._SL160_.jpg",
				ParentId: "root",
			},
			{
				ID:       "003",
				Name:     "The Stories of Anton Chekhov by Anton Checkhov",
				ImageUrl: "https://picsum.photos/100/150",
				ParentId: "root",
			},
			{
				ID:       "004",
				Name:     "War and Peace by Leo Tolstoy",
				ImageUrl: "https://picsum.photos/100/150",
				ParentId: "root",
			},

			{
				ID:       "005",
				Name:     " Madame Bovary by Gustav Flaubert",
				ImageUrl: "https://picsum.photos/100/150",
				ParentId: "root",
			},

			{
				ID:       "006",
				Name:     "The Adventures of Huckleberry Finn by Mark Twain",
				ImageUrl: "https://picsum.photos/100/150",
				ParentId: "root",
			},

			{
				ID:       "007",
				Name:     " The Brothers Karamazov by Fyodor Dostoyevsky",
				ImageUrl: "https://picsum.photos/100/150",
				ParentId: "root",
			},

			{
				ID:       "008",
				Name:     "Don Quixote by Miguel de Cervantes",
				ImageUrl: "https://m.media-amazon.com/images/I/51nBHIQv6zL._SL160_.jpg",
				ParentId: "root",
			},

			{
				ID:       "009",
				Name:     "Ulysses by James Joyce",
				ImageUrl: "https://picsum.photos/100/150",
				ParentId: "root",
			},
			{
				ID:       "010",
				Name:     "Crime and Punishment by Fyodor Dostoyevsky",
				ImageUrl: "https://picsum.photos/100/150",
				ParentId: "root",
			},
		},
	}
	c.JSON(http.StatusOK, documentList.Documents)
}

func (app *ReactAppWrapper) getAppUsers(c *gin.Context) {
	// Try to find the user
	users, err := app.userStorer.GetUsers()

	for _, u := range users {
		//FIXME: use a different object for ui
		u.Password = ""
	}

	if err != nil {
		log.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to get users."})
		return
	}

	c.JSON(http.StatusOK, users)
}
func (app *ReactAppWrapper) getUser(c *gin.Context) {
	uid := c.Param("userid")
	log.Info("Requested: ", uid)

	//TODO: check for admin role

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
