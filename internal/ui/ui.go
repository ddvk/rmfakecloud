package ui

import (
	"net/http"
	"path"

	log "github.com/sirupsen/logrus"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/db"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/webassets"
	"github.com/gin-gonic/gin"
)

type ReactAppWrapper struct {
	fs     http.FileSystem
	prefix string
}

const indexReplacement = "/default"

func (w ReactAppWrapper) Open(filepath string) (http.File, error) {
	fullpath := filepath
	//index.html hack
	if filepath != indexReplacement {
		fullpath = path.Join(w.prefix, filepath)
	} else {
		fullpath = "/index.html"
	}
	f, err := w.fs.Open(fullpath)
	return f, err
}
func (w ReactAppWrapper) Register(router *gin.Engine) {
	router.StaticFS(w.prefix, w)

	//hack for index.html
	router.NoRoute(func(c *gin.Context) {
		c.FileFromFS(indexReplacement, w)
	})

	router.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("/favicon.ico", webassets.Assets)
	})
}

type Document struct {
	ID   string `json:id`
	Name string `json:name`
}
type DocumentList struct {
	Documents []Document `json:documents`
}

func badReq(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
	c.Abort()
}

type loginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

/// RegisterUI add the react ui
func RegisterUI(e *gin.Engine, cfg *config.Config, userStorer db.UserStorer) {
	staticWrapper := ReactAppWrapper{
		fs:     webassets.Assets,
		prefix: "/static",
	}
	staticWrapper.Register(e)

	r := e.Group("/ui/api")
	r.POST("register", func(c *gin.Context) {
		if !cfg.RegistrationOpen {
			c.JSON(http.StatusForbidden, gin.H{"error": "Registrations are closed"})
			c.Abort()
			return
		}

		var form loginForm
		if err := c.ShouldBindJSON(&form); err != nil {
			log.Error(err)
			badReq(c, err.Error())
			return
		}

		// Check this user doesn't already exist
		users, err := userStorer.GetUsers()
		if err != nil {
			log.Error(err)
			badReq(c, err.Error())
			return
		}

		for _, u := range users {
			if u.Email == form.Email {
				badReq(c, form.Email+" is already registered.")
				return
			}
		}

		user, err := messages.NewUser(form.Email, form.Password)
		if err != nil {
			log.Error(err)
			badReq(c, err.Error())
			return
		}

		err = userStorer.RegisterUser(user)
		if err != nil {
			log.Error(err)
			badReq(c, err.Error())
			return
		}

		c.JSON(http.StatusOK, user)
	})

	r.POST("login", func(c *gin.Context) {
		var form loginForm
		if err := c.ShouldBindJSON(&form); err != nil {
			log.Error(err)
			badReq(c, err.Error())
			return
		}

		// Try to find the user
		users, err := userStorer.GetUsers()
		if err != nil {
			log.Error(err)
			badReq(c, err.Error())
			return
		}

		var user *messages.User
		for _, u := range users {
			if form.Email == u.Email {
				user = u
			}
		}

		if user == nil {
			log.Error(err)
			c.JSON(http.StatusUnauthorized, "Invalid email or password")
			return
		}

		if ok, err := user.CheckPassword(form.Password); err != nil || !ok {
			log.Error(err)
			c.JSON(http.StatusUnauthorized, "Invalid email or password")
			return
		}

		token := user.NewAuth0Token("ui", "")
		tokenString, err := token.SignedString(cfg.JWTSecretKey)
		if err != nil {
			badReq(c, err.Error())
			return
		}

		c.String(http.StatusOK, tokenString)
	})

}

func RegisterUIAuth(e *gin.RouterGroup, metaStorer db.MetadataStorer, userStorer db.UserStorer) {
	r := e.Group("/ui/api")

	r.GET("list", func(c *gin.Context) {
		docs, err := metaStorer.GetAllMetadata(false)
		if err != nil {
			log.Error(err)
			badReq(c, err.Error())
			return
		}

		c.JSON(http.StatusOK, docs)
	})
}
