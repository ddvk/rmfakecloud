package fs

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/blend/go-sdk/jwt"
	log "github.com/sirupsen/logrus"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/gin-gonic/gin"
)

// Storage file system document storage
type Storage struct {
	Cfg config.Config
}

func (fs *Storage) getSanitizedFileName(path string) string {
	return filepath.Join(fs.Cfg.DataDir, filepath.Base(path))
}

// GetDocument Opens a document by id
func (fs *Storage) GetDocument(uid, id string) (io.ReadCloser, error) {
	fullPath := fs.getSanitizedFileName(id + ".zip")
	log.Debugln("Fullpath:", fullPath)
	reader, err := os.Open(fullPath)
	return reader, err
}

// UpdateMetadata updates the metadata of a document
func (fs *Storage) UpdateMetadata(uid string, r *messages.RawDocument) error {
	filepath := fs.getSanitizedFileName(r.Id + ".metadata")

	js, err := json.Marshal(r)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath, js, 0600)
	return err

}

// RemoveDocument removes document (moves it to trash)
func (fs *Storage) RemoveDocument(uid, id string) error {
	//do not delete, move to trash
	trashDir := fs.Cfg.TrashDir
	meta := filepath.Base(fmt.Sprintf("%s.metadata", id))
	fullPath := fs.getSanitizedFileName(meta)
	err := os.Rename(fullPath, path.Join(trashDir, meta))
	if err != nil {
		return err
	}
	zipfile := filepath.Base(fmt.Sprintf("%s.zip", id))
	fullPath = fs.getSanitizedFileName(zipfile)
	err = os.Rename(fullPath, path.Join(trashDir, zipfile))
	if err != nil {
		return err
	}
	return nil
}

// GetStorageURL return a url for a file to store
func (fs *Storage) GetStorageURL(uid string, exp time.Time, id string) (string, error) {
	uploadRL := fs.Cfg.StorageURL
	log.Debugln("url", uploadRL)
	claim := &storageToken{
		DocumentId: id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: exp.Unix(),
			ID:        uid,
		},
	}
	signedToken, err := common.SignToken(claim, fs.Cfg.JWTSecretKey)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/storage/%s", uploadRL, url.QueryEscape(signedToken)), nil
}

// StoreDocument stores a document
func (fs *Storage) StoreDocument(uid string, stream io.ReadCloser, id string) error {
	dataDir := fs.Cfg.DataDir
	fullPath := path.Join(dataDir, filepath.Base(fmt.Sprintf("%s.zip", id)))
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()
	io.Copy(file, stream)
	return nil
}

// GetAllMetadata load all metadata
func (fs *Storage) GetAllMetadata(uid string, withBlob bool) (result []*messages.RawDocument, err error) {
	files, err := ioutil.ReadDir(fs.Cfg.DataDir)
	if err != nil {
		return
	}
	result = []*messages.RawDocument{}

	for _, f := range files {
		ext := filepath.Ext(f.Name())
		id := strings.TrimSuffix(f.Name(), ext)
		if ext != ".metadata" {
			continue
		}
		doc, err := fs.GetMetadata(uid, id, withBlob)
		if err != nil {
			log.Error(err)
			continue
		}

		result = append(result, doc)
	}
	return
}

// GetMetadata loads a document's metadata
func (fs *Storage) GetMetadata(uid, id string, withBlob bool) (*messages.RawDocument, error) {
	dataDir := fs.Cfg.DataDir
	filePath := id + ".metadata"
	fullPath := path.Join(dataDir, filepath.Base(filePath))
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	response := messages.RawDocument{}
	err = json.Unmarshal(content, &response)
	if err != nil {
		return nil, err
	}

	response.Success = true

	if withBlob {
		exp := time.Now().Add(time.Second * 60)
		storageURL, err := fs.GetStorageURL(uid, exp, response.Id)
		if err != nil {
			return nil, err
		}
		response.BlobURLGet = storageURL
		response.BlobURLGetExpires = exp.UTC().Format(time.RFC3339Nano)
	} else {
		response.BlobURLGetExpires = time.Time{}.Format(time.RFC3339Nano)

	}

	//fix time to utc
	tt, err := time.Parse(time.RFC3339, response.ModifiedClient)
	if err != nil {
		log.Errorln("cant parse time", err)
		tt = time.Now()
	}
	response.ModifiedClient = tt.UTC().Format(time.RFC3339)

	return &response, nil

}

const (
	userDir     = "users"
	profileName = ".userprofile"
)

// GetUser blah
func (fs *Storage) GetUser(id string) (response *model.User, err error) {
	dataDir := fs.Cfg.DataDir
	fullPath := path.Join(dataDir, userDir, id, profileName)

	var f *os.File
	f, err = os.Open(fullPath)
	if err != nil {
		return
	}
	defer f.Close()

	var content []byte
	content, err = ioutil.ReadAll(f)
	if err != nil {
		return
	}

	response = &model.User{}
	err = json.Unmarshal(content, response)
	if err != nil {
		return
	}

	return
}

// GetUsers blah
func (fs *Storage) GetUsers() (users []*model.User, err error) {
	dataDir := path.Join(fs.Cfg.DataDir, userDir)

	entries, err := ioutil.ReadDir(dataDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if user, err := fs.GetUser(entry.Name()); err == nil {
				users = append(users, user)
			}
		}
	}
	return
}

// RegisterUser blah
func (fs *Storage) RegisterUser(u *model.User) (err error) {
	userDir := path.Join(fs.Cfg.DataDir, userDir, u.Id)
	profilePath := path.Join(userDir, profileName)

	// Create the user's directory
	err = os.MkdirAll(userDir, 0700)
	if err != nil {
		return
	}

	// Create the profile file
	var js []byte
	js, err = json.Marshal(u)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(profilePath, os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(js)
	if err != nil {
		return err
	}

	return
}

func (fs *Storage) UpdateUser(u *model.User) (err error) {
	userDir := path.Join(fs.Cfg.DataDir, u.Id)
	profilePath := path.Join(userDir, profileName)

	// Erase the profile file
	var js []byte
	js, err = json.Marshal(u)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(profilePath, js, 0600)

	return
}

type storageToken struct {
	DocumentId string `json:"documentId"`
	jwt.StandardClaims
}

func (st *storageToken) Valid() error {
	return st.StandardClaims.Valid()
}

func parseToken(strToken string) (token *storageToken, err error) {
	return &storageToken{}, nil

}

func (fs *Storage) uploadDocument(c *gin.Context) {
	strToken := c.Query("token")
	token, err := parseToken(strToken)

	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	id := token.DocumentId
	log.Printf("Uploading id %s\n", id)
	body := c.Request.Body
	defer body.Close()

	err = fs.StoreDocument(token.StandardClaims.ID, body, id)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
func (fs *Storage) downloadDocument(c *gin.Context) {
	strToken := c.Query("token")
	token, err := parseToken(strToken)

	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	id := token.DocumentId

	//todo: storage provider
	log.Printf("Requestng Id: %s\n", id)

	reader, err := fs.GetDocument(token.StandardClaims.ID, id)
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

	router.GET("/storage/:token", fs.downloadDocument)
	router.PUT("/storage/:token", fs.uploadDocument)
}
