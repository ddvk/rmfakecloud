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
	sync10     = "sync:default"
	sync15     = "sync:fox"
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

		var isDefault = false
		for _, s := range scopes {
			if s == sync10 {
				isDefault = true
				break
			}
		}
		if isDefault {
			log.Warn("missing sync:")
			// c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		uid := strings.TrimPrefix(claims.Profile.UserID, "auth0|")
		c.Set(userIDKey, uid)
		c.Set(deviceIDKey, claims.DeviceID)
		log.Infof("%s UserId: %s deviceId: %s ", authLog, uid, claims.DeviceID)
		c.Next()
	}
}

var ignoreBodyLogging = []string{"/storage", "/blobstorage", "/api/v2/document", "/ui/api/documents/upload", "/v1/reports"}

func requestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// if log.IsLevelEnabled(log.DebugLevel) {
		// 	var str bytes.Buffer
		// 	for k, v := range c.Request.Header {
		// 		var ln string
		// 		if k != "Authorization" {
		// 			ln = fmt.Sprintf("%s\t%s\n", k, v)
		// 		} else {
		// 			ln = fmt.Sprintf("%s\t\n", k)
		// 		}
		// 		str.WriteString(ln)
		// 	}
		// 	log.Debugln(requestLog, "headers: \n", str.String())
		// }
		for _, skip := range ignoreBodyLogging {
			if strings.Index(c.Request.URL.Path, skip) == 0 {
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
