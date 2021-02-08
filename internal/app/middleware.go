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

func (app *App) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := app.getUserClaims(c)

		if err != nil {
			log.Warn("token parsing:", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or incorrect token"})
			return
		}

		uid := strings.TrimPrefix(claims.Profile.UserId, "auth0|")
		c.Set(userID, uid)
		log.Info("got a user from token: ", uid)
		c.Next()
	}
}

var ignoreBodyLogging = []string{"/storage", "/api/v2/document"}

func requestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		log.Debugln("header ", c.Request.Header)
		for _, skip := range ignoreBodyLogging {
			if skip == c.Request.URL.Path {
				log.Debugln("body logging ignored")
				c.Next()
				return
			}
		}

		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ := ioutil.ReadAll(tee)
		c.Request.Body = ioutil.NopCloser(&buf)
		log.Debugln("body: ", string(body))
		c.Next()
	}
}
