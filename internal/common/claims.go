package common

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type StorageClaim struct {
	DocumentId string `json:"documentId"`
	UserId     string `json:"userId"`
	jwt.StandardClaims
}
type DeviceClaims struct {
	UserID     string `json:"auth0-userid"`
	DeviceDesc string `json:"device-desc"`
	DeviceId   string `json:"device-id"`
	Scopes     string `json:"scopes,omitempty"`
	jwt.StandardClaims
}

// UserClaims is the oauth token struct.
type UserClaims struct {
	Profile    Auth0profile `json:"auth0-profile,omitempty"`
	DeviceDesc string       `json:"device-desc"`
	DeviceId   string       `json:"device-id"`
	Scopes     string       `json:"scopes,omitempty"`
	jwt.StandardClaims
}

// Auth0profile is the oauth user struct.
type Auth0profile struct {
	UserId        string `json:"UserID"`
	IsSocial      bool
	ClientId      string `json:"ClientID"`
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

type WebUserClaims struct {
	UserId string `json:"UserID"`
	Email  string
	Scopes string `json:"scopes,omitempty"`
	Roles  []string
	jwt.StandardClaims
}

const WebUsage = "web"
const StorageUsage = "storage"
const ApiUsage = "api"
