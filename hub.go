package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients []*Client
}

func (h *Hub) Send(msg wsMessage) {
	for _, c := range h.clients {
		c.ntf <- msg
	}
}

func (h *Hub) AddClient(c *Client) {
	h.clients = append(h.clients, c)
}
func NewHub() *Hub {
	h := Hub{}
	return &h
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
func (c *Client) Process(ws *websocket.Conn) {
	defer ws.Close()
	for {
		select {
		case m := <-c.ntf:
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
	client := Client{}
	client.ntf = make(chan wsMessage)
	h.AddClient(&client)
	go client.Read(c)
	client.Process(c)
}
