package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultPort     = 3000
	defaultDataDir  = "data"
	defaultTrashDir = "trash"
)

//todo: config
var dataDir = "data"
var defaultHost = "local.appspot.com"
var uploadUrl string = "https://" + defaultHost
var port = 3000

func init() {
}

func newWs(id string, name string, version int, typ string) wsMessage {
	msg := wsMessage{}
	msg.Message.MessageId = "1234"
	msg.Message.MessageId2 = "1234"
	msg.Message.Attributes.Auth0UserID = "auth0|12341234123412"
	msg.Message.Attributes.Event = typ
	msg.Message.Attributes.Id = id
	msg.Message.Attributes.Type = "DocumentType"
	msg.Message.Attributes.Version = fmt.Sprintf("%d", version)

	msg.Message.Attributes.VissibleName = name
	msg.Message.Attributes.SourceDeviceDesc = "some-client"
	msg.Message.Attributes.SourceDeviceID = "12345"
	tt, _ := time.Now().MarshalText()
	msg.Message.PublishTime = string(tt)
	msg.Message.PublishTime2 = string(tt)
	msg.Subscription = "dummy-subscription"
	return msg
}
func main() {
	gin.ForceConsoleColor()
	log.SetOutput(os.Stdout)

	hub := NewHub()
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.PUT("/test", func(c *gin.Context) {

		msg := newWs("1234", "test", 0, "DocAdded")
		hub.Send(msg)

	})

	r.GET("/", func(c *gin.Context) {
		c.String(200, "%s", "hi")
	})

	//service locator
	r.GET("/service/json/1/:service", func(c *gin.Context) {
		svc := c.Param("service")
		log.Printf("Requested: %s\n", svc)
		response := hostResponse{Host: defaultHost, Status: "OK"}
		c.JSON(200, response)
	})

	// register device
	r.POST("/token/json/2/device/new", func(c *gin.Context) {
		var json deviceTokenRequest
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("Request: %s\n", json)
		c.String(200, "some device token")
	})

	r.POST("/token/json/3/device/delete", func(c *gin.Context) {
		auth := c.Request.Header["Authorization"]

		log.Printf("Request: %s\n", auth)
		c.String(204, "")
	})
	// create new access token
	r.POST("/token/json/2/user/new", func(c *gin.Context) {
		auth := c.Request.Header["Authorization"]
		log.Printf("Request: %s\n", auth)
		c.String(200, "some user token")
	})

	// websocket notifications
	r.GET("/notifications/ws/json/1", func(c *gin.Context) {
		log.Println("before")
		hub.ConnectWs(c.Writer, c.Request)
		log.Println("after")
	})
	// live sync
	r.GET("/livesync/ws/json/2/:authid/sub", func(c *gin.Context) {
		hub.ConnectWs(c.Writer, c.Request)
	})

	r.PUT("/document-storage/json/2/upload/request", func(c *gin.Context) {
		var req []documentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response := []documentRequest{}
		for _, r := range req {
			id := r.Id
			if id == "" {
				id = "someid"
			}
			url := uploadUrl + "/storage?id=" + id
			log.Println(url)
			dr := documentRequest{BlobUrlPut: url, Id: id, Success: true, Version: 1}
			response = append(response, dr)
		}

		c.JSON(200, response)
	})

	r.PUT("/storage", func(c *gin.Context) {
		log.Println("Uploading...")

		id := c.Query("id")
		body := c.Request.Body
		defer body.Close()
		fullPath := path.Join(dataDir, fmt.Sprintf("%s.zip", id))
		file, err := os.Create(fullPath)
		if err != nil {
			log.Println(err)

			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		defer file.Close()
		io.Copy(file, body)

		log.Printf("Request: %s\n", id)
		c.JSON(200, gin.H{})
	})

	r.GET("/storage", func(c *gin.Context) {

		id := c.Query("id")
		log.Printf("Request: %s\n", id)

		fullPath := path.Join(dataDir, filepath.Base(fmt.Sprintf("%s.zip", id)))
		log.Printf("Fullpath", fullPath)

		c.File(fullPath)
	})

	r.PUT("/document-storage/json/2/upload/update-status", func(c *gin.Context) {
		var req []updateStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result := []statusResponse{}
		for _, r := range req {
			path := path.Join(dataDir, fmt.Sprintf("%s.metadata", r.Id))
			file, err := os.Create(path)
			if err != nil {
				log.Println(err)
				return
			}
			defer file.Close()

			js, err := json.Marshal(r)
			file.Write(js)
			if err != nil {
				log.Println(err)
			}
			log.Println(r)
			result = append(result, statusResponse{Id: r.Id, Success: true})
			//fix it: send not to the device
			msg := newWs(r.Id, r.VissibleName, 0, "DocAdded")
			hub.Send(msg)
		}

		c.JSON(200, result)
	})

	r.PUT("/document-storage/json/2/delete", func(c *gin.Context) {
		var req []idRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Println("bad request")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result := []statusResponse{}
		for _, r := range req {
			mt, err := loadMetadata(r.Id + ".metadata")
			ok := true
			if err == nil {
				err := deleteFile(r.Id)
				if err != nil {
					log.Println(err)
					ok = false
				}
				msg := newWs(r.Id, mt.VissibleName, mt.Version, "DocDeleted")
				hub.Send(msg)
			}
			result = append(result, statusResponse{Id: r.Id, Success: ok})
		}

		c.JSON(200, result)
	})

	r.GET("/document-storage/json/2/docs", func(c *gin.Context) {
		withBlob := c.Query("withBlob")
		docId := c.Query("docId")
		log.Println(withBlob, docId)
		result := []*rawDocument{}

		files, err := ioutil.ReadDir(dataDir)
		if err != nil {
			log.Println(err)
			return
		}

		if docId != "" {
			doc, err := loadMetadata(fmt.Sprintf("%s.metadata", docId))
			if err != nil {
				log.Println(err)
			} else {
				result = append(result, doc)
			}
		} else {

			for _, f := range files {
				ext := filepath.Ext(f.Name())
				if ext != ".metadata" {
					continue
				}
				doc, err := loadMetadata(f.Name())
				if err != nil {
					log.Println(err)
					continue
				}

				result = append(result, doc)
			}
		}

		c.JSON(200, result)
	})

	// send email
	r.POST("/api/v2/document", func(c *gin.Context) {
		log.Println("email")
		file, err := c.FormFile("attachment")
		if err != nil {
			log.Println("no file")
		}
		log.Println("file", file.Filename)
		log.Println("size", file.Size)
		reply := c.PostForm("reply-to")
		from := c.PostForm("from")
		subject := c.PostForm("subject")
		html := c.PostForm("html")

		log.Println("reply-to", reply)
		log.Println("from", from)
		log.Println("subject", subject)
		log.Println("body", html)

		c.String(200, "")
	})
	// hwr
	r.POST("/api/v1/page", func(c *gin.Context) {
		//return json
		c.String(200, "%s", "hi")
	})
	// configs
	var err error
	data := os.Getenv("DATADIR")
	if data != "" {
		dataDir, err = filepath.Abs(data)
		if err != nil {
			panic(err)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	host := os.Getenv("STORAGE_URL")
	if host == "" {
		h, err := os.Hostname()
		if err == nil {
			host = h
		} else {
			host = defaultHost
		}
	}

	uploadUrl = fmt.Sprintf("http://%s:%s", host, port)
	err = os.MkdirAll(path.Join(dataDir, "trash"), 0700)

	if err != nil {
		panic(err)
	}

	log.Println("File will be saved in: ", dataDir)
	log.Println("Url the device should use: ", uploadUrl)

	r.Run(":" + port)
}
