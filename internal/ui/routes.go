package ui

import (
	"net/http"

	"github.com/ddvk/rmfakecloud/internal/webassets"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// RegisterRoutes the apps routes
func (app *ReactAppWrapper) RegisterRoutes(router *gin.Engine) {
	router.StaticFS(app.prefix, app)

	router.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("/favicon.ico", webassets.Assets)
	})

	//hack for index.html
	router.NoRoute(func(c *gin.Context) {
		method := c.Request.Method
		if method == http.MethodGet {
			c.FileFromFS(indexReplacement, app)
		} else {
			c.AbortWithStatus(http.StatusNotFound)
		}
	})

	r := router.Group("/ui/api")
	r.POST("register", app.register)
	r.POST("login", app.login)

	//with authentication
	auth := r.Group("")
	auth.Use(app.authMiddleware())
	auth.GET("sync", func(c *gin.Context) {
		uid := c.GetString(userID)
		br := c.GetString(browserID)
		log.Info("browser", br)
		app.h.NotifySync(uid, br)
	})

	auth.GET("newcode", app.newCode)
	auth.POST("resetPassword", app.resetPassword)

	auth.GET("documents", app.listDocuments)
	auth.GET("documents/:docid", app.getDocument)
	auth.POST("documents/upload", app.createDocument)
	auth.DELETE("documents/:docid", app.deleteDocument)
	//move, rename
	auth.PUT("documents", app.updateDocument)

	//admin
	admin := auth.Group("")
	admin.Use(app.adminMiddleware())
	admin.GET("users/:userid", app.getUser)
	admin.GET("users", app.getAppUsers)
}
