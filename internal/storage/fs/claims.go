package fs

import "github.com/golang-jwt/jwt/v4"

// StorageClaim used for file retrieval
type StorageClaim struct {
	DocumentID string `json:"documentId"`
	UserID     string `json:"userId"`
	jwt.RegisteredClaims
}

const (
	WriteScope = "write"
	ReadScope  = "read"
)
