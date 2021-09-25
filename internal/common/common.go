package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
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

	return err
}

// SignClaims signs the claims i.e. creates a token
func SignClaims(claims jwt.Claims, key []byte) (string, error) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtToken.Header["kid"] = "1"
	return jwtToken.SignedString(key)
}

func Sign(parts []string, key []byte) string {
	h := hmac.New(sha256.New, key)
	for _, s := range parts {
		h.Write([]byte(s))
	}
	hs := h.Sum(nil)
	s := hex.EncodeToString(hs)
	return s
}

func VerifySignature(parts []string, exp, signature string, key []byte) error {
	expected := Sign(parts, key)
	expiration, err := strconv.Atoi(exp)
	if err != nil {
		return err
	}
	if expiration < int(time.Now().Unix()) {
		return errors.New("expired")
	}

	if subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) != 1 {
		return errors.New("wrong signature")
	}

	return nil
}

// CodeConnector matches a code to users
type CodeConnector interface {
	//NewCode generates one time code for a user
	NewCode(uid string) (code string, err error)

	//ConsumeCode a code and returns the uid if ofound
	ConsumeCode(code string) (uid string, err error)
}
