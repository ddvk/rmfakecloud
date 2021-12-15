package common

import (
	"errors"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var signingMethod = jwt.SigningMethodHS256

// GetToken gets the token from the headers
func GetToken(c *gin.Context) (string, error) {
	auth := c.Request.Header["Authorization"]

	if len(auth) < 1 {
		return "", errors.New("missing auth header")
	}
	token := strings.Split(auth[0], " ")
	if len(token) < 2 {
		return "", errors.New("wrong token format")
	}
	strToken := token[1]
	return strToken, nil

}

// ClaimsFromToken parses the claims from the token
func ClaimsFromToken(claim jwt.Claims, token string, key []byte) error {
	_, err := jwt.ParseWithClaims(token, claim,
		func(token *jwt.Token) (interface{}, error) {
			return key, nil
		}, jwt.WithValidMethods([]string{signingMethod.Name}))

	return err
}

// SignClaims signs the claims i.e. creates a token
func SignClaims(claims jwt.Claims, key []byte) (string, error) {
	jwtToken := jwt.NewWithClaims(signingMethod, claims)
	jwtToken.Header["kid"] = "1"
	return jwtToken.SignedString(key)
}

var nameSeparators = regexp.MustCompile(`[./\\]`)

func Sanitize(param string) string {
	return nameSeparators.ReplaceAllString(param, "")
}

// QueryS sanitize the param
func QueryS(param string, c *gin.Context) string {
	p := c.Query(param)
	return Sanitize(p)
}

// ParamS sanitize the param
func ParamS(param string, c *gin.Context) string {
	p := c.Param(param)
	return Sanitize(p)
}
