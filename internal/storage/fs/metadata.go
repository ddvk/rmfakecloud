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
	log "github.com/sirupsen/logrus"
)

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
