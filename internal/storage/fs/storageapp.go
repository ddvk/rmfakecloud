package fs

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	tokenParam            = "token"
	GenerationHeader      = "x-goog-generation"
	GenerationMatchHeader = "x-goog-if-generation-match"

	paramUid       = "uid"
	paramBlobId    = "blobid"
	paramExp       = "exp"
	paramSignature = "signature"
	routeBlob      = "/blobstorage"
)

// Storage file system document storage
type StorageApp struct {
	cfg  *config.Config
	fs   storage.DocumentStorer
	blob storage.BlobStorage
	// h   *hub.Hub
}

// NewApp StorageApp various storage routes
func NewApp(cfg *config.Config, fs storage.DocumentStorer, blob storage.BlobStorage) *StorageApp {
	staticWrapper := StorageApp{
		fs:   fs,
		blob: blob,
		cfg:  cfg,
	}
	return &staticWrapper
}

// RegisterRoutes blah
func (fs *StorageApp) RegisterRoutes(router *gin.Engine) {

	router.GET("/storage/:"+tokenParam, fs.downloadDocument)
	router.PUT("/storage/:"+tokenParam, fs.uploadDocument)

	//sync15
	router.GET(routeBlob, fs.downloadBlob)
	router.PUT(routeBlob, fs.uploadBlob)
}

func (app *StorageApp) parseToken(token string) (*common.StorageClaim, error) {
	claim := &common.StorageClaim{}
	err := common.ClaimsFromToken(claim, token, app.cfg.JWTSecretKey)
	if err != nil {
		return nil, err
	}
	if claim.StandardClaims.Audience != common.StorageUsage {
		return nil, errors.New("not a storage token")
	}
	return claim, nil
}

func (app *StorageApp) uploadDocument(c *gin.Context) {
	strToken := c.Param(tokenParam)
	log.Debug("[storage] uploading with token:", strToken)
	token, err := app.parseToken(strToken)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	id := token.DocumentID
	log.Debug("[storage] uploading documentId: ", id)
	body := c.Request.Body
	defer body.Close()

	err = app.fs.StoreDocument(token.UserID, id, body)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
func (app *StorageApp) downloadDocument(c *gin.Context) {
	strToken := c.Param(tokenParam)
	token, err := app.parseToken(strToken)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	id := token.DocumentID

	//todo: storage provider
	log.Info("Requestng Id: ", id)

	reader, err := app.fs.GetDocument(token.UserID, id)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer reader.Close()
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

func (app *StorageApp) downloadBlob(c *gin.Context) {
	uid := c.Query(paramUid)
	blobId := c.Query(paramBlobId)
	exp := c.Query(paramExp)
	signature := c.Query(paramSignature)

	err := VerifySignature([]string{uid, blobId, exp}, exp, signature, app.cfg.JWTSecretKey)
	if err != nil {
		log.Warn(err)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if blobId == "" {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	log.Info("Requestng blob: ", blobId)

	reader, generation, err := app.blob.LoadBlob(uid, blobId)
	if err != nil {
		if err == storage.ErrorNotFound {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	log.Debug("Sending gen: ", generation)
	c.Header(GenerationHeader, strconv.FormatInt(generation, 10))
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

func (app *StorageApp) uploadBlob(c *gin.Context) {
	uid := c.Query(paramUid)
	blobId := c.Query(paramBlobId)
	exp := c.Query(paramExp)
	signature := c.Query(paramSignature)

	err := VerifySignature([]string{uid, blobId, exp}, exp, signature, app.cfg.JWTSecretKey)
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
	}
	log.Info(exp, signature)

	if blobId == "" {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	body := c.Request.Body
	defer body.Close()

	generation := int64(0)
	gh := c.Request.Header.Get(GenerationMatchHeader)
	if gh != "" {
		log.Info("Client sent generation:", gh)
		var err error
		generation, err = strconv.ParseInt(gh, 10, 64)
		if err != nil {
			log.Warn(err)
		}
	}

	newgen, err := app.blob.StoreBlob(uid, blobId, body, generation)

	if err != nil {
		if err == storage.ErrorWrongGeneration {
			c.AbortWithStatus(http.StatusPreconditionFailed)
			return
		}
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header(GenerationHeader, strconv.FormatInt(newgen, 10))
	c.JSON(http.StatusOK, gin.H{})

}
func Sign(parts []string, key []byte) (string, error) {
	h := hmac.New(sha256.New, key)
	for _, s := range parts {
		if s == "" {
			return "", errors.New("empty part")
		}
		h.Write([]byte(s))
	}
	hs := h.Sum(nil)
	s := hex.EncodeToString(hs)
	return s, nil
}

func VerifySignature(parts []string, exp, signature string, key []byte) error {
	expected, err := Sign(parts, key)
	if err != nil {
		return err
	}
	expiration, err := strconv.Atoi(exp)
	if err != nil {
		return err
	}
	if expiration < int(time.Now().Unix()) {
		return errors.New("expired")
	}

	if subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) != 1 {
		return errors.New("wrong signature")
	}

	return nil
}
