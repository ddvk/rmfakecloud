package fs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	log "github.com/sirupsen/logrus"
)

// GetAllMetadata load all metadata
func (fs *FileSystemStorage) GetAllMetadata(uid string) (result []*messages.RawMetadata, err error) {
	result = []*messages.RawMetadata{}

	var files []os.FileInfo
	folder := fs.getUserPath(uid)
	files, err = ioutil.ReadDir(folder)

	for _, f := range files {
		ext := filepath.Ext(f.Name())
		id := strings.TrimSuffix(f.Name(), ext)
		if ext != models.MetadataFileExt {
			continue
		}
		doc, err := fs.GetMetadata(uid, id)
		if err != nil {
			log.Error(err)
			continue
		}

		result = append(result, doc)
	}
	return
}

// GetMetadata loads a document's metadata
func (fs *FileSystemStorage) GetMetadata(uid, id string) (*messages.RawMetadata, error) {
	fullPath := fs.getPathFromUser(uid, id+models.MetadataFileExt)
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	response := messages.RawMetadata{}
	err = json.Unmarshal(content, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil

}

// UpdateMetadata updates the metadata of a document
func (fs *FileSystemStorage) UpdateMetadata(uid string, r *messages.RawMetadata) error {
	filepath := fs.getPathFromUser(uid, r.ID+models.MetadataFileExt)

	js, err := json.Marshal(r)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath, js, 0600)
	return err

}
