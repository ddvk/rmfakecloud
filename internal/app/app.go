package app

import (
	"bytes"
	"context"
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
	"github.com/ddvk/rmfakecloud/internal/ui"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
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

// Stop the app
func (app *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
}

func getToken(c *gin.Context, jwtSecretKey []byte) (claims *messages.OAuthtoken, err error) {
	auth := c.Request.Header["Authorization"]

	if len(auth) < 1 {
		accessDenied(c, "missing token")
		return nil, errors.New("missing token")
	}
	token := strings.Split(auth[0], " ")
	if len(token) < 2 {
		return nil, errors.New("missing token")
	}

	claims = &messages.OAuthtoken{}
	_, err = jwt.ParseWithClaims(token[1], claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil
	})
	return
}
func authMiddleware(jwtSecretKey []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := getToken(c, jwtSecretKey)
		if err == nil {
			c.Set("userId", strings.TrimPrefix(claims.Profile.UserId, "oauth|"))
			log.Info("got a user token", claims.Profile.UserId)
		} else {
			log.Warn(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
		}
		c.Next()
	}
}

var ignored = []string{"/storage", "/api/v2/document"}

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
func NewApp(cfg *config.Config, metaStorer db.MetadataStorer, docStorer storage.DocumentStorer, userStorer db.UserStorer) App {
	hub := NewHub()
	gin.ForceConsoleColor()
	router := gin.Default()

	corsConfig := cors.DefaultConfig()

	corsConfig.AllowOrigins = []string{"*"}
	
	// To be able to send tokens to the server.
	corsConfig.AllowCredentials = true

	// OPTIONS method for ReactJS
	corsConfig.AddAllowMethods("OPTIONS")
	//corsConfig.AllowHeaders = []string{"*"};
	//corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"};

	corsConfig.AddAllowHeaders("Authorization");

	// Register the middleware
	router.Use(cors.New(corsConfig))

	// router.Use(ginlogrus.Logger(std.Out), gin.Recovery())

	if log.GetLevel() == log.DebugLevel {
		router.Use(requestLoggerMiddleware())
	}

	ui.RegisterUI(router, cfg, userStorer)

	router.Use(requestLoggerMiddleware())

	docStorer.RegisterRoutes(router)

	router.GET("/health", func(c *gin.Context) {
		count := hub.ClientCount()

		c.String(http.StatusOK, "Working, %d clients", count)
	})
	// register device
	router.POST("/token/json/2/device/new", func(c *gin.Context) {
		var json messages.DeviceTokenRequest
		if err := c.ShouldBindJSON(&json); err != nil {
			badReq(c, err.Error())
			return
		}

		log.Printf("Request: %s\n", json)

		// generate the JWT token
		expirationTime := time.Now().Add(356 * 24 * time.Hour)
		claims := &messages.OAuthtoken{
			DeviceDesc: json.DeviceDesc,
			DeviceId:   json.DeviceId,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
				Subject:   "rM Device Token",
			},
		}

		deviceToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := deviceToken.SignedString(cfg.JWTSecretKey)
		if err != nil {
			badReq(c, err.Error())
			return
		}

		c.String(200, tokenString)
	})

	// create new access token
	router.POST("/token/json/2/user/new", func(c *gin.Context) {
		deviceToken, err := getToken(c, cfg.JWTSecretKey)
		if err != nil {
			log.Warnln(err)
		}
		log.Debug(deviceToken)

		expirationTime := time.Now().Add(30 * 24 * time.Hour)
		claims := &messages.OAuthtoken{
			Profile: &messages.OAuthprofile{				
				UserId:        "oauth|1234",
				IsSocial:      false,
				Name:          "rmFake",
				Nickname:      "rmFake",
				Email:         "fake@rmfake",
				EmailVerified: true,
				Picture:       "image.png",
				CreatedAt:     time.Date(2020, 04, 29, 10, 48, 25, 936, time.UTC),
				UpdatedAt:     time.Now(),
			},
			DeviceDesc: deviceToken.DeviceDesc,
			DeviceId:   deviceToken.DeviceId,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
				Subject:   "rM User Token",
			},
		}

		userToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := userToken.SignedString(cfg.JWTSecretKey)
		if err != nil {
			badReq(c, err.Error())
			return
		}
		//TODO: do something with the token

		c.String(200, tokenString)
	})

	//service locator
	router.GET("/service/json/1/:service", func(c *gin.Context) {
		svc := c.Param("service")
		log.Printf("Requested: %s\n", svc)
		response := messages.HostResponse{Host: config.DefaultHost, Status: "OK"}
		c.JSON(http.StatusOK, response)
	})

	r := router.Group("/")
	r.Use(authMiddleware(cfg.JWTSecretKey))
	{
		ui.RegisterUIAuth(r, metaStorer, userStorer)

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

			c.JSON(http.StatusOK, response)
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

			c.JSON(http.StatusOK, result)
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

			c.JSON(http.StatusOK, result)
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

			c.JSON(http.StatusOK, result)
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
			c.String(http.StatusOK, "")
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
			c.Data(http.StatusOK, hwr.JIIX, response)
		})
	}

	app := App{
		router: router,
		cfg:    cfg,
	}

	return app
}
