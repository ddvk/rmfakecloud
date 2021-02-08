package common

import "github.com/dgrijalva/jwt-go"

func ClaimsFromToken(claim jwt.Claims, token string, key []byte) error {
	return nil

}

func SignToken(claim jwt.Claims, key []byte) (string, error) {
	return "", nil
}
