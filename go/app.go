package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

//todo: config
const (
	dataDir     = "data"
	uploadUrl   = "http://localhost:3000"
	defaultHost = "local.appspot.com"
)

func EchoServer(ws *websocket.Conn) {
	b := []byte("ABC")

	ws.Write(b)
}

type updateStatusRequest struct {
	ID             string `json:"ID"`
	Parent         string `json:"Parent"`
	Version        int    `json:"Version"`
	Message        string `json:"Message"`
	Success        bool   `json:"Success"`
	ModifiedClient string `json:"ModifiedClient"`
	Type           string `json:"Type"`
	VissibleName   string `json:"VissibleName"`
	CurrentPage    int    `json:"CurrentPage"`
	Bookmarked     bool   `json:"Bookmarked"`
}
type rawDocument struct {
	ID                string `json:"ID"`
	Version           int    `json:"Version"`
	Message           string `json:"Message"`
	Success           bool   `json:"Success"`
	BlobURLGet        string `json:"BlobURLGet"`
	BlobURLGetExpires string `json:"BlobURLGetExpires"`
	BlobURLPut        string `json:"BlobURLPut"`
	BlobURLPutExpires string `json:"BlobURLPutExpires"`
	ModifiedClient    string `json:"ModifiedClient"`
	Type              string `json:"Type"`
	VissibleName      string `json:"VissibleName"`
	CurrentPage       int    `json:"CurrentPage"`
	Bookmarked        bool   `json:"Bookmarked"`
	Parent            string `json:"Parent"`
}
type idRequest struct {
	Id string `json:"ID"`
}
type documentRequest struct {
	Id         string `json:"ID"`
	Message    string `json:"Mesasge"`
	Success    bool   `json:"Success"`
	BlobUrlPut string `json:"BlobURLPut"`
	Version    int    `json:"Version"`
}
type hostResponse struct {
	Host   string `json:"Host"`
	Status string `json:"Status"`
}

type deviceTokenRequest struct {
	Code       string `json:"code"`
	DeviceDesc string `json:"deviceDesc"`
	DeviceId   string `json:"deviceID"`
}

func main() {
	gin.ForceConsoleColor()
	r := gin.Default()
	fmt.Println("Running")
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.PUT("/test", func(c *gin.Context) {
		fmt.Println("name: " + c.Query("name"))
		out, err := os.Create("file.txt")
		if err != nil {
		}
		io.Copy(out, c.Request.Body)
		out.Close()

	})

	r.GET("/", func(c *gin.Context) {
		c.String(200, "%s", "hi")
	})

	r.GET("/service/json/1/:service", func(c *gin.Context) {
		svc := c.Param("service")
		fmt.Printf("Requested: %s\n", svc)
		response := hostResponse{Host: defaultHost, Status: "OK"}
		c.JSON(200, response)
	})

	r.POST("/token/json/2/device/new", func(c *gin.Context) {
		var json deviceTokenRequest
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fmt.Printf("Request: %s\n", json)
		c.String(200, "some device token")
	})

	r.POST("/token/json/2/user/new", func(c *gin.Context) {
		auth := c.Request.Header["Authorization"]

		fmt.Printf("Request: %s\n", auth)
		c.String(200, "some user token")
	})

	r.GET("/notifications/ws/json/1", func(c *gin.Context) {
		//TODO: channel
		handler := websocket.Handler(EchoServer)
		handler.ServeHTTP(c.Writer, c.Request)
	})

	r.PUT("/document-storage/json/2/upload/request", func(c *gin.Context) {
		var req []documentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response := []documentRequest{}
		for _, r := range req {
			id := r.Id
			if id == "" {
				id = "someid"
			}
			url := uploadUrl + "/upload?id=" + id
			fmt.Println(url)
			dr := documentRequest{BlobUrlPut: url, Id: id, Success: true, Version: 1}
			response = append(response, dr)
		}

		c.JSON(200, response)
	})

	r.PUT("/upload", func(c *gin.Context) {

		id := c.Query("id")
		body := c.Request.Body
		defer body.Close()
		fullPath := path.Join(dataDir, fmt.Sprintf("%s.zip", id))
		file, err := os.Create(fullPath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		defer file.Close()
		io.Copy(file, body)

		fmt.Printf("Request: %s\n", id)
		c.JSON(200, gin.H{})
	})

	r.GET("/download", func(c *gin.Context) {

		id := c.Query("id")
		fmt.Printf("Request: %s\n", id)

		fullPath := path.Join(dataDir, fmt.Sprintf("%s.zip", id))

		//content, err := ioutil.ReadFile("test.zip")
		// if err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// 	return
		// }
		c.File(fullPath)
		//c.Data(200, "octetstream/binary", content)
	})

	r.PUT("/document-storage/json/2/upload/update-status", func(c *gin.Context) {
		var req []updateStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		for _, r := range req {
			path := path.Join(dataDir, fmt.Sprintf("%s.metadata", r.ID))
			file, err := os.Create(path)
			if err != nil {
				return
				// log
			}
			defer file.Close()

			js, err := json.Marshal(r)
			file.Write(js)
			if err != nil {
				// log
			}
		}

		c.String(200, "")
	})

	r.PUT("/document-storage/json/2/delete", func(c *gin.Context) {
		var json []idRequest

		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		for _, j := range json {
			fmt.Println(j)
		}
		c.JSON(200, gin.H{})
	})

	r.GET("/document-storage/json/2/docs", func(c *gin.Context) {
		withBlob := c.Query("withBlob")
		docId := c.Query("docId")
		fmt.Println(withBlob, docId)
		result := []rawDocument{}

		files, err := ioutil.ReadDir(dataDir)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, f := range files {
			ext := filepath.Ext(f.Name())
			fmt.Println(ext)
			if ext != ".metadata" {
				continue
			}
			fullPath := path.Join(dataDir, f.Name())
			f, err := os.Open(fullPath)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer f.Close()
			content, err := ioutil.ReadAll(f)
			if err != nil {
				fmt.Println(err)
				return
			}

			response := rawDocument{}
			response.BlobURLGet = uploadUrl + "/download?id=" + response.ID
			err = json.Unmarshal(content, &response)
			if err != nil {
			}
			result = append(result, response)
		}

		c.JSON(200, result)
	})

	err := os.MkdirAll(dataDir, 0600)

	if err != nil {
		panic(err)
	}

	r.Run(":3000") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
