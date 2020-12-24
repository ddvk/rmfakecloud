package messages

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/argon2"
)

const (
	Argon2Config_time    = 1
	Argon2Config_memory  = 64 * 1024
	Argon2Config_threads = 4
	Argon2Config_keyLen  = 32
)

type User struct {
	Id            string `json:"userid"`
	Email         string `json:"email"`
	EmailVerified bool
	Password      string     `json:"password"`
	CurrentCode   *string    `json:"code"`
	CodeExpire    *time.Time `json:"code_exp"`
	Name          string
	Nickname      string
	GivenName     string
	FamilyName    string
	Picture       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func genId() (string, error) {
	b := make([]byte, 45)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base32.StdEncoding.EncodeToString(b), nil
}

func genPassword(raw string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(raw),
		salt,
		Argon2Config_time,
		Argon2Config_memory,
		Argon2Config_threads,
		Argon2Config_keyLen,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	format := "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s"
	full := fmt.Sprintf(format, argon2.Version, Argon2Config_memory, Argon2Config_time, Argon2Config_threads, b64Salt, b64Hash)

	return full, nil
}

func NewUser(email string, rawPassword string) (*User, error) {
	id, err := genId()
	if err != nil {
		return nil, err
	}

	password, err := genPassword(rawPassword)
	if err != nil {
		return nil, err
	}

	return &User{
		Id:            id,
		Email:         email,
		EmailVerified: true,
		Password:      password,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}

func (u *User) GenId() (err error) {
	u.Id, err = genId()
	return
}

func (u *User) SetPassword(raw string) (err error) {
	u.Password, err = genPassword(raw)
	return
}

func (u *User) CheckPassword(raw string) (bool, error) {
	parts := strings.Split(u.Password, "$")

	var (
		memory  uint32
		time    uint32
		threads uint8
	)

	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	var keyLen = uint32(len(decodedHash))

	comparisonHash := argon2.IDKey([]byte(raw), salt, time, memory, threads, keyLen)

	return (subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1), nil
}

func (u *User) NewUserCode() (code string, err error) {
	b := make([]byte, 5)

	if _, err = rand.Read(b); err != nil {
		return
	}

	code = base32.StdEncoding.EncodeToString(b)

	u.Code = code
	u.CodeExpire = time.Now().Add(10 * time.Minute)

	return
}

type Auth0token struct {
	Profile    *Auth0profile `json:"auth0-profile,omitempty"`
	DeviceDesc string        `json:"device-desc"`
	DeviceId   string        `json:"device-id"`
	Scopes     string        `json:"scopes,omitempty"`
	jwt.StandardClaims
}
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

func (u *User) NewAuth0Token(deviceDesc, deviceId string) *jwt.Token {
	expirationTime := time.Now().Add(30 * 24 * time.Hour)
	claims := &Auth0token{
		Profile: &Auth0profile{
			UserId:        "auth0|" + u.Id,
			IsSocial:      false,
			Name:          u.Name,
			Nickname:      u.Nickname,
			Email:         u.Email,
			EmailVerified: u.EmailVerified,
			Picture:       u.Picture,
			CreatedAt:     u.CreatedAt,
			UpdatedAt:     u.UpdatedAt,
		},
		DeviceDesc: deviceDesc,
		DeviceId:   deviceId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Subject:   "rM User Token",
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
}
