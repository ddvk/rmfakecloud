package common

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// DeviceClaims is the oauth token struct.
type DeviceClaims struct {
	Profile    Auth0profile `json:"auth0-profile,omitempty"`
	DeviceDesc string       `json:"device-desc"`
	DeviceId   string       `json:"device-id"`
	Scopes     string       `json:"scopes,omitempty"`
	jwt.StandardClaims
}

// Auth0profile is the oauth user struct.
type Auth0profile struct {
	UserId        string `json:"UserID'`
	IsSocial      bool
	ClientId      string `json:"ClientID'`
	Connection    string
	Name          string
	Nickname      string
	GivenName     string
	FamilyName    string
	Email         string
	EmailVerified bool
	Picture       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
