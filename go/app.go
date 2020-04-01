package main

import "io"
import "golang.org/x/net/websocket"
import "github.com/gin-gonic/gin"
import "fmt"
import "os"
import "net/http"
import "io/ioutil"

func EchoServer(ws *websocket.Conn) {
	b := []byte("ABC")

	ws.Write(b)
}

const defaultHost = "local.appspot.com"

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
		var json []documentRequest
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id := json[0].Id
		if id == "" {
			id = "someid"
		}

		response := documentRequest{BlobUrlPut: fmt.Sprintf("http://localhost:3000/upload?id=%v", id), Id: id, Success: true, Version: 1}

		c.JSON(200, response)
	})

	r.GET("/upload", func(c *gin.Context) {

		id := c.Query("id")
		fmt.Printf("Request: %s\n", id)
		c.JSON(200, gin.H{})
	})

	r.GET("/download", func(c *gin.Context) {

		id := c.Query("id")
		fmt.Printf("Request: %s\n", id)
		content, err := ioutil.ReadFile("test.zip")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Data(200, "octetstream/binary", content)
	})

	r.GET("/document-storage/json/2/upload/update-status", func(c *gin.Context) {

		c.String(200, "")
	})

	r.PUT("/document-storage/json/2/delete", func(c *gin.Context) {
		var json []idRequest

		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fmt.Println(json)
		c.JSON(200, gin.H{})
	})

	r.GET("/document-storage/json/2/docs", func(c *gin.Context) {
		withBlob := c.Query("withBlob")
		docId := c.Query("docId")
		fmt.Println(withBlob, docId)
		response := rawDocument{VissibleName: "stuff"}

		c.JSON(200, response)
	})

	r.Run(":3000") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
