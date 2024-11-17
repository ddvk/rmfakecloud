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

var emailWhiteList *regexp.Regexp

func init() {
	var err error
	emailWhiteList, err = regexp.Compile("[^a-zA-Z0-9.@-_]+")
	if err != nil {
		log.Fatal(err)
	}

}

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
	// IsAdmin indicates if the user can managed others users in this instance.
	IsAdmin bool
	// Sync15 if the user should use this sync type (which uses a lot less bandwidth).
	Sync15 bool
	// AdditionalScopes is a list of scopes to add to the user session.
	AdditionalScopes []string
	// Integrations stores the list of "Integrations" as shown on the tablet.
	Integrations []IntegrationConfig
}

// IntegrationConfig config for various integrations
type IntegrationConfig struct {
	ID       string
	Provider string
	Name     string

	// WebDav // FTP
	Username string
	Password string
	Address  string

	// FTP
	ActiveTransfers bool

	// Insecure ignore TLS cert errors
	Insecure bool

	// Dropbox
	Accesstoken string

	// Localfs
	//TODO: experimental, security blah blah
	Path string
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
	//remove all not whitelisted
	return emailWhiteList.ReplaceAllString(email, "")
}

// NewUser create a new user object
func NewUser(userID string, rawPassword string) (*User, error) {
	password, err := hashPassword(rawPassword)
	if err != nil {
		return nil, err
	}

	sanitizedID := sanitizeEmail(userID)
	return &User{
		ID:            sanitizedID,
		Email:         sanitizedID,
		EmailVerified: true,
		Password:      password,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Sync15:        true,
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
