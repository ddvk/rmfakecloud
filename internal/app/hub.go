package app

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/gorilla/websocket"
)

type Hub struct {
	clients      map[*Client]bool
	additions    chan *Client
	removals     chan *Client
	notification chan *messages.WsMessage
}

func (h *Hub) Send(msg messages.WsMessage) {
	for c := range h.clients {
		c.ntf <- msg
	}
}
func (h *Hub) ClientCount() int {
	return len(h.clients)
}

func NewHub() *Hub {
	h := Hub{
		clients:   make(map[*Client]bool),
		additions: make(chan *Client),
		removals:  make(chan *Client),
	}
	go h.start()
	return &h
}

func (h *Hub) removeClient(c *Client) {
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.ntf)
	}
}
func (h *Hub) start() {
	for {
		select {
		case c := <-h.additions:
			log.Printf("adding a client")
			h.clients[c] = true
		case c := <-h.removals:
			log.Printf("removing a client")
			h.removeClient(c)
		}
	}
}

type Client struct {
	ntf chan messages.WsMessage
	hub *Hub
}

func (c *Client) Read(ws *websocket.Conn) {
	defer ws.Close()
	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			log.Println("readng ", err)
			c.hub.removals <- c
			return
		}
		log.Println("Message: ", string(p))
	}
}
func (c *Client) Write(ws *websocket.Conn) {
	defer ws.Close()
	for {
		select {
		case m, ok := <-c.ntf:
			if !ok {
				log.Println("done")
				return
			}
			log.Println("sending notification")
			ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := ws.WriteJSON(m)
			if err != nil {
				c.hub.removals <- c
				return
			}
		}
	}
}

// upgrade the connection to websocket
func (h *Hub) ConnectWs(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	client := &Client{
		hub: h,
		ntf: make(chan messages.WsMessage),
	}
	go client.Read(c)
	h.additions <- client
	client.Write(c)
	h.removals <- client
}

func newWs(doc *messages.RawDocument, typ string) messages.WsMessage {
	tt := time.Now().UTC().Format(time.RFC3339Nano)
	msg := messages.WsMessage{
		Message: messages.NotificationMessage{
			MessageId:  "1234",
			MessageId2: "1234",
			Attributes: messages.Attributes{
				Auth0UserID:      "auth0|12341234123412",
				Event:            typ,
				Id:               doc.Id,
				Type:             doc.Type,
				Version:          strconv.Itoa(doc.Version),
				VissibleName:     doc.VissibleName,
				SourceDeviceDesc: "some-client",
				SourceDeviceID:   "12345",
				Parent:           doc.Parent,
			},
			PublishTime:  tt,
			PublishTime2: tt,
		},
		Subscription: "dummy-subscription",
	}

	return msg
}
