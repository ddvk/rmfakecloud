package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type Tokens struct {
	DeviceToken string
	UserToken   string
}
type HostResponse struct {
	Host   string `json:"Host"`
	Status string `json:"Status"`
}

const (
	notifications  = "/notifications/ws/json/1"
	serviceLocator = "/service/json/1/notifications?environment=production&group=auth0%7Crm2&apiVer=1"

	origin = "https://service-manager-production-dot-remarkable-production.appspot.com"
)

func getUrl(host string, tokens Tokens) (string, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", host+serviceLocator, nil)
	if err != nil {
		return "", fmt.Errorf("Got error %w", err)
	}
	req.Header.Add("Authorization", "Bearrer "+tokens.UserToken)
	req.Header.Add("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Got error %w", err)
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("response Body:", string(body))
	hostResponse := &HostResponse{}
	json.Unmarshal(body, hostResponse)
	return hostResponse.Host, nil

}

func auth(host string, tokens Tokens, withHttps bool) (*websocket.Conn, error) {
	schema := "ws://"
	if withHttps {
		schema = "wss://"
	}
	url := schema + host + notifications
	fmt.Println(url)

	header := http.Header{
		"Authorization": {"Bearer " + tokens.UserToken},
	}
	conn, response, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		fmt.Println("response headers:", response.Header)
		fmt.Println("response Body:", string(body))
	}

	return conn, err

}

func loadToken(configFile string) (*Tokens, error) {
	tokens := Tokens{}
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(content, &tokens)
	if err != nil {
		return nil, err
	}
	return &tokens, nil
}

func main() {
	logger := logrus.New()
	log.SetOutput(logger.Writer())

	host := flag.String("h", "http://localhost:3001", "host, use origin for the real ip")
	rmapiConf := flag.String("c", "", "rmapi .conf file")
	flag.Parse()

	if *rmapiConf == "" {
		fmt.Println("config file needed")
		os.Exit(1)
	}

	err := func() error {
		tokens, err := loadToken(*rmapiConf)
		if err != nil {
			return err
		}

		if *host == "origin" {
			*host = origin
		}

		url, err := getUrl(*host, *tokens)
		if err != nil {
			return err
		}

		withHttps := false
		if strings.Index(*host, "https") == 0 {
			withHttps = true
			*host = strings.TrimPrefix(*host, "https://")
		} else {
			*host = strings.TrimPrefix(*host, "http://")
		}

		//TODO:
		if url == "local.appspot.com" {
			url = *host
		}

		conn, err := auth(url, *tokens, withHttps)
		if err != nil {
			return err
		}

		done := make(chan struct{}, 1)
		//defer conn.Close()
		go func() {
			for {
				fmt.Print("Enter text: ")
				var text string
				fmt.Scanln(&text)
				if text == "" {
					text = "(null)"
				}
				err := conn.WriteMessage(websocket.TextMessage, []byte(text))
				if err != nil {
					fmt.Println("Cant send")
					done <- struct{}{}
					return
				}
				switch text {
				case "q":
					conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1000, "bye"))
					done <- struct{}{}
					return
				case "g":
					fmt.Println("exit with close...")
					conn.Close()
				}
			}
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, 1000) {
					return err
				}
				fmt.Println("closed")
				break
			}
			fmt.Println(string(message))
		}
		<-done
		return nil
	}()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
