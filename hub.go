package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients      map[*Client]bool
	additions    chan *Client
	removals     chan *Client
	notification chan *wsMessage
}

func (h *Hub) Send(msg wsMessage) {
	for c, _ := range h.clients {
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
	ntf chan wsMessage
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
		ntf: make(chan wsMessage),
	}
	go client.Read(c)
	h.additions <- client
	client.Write(c)
	h.removals <- client
}
