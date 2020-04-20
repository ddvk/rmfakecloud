package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/net/websocket"
)

func main() {
	log.SetOutput(os.Stdout)
	conn, err := websocket.Dial("ws://localhost:3000/notifications/ws/json/1", "wss", "http://blah")
	if err != nil {

		log.Fatal(err)
		// handle error
	}
	defer conn.Close()
	var message string
	for {
		if err := websocket.Message.Receive(conn, &message); err != nil {
			log.Fatal(err)
		}
		fmt.Println(message)
	}
}
