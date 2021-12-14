package app

import (
	"time"

	"github.com/golang-jwt/jwt"
)

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
	Version    int          `json:"version"`
	Level      string       `json:"level"`
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

const (
	// APIUsage for the device api
	APIUsage = "api"
)
