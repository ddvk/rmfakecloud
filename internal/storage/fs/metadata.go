package fs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/storage/fs/sync15"
	log "github.com/sirupsen/logrus"
)

const metadataExtension = ".metadata"
const zipExtension = ".zip"

func createMedatadata(name, id string) *messages.RawDocument {
	doc := messages.RawDocument{
		ID:             id,
		VissibleName:   name,
		Version:        1,
		ModifiedClient: time.Now().UTC().Format(time.RFC3339Nano),
		CurrentPage:    0,
		Type:           "DocumentType",
	}
	return &doc

}

func (fs *Storage) GetTree(uid string) (t *sync15.HashTree, err error) {

	syncFolder := fs.getUserSyncPath(uid)
	ls := &LocalStore{
		Folder: syncFolder,
	}

	cachePath := path.Join(fs.getUserPath(uid), ".tree")

	tree, err := sync15.LoadTree(cachePath)
	if err != nil {
		return nil, err
	}
	changed, err := tree.Mirror(ls)
	if err != nil {
		return nil, err
	}
	if changed {
		err = tree.Save(cachePath)
		if err != nil {
			return nil, err
		}
	}
	return tree, nil
}

// GetAllMetadata load all metadata
func (fs *Storage) GetAllMetadata(uid string) (result []*messages.RawDocument, err error) {
	result = []*messages.RawDocument{}

	var files []os.FileInfo
	folder := fs.getUserPath(uid)
	files, err = ioutil.ReadDir(folder)

	for _, f := range files {
		ext := filepath.Ext(f.Name())
		id := strings.TrimSuffix(f.Name(), ext)
		if ext != metadataExtension {
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
func (fs *Storage) GetMetadata(uid, id string) (*messages.RawDocument, error) {
	fullPath := fs.getPathFromUser(uid, id+metadataExtension)
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

	return &response, nil

}

// UpdateMetadata updates the metadata of a document
func (fs *Storage) UpdateMetadata(uid string, r *messages.RawDocument) error {
	filepath := fs.getPathFromUser(uid, r.ID+metadataExtension)

	js, err := json.Marshal(r)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath, js, 0600)
	return err

}
