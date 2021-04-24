package model

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/argon2"
	"gopkg.in/yaml.v3"
)

const (
	//TODO: make it configurable
	argon2configTime    = 5
	argon2configMemory  = 3 * 1024
	argon2configThreads = 4
	argon2configKeylen  = 32
)

// User holds the user profile
type User struct {
	ID            string
	Email         string
	EmailVerified bool
	Password      string
	Name          string
	Nickname      string
	GivenName     string
	FamilyName    string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	IsAdmin       bool
}

// GenPassword generates a new random password
func GenPassword() (string, error) {
	b := make([]byte, 10)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base32.HexEncoding.EncodeToString(b), nil
}

func hashPassword(raw string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(raw),
		salt,
		argon2configTime,
		argon2configMemory,
		argon2configThreads,
		argon2configKeylen,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	format := "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s"
	full := fmt.Sprintf(format, argon2.Version, argon2configMemory, argon2configTime, argon2configThreads, b64Salt, b64Hash)

	return full, nil
}

func sanitizeEmail(email string) string {
	rg, err := regexp.Compile("[^a-zA-Z0-9.@-_]+")
	if err != nil {
		log.Fatal(err)
	}
	//remove all not whitelisted
	return rg.ReplaceAllString(email, "")
}

// NewUser create a new user object
func NewUser(email string, rawPassword string) (*User, error) {
	// id, err := genId()
	// if err != nil {
	// 	return nil, err
	// }

	password, err := hashPassword(rawPassword)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:            sanitizeEmail(email),
		Email:         email,
		EmailVerified: true,
		Password:      password,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}

// GenID newid
func (u *User) GenID() (err error) {
	return errors.New("not implemented")
}

// SetPassword sets the user password (and hashes it)
func (u *User) SetPassword(raw string) (err error) {
	u.Password, err = hashPassword(raw)
	return
}

// CheckPassword checks the password
func (u *User) CheckPassword(raw string) (bool, error) {
	parts := strings.Split(u.Password, "$")
	if len(parts) < 3 {
		return false, errors.New("invalid password format")
	}

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

// Serialize gets a representation
func (u User) Serialize() ([]byte, error) {
	return yaml.Marshal(u)
}

// DeserializeUser deserializes
func DeserializeUser(b []byte) (*User, error) {
	usr := &User{}
	if err := yaml.Unmarshal(b, usr); err != nil {
		return nil, err
	}
	return usr, nil
}
