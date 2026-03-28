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
	// RegisteredDevices are reMarkable clients that completed pairing (see newDevice).
	RegisteredDevices []RegisteredDevice `yaml:"registereddevices,omitempty"`
}

// RegisteredDevice is a tablet client that obtained a device token.
type RegisteredDevice struct {
	DeviceID     string    `yaml:"deviceid,omitempty"`
	DeviceDesc   string    `yaml:"devicedesc,omitempty"`
	RegisteredAt time.Time `yaml:"registeredat,omitempty"`
	LastSeen     time.Time `yaml:"lastseen,omitempty"`
}

// UpsertRegisteredDevice records or updates a paired device for this user.
func (u *User) UpsertRegisteredDevice(deviceID, desc string) {
	now := time.Now()
	for i := range u.RegisteredDevices {
		if u.RegisteredDevices[i].DeviceID == deviceID {
			u.RegisteredDevices[i].DeviceDesc = desc
			u.RegisteredDevices[i].LastSeen = now
			u.UpdatedAt = now
			return
		}
	}
	u.RegisteredDevices = append(u.RegisteredDevices, RegisteredDevice{
		DeviceID:     deviceID,
		DeviceDesc:   desc,
		RegisteredAt: now,
		LastSeen:     now,
	})
	u.UpdatedAt = now
}

// GetRegisteredDevice returns a stored device entry if present.
func (u *User) GetRegisteredDevice(deviceID string) (RegisteredDevice, bool) {
	for _, d := range u.RegisteredDevices {
		if d.DeviceID == deviceID {
			return d, true
		}
	}
	return RegisteredDevice{}, false
}

// RemoveRegisteredDevice drops a device from the registry (e.g. tablet logout).
func (u *User) RemoveRegisteredDevice(deviceID string) {
	if deviceID == "" {
		return
	}
	j := 0
	for _, d := range u.RegisteredDevices {
		if d.DeviceID != deviceID {
			u.RegisteredDevices[j] = d
			j++
		}
	}
	u.RegisteredDevices = u.RegisteredDevices[:j]
	u.UpdatedAt = time.Now()
}

// IntegrationConfig config for various integrations
type IntegrationConfig struct {
	ID       string
	Provider string
	Name     string

	// WebDav // FTP
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Address  string `yaml:"address,omitempty"`

	// FTP
	ActiveTransfers bool `yaml:"activetransfers,omitempty"`

	// Insecure ignore TLS cert errors
	Insecure bool `yaml:"insecure,omitempty"`

	// Dropbox
	Accesstoken string `yaml:"accesstoken,omitempty"`

	// Localfs
	// really dangerous as it allows path traversal
	Path string `yaml:"path,omitempty"`

	// Webhook
	Endpoint string `yaml:"endpoint,omitempty"`
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
