package common

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// StorageClaim used for file retrieval
type StorageClaim struct {
	DocumentID string `json:"documentId"`
	UserID     string `json:"userId"`
	jwt.StandardClaims
}

// DeviceClaims device
type DeviceClaims struct {
	UserID     string `json:"auth0-userid"`
	DeviceDesc string `json:"device-desc"`
	DeviceID   string `json:"device-id"`
	Scopes     string `json:"scopes,omitempty"`
	jwt.StandardClaims
}

// UserClaims is the oauth token struct.
type UserClaims struct {
	Profile    Auth0profile `json:"auth0-profile,omitempty"`
	DeviceDesc string       `json:"device-desc"`
	DeviceID   string       `json:"device-id"`
	Scopes     string       `json:"scopes,omitempty"`
	jwt.StandardClaims
}

// Auth0profile is the oauth user struct.
type Auth0profile struct {
	UserID        string `json:"UserID"`
	IsSocial      bool
	ClientID      string `json:"ClientID"`
	Connection    string
	Name          string `json:"Name"`
	Nickname      string `json:"NickName"`
	GivenName     string
	FamilyName    string
	Email         string
	EmailVerified bool
	Picture       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// WebUserClaims the claims
type WebUserClaims struct {
	UserID string `json:"UserID"`
	Email  string
	Scopes string `json:"scopes,omitempty"`
	Roles  []string
	jwt.StandardClaims
}

const (
	// WebUsage used for the uid
	WebUsage = "web"
	// StorageUsage for file retrieval
	StorageUsage = "storage"
	// APIUSage for the device api
	APIUSage = "api"
)
