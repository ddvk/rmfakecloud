package model

import (
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

func TestWebAuthnCredentialRoundTrip(t *testing.T) {
	cred := &webauthn.Credential{
		ID:              []byte{1, 2, 3, 4},
		PublicKey:       []byte{9, 8, 7},
		AttestationType: "none",
		Flags: webauthn.CredentialFlags{
			UserPresent:  true,
			UserVerified: true,
		},
		Authenticator: webauthn.Authenticator{
			AAGUID:    []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			SignCount: 3,
		},
	}
	stored := FromLibraryCredential(cred, "Laptop")
	if stored.Name != "Laptop" || stored.SignCount != 3 {
		t.Fatalf("unexpected stored credential: %+v", stored)
	}
	if stored.CreatedAt.IsZero() {
		t.Fatal("CreatedAt should be set")
	}
	lib := stored.ToLibrary()
	if string(lib.ID) != string(cred.ID) || string(lib.PublicKey) != string(cred.PublicKey) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", lib, cred)
	}
}

func TestUpdateAndRemoveWebAuthnCredential(t *testing.T) {
	u := &User{
		ID: "alice@example.com",
		WebAuthnCredentials: []WebAuthnCredential{
			{ID: []byte("abc"), SignCount: 1, Name: "a", CreatedAt: time.Now()},
		},
	}
	updated := u.UpdateWebAuthnCredentialSignCount([]byte("abc"), &webauthn.Credential{
		Authenticator: webauthn.Authenticator{SignCount: 5},
		Flags:         webauthn.CredentialFlags{BackupState: true},
	})
	if !updated || u.WebAuthnCredentials[0].SignCount != 5 {
		t.Fatalf("sign count not updated: %+v", u.WebAuthnCredentials[0])
	}
	if !u.RemoveWebAuthnCredential([]byte("abc")) || len(u.WebAuthnCredentials) != 0 {
		t.Fatal("remove failed")
	}
}

func TestFindUserByWebAuthnHandle(t *testing.T) {
	users := []*User{
		{ID: "a@example.com"},
		{ID: "b@example.com"},
	}
	found := FindUserByWebAuthnHandle(users, []byte("b@example.com"))
	if found == nil || found.ID != "b@example.com" {
		t.Fatalf("expected b, got %#v", found)
	}
	if FindUserByWebAuthnHandle(users, []byte("missing")) != nil {
		t.Fatal("expected nil")
	}
}

func TestCredentialIDBase64(t *testing.T) {
	id := []byte{0xff, 0x00, 0x01}
	s := CredentialIDBase64(id)
	got, err := ParseCredentialIDBase64(s)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(id) {
		t.Fatalf("got %v want %v", got, id)
	}
}

func TestWebAuthnUserInterface(t *testing.T) {
	u := &User{ID: "u1", Email: "u1@example.com", Name: "User One"}
	wu := WebAuthnUser{User: u}
	if string(wu.WebAuthnID()) != "u1" {
		t.Fatal(wu.WebAuthnID())
	}
	if wu.WebAuthnName() != "u1@example.com" || wu.WebAuthnDisplayName() != "User One" {
		t.Fatalf("%s / %s", wu.WebAuthnName(), wu.WebAuthnDisplayName())
	}
	if len(wu.WebAuthnCredentials()) != 0 {
		t.Fatal("expected empty credentials")
	}
}
