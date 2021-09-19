package app

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/common"
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
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}


		if claims.Audience != common.APIUsage {
			log.Warn("wrong token audience: ", claims.Audience)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or incorrect token"})
			return
		}
        scopes := strings.Split(claims.Scopes, " ")

        var isDefault = false
        for _, s := range scopes {
            if s == "sync:default" {
                isDefault = true
            }
        }
        if !isDefault {
            log.Warn("missing sync:default scope")
        }


		uid := strings.TrimPrefix(claims.Profile.UserID, "auth0|")
		c.Set(userIDKey, uid)
		c.Set(deviceIDKey, claims.DeviceID)
		log.Infof("%s UserId: %s deviceId: %s ", authLog, uid, claims.DeviceID)
		c.Next()
	}
}

var ignoreBodyLogging = []string{"/storage", "/api/v2/document", "/ui/api/documents/upload"} //, "/v1/reports"}

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
