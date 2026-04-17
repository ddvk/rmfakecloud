package passcodestore

import (
	"testing"
	"time"

	"github.com/ddvk/rmfakecloud/internal/messages"
)

func newReq(id string, ttl time.Duration) messages.PasscodeReset {
	now := time.Now()
	return messages.PasscodeReset{
		DeviceID:   "dev-1",
		DeviceName: "reMarkable 2",
		RequestID:  id,
		Created:    now,
		Expires:    now.Add(ttl),
	}
}

func TestCreateGetApprove(t *testing.T) {
	s := NewInMemory()
	req := newReq("abc", time.Hour)
	if err := s.Create("user1", req); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := s.Get("abc")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.UID != "user1" || got.Approved {
		t.Fatalf("unexpected: %+v", got)
	}

	if _, err := s.Approve("user1", "abc"); err != nil {
		t.Fatalf("approve: %v", err)
	}
	got, err = s.Get("abc")
	if err != nil {
		t.Fatalf("get after approve: %v", err)
	}
	if !got.Approved {
		t.Fatalf("expected approved")
	}
}

func TestApproveOtherUserFails(t *testing.T) {
	s := NewInMemory()
	_ = s.Create("user1", newReq("abc", time.Hour))
	if _, err := s.Approve("user2", "abc"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestExpiredEntryDropped(t *testing.T) {
	s := NewInMemory()
	_ = s.Create("user1", newReq("abc", -time.Second))
	if _, err := s.Get("abc"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListForUserOmitsApprovedAndOther(t *testing.T) {
	s := NewInMemory()
	_ = s.Create("user1", newReq("pending", time.Hour))
	approved := newReq("done", time.Hour)
	_ = s.Create("user1", approved)
	_, _ = s.Approve("user1", "done")
	_ = s.Create("user2", newReq("other", time.Hour))

	list := s.ListForUser("user1")
	if len(list) != 1 || list[0].RequestID != "pending" {
		t.Fatalf("unexpected list: %+v", list)
	}
}
