package fs

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/gin-gonic/gin"
)

// Storage file system document storage
type Storage struct {
	Cfg *config.Config
}

func (fs *Storage) getUserPath(uid string) string {

	return filepath.Join(fs.Cfg.DataDir, filepath.Base(userDir), filepath.Base(uid))
}
func (fs *Storage) getPathFromUser(uid, path string) string {
	return filepath.Join(fs.getUserPath(uid), filepath.Base(path))
}

const tokenParam = "token"

// GetDocument Opens a document by id
func (fs *Storage) GetDocument(uid, id string) (io.ReadCloser, error) {
	fullPath := fs.getPathFromUser(uid, id+".zip")
	log.Debugln("Fullpath:", fullPath)
	reader, err := os.Open(fullPath)
	return reader, err
}

// RemoveDocument removes document (moves it to trash)
func (fs *Storage) RemoveDocument(uid, id string) error {

	trashDir := fs.getPathFromUser(uid, config.DefaultTrashDir)
	err := os.MkdirAll(trashDir, 0700)
	if err != nil {
		return err
	}
	//do not delete, move to trash
	log.Info(trashDir)
	meta := filepath.Base(fmt.Sprintf("%s.metadata", id))
	fullPath := fs.getPathFromUser(uid, meta)
	err = os.Rename(fullPath, path.Join(trashDir, meta))
	if err != nil {
		return err
	}
	zipfile := filepath.Base(fmt.Sprintf("%s.zip", id))
	fullPath = fs.getPathFromUser(uid, zipfile)
	err = os.Rename(fullPath, path.Join(trashDir, zipfile))
	if err != nil {
		return err
	}
	return nil
}

// GetStorageURL return a url for a file to store
func (fs *Storage) GetStorageURL(uid string, exp time.Time, id string) (string, error) {
	uploadRL := fs.Cfg.StorageURL
	log.Debugln("uploadUrl: ", uploadRL)
	claim := &common.StorageClaim{
		DocumentID: id,
		UserID:     uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: exp.Unix(),
			Audience:  common.StorageUsage,
		},
	}
	signedToken, err := common.SignClaims(claim, fs.Cfg.JWTSecretKey)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/storage/%s", uploadRL, url.QueryEscape(signedToken)), nil
}

// StoreDocument stores a document
func (fs *Storage) StoreDocument(uid string, stream io.ReadCloser, id string) error {
	fullPath := fs.getPathFromUser(uid, fmt.Sprintf("%s.zip", id))
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, stream)
	return err
}

func (fs *Storage) parseToken(token string) (*common.StorageClaim, error) {
	claim := &common.StorageClaim{}
	err := common.ClaimsFromToken(claim, token, fs.Cfg.JWTSecretKey)
	if err != nil {
		return nil, err
	}
	if claim.StandardClaims.Audience != common.StorageUsage {
		return nil, errors.New("not a storage token")
	}
	return claim, nil
}

func (fs *Storage) uploadDocument(c *gin.Context) {
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

	err = fs.StoreDocument(token.UserID, body, id)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
func (fs *Storage) downloadDocument(c *gin.Context) {
	strToken := c.Param(tokenParam)
	token, err := fs.parseToken(strToken)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	id := token.DocumentID

	//todo: storage provider
	log.Printf("Requestng Id: %s\n", id)

	reader, err := fs.GetDocument(token.UserID, id)
	defer reader.Close()

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

// RegisterRoutes blah
func (fs *Storage) RegisterRoutes(router *gin.Engine) {

	router.GET("/storage/:"+tokenParam, fs.downloadDocument)
	router.PUT("/storage/:"+tokenParam, fs.uploadDocument)
}
