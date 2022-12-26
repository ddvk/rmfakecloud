package app

import (
	"io/ioutil"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var mb uint64 = 2 << 20
var kb uint64 = 2 << 10

const (
	integrationKey = "integrationid"
	folderKey      = "folderid"
	fileKey        = "file"
)

func (app *App) registerRoutes(router *gin.Engine) {

	router.GET("/health", func(c *gin.Context) {
		count := app.hub.ClientCount()
		gnum := runtime.NumGoroutine()
		ms := runtime.MemStats{}
		runtime.ReadMemStats(&ms)
		live := (ms.Mallocs - ms.Frees) / kb
		sysmb := ms.Sys / mb
		c.String(http.StatusOK, "Working, %d clients, gn: %d, mem: %dkb sys: %dmb", count, gnum, live, sysmb)
	})
	// register  a new device
	router.POST("/token/json/2/device/new", app.newDevice)

	// renew device acces token
	router.POST("/token/json/2/user/new", app.newUserToken)

	//unregister device
	router.POST("/token/json/2/device/delete", app.deleteDevice)
	router.POST("/token/json/3/device/delete", app.deleteDevice)

	//service locator
	router.GET("/service/json/1/:service", app.locateService)

	//some beta stuff from internal.cloud
	router.GET("/settings/v1/beta", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"enrolled": false, "available": true})
	})

	router.POST("/settings/v1/beta", func(c *gin.Context) {
		body, _ := ioutil.ReadAll(c.Request.Body)
		log.Info("enrolling in the beta:", string(body))
		c.Status(http.StatusOK)
	})

	//some telemetry stuff from ping.
	router.POST("/v1/reports", func(c *gin.Context) {
		_, err := ioutil.ReadAll(c.Request.Body)

		if err != nil {
			log.Warn("cant parse telemetry, ignored")
			c.Status(http.StatusOK)
			return
		}
		c.Status(http.StatusOK)
	})

	//routes needing api authentitcation
	authRoutes := router.Group("/")
	authRoutes.Use(app.authMiddleware())
	{

		// document notifications
		authRoutes.GET("/notifications/ws/json/1", app.connectWebSocket)

		authRoutes.PUT("/document-storage/json/2/upload/request", app.uploadRequest)

		authRoutes.PUT("/document-storage/json/2/upload/update-status", app.updateStatus)

		authRoutes.PUT("/document-storage/json/2/delete", app.deleteDocument)

		authRoutes.GET("/document-storage/json/2/docs", app.listDocuments)

		// send email
		authRoutes.POST("/api/v2/document", app.sendEmail)
		authRoutes.POST("/share/v1/email", app.sendEmail)
		// hwr
		authRoutes.POST("/api/v1/page", app.handleHwr)
		authRoutes.POST("/convert/v1/handwriting", app.handleHwr)

		// read on remarkable extension
		authRoutes.POST("/doc/v1/files", app.uploadDoc)
		// v2
		authRoutes.POST("/doc/v2/files", app.uploadDocV2)
		authRoutes.OPTIONS("/doc/v2/files", func(c *gin.Context) {
			//TODO: seems to be a cors preflight
			c.Status(http.StatusOK)
		})

		// integrations
		authRoutes.GET("/integrations/v1/:"+integrationKey+"/folders/:"+folderKey, app.integrationsList)
		authRoutes.GET("/integrations/v1/:"+integrationKey+"/files/:"+fileKey+"/metadata", app.integrationsGetMetadata)
		authRoutes.GET("/integrations/v1/:"+integrationKey+"/files/:"+fileKey, app.integrationsGetFile)
		authRoutes.POST("/integrations/v1/:"+integrationKey+"/files/:"+folderKey, app.integrationsUpload)
		authRoutes.GET("/integrations/v1/", app.integrations)

		// sync15
		authRoutes.POST("/api/v1/signed-urls/downloads", app.blobStorageDownload)
		authRoutes.POST("/api/v1/signed-urls/uploads", app.blobStorageUpload)
		authRoutes.POST("/api/v1/sync-complete", app.syncComplete)

		authRoutes.POST("/sync/v2/signed-urls/downloads", app.blobStorageDownload)
		authRoutes.POST("/sync/v2/signed-urls/uploads", app.blobStorageUpload)
		authRoutes.POST("/sync/v2/sync-complete", app.syncCompleteV2)
	}
}
