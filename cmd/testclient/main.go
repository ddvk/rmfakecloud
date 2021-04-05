package main

import (
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

func main() {
	logger := logrus.New()

	log.SetOutput(logger.Writer())

	host := os.Getenv("RMAPI_AUTH")
	if host == "" {
		host = "localhost:3000"
	}
	host = "localhost:3000"
	config, err := websocket.NewConfig("ws://"+host+"/notifications/ws/json/1", "http://none")
	if err != nil {
		log.Fatal(err)
	}
	token := "blah.blah.blah"
	if len(os.Args) > 1 {
		token = os.Args[1]
	}

	config.Header = http.Header{
		"Authorization": {"Bearer " + token},
	}
	conn, err := websocket.DialConfig(config)

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
			log.Error(err)
			return
		}
		fmt.Println(message)
	}
}
