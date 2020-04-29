package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

const (
	defaultPort     = "3000"
	defaultDataDir  = "data"
	defaultTrashDir = "trash"
	defaultHost     = "local.appspot.com"
)

//todo: config
var dataDir string
var uploadUrl string
var port string

//todo: claims unmarshal
type MyCustomClaims struct {
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
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := getToken(c)
		if err != nil {
			log.Println(token)
			c.Set("userId", "abc")
		}
		c.Next()
	}
}
func RequestLoggerMiddleware() gin.HandlerFunc {
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
func main() {
	gin.ForceConsoleColor()
	log.SetOutput(os.Stdout)

	hub := NewHub()
	router := gin.Default()
	router.Use(RequestLoggerMiddleware())

	router.GET("/", func(c *gin.Context) {
		count := hub.ClientCount()
		c.String(200, "Working, %d clients", count)
	})
	// register device
	router.POST("/token/json/2/device/new", func(c *gin.Context) {
		var json deviceTokenRequest
		if err := c.ShouldBindJSON(&json); err != nil {
			badReq(c, err.Error())
			return
		}

		log.Printf("Request: %s\n", json)
		c.String(200, "some_device_token")
	})

	//todo: pass the token in the url
	router.PUT("/storage", func(c *gin.Context) {
		id := c.Query("id")
		log.Printf("Uploading id %s\n", id)
		body := c.Request.Body
		defer body.Close()

		err := saveUpload(body, id)
		if err != nil {
			fmt.Println(err)
			c.String(500, "set up us the bomb")
			c.Abort()
		}

		c.JSON(200, gin.H{})
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
		response := hostResponse{Host: defaultHost, Status: "OK"}
		c.JSON(200, response)
	})

	r := router.Group("/")
	r.Use(AuthMiddleware())
	{
		r.POST("/token/json/3/device/delete", func(c *gin.Context) {

			c.String(204, "")
		})

		// websocket notifications
		r.GET("/notifications/ws/json/1", func(c *gin.Context) {
			userId := c.GetString("userId")
			log.Println("accepting websocket", userId)
			hub.ConnectWs(c.Writer, c.Request)
			log.Println("closing the ws")
		})
		// live sync
		r.GET("/livesync/ws/json/2/:authid/sub", func(c *gin.Context) {
			hub.ConnectWs(c.Writer, c.Request)
		})

		r.PUT("/document-storage/json/2/upload/request", func(c *gin.Context) {
			var req []uploadRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				log.Println(err)
				badReq(c, err.Error())
				return
			}

			response := []uploadResponse{}
			for _, r := range req {
				id := r.Id
				if id == "" {
					badReq(c, "no id")
				}
				url := formatStorageUrl(id)
				log.Println(url)
				dr := uploadResponse{BlobUrlPut: url, Id: id, Success: true, Version: r.Version}
				response = append(response, dr)
			}

			c.JSON(200, response)
		})

		r.GET("/storage", func(c *gin.Context) {
			id := c.Query("id")
			if id == "" {
				badReq(c, "no id supplied")
				return
			}
			log.Printf("Requestng Id: %s\n", id)
			fullPath := path.Join(dataDir, filepath.Base(fmt.Sprintf("%s.zip", id)))
			log.Println("Fullpath:", fullPath)

			c.File(fullPath)
		})

		r.PUT("/document-storage/json/2/upload/update-status", func(c *gin.Context) {
			// b := c.Request.Body
			// bm, _ := ioutil.ReadAll(b)
			// log.Println(string(bm))
			// c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bm))
			var req []rawDocument
			if err := c.ShouldBindJSON(&req); err != nil {
				log.Println(err)
				badReq(c, err.Error())
				return
			}
			result := []statusResponse{}
			for _, r := range req {
				log.Println("For Id: ", r.Id)
				log.Println(" Name: ", r.VissibleName)
				path := path.Join(dataDir, fmt.Sprintf("%s.metadata", r.Id))

				ok := false
				event := "DocAdded"
				message := ""

				js, err := json.Marshal(r)
				if err != nil {
					log.Println(err)
				} else {
					err = ioutil.WriteFile(path, js, 0700)
					if err == nil {
						ok = true
						//fix it: id of subscriber
						msg := newWs(&r, event)
						hub.Send(msg)
					} else {
						message = err.Error()
						log.Println(err)
					}
				}
				result = append(result, statusResponse{Id: r.Id, Success: ok, Message: message})
			}

			c.JSON(200, result)
		})

		r.PUT("/document-storage/json/2/delete", func(c *gin.Context) {
			var req []idRequest

			if err := c.ShouldBindJSON(&req); err != nil {
				log.Println("bad request")
				badReq(c, err.Error())
				return
			}

			result := []statusResponse{}
			for _, r := range req {
				metadata, err := loadMetadata(r.Id+".metadata", false)
				ok := true
				if err == nil {
					err := deleteFile(r.Id)
					if err != nil {
						log.Println(err)
						ok = false
					}
					msg := newWs(metadata, "DocDeleted")
					hub.Send(msg)
				}
				result = append(result, statusResponse{Id: r.Id, Success: ok})
			}

			c.JSON(200, result)
		})

		r.GET("/document-storage/json/2/docs", func(c *gin.Context) {
			withBlob, err := strconv.ParseBool(c.Query("withBlob"))
			docId := c.Query("doc")
			log.Println(withBlob, docId)
			result := []*rawDocument{}

			files, err := ioutil.ReadDir(dataDir)
			if err != nil {
				log.Println(err)
				return
			}

			if docId != "" {
				doc, err := loadMetadata(fmt.Sprintf("%s.metadata", docId), withBlob)
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
					doc, err := loadMetadata(f.Name(), withBlob)
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
	}
	// configs
	var err error
	data := os.Getenv("DATADIR")
	if data != "" {
		dataDir = data
	} else {
		dataDir, err = filepath.Abs(defaultDataDir)
		if err != nil {
			panic(err)
		}
	}
	err = os.MkdirAll(path.Join(dataDir, defaultTrashDir), 0700)
	if err != nil {
		panic(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	uploadUrl = os.Getenv("STORAGE_URL")
	if uploadUrl == "" {
		host, err := os.Hostname()
		if err != nil {
			log.Println("cannot get hostname")
			host = defaultHost
		}
		uploadUrl = fmt.Sprintf("http://%s:%s", host, port)
	}

	if err != nil {
		panic(err)
	}

	log.Println("File will be saved in: ", dataDir)
	log.Println("Url the device should use: ", uploadUrl)

	router.Run(":" + port)
}
