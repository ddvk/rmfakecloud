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
	AuthLog    = "[auth-middleware]"
	RequestLog = "[requestlogging-middleware]"
)

func (app *App) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := app.getUserClaims(c)

		if err != nil {
			log.Warn(AuthLog, "token parsing:", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or incorrect token"})
			return
		}

		uid := strings.TrimPrefix(claims.Profile.UserId, "auth0|")
		c.Set(UserID, uid)
		c.Set(DeviceId, claims.DeviceId)
		log.Infof("%s got userId: %s deviceId: %s ", AuthLog, uid, claims.DeviceId)
		c.Next()
	}
}

var ignoreBodyLogging = []string{"/storage", "/api/v2/document"}

func requestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		log.Debugln(RequestLog, "header ", c.Request.Header)
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
		log.Debugln(RequestLog, "body: ", string(body))
		c.Next()
	}
}
