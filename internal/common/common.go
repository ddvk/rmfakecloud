package common

import (
	"errors"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

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

func SignClaims(claims jwt.Claims, key []byte) (string, error) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString(key)
}

type CodeConnector interface {
	NewCode(string) (string, error)
	ConsumeCode(string) (string, error)
}
