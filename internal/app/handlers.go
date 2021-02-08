package app

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/email"
	"github.com/ddvk/rmfakecloud/internal/hwr"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (app *App) getDeviceClaims(c *gin.Context) (*common.DeviceClaims, error) {
	token, err := common.GetToken(c)
	if err != nil {
		return nil, err
	}
	claims := &common.DeviceClaims{}
	err = common.ClaimsFromToken(claims, token, app.cfg.JWTSecretKey)
	if err != nil {
		return nil, err
	}
	if claims.UserID == "" {
		return nil, fmt.Errorf("wrong token, missing userid")
	}
	return claims, nil
}

func (app *App) getUserClaims(c *gin.Context) (*common.UserClaims, error) {
	token, err := common.GetToken(c)
	log.Info("Token:", token)
	if err != nil {
		return nil, err
	}
	claims := &common.UserClaims{}
	err = common.ClaimsFromToken(claims, token, app.cfg.JWTSecretKey)
	if err != nil {
		return nil, err
	}
	if claims.Profile.UserId == "" {
		return nil, fmt.Errorf("wrong token, missing userid")
	}
	return claims, nil
}

func (app *App) newDevice(c *gin.Context) {
	var tokenRequest messages.DeviceTokenRequest
	if err := c.ShouldBindJSON(&tokenRequest); err != nil {
		badReq(c, err.Error())
		return
	}

	code := strings.ToUpper(tokenRequest.Code)
	log.Info("Got code ", code)

	uid, err := app.codeConnector.ConsumeCode(code)
	if err != nil {
		log.Warn(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	log.Info("Request: ", tokenRequest)
	log.Info("Token for:", uid)

	// generate the JWT token
	claims := &common.DeviceClaims{
		DeviceDesc: tokenRequest.DeviceDesc,
		DeviceId:   tokenRequest.DeviceId,
		UserID:     uid,
		StandardClaims: jwt.StandardClaims{
			Audience: common.ApiUsage,
		},
	}

	tokenString, err := common.SignClaims(claims, app.cfg.JWTSecretKey)
	if err != nil {
		log.Warn(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.String(http.StatusOK, tokenString)
}

func (app *App) newUserToken(c *gin.Context) {
	deviceToken, err := app.getDeviceClaims(c)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	user, err := app.userStorer.GetUser(deviceToken.UserID)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if user == nil {
		log.Warn("User not found: ", deviceToken.UserID)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &common.UserClaims{
		Profile: common.Auth0profile{
			UserId:        deviceToken.UserID,
			IsSocial:      false,
			Name:          user.Name,
			Nickname:      user.Nickname,
			Email:         user.Email,
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

	tokenString, err := common.SignClaims(claims, app.cfg.JWTSecretKey)
	if err != nil {
		badReq(c, err.Error())
		return
	}
	c.String(http.StatusOK, tokenString)
}

func (app *App) sendEmail(c *gin.Context) {
	log.Println("Sending email")
	uid := c.GetString(userID)

	log.Info("Sending mail for: ", uid)

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
			app.hub.Send(uid, msg)
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
			app.hub.Send(uid, msg)
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
	app.hub.ConnectWs(uid, c.Writer, c.Request)
	log.Println("closing the ws")
}
