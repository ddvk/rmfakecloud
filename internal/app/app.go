package app

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/db"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// App web app
type App struct {
	router *gin.Engine
	cfg    *config.Config
}

// Start starts the app
func (app *App) Start() {
	app.router.Run(":" + app.cfg.Port)
}

type myCustomClaims struct {
	Foo string `json:"foo"`
	jwt.StandardClaims
}

func getToken(c *gin.Context) (string, error) {
	auth := c.Request.Header["Authorization"]

	if len(auth) < 1 {
		accessDenied(c, "missing token")
		return "", errors.New("missing token")
	}
	token := strings.Split(auth[0], " ")
	if len(token) < 2 {
		return "", errors.New("missing token")
	}
	parts := strings.Split(token[1], ".")
	if len(parts) != 3 {
		log.Println("not jwt")
		return "", nil
	}

	payload, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		log.Println(err)
		return string(payload), nil
	}
	return "", nil
}
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := getToken(c)
		if err != nil {
			log.Println(token)
			c.Set("userId", "abc")
		}
		c.Next()
	}
}
func requestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path != "/storage" {
			var buf bytes.Buffer
			tee := io.TeeReader(c.Request.Body, &buf)
			body, _ := ioutil.ReadAll(tee)
			c.Request.Body = ioutil.NopCloser(&buf)
			log.Println(c.Request.Header)
			log.Println(string(body))
		}
		c.Next()
	}
}
func accessDenied(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{"error": message})
	c.Abort()
}

func badReq(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
	c.Abort()
}

func internalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
	c.Abort()
}

// NewApp constructs an app
func NewApp(cfg *config.Config, metaStorer db.MetadataStorer, docStorer storage.DocumentStorer) App {
	hub := NewHub()
	gin.ForceConsoleColor()
	router := gin.Default()

	router.Use(requestLoggerMiddleware())

	docStorer.RegisterRoutes(router)

	router.GET("/", func(c *gin.Context) {
		count := hub.ClientCount()
		c.String(200, "Working, %d clients", count)
	})
	// register device
	router.POST("/token/json/2/device/new", func(c *gin.Context) {
		var json messages.DeviceTokenRequest
		if err := c.ShouldBindJSON(&json); err != nil {
			badReq(c, err.Error())
			return
		}

		log.Printf("Request: %s\n", json)
		c.String(200, "some_device_token")
	})

	// create new access token
	router.POST("/token/json/2/user/new", func(c *gin.Context) {
		token, err := getToken(c)
		if err != nil {
			log.Println("Got: ", token)
		}
		c.String(200, "some_user_token")
	})

	//service locator
	router.GET("/service/json/1/:service", func(c *gin.Context) {
		svc := c.Param("service")
		log.Printf("Requested: %s\n", svc)
		response := messages.HostResponse{Host: config.DefaultHost, Status: "OK"}
		c.JSON(200, response)
	})

	r := router.Group("/")
	r.Use(authMiddleware())
	{
		//unregister device
		r.POST("/token/json/3/device/delete", func(c *gin.Context) {
			c.String(204, "")
		})

		// websocket notifications
		r.GET("/notifications/ws/json/1", func(c *gin.Context) {
			userID := c.GetString("userId")
			log.Println("accepting websocket", userID)
			hub.ConnectWs(c.Writer, c.Request)
			log.Println("closing the ws")
		})
		// live sync
		r.GET("/livesync/ws/json/2/:authid/sub", func(c *gin.Context) {
			hub.ConnectWs(c.Writer, c.Request)
		})

		r.PUT("/document-storage/json/2/upload/request", func(c *gin.Context) {
			var req []messages.UploadRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				log.Println(err)
				badReq(c, err.Error())
				return
			}

			response := []messages.UploadResponse{}

			for _, r := range req {
				id := r.Id
				if id == "" {
					badReq(c, "no id")
				}
				url := docStorer.GetStorageURL(id)
				log.Println(url)
				dr := messages.UploadResponse{BlobUrlPut: url, Id: id, Success: true, Version: r.Version}
				response = append(response, dr)
			}

			c.JSON(200, response)
		})

		r.PUT("/document-storage/json/2/upload/update-status", func(c *gin.Context) {
			var req []messages.RawDocument
			if err := c.ShouldBindJSON(&req); err != nil {
				log.Println(err)
				badReq(c, err.Error())
				return
			}
			result := []messages.StatusResponse{}
			for _, r := range req {
				log.Println("For Id: ", r.Id)
				log.Println(" Name: ", r.VissibleName)

				ok := false
				event := "DocAdded"
				message := ""

				err := metaStorer.UpdateMetadata(&r)
				if err == nil {
					ok = true
					//fix it: id of subscriber
					msg := newWs(&r, event)
					hub.Send(msg)
				} else {
					message = err.Error()
					log.Println(err)
				}
				result = append(result, messages.StatusResponse{Id: r.Id, Success: ok, Message: message})
			}

			c.JSON(200, result)
		})

		r.PUT("/document-storage/json/2/delete", func(c *gin.Context) {
			var req []messages.IdRequest

			if err := c.ShouldBindJSON(&req); err != nil {
				log.Println("bad request")
				badReq(c, err.Error())
				return
			}

			result := []messages.StatusResponse{}
			for _, r := range req {
				metadata, err := metaStorer.GetMetadata(r.Id, false)
				ok := true
				if err == nil {
					err := docStorer.RemoveDocument(r.Id)
					if err != nil {
						log.Println(err)
						ok = false
					}
					msg := newWs(metadata, "DocDeleted")
					hub.Send(msg)
				}
				result = append(result, messages.StatusResponse{Id: r.Id, Success: ok})
			}

			c.JSON(200, result)
		})

		r.GET("/document-storage/json/2/docs", func(c *gin.Context) {
			withBlob, _ := strconv.ParseBool(c.Query("withBlob"))
			docID := c.Query("doc")
			log.Println(withBlob, docID)
			result := []*messages.RawDocument{}

			var err error
			if docID != "" {
				//load single document
				var doc *messages.RawDocument
				doc, err = metaStorer.GetMetadata(docID, withBlob)
				if err == nil {
					result = append(result, doc)
				}
			} else {
				//load all
				result, err = metaStorer.GetAllMetadata(withBlob)
			}

			if err != nil {
				log.Println(err)
				internalError(c, "blah")
				return
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
			//todo: pass to the hwr endpoint
			//return json
			c.String(200, "%s", "hi")
		})
	}

	app := App{
		router: router,
		cfg:    cfg,
	}

	return app
}
