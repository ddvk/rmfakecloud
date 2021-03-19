package app

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	authLog    = "[auth-middleware]"
	requestLog = "[requestlogging-middleware]"
)

func (app *App) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := app.getUserClaims(c)

		if err != nil {
			log.Warn(authLog, "token parsing:", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or incorrect token"})
			return
		}

		if claims.Scopes != "sync:default" {
			log.Warn(authLog, " wrong scope, proably old token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing scope"})
			return
		}

		uid := strings.TrimPrefix(claims.Profile.UserID, "auth0|")
		c.Set(userIDKey, uid)
		c.Set(deviceIDKey, claims.DeviceID)
		log.Infof("%s got userId: %s deviceId: %s ", authLog, uid, claims.DeviceID)
		c.Next()
	}
}

var ignoreBodyLogging = []string{"/storage", "/api/v2/document"}

func requestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		log.Debugln(requestLog, "header ", c.Request.Header)
		for _, skip := range ignoreBodyLogging {
			if strings.Index(c.Request.URL.Path, skip) == 0 {
				log.Debugln("body logging ignored")
				c.Next()
				return
			}
		}

		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ := ioutil.ReadAll(tee)
		c.Request.Body = ioutil.NopCloser(&buf)
		log.Debugln(requestLog, "body: ", string(body))
		c.Next()
	}
}
