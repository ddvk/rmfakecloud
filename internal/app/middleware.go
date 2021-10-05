package app

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	authLog     = "[auth-middleware]"
	requestLog  = "[requestlogging-middleware]"
	syncDefault = "sync:default"
	syncFox     = "sync:fox"
)

func (app *App) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := app.getUserClaims(c)

		if err != nil {
			log.Warn(authLog, "token parsing:", err)
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		scopes := strings.Fields(claims.Scopes)

		var isSync15 = false
		for _, s := range scopes {
			if s == syncFox {
				isSync15 = true
				break
			}
		}
		if isSync15 {
			log.Info("Using sync 1.5")
		}

		uid := strings.TrimPrefix(claims.Profile.UserID, "auth0|")
		c.Set(userIDKey, uid)
		c.Set(deviceIDKey, claims.DeviceID)
		log.Infof("%s UserId: %s deviceId: %s newSync: %t", authLog, uid, claims.DeviceID, isSync15)
		c.Next()
	}
}

var dontLogBody = map[string]bool{
	"/storage":                 true,
	"/blobstorage":             true,
	"/api/v2/document":         true,
	"/ui/api/documents/upload": true,
	"/v1/reports":              true,
}

func requestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		if log.IsLevelEnabled(log.TraceLevel) {
			var str bytes.Buffer
			for k, v := range c.Request.Header {
				var ln string
				if k != "Authorization" {
					ln = fmt.Sprintf("%s\t%s\n", k, v)
				} else {
					ln = fmt.Sprintf("%s\t\n", k)
				}
				str.WriteString(ln)
			}
			log.Traceln(requestLog, "headers: \n", str.String())
		}

		if _, ok := dontLogBody[c.Request.URL.Path]; ok {
			c.Next()
			return
		}

		if log.IsLevelEnabled(log.DebugLevel) {
			var buf bytes.Buffer
			tee := io.TeeReader(c.Request.Body, &buf)
			body, _ := ioutil.ReadAll(tee)
			c.Request.Body = ioutil.NopCloser(&buf)
			log.Debugln(requestLog, "body: ", string(body))
		}
		c.Next()
	}
}
