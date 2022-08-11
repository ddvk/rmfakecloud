package ui

import (
	"net/http"
	"strings"

	"github.com/zgs225/rmfakecloud/internal/common"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func IsAdmin(c *gin.Context) bool {
	return c.GetBool(AdminRole)
}

func (app *ReactAppWrapper) adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAdmin(c) {
			log.Warn("not admin")
			c.AbortWithStatus(http.StatusForbidden)
		}
	}
}

func (app *ReactAppWrapper) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(cookieName)
		if err == http.ErrNoCookie {
			log.Warn("missing cookie, trying headers")
			token, err = common.GetToken(c)
		}

		if err != nil {
			log.Warn("[ui-authmiddleware] token parsing, ", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or incorrect token"})
			return
		}
		claims := &WebUserClaims{}
		err = common.ClaimsFromToken(claims, token, app.cfg.JWTSecretKey)
		if err != nil {
			log.Warn("[ui-authmiddleware] token verification, ", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or incorrect token"})
			return
		}

		if claims.Audience != WebUsage {
			log.Warn("wrong token audience: ", claims.Audience)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or incorrect token"})
			return
		}

		scopes := strings.Fields(claims.Scopes)
		newsync := false
		for _, s := range scopes {
			switch s {
			case isSync15Key:
				newsync = true
			}
		}

		if newsync {
			c.Set("backend", app.backend15)
		} else {
			c.Set("backend", app.backend10)

		}
		uid := claims.UserID
		brid := claims.BrowserID
		c.Set(userIDContextKey, uid)
		c.Set(browserIDContextKey, brid)
		c.Set(isSync15Key, newsync)
		for _, r := range claims.Roles {
			if r == AdminRole {
				c.Set(AdminRole, true)
				break
			}
		}
		log.Info("[ui-authmiddleware] User from token: ", uid)
		c.Next()
	}
}
