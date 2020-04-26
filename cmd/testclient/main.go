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
	host := os.Getenv("RMAPI_AUTH")
	if host == "" {
		host = "localhost:3000"
	}
	conn, err := websocket.Dial("ws://"+host+"/notifications/ws/json/1", "", "http://none")

	if err != nil {

		log.Fatal(err)
		// handle error
	}
	//defer conn.Close()
	go func() {
		for {
			fmt.Print("Enter text: ")
			var text string
			fmt.Scanln(&text)
			if text == "" {
				text = "(null)"
			}
			websocket.Message.Send(conn, text)
			if text == "q" {
				conn.Close()
				os.Exit(0)
				break
			}
			if text == "g" {
				fmt.Println("exit...")
				os.Exit(0)
			}
		}
	}()
	var message string
	for {
		if err := websocket.Message.Receive(conn, &message); err != nil {
			log.Println(err)
			return
		}
		fmt.Println(message)
	}
}
