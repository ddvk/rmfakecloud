package ui

import "github.com/golang-jwt/jwt/v4"

// WebUserClaims the claims
type WebUserClaims struct {
	UserID    string `json:"UserID"`
	BrowserID string `json:"BrowserID"`
	Email     string
	Scopes    string `json:"scopes,omitempty"`
	Roles     []string
	jwt.RegisteredClaims
}

// WebUsage used for the uid
const WebUsage = "web"

// AdminRole is admin
const AdminRole = "Admin"
