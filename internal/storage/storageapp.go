package storage

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const tokenParam = "token"

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

const root = "root"

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

	reader, err := fs.fs.LoadBlob(token.UserID, id)
	//TODO: empty root
	if id == root {
		if t, ok := err.(*os.PathError); ok && id == root {
			log.Warn(t.Err)
			c.Status(http.StatusOK)
			return
		}
		gen := fs.fs.RootGen(uid)

		log.Info("Root gen: ", gen)

		//TODO: read generation
		c.Header("x-goog-generation", strconv.Itoa(gen))
	}

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer reader.Close()
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
	id := token.DocumentID
	body := c.Request.Body
	defer body.Close()
	gen := 0
	if id == root {
		gh := c.Request.Header.Get("x-goog-if-generation-match")
		log.Warn("Client sent generation:", gh)
		gen, err = strconv.Atoi(gh)
		if err != nil {
			log.Warn(err)
		}
		lastgen := fs.fs.RootGen(uid)
		if gen != lastgen {
			log.Warn("Wrong generation")
			c.AbortWithStatus(http.StatusPreconditionFailed)
			return
		}
	}

	err = fs.fs.StoreBlob(token.UserID, id, body)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if id == root {
		fs.h.Notify(uid, "TODO", nil, hub.SyncUpdated)
	}

	c.JSON(http.StatusOK, gin.H{})

}
