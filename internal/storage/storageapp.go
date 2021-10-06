package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	tokenParam            = "token"
	GenerationHeader      = "x-goog-generation"
	GenerationMatchHeader = "x-goog-if-generation-match"

	ParamUID       = "uid"
	ParamBlobID    = "blobid"
	ParamExp       = "exp"
	ParamSignature = "signature"
	RouteBlob      = "/blobstorage"
	Storage        = "/storage"
)

// App file system document storage
type App struct {
	cfg  *config.Config
	fs   DocumentStorer
	blob BlobStorage
	// h   *hub.Hub
}

// NewApp StorageApp various storage routes
func NewApp(cfg *config.Config, fs DocumentStorer, blob BlobStorage) *App {
	staticWrapper := App{
		fs:   fs,
		blob: blob,
		cfg:  cfg,
	}
	return &staticWrapper
}

// RegisterRoutes blah
func (app *App) RegisterRoutes(router *gin.Engine) {

	router.GET(Storage+"/:"+tokenParam, app.downloadDocument)
	router.PUT(Storage+"/:"+tokenParam, app.uploadDocument)

	//sync15
	router.GET(RouteBlob, app.downloadBlob)
	router.PUT(RouteBlob, app.uploadBlob)
}

func (app *App) parseToken(token string) (*common.StorageClaim, error) {
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

func (app *App) uploadDocument(c *gin.Context) {
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
func (app *App) downloadDocument(c *gin.Context) {
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

var nameSeparators = regexp.MustCompile(`[./\\]`)

func sanitized(param string, c *gin.Context) string {
	p := c.Query(param)
	return nameSeparators.ReplaceAllString(p, "")
}
func (app *App) downloadBlob(c *gin.Context) {
	uid := sanitized(ParamUID, c)
	blobID := sanitized(ParamBlobID, c)
	exp := sanitized(ParamExp, c)
	signature := sanitized(ParamSignature, c)

	err := VerifySignature([]string{uid, blobID, exp}, exp, signature, app.cfg.JWTSecretKey)
	if err != nil {
		log.Warn(err)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if blobID == "" {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	log.Info("Requestng blob: ", blobID)

	reader, generation, err := app.blob.LoadBlob(uid, blobID)
	if err != nil {
		if err == ErrorNotFound {
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

func (app *App) uploadBlob(c *gin.Context) {
	uid := sanitized(ParamUID, c)
	blobID := sanitized(ParamBlobID, c)
	exp := sanitized(ParamExp, c)
	signature := sanitized(ParamSignature, c)

	err := VerifySignature([]string{uid, blobID, exp}, exp, signature, app.cfg.JWTSecretKey)
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
	}
	log.Info(exp, signature)

	if blobID == "" {
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

	newgen, err := app.blob.StoreBlob(uid, blobID, body, generation)

	if err != nil {
		if err == ErrorWrongGeneration {
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
	for i, s := range parts {
		if s == "" {
			return "", fmt.Errorf("index %d is empty", i)
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
