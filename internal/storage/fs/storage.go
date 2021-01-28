package fs

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/messages"
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
func (fs *Storage) GetDocument(id string) (io.ReadCloser, error) {
	fullPath := fs.getSanitizedFileName(id + ".zip")
	log.Debugln("Fullpath:", fullPath)
	reader, err := os.Open(fullPath)
	return reader, err
}

// UpdateMetadata updates the metadata of a document
func (fs *Storage) UpdateMetadata(r *messages.RawDocument) error {
	filepath := fs.getSanitizedFileName(r.Id + ".metadata")

	js, err := json.Marshal(r)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath, js, 0600)
	return err

}

// RemoveDocument removes document (moves it to trash)
func (fs *Storage) RemoveDocument(id string) error {
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
func (fs *Storage) GetStorageURL(id string) string {
	uploadRL := fs.Cfg.StorageURL
	log.Debugln("url", uploadRL)
	return fmt.Sprintf("%s/storage?id=%s", uploadRL, id)
}

// StoreDocument stores a document
func (fs *Storage) StoreDocument(stream io.ReadCloser, id string) error {
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
func (fs *Storage) GetAllMetadata(withBlob bool) (result []*messages.RawDocument, err error) {
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
		doc, err := fs.GetMetadata(id, withBlob)
		if err != nil {
			log.Error(err)
			continue
		}

		result = append(result, doc)
	}
	return
}

// GetMetadata loads a document's metadata
func (fs *Storage) GetMetadata(id string, withBlob bool) (*messages.RawDocument, error) {
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
		exp := time.Now().Add(time.Minute * 5)
		response.BlobURLGet = fs.GetStorageURL(response.Id)
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

func (fs *Storage) GetUser(id string) (response *messages.User, err error) {
	dataDir := fs.Cfg.DataDir
	filePath := ".userprofile"
	fullPath := path.Join(dataDir, id, filepath.Base(filePath))

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

	response = &messages.User{}
	err = json.Unmarshal(content, response)
	if err != nil {
		return
	}

	return
}

func (fs *Storage) GetUsers() (users []*messages.User, err error) {
	dataDir := fs.Cfg.DataDir

	err = filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if user, err := fs.GetUser(info.Name()); err == nil {
				users = append(users, user)
			}
		}

		return nil
	})

	return
}

func (fs *Storage) RegisterUser(u *messages.User) (err error) {
	userDir := path.Join(fs.Cfg.DataDir, u.Id)
	filePath := ".userprofile"
	fullPath := path.Join(userDir, filepath.Base(filePath))

	// Create the user's directory
	err = os.MkdirAll(userDir, 0755)
	if err != nil {
		return
	}

	// Create the profile file
	var js []byte
	js, err = json.Marshal(u)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fullPath, js, 0644)

	return
}

func (fs *Storage) UpdateUser(u *messages.User) (err error) {
	userDir := path.Join(fs.Cfg.DataDir, u.Id)
	filePath := ".userprofile"
	fullPath := path.Join(userDir, filepath.Base(filePath))

	// Erase the profile file
	var js []byte
	js, err = json.Marshal(u)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fullPath, js, 0644)

	return
}

func (fs *Storage) RegisterRoutes(router *gin.Engine) {

	router.GET("/storage", func(c *gin.Context) {
		id := c.Query("id")
		if id == "" {
			c.String(400, "set up us the bomb")
			return
		}

		//todo: storage provider
		log.Printf("Requestng Id: %s\n", id)

		reader, err := fs.GetDocument(id)
		defer reader.Close()

		if err != nil {
			log.Error(err)
			c.String(500, "internal error")
			c.Abort()
			return
		}

		c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
	})
	//todo: pass the token in the url
	router.PUT("/storage", func(c *gin.Context) {
		id := c.Query("id")
		log.Printf("Uploading id %s\n", id)
		body := c.Request.Body
		defer body.Close()

		err := fs.StoreDocument(body, id)
		if err != nil {
			log.Error(err)
			c.String(500, "set up us the bomb")
			c.Abort()
			return
		}

		c.JSON(200, gin.H{})
	})

}
