package app

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/email"
	"github.com/ddvk/rmfakecloud/internal/hwr"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (app *App) getToken(c *gin.Context) (claims *messages.Auth0token, err error) {
	auth := c.Request.Header["Authorization"]

	if len(auth) < 1 {
		return nil, errors.New("missing token")
	}
	token := strings.Split(auth[0], " ")
	if len(token) < 2 {
		return nil, errors.New("missing token")
	}
	strToken := token[1]

	claims = &messages.Auth0token{}
	_, err = jwt.ParseWithClaims(strToken, claims,
		func(token *jwt.Token) (interface{}, error) {
			return app.cfg.JWTSecretKey, nil
		})
	log.Info("token parsed")
	return
}

func (app *App) newDevice(c *gin.Context) {
	var json messages.DeviceTokenRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		badReq(c, err.Error())
		return
	}

	log.Printf("Request: %s\n", json)

	// generate the JWT token
	expirationTime := time.Now().Add(356 * 24 * time.Hour)
	claims := &messages.Auth0token{
		DeviceDesc: json.DeviceDesc,
		DeviceId:   json.DeviceId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Subject:   "rM Device Token",
		},
	}

	deviceToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := deviceToken.SignedString(app.cfg.JWTSecretKey)
	if err != nil {
		badReq(c, err.Error())
		return
	}

	c.String(http.StatusOK, tokenString)
}

func (app *App) newUserToken(c *gin.Context) {
	key := app.cfg.JWTSecretKey
	deviceToken, err := app.getToken(c)
	if err != nil {
		log.Warnln(err)
	}
	log.Debug(deviceToken)

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &messages.Auth0token{
		Profile: messages.Auth0profile{
			UserId:        "auth0|1234",
			IsSocial:      false,
			Name:          "rmFake",
			Nickname:      "rmFake",
			Email:         "fake@rmfake",
			EmailVerified: true,
			Picture:       "image.png",
			CreatedAt:     time.Now(),
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
	tokenString, err := userToken.SignedString(key)
	if err != nil {
		badReq(c, err.Error())
		return
	}
	//TODO: do something with the token

	c.String(http.StatusOK, tokenString)
}

func (app *App) sendEmail(c *gin.Context) {
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
}
func (app *App) listDocuments(c *gin.Context) {

	uid := c.GetString(userID)
	withBlob, _ := strconv.ParseBool(c.Query("withBlob"))
	docID := c.Query("doc")
	log.Println("params: withBlob, docId", withBlob, docID)
	result := []*messages.RawDocument{}

	var err error
	if docID != "" {
		//load single document
		var doc *messages.RawDocument
		doc, err = app.metaStorer.GetMetadata(uid, docID, withBlob)
		if err == nil {
			result = append(result, doc)
		}
	} else {
		//load all
		result, err = app.metaStorer.GetAllMetadata(uid, withBlob)
	}

	if err != nil {
		log.Error(err)
		internalError(c, "blah")
		return
	}

	c.JSON(http.StatusOK, result)
}
func (app *App) deleteDocument(c *gin.Context) {
	uid := c.GetString(userID)

	var req []messages.IdRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("bad request")
		badReq(c, err.Error())
		return
	}

	result := []messages.StatusResponse{}
	for _, r := range req {
		metadata, err := app.metaStorer.GetMetadata(uid, r.Id, false)
		ok := true
		if err == nil {
			err := app.docStorer.RemoveDocument(uid, r.Id)
			if err != nil {
				log.Error(err)
				ok = false
			}
			msg := newWs(metadata, "DocDeleted")
			app.hub.Send(msg)
		}
		result = append(result, messages.StatusResponse{Id: r.Id, Success: ok})
	}

	c.JSON(http.StatusOK, result)
}
func (app *App) updateStatus(c *gin.Context) {
	uid := c.GetString(userID)
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

		err := app.metaStorer.UpdateMetadata(uid, &r)
		if err == nil {
			ok = true
			//fix it: id of subscriber
			msg := newWs(&r, event)
			app.hub.Send(msg)
		} else {
			message = err.Error()
			log.Error(err)
		}
		result = append(result, messages.StatusResponse{Id: r.Id, Success: ok, Message: message})
	}

	c.JSON(http.StatusOK, result)
}

func (app *App) locateService(c *gin.Context) {
	svc := c.Param("service")
	log.Printf("Requested: %s\n", svc)
	response := messages.HostResponse{Host: config.DefaultHost, Status: "OK"}
	c.JSON(http.StatusOK, response)
}
func (app *App) uploadRequest(c *gin.Context) {
	uid := c.GetString(userID)
	var req []messages.UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}

	response := []messages.UploadResponse{}

	for _, r := range req {
		documentID := r.Id
		if documentID == "" {
			badReq(c, "no id")
		}
		exp := time.Now().Add(time.Minute)
		url, err := app.docStorer.GetStorageURL(uid, exp, documentID)
		if err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		log.Debugln("StorageUrl: ", url)
		dr := messages.UploadResponse{BlobUrlPut: url, Id: documentID, Success: true, Version: r.Version}
		response = append(response, dr)
	}

	c.JSON(http.StatusOK, response)
}

func (app *App) handleHwr(c *gin.Context) {
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
}
func (app *App) connectWebSocket(c *gin.Context) {
	uid := c.GetString(userID)
	log.Info("accepting websocket from:", uid)
	app.hub.ConnectWs(c.Writer, c.Request)
	log.Println("closing the ws")
}
