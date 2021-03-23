package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/db"
	"github.com/ddvk/rmfakecloud/internal/email"
	"github.com/ddvk/rmfakecloud/internal/hwr"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/storage"

	//	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// App web app
type App struct {
	router *gin.Engine
	cfg    *config.Config
	srv    *http.Server
}

// Start starts the app
func (app *App) Start() {
	app.srv = &http.Server{
		Addr:    ":" + app.cfg.Port,
		Handler: app.router,
	}

	if err := app.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}
func (app *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
}

type auth0token struct {
	Profile auth0profile `json:"auth0-profile"`
}
type auth0profile struct {
	UserId string `json:"UserID'`
}

func getToken(c *gin.Context) (parsed *auth0token, err error) {
	auth := c.Request.Header["Authorization"]

	if len(auth) < 1 {
		accessDenied(c, "missing token")
		return nil, errors.New("missing token")
	}
	token := strings.Split(auth[0], " ")
	if len(token) < 2 {
		return nil, errors.New("missing token")
	}
	parts := strings.Split(token[1], ".")
	length := len(parts)
	if length != 3 {
		return nil, errors.New("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		log.Warnln("decode token err", err)
		return nil, err
	}

	parsed = &auth0token{}
	err = json.Unmarshal(payload, &parsed)
	if err != nil {
		return nil, err
	}
	return
}
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := getToken(c)
		if err == nil {
			if err != nil {
				log.Warnln(err)
			}
			c.Set("userId", "abc")
			log.Info("got a user token", token.Profile.UserId)
		} else {
			c.Set("userId", "annon")
			log.Warn(err)
		}
		c.Next()
	}
}

var ignored = []string{"/storage", "/api/v2/document", "/v1/reports"}

func requestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		log.Debugln("header ", c.Request.Header)
		for _, skip := range ignored {
			if skip == c.Request.URL.Path {
				log.Debugln("body logging ignored")
				c.Next()
				return
			}
		}

		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ := ioutil.ReadAll(tee)
		c.Request.Body = ioutil.NopCloser(&buf)
		log.Debugln("body: ", string(body))
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

/// remove remarkable ads
func stripAds(msg string) string {
	br := "<br>--<br>"
	i := strings.Index(msg, br)
	if i > 0 {
		return msg[:i]
	}
	return msg
}

