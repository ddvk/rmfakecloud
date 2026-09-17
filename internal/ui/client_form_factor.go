package ui

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var mobileUA = regexp.MustCompile(`(?i)(android|webos|iphone|ipad|ipod|blackberry|iemobile|opera mini|mobile|silk|kindle|fennec)`)

// clientFormFactor returns "mobile" or "desktop" from request headers.
// Prefers the Sec-CH-UA-Mobile client hint (?1 / ?0), then User-Agent.
func clientFormFactor(r *http.Request) string {
	if r == nil {
		return "desktop"
	}
	if ch := strings.TrimSpace(r.Header.Get("Sec-CH-UA-Mobile")); ch != "" {
		if ch == "?1" || strings.EqualFold(ch, "1") || strings.EqualFold(ch, "true") {
			return "mobile"
		}
		if ch == "?0" || strings.EqualFold(ch, "0") || strings.EqualFold(ch, "false") {
			return "desktop"
		}
	}
	ua := r.Header.Get("User-Agent")
	if ua != "" && mobileUA.MatchString(ua) {
		return "mobile"
	}
	return "desktop"
}

func setFormFactorHeaders(c *gin.Context) {
	c.Header("Accept-CH", "Sec-CH-UA-Mobile")
	c.Header("Vary", "User-Agent, Sec-CH-UA-Mobile")
	c.Header("Critical-CH", "Sec-CH-UA-Mobile")
}
