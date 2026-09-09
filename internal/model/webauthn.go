package model

import (
	"bytes"
	"encoding/base64"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// WebAuthnCredential is a stored passkey / WebAuthn public-key credential.
type WebAuthnCredential struct {
	ID              []byte    `yaml:"id"`
	PublicKey       []byte    `yaml:"publickey"`
	AttestationType string    `yaml:"attestationtype,omitempty"`
	Transport       []string  `yaml:"transport,omitempty"`
	UserPresent     bool      `yaml:"userpresent,omitempty"`
	UserVerified    bool      `yaml:"userverified,omitempty"`
	BackupEligible  bool      `yaml:"backupeligible,omitempty"`
	BackupState     bool      `yaml:"backupstate,omitempty"`
	AAGUID          []byte    `yaml:"aaguid,omitempty"`
	SignCount       uint32    `yaml:"signcount"`
	Attachment      string    `yaml:"attachment,omitempty"`
	Name            string    `yaml:"name,omitempty"`
	CreatedAt       time.Time `yaml:"createdat,omitempty"`
}

// WebAuthnUser adapts *User to webauthn.User.
type WebAuthnUser struct {
	*User
}

func (u WebAuthnUser) WebAuthnID() []byte {
	if u.User == nil {
		return nil
	}
	return []byte(u.ID)
}

func (u WebAuthnUser) WebAuthnName() string {
	if u.User == nil {
		return ""
	}
	return u.Email
}

func (u WebAuthnUser) WebAuthnDisplayName() string {
	if u.User == nil {
		return ""
	}
	if u.Name != "" {
		return u.Name
	}
	return u.Email
}

func (u WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	if u.User == nil {
		return nil
	}
	out := make([]webauthn.Credential, 0, len(u.User.WebAuthnCredentials))
	for _, c := range u.User.WebAuthnCredentials {
		out = append(out, c.ToLibrary())
	}
	return out
}

// ToLibrary converts a stored credential to the go-webauthn type.
func (c WebAuthnCredential) ToLibrary() webauthn.Credential {
	transports := make([]protocol.AuthenticatorTransport, 0, len(c.Transport))
	for _, t := range c.Transport {
		transports = append(transports, protocol.AuthenticatorTransport(t))
	}
	return webauthn.Credential{
		ID:              c.ID,
		PublicKey:       c.PublicKey,
		AttestationType: c.AttestationType,
		Transport:       transports,
		Flags: webauthn.CredentialFlags{
			UserPresent:    c.UserPresent,
			UserVerified:   c.UserVerified,
			BackupEligible: c.BackupEligible,
			BackupState:    c.BackupState,
		},
		Authenticator: webauthn.Authenticator{
			AAGUID:     c.AAGUID,
			SignCount:  c.SignCount,
			Attachment: protocol.AuthenticatorAttachment(c.Attachment),
		},
	}
}

// FromLibraryCredential builds a stored credential from a successful registration.
func FromLibraryCredential(cred *webauthn.Credential, name string) WebAuthnCredential {
	if cred == nil {
		return WebAuthnCredential{}
	}
	transports := make([]string, 0, len(cred.Transport))
	for _, t := range cred.Transport {
		transports = append(transports, string(t))
	}
	return WebAuthnCredential{
		ID:              append([]byte(nil), cred.ID...),
		PublicKey:       append([]byte(nil), cred.PublicKey...),
		AttestationType: cred.AttestationType,
		Transport:       transports,
		UserPresent:     cred.Flags.UserPresent,
		UserVerified:    cred.Flags.UserVerified,
		BackupEligible:  cred.Flags.BackupEligible,
		BackupState:     cred.Flags.BackupState,
		AAGUID:          append([]byte(nil), cred.Authenticator.AAGUID...),
		SignCount:       cred.Authenticator.SignCount,
		Attachment:      string(cred.Authenticator.Attachment),
		Name:            name,
		CreatedAt:       time.Now().UTC(),
	}
}

// UpdateWebAuthnCredentialSignCount updates sign count / backup flags after login.
func (u *User) UpdateWebAuthnCredentialSignCount(credID []byte, cred *webauthn.Credential) bool {
	if u == nil || cred == nil {
		return false
	}
	for i := range u.WebAuthnCredentials {
		if bytes.Equal(u.WebAuthnCredentials[i].ID, credID) {
			u.WebAuthnCredentials[i].SignCount = cred.Authenticator.SignCount
			u.WebAuthnCredentials[i].BackupState = cred.Flags.BackupState
			u.WebAuthnCredentials[i].UserPresent = cred.Flags.UserPresent
			u.WebAuthnCredentials[i].UserVerified = cred.Flags.UserVerified
			return true
		}
	}
	return false
}

// RemoveWebAuthnCredential deletes a credential by raw id. Returns true if removed.
func (u *User) RemoveWebAuthnCredential(credID []byte) bool {
	if u == nil {
		return false
	}
	for i := range u.WebAuthnCredentials {
		if bytes.Equal(u.WebAuthnCredentials[i].ID, credID) {
			u.WebAuthnCredentials = append(u.WebAuthnCredentials[:i], u.WebAuthnCredentials[i+1:]...)
			return true
		}
	}
	return false
}

// FindUserByWebAuthnHandle finds a user whose WebAuthnID matches userHandle.
func FindUserByWebAuthnHandle(users []*User, userHandle []byte) *User {
	for _, u := range users {
		if u != nil && bytes.Equal([]byte(u.ID), userHandle) {
			return u
		}
	}
	return nil
}

// CredentialIDBase64 returns URL-safe base64 without padding for API paths.
func CredentialIDBase64(id []byte) string {
	return base64.RawURLEncoding.EncodeToString(id)
}

// ParseCredentialIDBase64 decodes a URL-safe credential id.
func ParseCredentialIDBase64(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
