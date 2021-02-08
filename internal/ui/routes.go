package ui

import (
	"github.com/ddvk/rmfakecloud/internal/webassets"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes the apps routes
func (app *ReactAppWrapper) RegisterRoutes(router *gin.Engine) {
	router.StaticFS(app.prefix, app)

	router.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("/favicon.ico", webassets.Assets)
	})

	//hack for index.html
	router.NoRoute(func(c *gin.Context) {
		c.FileFromFS(indexReplacement, app)
	})

	r := router.Group("/ui/api")
	r.POST("register", app.register)
	r.POST("login", app.login)
}

// RegisterAuthRoutes routes needint auth
func (app *ReactAppWrapper) RegisterAuthRoutes(e *gin.RouterGroup) {
	r := e.Group("/ui/api")

	r.GET("newcode", app.newCode)
	r.GET("list", app.listDocuments)
	r.GET("users", app.getAppUsers)
	r.GET("users/:userid", app.getUser)
}
