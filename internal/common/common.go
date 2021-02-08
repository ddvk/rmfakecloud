package common

import "github.com/dgrijalva/jwt-go"

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

func SignToken(claims jwt.Claims, key []byte) (string, error) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString(key)
}
