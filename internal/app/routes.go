package app

import (
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var mb uint64 = 2 << 20
var kb uint64 = 2 << 10

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

	router.GET("/list", func(c *gin.Context) {
		jsn := `{"kind": "storage#object", "prefixes": [], "items": [{"kind": "storage#object", "id": "test-bucket/blob1/1626900813", "selfLink": "/storage/v1/b/test-bucket/o/blob1", "name": "blob1", "bucket": "test-bucket", "generation": "1626900813", "metageneration": "3", "contentType": "application/octet-stream", "timeCreated": "2021-07-21T20:53:33.584223Z", "updated": "2021-07-21T20:53:33.584223Z", "storageClass": "STANDARD", "timeStorageClassUpdated": "2021-07-21T20:53:33.584223Z", "size": "10", "md5Hash": "al+569bI6n77U9BxBT73eA==", "mediaLink": "http://0.0.0.0:9023/download/storage/v1/b/test-bucket/o/blob1?generation=1626900813&alt=media", "crc32c": "hEGyZA==", "etag": "al+569bI6n77U9BxBT73eA=="}, {"kind": "storage#object", "id": "test-bucket/blob2/1626822255", "selfLink": "/storage/v1/b/test-bucket/o/blob2", "name": "blob2", "bucket": "test-bucket", "generation": "1626822255", "metageneration": "2", "contentType": "text/plain", "timeCreated": "2021-07-20T23:04:15.499489Z", "updated": "2021-07-20T23:04:15.499489Z", "storageClass": "STANDARD", "timeStorageClassUpdated": "2021-07-20T23:04:15.499489Z", "size": "5", "md5Hash": "rQI0gpIFuQMxlrqBj3qHKw==", "mediaLink": "http://0.0.0.0:9023/download/storage/v1/b/test-bucket/o/blob2?generation=1626822255&alt=media", "crc32c": "QK7soQ==", "etag": "rQI0gpIFuQMxlrqBj3qHKw=="}, {"kind": "storage#object", "id": "test-bucket/abcd/1626900869", "selfLink": "/storage/v1/b/test-bucket/o/abcd", "name": "abcd", "bucket": "test-bucket", "generation": "1626900869", "metageneration": "2", "contentType": "text/plain", "timeCreated": "2021-07-21T20:54:29.504431Z", "updated": "2021-07-21T20:54:29.504431Z", "storageClass": "STANDARD", "timeStorageClassUpdated": "2021-07-21T20:54:29.504431Z", "size": "4", "md5Hash": "4vxxTEcn7pOV8yTNLn8zHw==", "mediaLink": "http://0.0.0.0:9023/download/storage/v1/b/test-bucket/o/abcd?generation=1626900869&alt=media", "crc32c": "ksgKMQ==", "etag": "4vxxTEcn7pOV8yTNLn8zHw=="}]}`
		c.String(http.StatusOK, jsn)
	})
	router.PUT("/list", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/download/:blob", func(c *gin.Context) {
		body, _ := ioutil.ReadAll(c.Request.Body)
		log.Info("blob:", string(body))
	})
	router.PUT("/upload/:blob", func(c *gin.Context) {

		body, _ := ioutil.ReadAll(c.Request.Body)
		log.Info("blob:", string(body))
		c.Status(http.StatusOK)
	})

	//some telemetry stuff from ping.
	router.POST("/v1/reports", func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)

		if err != nil {
			log.Warn("cant parse telemetry, ignored")
			c.Status(http.StatusOK)
			return
		}
		log.Info(hex.Dump(body))
		c.Status(http.StatusOK)
	})

	//routes needing api authentitcation
	authRoutes := router.Group("/")
	authRoutes.Use(app.authMiddleware())
	{

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
			//TODO: not implemented yet
			authid := c.Param("authid")
			log.Info("authid: ", authid)
			c.AbortWithStatus(http.StatusNoContent)
		})

		authRoutes.POST("/api/v1/signed-urls/downloads", app.blobStorageDownload)
		authRoutes.POST("/api/v1/signed-urls/uploads", app.blobStorageUpload)
	}
}
