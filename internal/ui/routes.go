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
		c.FileFromFS(indexReplacement, app)
	})

	r := router.Group("/ui/api")
	r.POST("register", app.register)
	r.POST("login", app.login)

	gr := r.Group("")
	gr.Use(app.authMiddleware())

	gr.GET("newcode", app.newCode)
	gr.GET("list", app.listDocuments)
	gr.GET("users", app.getAppUsers)
	gr.GET("users/:userid", app.getUser)
}

func (app *ReactAppWrapper) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := common.GetToken(c)

		if err != nil {
			log.Warn("token parsing", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or incorrect token"})
			return
		}
		claims := &common.WebUserClaims{}
		err = common.ClaimsFromToken(claims, token, app.cfg.JWTSecretKey)
		if err != nil {
			log.Warn("token verification, ", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or incorrect token"})
			return
		}

		if claims.Subject != common.WebUsage {
			log.Warn("wrong token subject: ", claims.Subject)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or incorrect token"})
			return
		}
		log.Debug(claims)

		uid := claims.UserId
		c.Set(userID, uid)
		log.Info("got a user from token: ", uid)
		c.Next()
	}
}
