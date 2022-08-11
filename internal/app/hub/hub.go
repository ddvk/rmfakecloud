package hub

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/zgs225/rmfakecloud/internal/messages"
	"github.com/gorilla/websocket"
)

const (
	//DocAddedEvent addded
	DocAddedEvent = "DocAdded"
	//DocDeletedEvent deleted
	DocDeletedEvent = "DocDeleted"
	//SyncCompleted somplete
	SyncCompleted = "SyncComplete"
)

type ntf struct {
	msg  *messages.WsMessage
	uid  string
	from string
}

// Hub ws notificaiton hub
type Hub struct {
	allClients    map[*wsClient]bool
	userClients   map[string]map[*wsClient]bool
	additions     chan *wsClient
	removals      chan *wsClient
	notifications chan ntf
}

// NotifySync sends a message to all connected clients 1.5
func (h *Hub) NotifySync(uid, deviceID string) string {
	log.Info("notify sync from: ", deviceID)
	timestamp := time.Now().UnixNano()
	msgid := strconv.Itoa(int(timestamp))
	msg := messages.WsMessage{
		Message: messages.NotificationMessage{
			MessageID3: msgid,
			Attributes: messages.Attributes{
				Auth0UserID:    uid,
				Event:          SyncCompleted,
				SourceDeviceID: deviceID,
			},
		},
	}

	h.notifications <- ntf{
		uid:  uid,
		from: deviceID,
		msg:  &msg,
	}
	return msgid
}

// DocumentNotification notification of something
type DocumentNotification struct {
	ID      string
	Type    string
	Version int
	Parent  string
	Name    string
}

// Notify sends a message to all connected clients
func (h *Hub) Notify(uid, deviceID string, doc DocumentNotification, eventType string) {
	timeStamp := time.Now().UTC().Format(time.RFC3339Nano)
	messageID := uuid.New().String()

	msg := messages.WsMessage{
		Message: messages.NotificationMessage{
			MessageID:  messageID,
			MessageID2: messageID,
			MessageID3: messageID,
			Attributes: messages.Attributes{
				Auth0UserID:      uid,
				Event:            eventType,
				SourceDeviceDesc: "some-client",
				SourceDeviceID:   deviceID,
			},
			PublishTime:  timeStamp,
			PublishTime2: timeStamp,
		},
		Subscription: "dummy-subscription",
	}
	msg.Message.Attributes.ID = doc.ID
	msg.Message.Attributes.Type = doc.Type
	msg.Message.Attributes.Version = strconv.Itoa(doc.Version)
	msg.Message.Attributes.VissibleName = doc.Name
	msg.Message.Attributes.Parent = doc.Parent

	h.notifications <- ntf{
		uid:  uid,
		from: deviceID,
		msg:  &msg,
	}
}
func (h *Hub) send(n ntf) {
	uid := n.uid
	msg := n.msg
	log.Info("Broadcast notification, for all devices of  uid:", uid, " id ", n.msg.Message.MessageID3)

	if clients, ok := h.userClients[uid]; ok {
		for c := range clients {
			if c.deviceID == n.from {
				continue
				// log.Warn("sending to same device: ", c.deviceID)
			}
			select {
			case c.notifications <- msg:
			case <-c.done:
				return
			default:
				log.Warn("dropping notification")
			}
		}
	}

}

// ClientCount number of connected clients
func (h *Hub) ClientCount() int {
	return len(h.allClients)
}

// NewHub construct a hub
func NewHub() *Hub {
	h := Hub{
		allClients:  make(map[*wsClient]bool),
		userClients: make(map[string]map[*wsClient]bool),

		additions:     make(chan *wsClient),
		removals:      make(chan *wsClient),
		notifications: make(chan ntf, 5),
	}
	go h.start()
	return &h
}

func (h *Hub) removeClient(c *wsClient) {
	if _, ok := h.allClients[c]; ok {
		delete(h.allClients, c)
		close(c.notifications)
	}
	if userclients, ok := h.userClients[c.uid]; ok {
		delete(userclients, c)
		if len(userclients) == 0 {
			delete(h.userClients, c.uid)
		}
	}
}
func (h *Hub) start() {
	for {
		select {
		case c := <-h.additions:
			log.Debugln("hub: adding a client")
			h.allClients[c] = true
			clients, ok := h.userClients[c.uid]
			if !ok {
				clients = make(map[*wsClient]bool)
				h.userClients[c.uid] = clients
			}
			clients[c] = true
		case c := <-h.removals:
			log.Info("hub: removing client")
			h.removeClient(c)
		case c := <-h.notifications:
			log.Info("hub: dispatching notification")
			h.send(c)
		}
	}
}

type wsClient struct {
	uid           string
	deviceID      string
	notifications chan *messages.WsMessage
	done          chan struct{}
	hub           *Hub
}

func (c *wsClient) readMessages(done chan<- struct{}, ws *websocket.Conn) {
	defer ws.Close()

	for {
		_, p, err := ws.ReadMessage()

		if err != nil {
			if !websocket.IsCloseError(err, 1000) {
				log.Warn("Can't read from ws ", err)
			}
			done <- struct{}{}
			return
		}

		log.Debugln("Message: ", string(p))
	}
}
func (c *wsClient) writeMessages(done chan<- struct{}, ws *websocket.Conn) {
	defer ws.Close()

outer:
	for {
		select {
		case m, ok := <-c.notifications:
			if !ok {
				break outer
			}
			log.Debugln("sending notification to:", c.deviceID)
			ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := ws.WriteJSON(m)
			if err != nil {
				log.Warn("Cant write to ws ", err)
				break outer
			}
			log.Debugln("notification sent: ", c.deviceID)
		case <-c.done:
			break outer
		}
	}
	done <- struct{}{}
}

// ConnectWs upgrade the connection to websocket
func (h *Hub) ConnectWs(uid, deviceID string, connection *websocket.Conn) {
	client := &wsClient{
		uid:           uid,
		deviceID:      deviceID,
		hub:           h,
		notifications: make(chan *messages.WsMessage),
		done:          make(chan struct{}),
	}
	h.additions <- client

	done := make(chan struct{}, 2)
	go client.readMessages(done, connection)
	go client.writeMessages(done, connection)
	<-done
	close(client.done)

	h.removals <- client
}
