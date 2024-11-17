package ui

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// RegisterRoutes the apps routes
func (app *ReactAppWrapper) RegisterRoutes(router *gin.Engine) {
	router.StaticFS(app.prefix, app)

	router.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("/favicon.ico", app.fs)
	})
	router.GET("/robots.txt", func(c *gin.Context) {
		c.FileFromFS("/robots.txt", app.fs)
	})
	router.GET("/pdf.worker.js", func(c *gin.Context) {
		c.FileFromFS("/pdf.worker.js", app.fs)
	})

	//hack for index.html
	router.NoRoute(func(c *gin.Context) {
		uri := c.Request.RequestURI
		log.Info(uri)
		if strings.HasPrefix(uri, "/api") ||
			strings.HasPrefix(uri, "/ui/api") ||
			c.Request.Method != http.MethodGet {

			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.FileFromFS(indexReplacement, app)
	})

	r := router.Group("/ui/api")
	r.POST("register", app.register)
	r.POST("login", app.login)
	r.GET("logout", func(c *gin.Context) {
		c.SetCookie(cookieName, "/", -1, "", "", false, true)
		c.Status(http.StatusOK)
	})
	//with authentication
	auth := r.Group("")
	auth.Use(app.authMiddleware())
	auth.HEAD("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	auth.GET("sync", func(c *gin.Context) {
		uid := c.GetString(userIDContextKey)
		br := c.GetString(browserIDContextKey)
		log.Info("browser", br)
		app.h.NotifySync(uid, br)
	})

	auth.GET("newcode", app.newCode)
	// auth.GET("profile", app.newCode)
	auth.POST("profile", app.changePassword)
	// auth.POST("changeEmail", app.changePassword)

	auth.GET("documents", app.listDocuments)
	auth.GET("documents/:docid", app.getDocument)
	auth.POST("documents/upload", app.createDocument)

	//move, rename
	auth.DELETE("documents/:docid", app.deleteDocument)
	auth.PUT("documents", app.updateDocument)
	auth.POST("folders", app.createFolder)
	auth.GET("documents/:docid/metadata", app.getDocumentMetadata)

	// integrations
	auth.GET("integrations", app.listIntegrations)
	auth.POST("integrations", app.createIntegration)
	auth.GET("integrations/:intid", app.getIntegration)
	auth.PUT("integrations/:intid", app.updateIntegration)
	auth.DELETE("integrations/:intid", app.deleteIntegration)

	auth.GET("integrations/:intid/explore/*path", app.exploreIntegration)
	auth.GET("integrations/:intid/metadata/*path", app.getMetadataIntegration)
	auth.GET("integrations/:intid/download/*path", app.downloadThroughIntegration)

	//admin
	admin := auth.Group("")
	admin.Use(app.adminMiddleware())
	admin.GET("users/:userid", app.getUser)
	admin.DELETE("users/:userid", app.deleteUser)
	admin.PUT("users", app.updateUser)
	admin.POST("users", app.createUser)
	admin.GET("users", app.getAppUsers)
}
