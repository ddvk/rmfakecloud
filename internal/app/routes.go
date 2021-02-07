package app

import (
	"net/http"

	"github.com/ddvk/rmfakecloud/internal/ui"
	"github.com/gin-gonic/gin"
)

func (self *App) registerRoutes(router *gin.Engine) {

	router.GET("/health", func(c *gin.Context) {
		count := self.hub.ClientCount()
		c.String(http.StatusOK, "Working, %d clients", count)
	})
	// register  a new device
	router.POST("/token/json/2/device/new", self.newDevice)

	//service locator
	router.GET("/service/json/1/:service", self.locateService)

	//routes needing authentitcation
	authRoute := router.Group("/")
	authRoute.Use(self.authMiddleware())
	{
		ui.RegisterUIAuth(authRoute, self.metaStorer, self.userStorer)

		// renew device acces token
		authRoute.POST("/token/json/2/user/new", self.newUserToken)

		//unregister device
		authRoute.POST("/token/json/3/device/delete", func(c *gin.Context) {
			c.String(http.StatusNoContent, "")
		})

		// doucment notifications
		authRoute.GET("/notifications/ws/json/1", self.connectWebSocket)

		authRoute.PUT("/document-storage/json/2/upload/request", self.uploadRequest)

		authRoute.PUT("/document-storage/json/2/upload/update-status", self.updateStatus)

		authRoute.PUT("/document-storage/json/2/delete", self.deleteDocument)

		authRoute.GET("/document-storage/json/2/docs", self.listDocuments)

		// send email
		authRoute.POST("/api/v2/document", self.sendEmail)
		// hwr
		authRoute.POST("/api/v1/page", self.handleHwr)
		//livesync
		authRoute.GET("/livesync/ws/json/2/:authid/sub", func(c *gin.Context) {
			self.hub.ConnectWs(c.Writer, c.Request)
		})
	}
}
