package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *App) registerRoutes(router *gin.Engine) {

	router.GET("/health", func(c *gin.Context) {
		count := app.hub.ClientCount()
		c.String(http.StatusOK, "Working, %d clients", count)
	})
	// register  a new device
	router.POST("/token/json/2/device/new", app.newDevice)

	// renew device acces token
	router.POST("/token/json/2/user/new", app.newUserToken)

	//service locator
	router.GET("/service/json/1/:service", app.locateService)

	app.docStorer.RegisterRoutes(router)
	app.ui.RegisterRoutes(router)

	//routes needing api authentitcation
	authRoutes := router.Group("/")
	authRoutes.Use(app.authMiddleware())
	{

		//unregister device
		authRoutes.POST("/token/json/3/device/delete", func(c *gin.Context) {
			c.String(http.StatusNoContent, "")
		})

		// doucment notifications
		authRoutes.GET("/notifications/ws/json/1", app.connectWebSocket)

		authRoutes.PUT("/document-storage/json/2/upload/request", app.uploadRequest)

		authRoutes.PUT("/document-storage/json/2/upload/update-status", app.updateStatus)

		authRoutes.PUT("/document-storage/json/2/delete", app.deleteDocument)

		authRoutes.GET("/document-storage/json/2/docs", app.listDocuments)

		// send email
		authRoutes.POST("/api/v2/document", app.sendEmail)
		// hwr
		authRoutes.POST("/api/v1/page", app.handleHwr)
		//livesync
		authRoutes.GET("/livesync/ws/json/2/:authid/sub", func(c *gin.Context) {
			app.hub.ConnectWs(c.Writer, c.Request)
		})
	}
}
