package app

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/email"
	"github.com/ddvk/rmfakecloud/internal/hwr"
	"github.com/ddvk/rmfakecloud/internal/integrations"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	internalErrorMessage = "Internal Error"
	handlerLog           = "[handler] "
	// a way to invalidate the user token
	tokenVersion   = 10
	maxRequestSize = 7000000000
)

func (app *App) getDeviceClaims(c *gin.Context) (*DeviceClaims, error) {
	token, err := common.GetToken(c)
	if err != nil {
		return nil, err
	}
	claims := &DeviceClaims{}
	err = common.ClaimsFromToken(claims, token, app.cfg.JWTSecretKey)
	if err != nil {
		return nil, err
	}
	if claims.UserID == "" {
		return nil, fmt.Errorf("wrong token, missing userid")
	}
	return claims, nil
}

func (app *App) getUserClaims(c *gin.Context) (*UserClaims, error) {
	token, err := common.GetToken(c)
	// log.Debug(handlerLog, "Token: ", token)
	if err != nil {
		return nil, err
	}
	claims := &UserClaims{}
	err = common.ClaimsFromToken(claims, token, app.cfg.JWTSecretKey)
	if err != nil {
		return nil, err
	}
	if claims.Profile.UserID == "" {
		return nil, fmt.Errorf("wrong token, missing userid")
	}
	if claims.Version != tokenVersion {
		return nil, fmt.Errorf("wrong token version, something has changed")
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
	claims := &DeviceClaims{
		DeviceDesc: tokenRequest.DeviceDesc,
		DeviceID:   tokenRequest.DeviceID,
		UserID:     uid,
		StandardClaims: jwt.StandardClaims{
			Audience: APIUsage,
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

	scopes := []string{"intgr", "screenshare", "hwcmail:-1", "mail:-1"}

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
	claims := &UserClaims{
		Profile: Auth0profile{
			UserID:        deviceToken.UserID,
			IsSocial:      false,
			Connection:    "Username-Password-Authentication",
			Name:          user.Name,
			Nickname:      user.Nickname,
			Email:         fmt.Sprintf("%s (via %s)", user.Email, app.cfg.StorageURL),
			EmailVerified: true,
			Picture:       "image.png",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		DeviceDesc: deviceToken.DeviceDesc,
		DeviceID:   deviceToken.DeviceID,
		Scopes:     scopesStr,
		Level:      "connect",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			NotBefore: now.Unix(),
			IssuedAt:  now.Unix(),
			Subject:   "rM User Token",
			Issuer:    "rM WebApp",
			Id:        user.ID,
			Audience:  APIUsage,
		},
		Version: tokenVersion,
	}

	tokenString, err := common.SignClaims(claims, app.cfg.JWTSecretKey)
	if err != nil {
		badReq(c, err.Error())
		return
	}
	c.String(http.StatusOK, tokenString)
}

type metapayload struct {
	FileName string `json:"file_name"`
}

func extFromContentType(contentType string) (string, error) {
	switch contentType {

	case "application/epub+zip":
		return storage.EpubFileExt, nil
	case "application/pdf":
		return storage.PdfFileExt, nil
	}
	return "", fmt.Errorf("unsupported content type %s", contentType)
}

func (app *App) uploadDoc(c *gin.Context) {
	uid := c.GetString(userIDKey)
	deviceID := c.GetString(deviceIDKey)
	syncVer := getSyncVersion(c)

	log.Info("uploading file for: ", uid)

	form, err := c.MultipartForm()
	if err != nil {
		log.Error(err)
		badReq(c, "not multiform")
		return
	}

	meta := form.Value["meta"][0]
	if meta == "" {
		log.Warn(handlerLog, " missing 'meta'")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	m := metapayload{}
	err = json.Unmarshal([]byte(meta), &m)
	if err != nil {
		log.Warn(handlerLog, " meta not json")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if len(form.File["file"]) < 1 {
		log.Warn(handlerLog, " missing 'file'")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	file := form.File["file"][0]
	if file == nil {
		log.Error(handlerLog, "no files")
		badReq(c, "mising file")
		return
	}
	contentType := file.Header.Get("Content-Type")
	ext, err := extFromContentType(contentType)
	if err != nil {
		log.Error(handlerLog, err)
		badReq(c, "unsupported content type")
		return
	}

	f, err := file.Open()
	if err != nil {
		log.Error(handlerLog, err)
		badReq(c, "cant open attachment")
		return
	}
	defer f.Close()

	if err != nil {
		log.Error(handlerLog, err)
		internalError(c, "cant upload document")
		return
	}
	fileName := m.FileName + ext
	log.Info("Uploading: ", fileName)

	err = saveUpload(app, syncVer, uid, deviceID, fileName, f)

	if err != nil {
		log.Error(handlerLog, err)
		internalError(c, "can't upload")
		return
	}
	c.Status(http.StatusOK)
}

func getSyncVersion(c *gin.Context) common.SyncVersion {
	syncVer, ok := c.Get(syncVersionKey)
	if !ok {
		panic("should have a sync version")
	}
	return syncVer.(common.SyncVersion)
}

// new read on rm api
func (app *App) uploadDocV2(c *gin.Context) {
	uid := c.GetString(userIDKey)
	deviceID := c.GetString(deviceIDKey)
	log.Info("uploading file for: ", uid)
	syncVer := getSyncVersion(c)

	metaHeader := c.Request.Header["Rm-Meta"]
	if len(metaHeader) < 1 {
		log.Warn(handlerLog, "missing 'meta' header")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	meta := metaHeader[0]
	if meta == "" {
		log.Warn(handlerLog, "empty 'meta' header")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	metaJSON, err := base64.StdEncoding.DecodeString(meta)
	if err != nil {
		log.Warn(handlerLog, "meta not base64 encoded ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	m := metapayload{}
	err = json.Unmarshal(metaJSON, &m)
	if err != nil {
		log.Warn(handlerLog, "meta not json ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	contentType := c.Request.Header["Content-Type"]
	if len(contentType) < 1 {
		log.Warn(handlerLog, "missing content-type")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	ext, err := extFromContentType(contentType[0])
	if err != nil {
		log.Error(handlerLog, err)
		badReq(c, "unsupported content type")
		return
	}

	f := c.Request.Body

	fileName := m.FileName + ext
	log.Info("Uploading: ", fileName)

	err = saveUpload(app, syncVer, uid, deviceID, fileName, f)
	if err != nil {
		log.Error(handlerLog, err)
		internalError(c, "can't upload")
		return
	}

	c.Status(http.StatusOK)
}

func saveUpload(app *App, syncVer common.SyncVersion, uid, deviceID, fileName string, f io.ReadCloser) error {
	//HACK:
	if syncVer == common.Sync15 {
		log.Info("sync 15 upload")
		_, err := app.blobStorer.CreateBlobDocument(uid, fileName, "", f)
		if err != nil {
			return err
		}
		app.hub.NotifySync(uid, deviceID)
	} else {
		log.Info("sync 10 upload")
		d, err := app.docStorer.CreateDocument(uid, fileName, "", f)
		if err != nil {
			return err
		}
		ntf := hub.DocumentNotification{
			Parent:  "",
			ID:      d.ID,
			Type:    d.Type,
			Name:    d.Name,
			Version: 1,
		}
		app.hub.Notify(uid, deviceID, ntf, messages.DocAddedEvent)
	}
	return nil
}

type emailForm struct {
	To         string                  `form:"to"`
	From       string                  `form:"from"`
	Subject    string                  `form:"subject"`
	Body       string                  `form:"html"`
	Attachment []*multipart.FileHeader `form:"attachment"`
}

func (app *App) sendEmail(c *gin.Context) {
	uid := c.GetString(userIDKey)
	log.Info("Sending mail for: ", uid)

	if app.cfg.SMTPConfig == nil {
		log.Error("smtp not configured")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var req emailForm

	if err := c.Bind(&req); err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if log.IsLevelEnabled(log.DebugLevel) {
		log.Debug("from: ", req.From)
		log.Debug("to: ", req.To)
		log.Debug("body: ", req.Body)
		for i, a := range req.Attachment {
			log.Debug(" Attachment: ", i)
			log.Debug(" FileName: ", a.Filename)
			log.Debug(" FileHeader: ", a.Header)
			log.Debug(" FileSize: ", a.Size)
		}
	}

	var from *mail.Address
	if app.cfg.SMTPConfig.FromOverride != nil {
		from = app.cfg.SMTPConfig.FromOverride
	} else {
		// try to use the user's email address if in the correct format
		if user, err := app.userStorer.GetUser(uid); err == nil && user.Email != "" {
			from, err = mail.ParseAddress(user.Email)
			if err != nil {
				log.Warn(handlerLog, "user: ", uid, " has invalid email address: ", user.Email)
			} else {
				log.Debug("using user's email address")
			}
		}

		// fallback FROM the request from the tablet
		if from == nil {
			var err error
			from, err = mail.ParseAddress(req.From)
			if err != nil {
				log.Warn(handlerLog, err)
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			log.Debug("using from, from the request")
		}
	}
	//parse TO addresses
	to, err := mail.ParseAddressList(email.TrimAddresses(req.To))
	if err != nil {
		log.Warn(handlerLog, err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	emailClient := email.Builder{
		Subject: req.Subject,
		To:      to,
		From:    from,
		Body:    stripAds(req.Body),
	}

	for _, file := range req.Attachment {
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
	docID := common.QueryS("doc", c)
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
				app.hub.Notify(uid, deviceID, ntf, messages.DocDeletedEvent)
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

			app.hub.Notify(uid, deviceID, ntf, messages.DocAddedEvent)
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

func (app *App) syncCompleteV2(c *gin.Context) {
	log.Info("Sync completeV2")
	uid := c.GetString(userIDKey)
	deviceID := c.GetString(deviceIDKey)

	var req messages.SyncCompletedRequestV2
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		badReq(c, err.Error())
		return
	}
	log.Info("got sync completed, gen: ", req.Generation)

	notificationID := app.hub.NotifySync(uid, deviceID)

	res := messages.SyncCompleted{
		ID: notificationID,
	}
	c.JSON(http.StatusOK, res)
}
func formatExpires(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
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

	url, exp, err := app.blobStorer.GetBlobURL(uid, req.RelativePath, false)
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
	url, exp, err := app.blobStorer.GetBlobURL(uid, req.RelativePath, true)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	response := messages.BlobStorageResponse{
		Method:         http.MethodPut,
		RelativePath:   req.RelativePath,
		URL:            url,
		Expires:        formatExpires(exp),
		MaxRequestSize: maxRequestSize,
	}

	c.JSON(http.StatusOK, response)
}

func (app *App) syncUpdateRootV3(c *gin.Context) {
	var rootv3 messages.SyncRootV3
	err := json.NewDecoder(c.Request.Body).Decode(&rootv3)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	uid := c.GetString(userIDKey)
	newgeneration, err := app.blobStorer.StoreBlob(uid, "root", bytes.NewBufferString(rootv3.Hash), rootv3.Generation)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, messages.SyncRootV3{
		Generation: newgeneration,
		Hash:       rootv3.Hash,
	})
}

func (app *App) syncGetRootV3(c *gin.Context) {
	uid := c.GetString(userIDKey)

	reader, generation, _, err := app.blobStorer.LoadBlob(uid, "root")
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	roothash, err := io.ReadAll(reader)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, messages.SyncRootV3{
		Generation: generation,
		Hash:       string(roothash),
	})
}

func (app *App) integrationsGetMetadata(c *gin.Context) {
	var metadata messages.IntegrationMetadata
	metadata.Thumbnail = ""
	c.JSON(http.StatusOK, &metadata)
}

func (app *App) integrationsUpload(c *gin.Context) {
	log.Info("uploading...")
	uid := c.GetString(userIDKey)
	integrationID := common.ParamS(integrationKey, c)
	folderID := common.ParamS(folderKey, c)
	name := common.QueryS("name", c)
	fileType := common.QueryS("fileType", c)

	integrationProvider, err := integrations.GetIntegrationProvider(app.userStorer, uid, integrationID)

	if err != nil {
		log.Error(fmt.Errorf("can't get integration, %v", err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	body := c.Request.Body
	id, err := integrationProvider.Upload(folderID, name, fileType, body)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (app *App) integrationsGetFile(c *gin.Context) {
	uid := c.GetString(userIDKey)
	integrationID := common.ParamS(integrationKey, c)
	fileID := common.ParamS(fileKey, c)

	integrationProvider, err := integrations.GetIntegrationProvider(app.userStorer, uid, integrationID)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	reader, err := integrationProvider.Download(fileID)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	defer reader.Close()

	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}
func (app *App) integrationsList(c *gin.Context) {
	uid := c.GetString(userIDKey)
	integrationID := common.ParamS(integrationKey, c)
	folder := common.ParamS(folderKey, c)
	folderDepthStr := c.Query("folderDepth")
	folderDepth := 1
	if folderDepthStr != "" {
		folderDepth, _ = strconv.Atoi(folderDepthStr)
	}

	integrationProvider, err := integrations.GetIntegrationProvider(app.userStorer, uid, integrationID)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	response, err := integrationProvider.List(folder, folderDepth)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}
func (app *App) integrations(c *gin.Context) {
	uid := c.GetString(userIDKey)

	response, err := integrations.List(app.userStorer, uid)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	c.JSON(http.StatusOK, response)
}
func (app *App) uploadRequest(c *gin.Context) {
	uid := c.GetString(userIDKey)
	var req []messages.UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("could not bind %v", err)
		c.AbortWithStatus(http.StatusBadRequest)
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
	response, err := app.hwrClient.SendRequest(body)
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

// / remove remarkable ads
func stripAds(msg string) string {
	br := "<br>--<br>"
	i := strings.Index(msg, br)
	if i > 0 {
		return msg[:i]
	}
	return msg
}
