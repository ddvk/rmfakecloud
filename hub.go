package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients      map[*Client]bool
	additions    chan *Client
	removals     chan *Client
	notification chan *wsMessage
}

func (h *Hub) Send(msg wsMessage) {
	for k, _ := range h.clients {
		k.ntf <- msg
	}
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

//todo O(n)
func (h *Hub) removeClient(c *Client) {
	delete(h.clients, c)
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
}

func (c *Client) Read(ws *websocket.Conn) {
	defer ws.Close()
	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		log.Println("Message: ", string(p))

	}
}
func (c *Client) Write(ws *websocket.Conn) {
	defer ws.Close()
	for {
		select {
		case m := <-c.ntf:
			log.Println("sending notification")
			err := ws.WriteJSON(m)
			if err != nil {
				break
			}
		}
	}
}

// upgrade the connection to websocket
func (h *Hub) ConnectWs(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			log.Printf("check origin")
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
		ntf: make(chan wsMessage),
	}
	h.additions <- client
	go client.Read(c)
	client.Write(c)
	h.removals <- client
	log.Println("done with this client")
}
