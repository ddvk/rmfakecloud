package common

import (
	"errors"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

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
		})

	if err != nil {
		return err
	}
	return nil

}

// SignClaims signs the claims i.e. creates a token
func SignClaims(claims jwt.Claims, key []byte) (string, error) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtToken.Header["kid"] = "1"
	return jwtToken.SignedString(key)
}

// CodeConnector matches a code a to users
type CodeConnector interface {
	NewCode(string) (string, error)
	ConsumeCode(string) (string, error)
}