// NewApp constructs an app
func NewApp(cfg *config.Config, metaStorer db.MetadataStorer, docStorer storage.DocumentStorer) App {
	hub := NewHub()
	isDebug := log.GetLevel() == log.DebugLevel

	if !isDebug {
		gin.SetMode(gin.ReleaseMode)
	}

	gin.ForceConsoleColor()
	router := gin.Default()

	if isDebug {
		router.Use(requestLoggerMiddleware())
	}

	docStorer.RegisterRoutes(router)

	router.GET("/settings/v1/beta", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"enrolled": true})
	})

	router.POST("/v1/reports", func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
		log.Info(hex.Dump(body))
		c.Status(400)
	})
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
		//TODO: generate the token
		c.String(200, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRoMC11c2VyaWQiOiJhdXRoMHxhdXRoIiwiZGV2aWNlLWRlc2MiOiJyZW1hcmthYmxlIiwiZGV2aWNlLWlkIjoiUk0xMDAtMDAwLTAwMDAwIiwiaWF0IjoxMjM0MTIzNCwiaXNzIjoick0gV2ViQXBwIiwic3ViIjoick0gRGV2aWNlIFRva2VuIn0.nf3D0dF1c_QbOAqh8e7R4cFQJp_wFa-rVa-PpOe80mw")
	})

	// create new access token
	router.POST("/token/json/2/user/new", func(c *gin.Context) {
		token, err := getToken(c)
		if err != nil {
			log.Warnln(err)
		}
		log.Debug(token)
		//TODO: do something with the token
		c.String(200, "eyJhbGciOiJIUzI1NiIsImtpZCI6IjEiLCJ0eXAiOiJKV1QifQ.eyJhdXRoMC1wcm9maWxlIjp7IlVzZXJJRCI6ImF1dGgwfDEyMzQiLCJJc1NvY2lhbCI6ZmFsc2UsIkNvbm5lY3Rpb24iOiJVc2VybmFtZS1QYXNzd29yZC1BdXRoZW50aWNhdGlvbiIsIk5hbWUiOiJmYWtlQHJtZmFrZSIsIk5pY2tuYW1lIjoiZmFrZSIsIkdpdmVuTmFtZSI6IiIsIkZhbWlseU5hbWUiOiIiLCJFbWFpbCI6InJtZmFrZUBybWFma2UiLCJFbWFpbFZlcmlmaWVkIjp0cnVlLCJQaWN0dXJlIjoiIiwiQ3JlYXRlZEF0IjoiMjAyMC0wNC0yOVQxMDo0ODoyNS45MzZaIiwiVXBkYXRlZEF0IjoiMjAyMS0wMy0xOVQxMTo0MDo0Ni45NzZaIn0sImRldmljZS1kZXNjIjoicmVtYXJrYWJsZSIsImRldmljZS1pZCI6ImRldmljZWlkIiwiZXhwIjoyNjE2MTk3Mjc0LCJpYXQiOjE2MTYxNTQwNzQsImlzcyI6InJNIFdlYkFwcCIsImp0aSI6IiIsIm5iZiI6MTYxNjE1NDA3NCwic2NvcGVzIjoic3luYzpkZWZhdWx0Iiwic3ViIjoick0gVXNlciBUb2tlbiJ9.PIPYgAQvlvKi7YHfxp5PnAYfmN5JdWmANaCxWs0Oo1M")
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
				log.Error(err)
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
				log.Debugln("StorageUrl: ", url)
				dr := messages.UploadResponse{BlobUrlPut: url, Id: id, Success: true, Version: r.Version}
				response = append(response, dr)
			}

			c.JSON(200, response)
		})

		r.PUT("/document-storage/json/2/upload/update-status", func(c *gin.Context) {
			var req []messages.RawDocument
			if err := c.ShouldBindJSON(&req); err != nil {
				log.Error(err)
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
					log.Error(err)
				}
				result = append(result, messages.StatusResponse{Id: r.Id, Success: ok, Message: message})
			}

			c.JSON(200, result)
		})

		r.PUT("/document-storage/json/2/delete", func(c *gin.Context) {
			var req []messages.IdRequest

			if err := c.ShouldBindJSON(&req); err != nil {
				log.Warn("bad request")
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
						log.Error(err)
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
			log.Println("params: withBlob, docId", withBlob, docID)
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
				log.Error(err)
				internalError(c, "blah")
				return
			}

			c.JSON(200, result)
		})

		// send email
		r.POST("/api/v2/document", func(c *gin.Context) {
			log.Println("Sending email")

			form, err := c.MultipartForm()
			if err != nil {
				log.Error(err)
				internalError(c, "not multipart form")
				return
			}
			for k := range form.File {
				log.Debugln("form field", k)
			}
			for k := range form.Value {
				log.Debugln("form value", k)
			}

			emailClient := email.EmailBuilder{
				Subject: form.Value["subject"][0],
				ReplyTo: form.Value["reply-to"][0],
				From:    form.Value["from"][0],
				To:      form.Value["to"][0],
				Body:    stripAds(form.Value["html"][0]),
			}

			for _, file := range form.File["attachment"] {
				f, err := file.Open()
				defer f.Close()
				if err != nil {
					log.Error(err)
					internalError(c, "cant open attachment")
					return
				}
				data, err := ioutil.ReadAll(f)
				if err != nil {
					log.Error(err)
					internalError(c, "cant read attachment")
					return
				}
				emailClient.AddFile(file.Filename, data, file.Header.Get("Content-Type"))
			}
			err = emailClient.Send()
			if err != nil {
				log.Error(err)
				internalError(c, "cant send email")
				return
			}
			c.String(200, "")
		})
		// hwr
		r.POST("/api/v1/page", func(c *gin.Context) {
			body, err := ioutil.ReadAll(c.Request.Body)
			if err != nil || len(body) < 1 {
				log.Warn("no body")
				badReq(c, "missing bbody")
				return
			}
			response, err := hwr.SendRequest(body)
			if err != nil {
				log.Error(err)
				internalError(c, "cannot send")
				return
			}
			c.Data(200, hwr.JIIX, response)

		})
	}

	app := App{
		router: router,
		cfg:    cfg,
	}

	return app
}
