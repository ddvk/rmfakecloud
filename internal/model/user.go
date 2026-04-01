package model

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
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
var yearRegex *regexp.Regexp
var serialLikeRegex *regexp.Regexp

func init() {
	var err error
	emailWhiteList, err = regexp.Compile("[^a-zA-Z0-9.@-_]+")
	if err != nil {
		log.Fatal(err)
	}
	yearRegex, err = regexp.Compile(`\b(19|20)\d{2}\b`)
	if err != nil {
		log.Fatal(err)
	}
	serialLikeRegex, err = regexp.Compile(`\bRM[0-9A-Z]{3,}\b`)
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
	DeviceLink   string    `yaml:"devicelink,omitempty"`
	Make         string    `yaml:"make,omitempty"`
	Model        string    `yaml:"model,omitempty"`
	Year         string    `yaml:"year,omitempty"`
	RegisteredAt time.Time `yaml:"registeredat,omitempty"`
	LastSeen     time.Time `yaml:"lastseen,omitempty"`
}

// UpsertRegisteredDevice records or updates a paired device for this user.
func (u *User) UpsertRegisteredDevice(deviceID, desc, link string) {
	make, model, year := inferDeviceInfo(deviceID, desc, link)
	now := time.Now()
	for i := range u.RegisteredDevices {
		if u.RegisteredDevices[i].DeviceID == deviceID {
			u.RegisteredDevices[i].DeviceDesc = desc
			if link != "" {
				u.RegisteredDevices[i].DeviceLink = link
			}
			if make != "" {
				u.RegisteredDevices[i].Make = make
			}
			if model != "" {
				u.RegisteredDevices[i].Model = model
			}
			if year != "" {
				u.RegisteredDevices[i].Year = year
			}
			u.RegisteredDevices[i].LastSeen = now
			u.UpdatedAt = now
			return
		}
	}
	u.RegisteredDevices = append(u.RegisteredDevices, RegisteredDevice{
		DeviceID:     deviceID,
		DeviceDesc:   desc,
		DeviceLink:   link,
		Make:         make,
		Model:        model,
		Year:         year,
		RegisteredAt: now,
		LastSeen:     now,
	})
	u.UpdatedAt = now
}

func inferDeviceInfo(deviceID, desc, link string) (string, string, string) {
	make := ""
	model := ""
	year := ""

	ls := strings.ToLower(desc + " " + link)
	if strings.Contains(ls, "remarkable") || strings.Contains(ls, "re-markable") || strings.Contains(ls, "rm2") || strings.Contains(ls, "rm1") {
		make = "reMarkable"
	}
	if strings.Contains(ls, "paper pro") || strings.Contains(ls, "paperpro") {
		model = "Paper Pro"
	} else if strings.Contains(ls, "remarkable 2") || strings.Contains(ls, "rm2") {
		model = "2"
	} else if strings.Contains(ls, "remarkable 1") || strings.Contains(ls, "rm1") {
		model = "1"
	}

	if u, err := url.Parse(link); err == nil {
		q := u.Query()
		if v := strings.TrimSpace(q.Get("make")); v != "" {
			make = v
		}
		if v := strings.TrimSpace(q.Get("manufacturer")); v != "" {
			make = v
		}
		if v := strings.TrimSpace(q.Get("brand")); v != "" {
			make = v
		}
		if v := strings.TrimSpace(q.Get("model")); v != "" {
			model = v
		}
		if v := strings.TrimSpace(q.Get("year")); v != "" {
			year = v
		}
	}
	for _, cand := range serialCandidates(deviceID, desc, link) {
		if mapped, ok := modelFromSerial(cand); ok {
			model = mapped
			if make == "" {
				make = "reMarkable"
			}
			break
		}
	}
	if year == "" {
		if m := yearRegex.FindString(ls); m != "" {
			year = m
		}
	}
	return make, model, year
}

func serialCandidates(deviceID, desc, link string) []string {
	out := make([]string, 0, 8)
	push := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		out = append(out, s)
	}
	push(deviceID)
	if u, err := url.Parse(link); err == nil {
		q := u.Query()
		for _, k := range []string{"serial", "serialNumber", "deviceSerial", "sn"} {
			push(q.Get(k))
		}
	}
	for _, m := range serialLikeRegex.FindAllString(strings.ToUpper(desc+" "+link), -1) {
		push(m)
	}
	return out
}

func modelFromSerial(serial string) (string, bool) {
	s := strings.ToUpper(strings.TrimSpace(serial))
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, " ", "")
	prefix6 := s
	if len(prefix6) > 6 {
		prefix6 = prefix6[:6]
	}
	if len(prefix6) >= 5 {
		key := prefix6[:5]
		switch key {
		case "RM02A":
			return "reMarkable Paper Pro", true
		case "RM03A":
			return "reMarkable Paper Pro Move", true
		case "RM110":
			return "reMarkable 2", true
		case "RM102":
			return "reMarkable 1", true
		case "RM12A":
			return "TBA", true
		}
	}
	return "", false
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
