package ui

import (
	"net/http"

	"github.com/ddvk/rmfakecloud/internal/common"
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

func (app *ReactAppWrapper) adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !c.GetBool("Admin") {
			log.Warn("not admin")
			c.AbortWithStatus(http.StatusForbidden)
		}
	}
}

func (app *ReactAppWrapper) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := common.GetToken(c)

		if err != nil {
			log.Warn("[ui-authmiddleware] token parsing, ", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or incorrect token"})
			return
		}
		claims := &common.WebUserClaims{}
		err = common.ClaimsFromToken(claims, token, app.cfg.JWTSecretKey)
		if err != nil {
			log.Warn("[ui-authmiddleware] token verification, ", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or incorrect token"})
			return
		}

		if claims.Audience != common.WebUsage {
			log.Warn("wrong token audience: ", claims.Audience)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or incorrect token"})
			return
		}
		uid := claims.UserID
		brid := claims.BrowserID
		c.Set(userID, uid)
		c.Set(browserID, brid)
		for _, r := range claims.Roles {
			if r == "Admin" {
				c.Set("Admin", true)
				break
			}
		}
		log.Info("[ui-authmiddleware] User from token: ", uid)
		c.Next()
	}
}
