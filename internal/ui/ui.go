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
	ImageUrl string `json:imageUrl`
	ParentId string `json:parentId`
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

	r.GET("newcode", func(c *gin.Context) {
		uid, ok := c.Get("userId")
		if !ok {
			log.Error("Unable to find userId in context")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			c.Abort()
			return
		}
		userId, ok := uid.(string)
		if !ok {
			log.Error("Unable to find valid userId in context")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			c.Abort()
			return
		}

		user, err := userStorer.GetUser(userId)
		if err != nil {
			log.Error("Unable to find user: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		code, err := user.NewUserCode()
		if err != nil {
			log.Error("Unable to generate new device code: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to generate new code"})
			c.Abort()
			return
		}

		userStorer.UpdateUser(user)

		c.JSON(http.StatusOK, code)
	})

	r.GET("list", func(c *gin.Context) {
		documentList := DocumentList{
			Documents: []Document{
				Document{
					ID:   "001",
					Name: "The Adventures of Huckleberry Finn by Mark Twain",
					ImageUrl: "https://picsum.photos/100/150",
					ParentId: "root",
				},
				Document{
					ID:   "002",
					Name: "The Great Gatsby by F. Scott Fizgerald",
					ImageUrl: "https://images-na.ssl-images-amazon.com/images/I/41iers%2BHLSL._SL160_.jpg",
					ParentId: "root",
				},
				Document{
					ID:   "003",
					Name: "The Stories of Anton Chekhov by Anton Checkhov",
					ImageUrl: "https://picsum.photos/100/150",
					ParentId: "root",
				},
				Document{
					ID:   "004",
					Name: "War and Peace by Leo Tolstoy",
					ImageUrl: "https://picsum.photos/100/150",
					ParentId: "root",
				},

				Document{
					ID:   "005",
					Name: " Madame Bovary by Gustav Flaubert",
					ImageUrl: "https://picsum.photos/100/150",
					ParentId: "root",
				},

				Document{
					ID:   "006",
					Name: "The Adventures of Huckleberry Finn by Mark Twain",
					ImageUrl: "https://picsum.photos/100/150",
					ParentId: "root",
				},

				Document{
					ID:   "007",
					Name: " The Brothers Karamazov by Fyodor Dostoyevsky",
					ImageUrl: "https://picsum.photos/100/150",
					ParentId: "root",
				},

				Document{
					ID:   "008",
					Name: "Don Quixote by Miguel de Cervantes",
					ImageUrl: "https://m.media-amazon.com/images/I/51nBHIQv6zL._SL160_.jpg",
					ParentId: "root",
				},

				Document{
					ID:   "009",
					Name: "Ulysses by James Joyce",
					ImageUrl: "https://picsum.photos/100/150",
					ParentId: "root",
				},
				Document{
					ID:   "010",
					Name: "Crime and Punishment by Fyodor Dostoyevsky",
					ImageUrl: "https://picsum.photos/100/150",
					ParentId: "root",
				},
			},
		}
		c.JSON(http.StatusOK, documentList.Documents)
	})
}
