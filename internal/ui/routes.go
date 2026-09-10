package ui

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// RegisterRoutes the apps routes
func (app *ReactAppWrapper) RegisterRoutes(router *gin.Engine) {
	router.StaticFS(app.prefix, app.fs)

	router.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("/favicon.ico", app.fs)
	})
	router.GET("/robots.txt", func(c *gin.Context) {
		c.FileFromFS("/robots.txt", app.fs)
	})

	// HTML pages (XML → XSLT)
	router.GET("/", app.pageHome)
	router.GET("/help", app.pageHelp)
	router.GET("/login", app.pageLogin)
	router.POST("/login", app.formLogin)
	router.POST("/logout", app.formLogout)
	router.GET("/logout", app.formLogout)
	router.GET("/connect", app.pageConnect)
	router.GET("/profile", app.pageProfile)
	router.POST("/profile/theme", app.formProfileTheme)
	router.POST("/profile/password", app.formProfilePassword)
	router.POST("/profile/passkeys/:id/delete", app.formPasskeyDelete)
	router.GET("/documents", app.pageDocuments)
	router.GET("/documents/:docid", app.pagePDF)
	router.POST("/documents/upload", app.formUploadDocument)
	router.POST("/documents/folder", app.formCreateFolder)
	router.POST("/documents/:docid/delete", app.formDeleteDocument)
	router.GET("/integrations", app.pageIntegrations)
	router.POST("/integrations", app.formCreateIntegration)
	router.POST("/integrations/:intid/delete", app.formDeleteIntegration)
	router.GET("/admin", app.pageAdmin)
	router.POST("/admin/users", app.formCreateUser)
	router.POST("/admin/users/:userid/delete", app.formDeleteUser)
	router.GET("/admin/themes", app.pageThemeStudio)
	router.POST("/admin/themes/save", app.formSaveTheme)
	router.POST("/admin/themes/:id/publish", app.formPublishTheme)
	router.POST("/admin/themes/:id/delete", app.formDeleteTheme)
	router.GET("/screenshare", app.pageScreenShare)

	router.NoRoute(func(c *gin.Context) {
		uri := c.Request.RequestURI
		log.Info(uri)
		if strings.HasPrefix(uri, "/api") ||
			strings.HasPrefix(uri, "/ui/api") ||
			c.Request.Method != http.MethodGet {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		app.page404(c)
	})

	r := router.Group("/ui/api")
	r.POST("register", app.register)
	r.POST("login", app.login)
	r.GET("webauthn/status", app.webAuthnStatus)
	r.POST("webauthn/login/begin", app.webAuthnLoginBegin)
	r.POST("webauthn/login/finish", app.webAuthnLoginFinish)
	r.GET("themes/assets/:name", app.getThemeAsset)
	r.GET("themes/:id", app.getThemePublic)
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
		uid := userID(c)
		br := c.GetString(browserIDContextKey)
		log.Info("browser", br)
		app.h.NotifySync(uid, br)
	})

	auth.GET("newcode", app.newCode)
	auth.GET("code", app.codeStatus)

	auth.GET("passcode/resets", app.listPasscodeResets)
	auth.POST("passcode/resets/:uuid/approve", app.approvePasscodeReset)
	auth.DELETE("passcode/resets/:uuid", app.dismissPasscodeReset)

	auth.POST("profile", app.changePassword)
	auth.GET("profile/theme", app.getProfileTheme)
	auth.PUT("profile/theme", app.putProfileTheme)

	auth.POST("webauthn/register/begin", app.webAuthnRegisterBegin)
	auth.POST("webauthn/register/finish", app.webAuthnRegisterFinish)
	auth.GET("webauthn/credentials", app.webAuthnListCredentials)
	auth.DELETE("webauthn/credentials/:id", app.webAuthnDeleteCredential)

	auth.GET("themes", app.listThemes)

	auth.GET("documents", app.listDocuments)
	auth.GET("documents/:docid", app.getDocument)
	auth.GET("documents/:docid/epub/*path", app.getEpubPath)
	auth.GET("documents/:docid/page/:pagenum", app.getDocumentPage)
	auth.POST("documents/upload", app.createDocument)

	auth.DELETE("documents/:docid", app.deleteDocument)
	auth.PUT("documents", app.updateDocument)
	auth.POST("folders", app.createFolder)
	auth.GET("documents/:docid/metadata", app.getDocumentMetadata)

	auth.GET("integrations", app.listIntegrations)
	auth.POST("integrations", app.createIntegration)
	auth.GET("integrations/:intid", app.getIntegration)
	auth.PUT("integrations/:intid", app.updateIntegration)
	auth.DELETE("integrations/:intid", app.deleteIntegration)

	auth.GET("integrations/:intid/explore/*path", app.exploreIntegration)
	auth.GET("integrations/:intid/metadata/*path", app.getMetadataIntegration)
	auth.GET("integrations/:intid/download/*path", app.downloadThroughIntegration)

	ss := auth.Group("screenshare")
	ss.GET("room", app.screenshareJoinActive)
	ss.GET("room/:roomId", app.screenshareGetRoom)
	ss.GET("offer", app.screenshareGetOffer)
	ss.POST("room/:roomId/answer", app.screenshareSendAnswer)
	ss.DELETE("room/:roomId", app.screenshareDeleteRoom)

	admin := auth.Group("")
	admin.Use(app.adminMiddleware())
	admin.GET("users/:userid", app.getUser)
	admin.DELETE("users/:userid", app.deleteUser)
	admin.PUT("users", app.updateUser)
	admin.POST("users", app.createUser)
	admin.GET("users", app.getAppUsers)
	admin.GET("logs", app.getAdminLogs)
	admin.POST("themes", app.saveTheme)
	admin.PUT("themes/:id", app.updateTheme)
	admin.POST("themes/:id/publish", app.publishTheme)
	admin.DELETE("themes/:id", app.deleteTheme)
}
