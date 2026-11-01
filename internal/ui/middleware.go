package ui

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	backendVersionKey string = "BackendVersion"
)

// parseWebSessionCookie reads the auth cookie and returns verified WebUserClaims.
// Returns an error if the cookie is missing, unparseable, or carries the wrong audience.
func (app *ReactAppWrapper) parseWebSessionCookie(c *gin.Context) (*WebUserClaims, error) {
	token, err := c.Cookie(cookieName)
	if err != nil {
		return nil, err
	}
	claims := &WebUserClaims{}
	if err := common.ClaimsFromToken(claims, token, app.cfg.JWTSecretKey); err != nil {
		return nil, err
	}
	if !slices.Contains(claims.Audience, WebUsage) {
		return nil, errors.New("wrong token audience")
	}
	return claims, nil
}

// IsAdmin checks if admin
func IsAdmin(c *gin.Context) bool {
	return c.GetBool(AdminRole)
}

// webAuthenticated reports whether the request carries a valid web session cookie.
// Used to decide server-side redirects without aborting the request.
func (app *ReactAppWrapper) webAuthenticated(c *gin.Context) bool {
	_, err := app.parseWebSessionCookie(c)
	return err == nil
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
		claims, err := app.parseWebSessionCookie(c)
		if errors.Is(err, http.ErrNoCookie) {
			log.Warn("missing cookie, trying headers")
			var token string
			token, err = common.GetToken(c)
			if err == nil {
				claims = &WebUserClaims{}
				if herr := common.ClaimsFromToken(claims, token, app.cfg.JWTSecretKey); herr != nil {
					err = herr
				} else if !slices.Contains(claims.Audience, WebUsage) {
					err = errors.New("wrong token audience")
				}
			}
		}
		if err != nil {
			log.Warn("[ui-authmiddleware] auth failed: ", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or incorrect token"})
			return
		}

		scopes := strings.Fields(claims.Scopes)
		c.Set(backendVersionKey, common.Sync10)
		for _, s := range scopes {
			switch s {
			case isSync15Key:
				c.Set(backendVersionKey, common.Sync15)
				break
			}
		}

		uid := common.SanitizeUid(claims.UserID)
		c.Set(userIDContextKey, uid)

		brid := claims.BrowserID
		c.Set(browserIDContextKey, brid)
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
