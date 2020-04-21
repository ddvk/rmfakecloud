package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/net/websocket"
)

type Test struct {
	name *string
}

func main() {
	log.SetOutput(os.Stdout)
	n := "blah"
	n2 := "bongo"
	t1 := Test{name: &n}
	t2 := t1
	t1.name = &n2

	fmt.Println(*t2.name)
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
