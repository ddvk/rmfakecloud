package fs

import "github.com/golang-jwt/jwt"

// StorageClaim used for file retrieval
type StorageClaim struct {
	DocumentID string `json:"documentId"`
	UserID     string `json:"userId"`
	jwt.StandardClaims
}
