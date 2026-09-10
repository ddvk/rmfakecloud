package ui

import (
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

func TestWebAuthnSessionStorePutTake(t *testing.T) {
	s := newWebAuthnSessionStore()
	id := s.Put(&webauthn.SessionData{Challenge: "abc"})
	if id == "" {
		t.Fatal("empty session id")
	}
	data, ok := s.Take(id)
	if !ok || data.Challenge != "abc" {
		t.Fatalf("take failed: ok=%v data=%+v", ok, data)
	}
	if _, ok := s.Take(id); ok {
		t.Fatal("session should be single-use")
	}
}

func TestWebAuthnSessionStoreExpiry(t *testing.T) {
	s := newWebAuthnSessionStore()
	id := "expired"
	s.mu.Lock()
	s.data[id] = webAuthnSessionEntry{
		Data:    webauthn.SessionData{Challenge: "x"},
		Expires: time.Now().Add(-time.Second),
	}
	s.mu.Unlock()
	if _, ok := s.Take(id); ok {
		t.Fatal("expired session should not be returned")
	}
	if s.LenForTest() != 0 {
		t.Fatalf("expected purged store, len=%d", s.LenForTest())
	}
}
