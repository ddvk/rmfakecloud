package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	tokenParam            = "token"
	generationHeader      = "x-goog-generation"
	generationMatchHeader = "x-goog-if-generation-match"
	StorageUsage          = "storage"

	ParamUID       = "uid"
	ParamBlobID    = "blobid"
	ParamExp       = "exp"
	ParamSignature = "signature"
	ParamScope     = "scope"
	RouteBlob      = "/blobstorage"
	RouteStorage   = "/storage"

	RootBlob = "root"
)

// ErrorNotFound not found
var ErrorNotFound = errors.New("not found")

// ErrorWrongGeneration the geration did not match
var ErrorWrongGeneration = errors.New("wrong generation")

// App file system document storage
type App struct {
	cfg        *config.Config
	docStorer  DocumentStorer
	userStorer UserStorer
	blobStorer BlobStorage
}

// NewApp StorageApp various storage routes
func NewApp(cfg *config.Config, docStorer DocumentStorer, userStorer UserStorer, blobStorer BlobStorage) *App {
	staticWrapper := App{
		cfg:        cfg,
		docStorer:  docStorer,
		blobStorer: blobStorer,
	}
	return &staticWrapper
}

// RegisterRoutes blah
func (app *App) RegisterRoutes(router *gin.Engine) {
	router.GET(RouteStorage+"/:"+tokenParam, app.downloadDocument)
	router.PUT(RouteStorage+"/:"+tokenParam, app.uploadDocument)

	//sync15
	router.GET(RouteBlob, app.downloadBlob)
	router.PUT(RouteBlob, app.uploadBlob)
}

func (app *App) parseToken(token string) (*StorageClaim, error) {
	claim := &StorageClaim{}
	err := common.ClaimsFromToken(claim, token, app.cfg.JWTSecretKey)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(claim.Audience, StorageUsage) {
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

	err = app.docStorer.StoreDocument(token.UserID, id, body)
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

	reader, err := app.docStorer.GetDocument(token.UserID, id)

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
	uid := c.Query(ParamUID)

	blobID := common.QueryS(ParamBlobID, c)
	exp := common.QueryS(ParamExp, c)
	signature := common.QueryS(ParamSignature, c)
	scope := common.QueryS(ParamScope, c)

	err := VerifyURLParams([]string{uid, blobID, exp, scope}, exp, signature, app.cfg.JWTSecretKey)
	if err != nil {
		log.Warn(err)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if scope != ReadScope {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if blobID == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	log.Info("Requestng blob: ", blobID)

	if blobID == RootBlob {
		root, generation, err := app.userStorer.GetRoot(uid)
		if err != nil && err != ErrorNotFound {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		log.Debug("Sending gen for root: ", generation)
		c.Header(generationHeader, strconv.FormatInt(generation, 10))

		c.Data(http.StatusOK, "application/octet-stream", []byte(root))
	} else {
		reader, size, crc32c, err := app.blobStorer.LoadBlob(uid, blobID)
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

		common.AddHashHeader(c, crc32c)

		c.DataFromReader(http.StatusOK, size, "application/octet-stream", reader, nil)
	}
}

func (app *App) uploadBlob(c *gin.Context) {
	//not sanitized, email address etc
	uid := c.Query(ParamUID)

	blobID := common.QueryS(ParamBlobID, c)
	exp := common.QueryS(ParamExp, c)
	signature := common.QueryS(ParamSignature, c)
	scope := common.QueryS(ParamScope, c)

	err := VerifyURLParams([]string{uid, blobID, exp, scope}, exp, signature, app.cfg.JWTSecretKey)
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	log.Info(exp, signature)

	if blobID == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if scope != WriteScope {
		log.Warn("wrong scope: " + scope)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	body := c.Request.Body
	defer body.Close()

	var generation int64
	gh := c.Request.Header.Get(generationMatchHeader)
	if gh != "" {
		log.Info("Client sent generation:", gh)
		var err error
		generation, err = strconv.ParseInt(gh, 10, 64)
		if err != nil {
			log.Warn(err)
		}
	}

	if blobID == RootBlob {
		var newgen int64
		newgen, err = app.userStorer.UpdateRoot(uid, body, generation)
		if err == ErrorWrongGeneration {
			c.AbortWithStatus(http.StatusPreconditionFailed)
			return
		}

		c.Header(generationHeader, strconv.FormatInt(newgen, 10))
	} else {
		err = app.blobStorer.StoreBlob(uid, blobID, "", "", body)
	}

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

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
