package screenshare

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPeekOfferMessagesFindsWebtrcOffer(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("user1", "tablet1")

	offer, _ := json.Marshal(map[string]interface{}{
		"type": "webtrc",
		"payload": map[string]interface{}{
			"type":        "offer",
			"description": "v=0",
		},
	})
	rm.AddDirect(room.RoomID, "tablet1", "browser1", offer)

	msgs := rm.PeekOfferMessages(room.RoomID)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 offer message, got %d", len(msgs))
	}
}

func TestAddBroadcastPreservesDirectOffers(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("user1", "tablet1")

	offer, _ := json.Marshal(map[string]interface{}{
		"type": "webtrc",
		"payload": map[string]interface{}{
			"type":        "offer",
			"description": "v=0",
		},
	})
	rm.AddDirect(room.RoomID, "tablet1", "browser1", offer)
	rm.AddBroadcast(room.RoomID, "browser1", json.RawMessage(`{"type":"request-offer","clientId":"browser1"}`))

	msgs := rm.PeekOfferMessages(room.RoomID)
	if len(msgs) == 0 {
		t.Fatal("request-offer broadcast cleared the pending async offer")
	}
}

func TestWaitForOfferReturnsExisting(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("user1", "tablet1")

	offer, _ := json.Marshal(map[string]interface{}{
		"type": "webtrc",
		"payload": map[string]interface{}{
			"type":        "offer",
			"description": "v=0",
		},
	})
	rm.AddDirect(room.RoomID, "tablet1", "browser1", offer)

	start := time.Now()
	msgs := rm.WaitForOffer(room.RoomID, 5*time.Second)
	if len(msgs) == 0 {
		t.Fatal("expected offer")
	}
	if time.Since(start) > time.Second {
		t.Fatalf("WaitForOffer took too long for existing offer: %s", time.Since(start))
	}
}

func TestWaitForOfferAbortsWhenRoomDeleted(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("user1", "tablet1")

	done := make(chan []Message, 1)
	go func() {
		done <- rm.WaitForOffer(room.RoomID, 30*time.Second)
	}()

	time.Sleep(50 * time.Millisecond)
	rm.DeleteRoom(room.RoomID)

	select {
	case msgs := <-done:
		if msgs != nil {
			t.Fatalf("expected nil after delete, got %d messages", len(msgs))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("WaitForOffer did not abort after room delete")
	}
}

func TestWaitForOfferReceivesLateDirect(t *testing.T) {
	rm := NewRoomManager()
	room := rm.CreateRoom("user1", "tablet1")

	done := make(chan []Message, 1)
	go func() {
		done <- rm.WaitForOffer(room.RoomID, 5*time.Second)
	}()

	time.Sleep(50 * time.Millisecond)
	offer, _ := json.Marshal(map[string]interface{}{
		"type": "webtrc",
		"payload": map[string]interface{}{
			"type":        "offer",
			"description": "v=0",
		},
	})
	rm.AddDirect(room.RoomID, "tablet1", "browser1", offer)

	select {
	case msgs := <-done:
		if len(msgs) == 0 {
			t.Fatal("expected late offer")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("WaitForOffer did not receive late offer")
	}
}
