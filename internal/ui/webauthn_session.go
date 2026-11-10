package ui

import (
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

const webAuthnSessionTTL = 2 * time.Minute

type webAuthnSessionStore struct {
	mu   sync.Mutex
	data map[string]webAuthnSessionEntry
}

type webAuthnSessionEntry struct {
	Data    webauthn.SessionData
	Expires time.Time
}

func newWebAuthnSessionStore() *webAuthnSessionStore {
	return &webAuthnSessionStore{data: make(map[string]webAuthnSessionEntry)}
}

func (s *webAuthnSessionStore) Put(session *webauthn.SessionData) string {
	id := uuid.NewString()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeLocked(time.Now())
	s.data[id] = webAuthnSessionEntry{
		Data:    *session,
		Expires: time.Now().Add(webAuthnSessionTTL),
	}
	return id
}

func (s *webAuthnSessionStore) Take(id string) (webauthn.SessionData, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.purgeLocked(now)
	entry, ok := s.data[id]
	if !ok {
		return webauthn.SessionData{}, false
	}
	delete(s.data, id)
	if now.After(entry.Expires) {
		return webauthn.SessionData{}, false
	}
	return entry.Data, true
}

func (s *webAuthnSessionStore) purgeLocked(now time.Time) {
	for k, v := range s.data {
		if now.After(v.Expires) {
			delete(s.data, k)
		}
	}
}

// LenForTest returns current session count (tests only).
func (s *webAuthnSessionStore) LenForTest() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeLocked(time.Now())
	return len(s.data)
}
