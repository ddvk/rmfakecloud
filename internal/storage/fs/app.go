package fs

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/zgs225/rmfakecloud/internal/common"
	"github.com/zgs225/rmfakecloud/internal/config"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	tokenParam            = "token"
	generationHeader      = "x-goog-generation"
	generationMatchHeader = "x-goog-if-generation-match"
	storageUsage          = "storage"

	paramUID       = "uid"
	paramBlobID    = "blobid"
	paramExp       = "exp"
	paramSignature = "signature"
	paramScope     = "scope"
	routeBlob      = "/blobstorage"
	routeStorage   = "/storage"
)

// ErrorNotFound not found
var ErrorNotFound = errors.New("not found")

// ErrorWrongGeneration the geration did not match
var ErrorWrongGeneration = errors.New("wrong generation")

// App file system document storage
type App struct {
	cfg *config.Config
	fs  *FileSystemStorage
}

// NewApp StorageApp various storage routes
func NewApp(cfg *config.Config, fs *FileSystemStorage) *App {
	staticWrapper := App{
		fs:  fs,
		cfg: cfg,
	}
	return &staticWrapper
}

// RegisterRoutes blah
func (app *App) RegisterRoutes(router *gin.Engine) {

	router.GET(routeStorage+"/:"+tokenParam, app.downloadDocument)
	router.PUT(routeStorage+"/:"+tokenParam, app.uploadDocument)

	//sync15
	router.GET(routeBlob, app.downloadBlob)
	router.PUT(routeBlob, app.uploadBlob)
}

func (app *App) parseToken(token string) (*StorageClaim, error) {
	claim := &StorageClaim{}
	err := common.ClaimsFromToken(claim, token, app.cfg.JWTSecretKey)
	if err != nil {
		return nil, err
	}
	if claim.StandardClaims.Audience != storageUsage {
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

func (app *App) downloadBlob(c *gin.Context) {
	//not sanitized, email address etc
	uid := c.Query(paramUID)

	blobID := common.QueryS(paramBlobID, c)
	exp := common.QueryS(paramExp, c)
	signature := common.QueryS(paramSignature, c)
	scope := common.QueryS(paramScope, c)

	err := VerifyURLParams([]string{uid, blobID, exp, scope}, exp, signature, app.cfg.JWTSecretKey)
	if err != nil {
		log.Warn(err)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if scope != "read" {
		c.AbortWithStatus(http.StatusForbidden)
	}

	if blobID == "" {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	log.Info("Requestng blob: ", blobID)

	reader, generation, size, err := app.fs.LoadBlob(uid, blobID)
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

	if blobID == "root" {
		log.Debug("Sending gen: ", generation)
	}
	c.Header(generationHeader, strconv.FormatInt(generation, 10))
	c.DataFromReader(http.StatusOK, size, "application/octet-stream", reader, nil)
}

func (app *App) uploadBlob(c *gin.Context) {
	//not sanitized, email address etc
	uid := c.Query(paramUID)

	blobID := common.QueryS(paramBlobID, c)
	exp := common.QueryS(paramExp, c)
	signature := common.QueryS(paramSignature, c)
	scope := common.QueryS(paramScope, c)

	err := VerifyURLParams([]string{uid, blobID, exp, scope}, exp, signature, app.cfg.JWTSecretKey)
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
	}
	log.Info(exp, signature)

	if blobID == "" {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	if scope != "write" {
		log.Warn("wrong scope: " + scope)
		c.AbortWithStatus(http.StatusForbidden)
	}

	body := c.Request.Body
	defer body.Close()

	generation := int64(0)
	gh := c.Request.Header.Get(generationMatchHeader)
	if gh != "" {
		log.Info("Client sent generation:", gh)
		var err error
		generation, err = strconv.ParseInt(gh, 10, 64)
		if err != nil {
			log.Warn(err)
		}
	}

	newgen, err := app.fs.StoreBlob(uid, blobID, body, generation)

	if err != nil {
		if err == ErrorWrongGeneration {
			c.AbortWithStatus(http.StatusPreconditionFailed)
			return
		}
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header(generationHeader, strconv.FormatInt(newgen, 10))
	c.JSON(http.StatusOK, gin.H{})
}

// SignURLParams signs url params
func SignURLParams(parts []string, key []byte) (string, error) {
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

// VerifyURLParams verify the signature and expiry
func VerifyURLParams(parts []string, exp, signature string, key []byte) error {
	expected, err := SignURLParams(parts, key)
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
