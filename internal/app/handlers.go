package app

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/email"
	"github.com/ddvk/rmfakecloud/internal/hwr"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	internalErrorMessage = "Internal Error"
	handlerLog           = "[handler] "
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
	// log.Debug(handlerLog, "Token: ", token)
	if err != nil {
		return nil, err
	}
	claims := &common.UserClaims{}
	err = common.ClaimsFromToken(claims, token, app.cfg.JWTSecretKey)
	if err != nil {
		return nil, err
	}
	if claims.Profile.UserID == "" {
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

	code := strings.ToLower(tokenRequest.Code)
	log.Info("Got code ", code)

	uid, err := app.codeConnector.ConsumeCode(code)
	if err != nil {
		log.Warn(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	log.Info("Request: ", tokenRequest, "Token for:", uid)

	// generate the JWT token
	claims := &common.DeviceClaims{
		DeviceDesc: tokenRequest.DeviceDesc,
		DeviceID:   tokenRequest.DeviceID,
		UserID:     uid,
		StandardClaims: jwt.StandardClaims{
			Audience: common.APIUsage,
		},
	}

	tokenString, err := common.SignClaims(claims, app.cfg.JWTSecretKey)
	if err != nil {
		log.Warn(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.String(http.StatusOK, tokenString)
}

func (app *App) deleteDevice(c *gin.Context) {
	deviceToken, err := app.getDeviceClaims(c)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	log.Info("Logging out: ", deviceToken.UserID)
	c.Status(http.StatusNoContent)
}

func (app *App) newUserToken(c *gin.Context) {
	deviceToken, err := app.getDeviceClaims(c)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	uid := strings.TrimPrefix(deviceToken.UserID, "auth0|")

	user, err := app.userStorer.GetUser(uid)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if user == nil {
		log.Warn("User not found: ", uid)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	scopes := []string{"hsu", "intgr", "screenshare", "hwcmail:-1", "mail:-1"}

	if user.Sync15 {
		log.Info("Using sync 1.5")
		scopes = append(scopes, syncNew)
	} else {
		scopes = append(scopes, syncDefault)
	}
	scopesStr := strings.Join(scopes, " ")
	log.Info("setting scopes: ", scopesStr)
	now := time.Now()
	expirationTime := now.Add(24 * time.Hour)
	claims := &common.UserClaims{
		Profile: common.Auth0profile{
			UserID:        deviceToken.UserID,
			IsSocial:      false,
			Connection:    "Username-Password-Authentication",
			Name:          user.Email,
			Nickname:      user.Email, // user.Nickname,
			Email:         user.Email,
			EmailVerified: true,
			Picture:       "image.png",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		DeviceDesc: deviceToken.DeviceDesc,
		DeviceID:   deviceToken.DeviceID,
		Scopes:     scopesStr,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			NotBefore: now.Unix(),
			IssuedAt:  now.Unix(),
			Subject:   "rM User Token",
			Issuer:    "rM WebApp",
			Id:        user.Email,
			Audience:  common.APIUsage,
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
	uid := c.GetString(userIDKey)
	log.Info("Sending mail for: ", uid)

	form, err := c.MultipartForm()
	if err != nil {
		log.Error(err)
		badReq(c, "not multiform")
		return
	}
	if log.IsLevelEnabled(log.DebugLevel) {
		for k := range form.File {
			log.Debugln("form field", k)
		}
		for k := range form.Value {
			log.Debugln("form value", k)
		}
	}

	emailClient := email.Builder{
		Subject: form.Value["subject"][0],
		ReplyTo: form.Value["reply-to"][0],
		From:    form.Value["from"][0],
		To:      form.Value["to"][0],
		Body:    stripAds(form.Value["html"][0]),
	}

	for _, file := range form.File["attachment"] {
		f, err := file.Open()
		if err != nil {
			log.Error(handlerLog, err)
			badReq(c, "cant open attachment")
			return
		}
		defer f.Close()

		emailClient.AddFile(file.Filename, f, file.Header.Get("Content-Type"))
	}
	err = emailClient.Send(app.cfg.SMTPConfig)
	if err != nil {
		log.Error(handlerLog, err)
		internalError(c, "cant send email")
		return
	}
	c.Status(http.StatusOK)
}
func (app *App) listDocuments(c *gin.Context) {

	uid := c.GetString(userIDKey)
	withBlob, _ := strconv.ParseBool(c.Query("withBlob"))
	docID := c.Query("doc")
	log.Debug(handlerLog, "params: withBlob: ", withBlob, ", DocId: ", docID)
	result := []*messages.RawMetadata{}

	var err error
	if docID != "" {
		//load single document
		var doc *messages.RawMetadata
		doc, err = app.metaStorer.GetMetadata(uid, docID)
		if err == nil {
			result = append(result, doc)
		}
	} else {
		//load all
		result, err = app.metaStorer.GetAllMetadata(uid)
	}

	if err != nil {
		log.Error(err)
		internalError(c, "cant get metadata")
		return
	}

	for _, response := range result {
		if withBlob {
			storageURL, exp, err := app.docStorer.GetStorageURL(uid, response.ID)
			if err != nil {
				response.Success = false
				log.Warn("Cant get storage url for : ", response.ID)
				continue
			}
			response.BlobURLGet = storageURL
			response.BlobURLGetExpires = exp.UTC().Format(time.RFC3339Nano)
		} else {
			response.BlobURLGetExpires = time.Time{}.Format(time.RFC3339Nano)
		}
		response.Success = true
	}

	c.JSON(http.StatusOK, result)
}
func (app *App) deleteDocument(c *gin.Context) {
	uid := c.GetString(userIDKey)
	deviceID := c.GetString(deviceIDKey)

	var req []messages.IDRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn(err)
		badReq(c, err.Error())
		return
	}

	result := []messages.StatusResponse{}
	for _, r := range req {
		doc, err := app.metaStorer.GetMetadata(uid, r.ID)
		ok := false
		if err == nil {
			err := app.docStorer.RemoveDocument(uid, r.ID)
			if err != nil {
				log.Error(err)
			} else {
				ok = true

				ntf := hub.DocumentNotification{
					ID:      doc.ID,
					Type:    doc.Type,
					Version: doc.Version,
					Parent:  doc.Parent,
					Name:    doc.VissibleName,
				}
				app.hub.Notify(uid, deviceID, ntf, hub.DocDeletedEvent)
			}
		}
		result = append(result, messages.StatusResponse{ID: r.ID, Success: ok})
	}

	c.JSON(http.StatusOK, result)
}
func (app *App) updateStatus(c *gin.Context) {
	uid := c.GetString(userIDKey)
	deviceID := c.GetString(deviceIDKey)
	var req []messages.RawMetadata

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}
	result := []messages.StatusResponse{}
	for _, doc := range req {
		log.Info("Id: ", doc.ID, " Name: ", doc.VissibleName)

		message := ""

		ok := false
		err := app.metaStorer.UpdateMetadata(uid, &doc)
		if err != nil {
			message = internalErrorMessage
			log.Error(err)
		} else {
			ok = true

			ntf := hub.DocumentNotification{
				ID:      doc.ID,
				Type:    doc.Type,
				Version: doc.Version,
				Parent:  doc.Parent,
				Name:    doc.VissibleName,
			}

			app.hub.Notify(uid, deviceID, ntf, hub.DocAddedEvent)
		}
		result = append(result, messages.StatusResponse{ID: doc.ID, Success: ok, Message: message, Version: doc.Version})
	}

	c.JSON(http.StatusOK, result)
}

func (app *App) locateService(c *gin.Context) {
	svc := c.Param("service")
	log.Infof("Requested: %s\n", svc)
	host := config.DefaultHost
	if svc == "blob-storage" {
		host = "https://" + config.DefaultHost
	}
	response := messages.HostResponse{Host: host, Status: "OK"}
	c.JSON(http.StatusOK, response)
}
func (app *App) syncComplete(c *gin.Context) {
	log.Info("Sync complete")
	uid := c.GetString(userIDKey)
	deviceID := c.GetString(deviceIDKey)

	var res messages.SyncCompleted
	res.ID = app.hub.NotifySync(uid, deviceID)
	c.JSON(http.StatusOK, res)
}

func formatExpires(t time.Time) string {
	return strconv.FormatInt(t.Unix(), 10)
}

func (app *App) blobStorageDownload(c *gin.Context) {
	uid := c.GetString(userIDKey)
	var req messages.BlobStorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}
	if req.RelativePath == "" {
		badReq(c, "no rel")
		return
	}

	url, exp, err := app.blobStorer.GetBlobURL(uid, req.RelativePath)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	response := messages.BlobStorageResponse{
		Method:       http.MethodGet,
		RelativePath: req.RelativePath,
		URL:          url,
		Expires:      formatExpires(exp),
	}
	c.JSON(http.StatusOK, response)
}

func (app *App) blobStorageUpload(c *gin.Context) {
	var req messages.BlobStorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}
	if req.RelativePath == "" {
		badReq(c, "no rel")
		return
	}
	if req.Initial {
		log.Info("--- Initial Sync ---")
	}
	uid := c.GetString(userIDKey)
	url, exp, err := app.blobStorer.GetBlobURL(uid, req.RelativePath)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	response := messages.BlobStorageResponse{
		Method:       http.MethodPut,
		RelativePath: req.RelativePath,
		URL:          url,
		Expires:      formatExpires(exp),
	}
	c.JSON(http.StatusOK, response)
}

func (app *App) integrations(c *gin.Context) {
	// uid := c.GetString(userIDKey)
	var res messages.IntegrationsResponse
	c.JSON(http.StatusOK, &res)
}
func (app *App) uploadRequest(c *gin.Context) {
	uid := c.GetString(userIDKey)
	var req []messages.UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}

	response := []messages.UploadResponse{}

	for _, r := range req {
		documentID := r.ID
		if documentID == "" {
			badReq(c, "no id")
		}
		url, exp, err := app.docStorer.GetStorageURL(uid, documentID)
		if err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		log.Debugln("StorageUrl: ", url)
		dr := messages.UploadResponse{
			BlobURLPut:        url,
			BlobURLPutExpires: exp.UTC().Format(time.RFC3339Nano),
			ID:                documentID,
			Success:           true,
			Version:           r.Version,
		}
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
	uid := c.GetString(userIDKey)
	deviceID := c.GetString(deviceIDKey)

	log.Info("connecting websocket from: ", uid)

	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	connection, err := upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		log.Warn("can't upgrade websocket to ws ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	go app.hub.ConnectWs(uid, deviceID, connection)
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
