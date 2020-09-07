package fs

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/gin-gonic/gin"
)

// Storage file system document storage
type Storage struct {
	Cfg config.Config
}

// GetDocument Opens a document by id
func (fs *Storage) GetDocument(id string) (io.ReadCloser, error) {
	fullPath := path.Join(fs.Cfg.DataDir, filepath.Base(fmt.Sprintf("%s.zip", id)))
	log.Println("Fullpath:", fullPath)
	reader, err := os.Open(fullPath)
	return reader, err
}

// UpdateMetadata updates the metadata of a document
func (fs *Storage) UpdateMetadata(r *messages.RawDocument) error {
	filepath := path.Join(fs.Cfg.DataDir, fmt.Sprintf("%s.metadata", r.Id))

	js, err := json.Marshal(r)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath, js, 0700)
	return err

}

// RemoveDocument remove document
func (fs *Storage) RemoveDocument(id string) error {
	//do not delete, move to trash
	dataDir := fs.Cfg.DataDir
	trashDir := fs.Cfg.TrashDir
	meta := fmt.Sprintf("%s.metadata", id)
	fullPath := path.Join(dataDir, meta)
	err := os.Rename(fullPath, path.Join(dataDir, trashDir, meta))
	if err != nil {
		return err
	}
	meta = fmt.Sprintf("%s.zip", id)
	fullPath = path.Join(dataDir, meta)
	err = os.Rename(fullPath, path.Join(dataDir, trashDir, meta))
	if err != nil {
		return err
	}
	return nil
}

func (fs *Storage) GetStorageURL(id string) string {
	uploadRL := fs.Cfg.StorageURL
	fmt.Println("url", uploadRL)
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
			log.Println(err)
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
	fullPath := path.Join(dataDir, filePath)
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
	tt, err := time.Parse(response.ModifiedClient, time.RFC3339Nano)
	if err != nil {
		tt = time.Now()
	}
	response.ModifiedClient = tt.UTC().Format(time.RFC3339)

	return &response, nil

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
			log.Println(err)
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
			fmt.Println(err)
			c.String(500, "set up us the bomb")
			c.Abort()
			return
		}

		c.JSON(200, gin.H{})
	})

}
