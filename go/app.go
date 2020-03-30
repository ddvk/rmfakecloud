package main

import "io"
import "golang.org/x/net/websocket"
import "github.com/gin-gonic/gin"
import "fmt"
import "os"

func EchoServer(ws *websocket.Conn) {
	b := []byte("ABC")

	ws.Write(b)
}

func main() {
	gin.ForceConsoleColor()
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/ws", func(c *gin.Context) {
		handler := websocket.Handler(EchoServer)
		handler.ServeHTTP(c.Writer, c.Request)
	})

	r.PUT("/test", func(c *gin.Context) {
		fmt.Println("name: " + c.Query("name"))
		out, err := os.Create("file.txt")
		if err != nil {
		}
		io.Copy(out, c.Request.Body)
		out.Close()

	})
	r.Run(":3000") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
