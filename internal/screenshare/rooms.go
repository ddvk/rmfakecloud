package screenshare

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

type Room struct {
	RoomID       string
	CreatedAt    time.Time
	lastActivity time.Time
	participants map[string]*RoomClient
	ownerUserID  string
	messages     []Message
	notify       chan struct{}
}

type RoomClient struct {
	ClientID string `json:"clientId"`
	UserID   string `json:"userId"`
	IsOwner  bool   `json:"isOwner"`
}

type Message struct {
	Type           string          `json:"type"`
	SenderClientID string          `json:"clientId,omitempty"`
	TargetClientID string          `json:"targetClientId,omitempty"`
	Payload        json.RawMessage `json:"payload"`
}

const roomTimeout = 60 * time.Second

func NewRoomManager() *RoomManager {
	rm := &RoomManager{
		rooms: make(map[string]*Room),
	}
	go rm.expireLoop()
	return rm
}

func (rm *RoomManager) expireLoop() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		rm.mu.Lock()
		now := time.Now().UTC()
		for id, room := range rm.rooms {
			if now.Sub(room.lastActivity) > roomTimeout {
				rm.wakeRoom(room)
				delete(rm.rooms, id)
				log.Infof("Screenshare: expired room=%s (no keepalive for %s)", id, roomTimeout)
			}
		}
		rm.mu.Unlock()
	}
}

func (rm *RoomManager) Keepalive(roomID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if room, exists := rm.rooms[roomID]; exists {
		room.lastActivity = time.Now().UTC()
	}
}

func (rm *RoomManager) CreateRoom(userID, deviceID string) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	roomID := uuid.New().String()
	now := time.Now().UTC()
	room := &Room{
		RoomID:       roomID,
		CreatedAt:    now,
		lastActivity: now,
		participants: make(map[string]*RoomClient),
		ownerUserID:  userID,
		notify:       make(chan struct{}, 1),
	}
	room.participants[deviceID] = &RoomClient{
		ClientID: deviceID,
		UserID:   userID,
		IsOwner:  true,
	}
	rm.rooms[roomID] = room
	log.Infof("Screenshare: created room=%s user=%s device=%s", roomID, userID, deviceID)
	return room
}

func (rm *RoomManager) GetRoom(roomID string) *Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.rooms[roomID]
}

func (rm *RoomManager) GetClients(roomID string) []RoomClient {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return nil
	}
	clients := make([]RoomClient, 0, len(room.participants))
	for _, c := range room.participants {
		clients = append(clients, *c)
	}
	return clients
}

func (rm *RoomManager) AddBroadcast(roomID, senderClientID string, payload json.RawMessage) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return
	}
	// Keep pending directs (offer/ICE) so an async browser join can still consume them.
	// Drop prior broadcasts to avoid stacking stale request-offer / control messages.
	kept := make([]Message, 0, len(room.messages)+1)
	for _, m := range room.messages {
		if m.Type == "direct" {
			kept = append(kept, m)
		}
	}
	room.messages = append(kept, Message{
		Type:           "broadcast",
		SenderClientID: senderClientID,
		Payload:        payload,
	})
	select {
	case room.notify <- struct{}{}:
	default:
	}
	log.Debugf("Screenshare: broadcast in room=%s from=%s", roomID, senderClientID)
}

func (rm *RoomManager) AddDirect(roomID, senderClientID, targetClientID string, payload json.RawMessage) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return
	}
	room.messages = append(room.messages, Message{
		Type:           "direct",
		SenderClientID: senderClientID,
		TargetClientID: targetClientID,
		Payload:        payload,
	})
	select {
	case room.notify <- struct{}{}:
	default:
	}
	log.Debugf("Screenshare: direct in room=%s from=%s to=%s", roomID, senderClientID, targetClientID)
}

func messageHasOffer(m Message) bool {
	var p map[string]interface{}
	if json.Unmarshal(m.Payload, &p) != nil {
		return false
	}
	if t, _ := p["type"].(string); t == "offer" {
		return true
	}
	if t, _ := p["type"].(string); t == "webtrc" {
		if inner, ok := p["payload"].(map[string]interface{}); ok {
			if it, _ := inner["type"].(string); it == "offer" {
				return true
			}
		}
	}
	return false
}

