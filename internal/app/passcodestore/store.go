package passcodestore

import (
	"errors"
	"sync"
	"time"

	"github.com/ddvk/rmfakecloud/internal/messages"
)

// ResetTTL is how long a pending reset request stays valid.
const ResetTTL = 24 * time.Hour

// ErrNotFound is returned when a reset request does not exist or has expired.
var ErrNotFound = errors.New("reset request not found")

// OwnedReset pairs a reset request with the user that owns it.
type OwnedReset struct {
	UID string
	messages.PasscodeReset
}

// Store persists pending passcode reset requests.
type Store interface {
	Create(uid string, req messages.PasscodeReset) error
	Get(requestID string) (OwnedReset, error)
	ListForUser(uid string) []messages.PasscodeReset
	Approve(uid, requestID string) (messages.PasscodeReset, error)
	Delete(uid, requestID string) error
}

// InMemory is an in-process Store. Entries are dropped lazily on read once expired.
type InMemory struct {
	mu    sync.RWMutex
	items map[string]OwnedReset
}

// NewInMemory returns an empty in-memory store.
func NewInMemory() *InMemory {
	return &InMemory{items: make(map[string]OwnedReset)}
}

// Create inserts (or replaces) a reset request for uid keyed by req.RequestID.
func (s *InMemory) Create(uid string, req messages.PasscodeReset) error {
	if req.RequestID == "" {
		return errors.New("missing RequestID")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[req.RequestID] = OwnedReset{UID: uid, PasscodeReset: req}
	return nil
}

// Get returns the reset request for requestID, dropping it if it has expired.
func (s *InMemory) Get(requestID string) (OwnedReset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[requestID]
	if !ok {
		return OwnedReset{}, ErrNotFound
	}
	if time.Now().After(item.Expires) {
		delete(s.items, requestID)
		return OwnedReset{}, ErrNotFound
	}
	return item, nil
}

// ListForUser returns all non-expired pending (unapproved) requests for uid.
func (s *InMemory) ListForUser(uid string) []messages.PasscodeReset {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	out := make([]messages.PasscodeReset, 0)
	for id, item := range s.items {
		if now.After(item.Expires) {
			delete(s.items, id)
			continue
		}
		if item.UID != uid || item.Approved {
			continue
		}
		out = append(out, item.PasscodeReset)
	}
	return out
}

// Delete removes a pending request if it belongs to uid.
func (s *InMemory) Delete(uid, requestID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[requestID]
	if !ok || item.UID != uid {
		return ErrNotFound
	}
	delete(s.items, requestID)
	return nil
}

// Approve flips the request to approved if it belongs to uid and is not expired.
func (s *InMemory) Approve(uid, requestID string) (messages.PasscodeReset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[requestID]
	if !ok {
		return messages.PasscodeReset{}, ErrNotFound
	}
	if time.Now().After(item.Expires) {
		delete(s.items, requestID)
		return messages.PasscodeReset{}, ErrNotFound
	}
	if item.UID != uid {
		return messages.PasscodeReset{}, ErrNotFound
	}
	item.Approved = true
	s.items[requestID] = item
	return item.PasscodeReset, nil
}
