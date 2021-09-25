package storage

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	tokenParam            = "token"
	GenerationHeader      = "x-goog-generation"
	GenerationMatchHeader = "x-goog-if-generation-match"
)

// Storage file system document storage
type StorageApp struct {
	cfg *config.Config
	fs  DocumentStorer
	h   *hub.Hub
}

// New Storage
func NewApp(cfg *config.Config, fs DocumentStorer,
	h *hub.Hub) *StorageApp {
	staticWrapper := StorageApp{
		fs:  fs,
		cfg: cfg,
		h:   h,
	}
	return &staticWrapper
}

// RegisterRoutes blah
func (fs *StorageApp) RegisterRoutes(router *gin.Engine) {

	router.GET("/storage/:"+tokenParam, fs.downloadDocument)
	router.PUT("/storage/:"+tokenParam, fs.uploadDocument)

	//sync15
	router.GET("/blobstorage/:"+tokenParam, fs.downloadBlob)
	router.PUT("/blobstorage/:"+tokenParam, fs.uploadBlob)
}

func (fs *StorageApp) parseToken(token string) (*common.StorageClaim, error) {
	claim := &common.StorageClaim{}
	err := common.ClaimsFromToken(claim, token, fs.cfg.JWTSecretKey)
	if err != nil {
		return nil, err
	}
	if claim.StandardClaims.Audience != common.StorageUsage {
		return nil, errors.New("not a storage token")
	}
	return claim, nil
}

func (fs *StorageApp) uploadDocument(c *gin.Context) {
	strToken := c.Param(tokenParam)
	log.Debug("[storage] uploading with token:", strToken)
	token, err := fs.parseToken(strToken)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	id := token.DocumentID
	log.Debug("[storage] uploading documentId: ", id)
	body := c.Request.Body
	defer body.Close()

	err = fs.fs.StoreDocument(token.UserID, id, body)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
func (fs *StorageApp) downloadDocument(c *gin.Context) {
	strToken := c.Param(tokenParam)
	token, err := fs.parseToken(strToken)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	id := token.DocumentID

	//todo: storage provider
	log.Info("Requestng Id: ", id)

	reader, err := fs.fs.GetDocument(token.UserID, id)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer reader.Close()
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

func (fs *StorageApp) downloadBlob(c *gin.Context) {
	strToken := c.Param(tokenParam)
	token, err := fs.parseToken(strToken)
	uid := token.UserID

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	id := token.DocumentID
	log.Info("Requestng blob Id: ", id)

	reader, generation, err := fs.fs.LoadBlob(uid, id)
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
	c.Header(GenerationHeader, strconv.Itoa(generation))
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

func (fs *StorageApp) uploadBlob(c *gin.Context) {
	strToken := c.Param(tokenParam)
	token, err := fs.parseToken(strToken)
	uid := token.UserID

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	blobId := token.DocumentID
	body := c.Request.Body
	defer body.Close()

	generation := 0
	gh := c.Request.Header.Get(GenerationMatchHeader)
	if gh != "" {
		log.Warn("Client sent generation:", gh)
		generation, err = strconv.Atoi(gh)
		if err != nil {
			log.Warn(err)
		}
	}

	newgen, err := fs.fs.StoreBlob(uid, blobId, body, generation)

	if err != nil {
		if err == ErrorWrongGeneration {
			c.AbortWithStatus(http.StatusPreconditionFailed)
			return
		}
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header(GenerationHeader, strconv.Itoa(newgen))
	c.JSON(http.StatusOK, gin.H{})

}