// PeekOfferMessages returns the offer and subsequent messages (e.g. ICE) if present.
func (rm *RoomManager) PeekOfferMessages(roomID string) []Message {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return nil
	}
	start := -1
	for i, m := range room.messages {
		if messageHasOffer(m) {
			start = i
			break
		}
	}
	if start < 0 {
		return nil
	}
	msgs := make([]Message, len(room.messages)-start)
	copy(msgs, room.messages[start:])
	return msgs
}

func (rm *RoomManager) WaitForMessages(roomID string, after int, timeout time.Duration) []Message {
	deadline := time.After(timeout)
	for {
		if !rm.RoomExists(roomID) {
			return nil
		}
		msgs := rm.GetMessages(roomID, after)
		if len(msgs) > 0 {
			// Wait briefly for additional messages (ICE candidates follow the offer)
			time.Sleep(200 * time.Millisecond)
			return rm.GetMessages(roomID, after)
		}
		rm.mu.RLock()
		room, exists := rm.rooms[roomID]
		var notify <-chan struct{}
		if exists {
			notify = room.notify
		}
		rm.mu.RUnlock()
		if !exists {
			return nil
		}
		select {
		case <-notify:
		case <-deadline:
			return nil
		case <-time.After(500 * time.Millisecond):
		}
	}
}

// WaitForOffer blocks until an offer is present, the room is gone, or timeout.
func (rm *RoomManager) WaitForOffer(roomID string, timeout time.Duration) []Message {
	deadline := time.After(timeout)
	for {
		if !rm.RoomExists(roomID) {
			return nil
		}
		if msgs := rm.PeekOfferMessages(roomID); len(msgs) > 0 {
			time.Sleep(200 * time.Millisecond)
			return rm.PeekOfferMessages(roomID)
		}
		rm.mu.RLock()
		room, exists := rm.rooms[roomID]
		var notify <-chan struct{}
		if exists {
			notify = room.notify
		}
		rm.mu.RUnlock()
		if !exists {
			return nil
		}
		select {
		case <-notify:
		case <-deadline:
			return nil
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func (rm *RoomManager) wakeRoom(room *Room) {
	if room == nil {
		return
	}
	select {
	case room.notify <- struct{}{}:
	default:
	}
}

func (rm *RoomManager) GetMessages(roomID string, after int) []Message {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return nil
	}
	if after >= len(room.messages) {
		return nil
	}
	msgs := make([]Message, len(room.messages)-after)
	copy(msgs, room.messages[after:])
	return msgs
}

func (rm *RoomManager) AddParticipant(roomID, clientID, userID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return
	}
	room.participants[clientID] = &RoomClient{
		ClientID: clientID,
		UserID:   userID,
		IsOwner:  false,
	}
	log.Debugf("Screenshare: added participant to room=%s client=%s user=%s total=%d",
		roomID, clientID, userID, len(room.participants))
}

func (rm *RoomManager) RemoveParticipant(clientID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for roomID, room := range rm.rooms {
		if _, exists := room.participants[clientID]; exists {
			delete(room.participants, clientID)
			log.Debugf("Screenshare: removed participant from room=%s client=%s remaining=%d",
				roomID, clientID, len(room.participants))

			if len(room.participants) == 0 {
				delete(rm.rooms, roomID)
				log.Infof("Screenshare: removed empty room=%s", roomID)
			}
			return
		}
	}
}

func (rm *RoomManager) DeleteAllForUser(userID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	for id, room := range rm.rooms {
		if room.ownerUserID == userID {
			rm.wakeRoom(room)
			delete(rm.rooms, id)
			log.Infof("Screenshare: deleted room=%s for user=%s", id, userID)
		}
	}
}

func (rm *RoomManager) DeleteRoom(roomID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if room, exists := rm.rooms[roomID]; exists {
		rm.wakeRoom(room)
		delete(rm.rooms, roomID)
		log.Infof("Screenshare: deleted room=%s", roomID)
	}
}

func (rm *RoomManager) FindActiveRoom(userID string) string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var newest *Room
	for _, room := range rm.rooms {
		if room.ownerUserID == userID {
			if newest == nil || room.CreatedAt.After(newest.CreatedAt) {
				newest = room
			}
		}
	}
	if newest != nil {
		return newest.RoomID
	}
	return ""
}

func (rm *RoomManager) RoomExists(roomID string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	_, exists := rm.rooms[roomID]
	return exists
}
