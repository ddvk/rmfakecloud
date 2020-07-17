package fs

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/messages"
)

type FSStorage struct {
	Cfg config.Config
}

func (fs *FSStorage) GetContent(id string) (io.ReadCloser, error) {
	fullPath := path.Join(fs.Cfg.DataDir, filepath.Base(fmt.Sprintf("%s.zip", id)))
	log.Println("Fullpath:", fullPath)
	reader, err := os.Open(fullPath)
	return reader, err
}

func (fs *FSStorage) UpdateMetadata(r *messages.RawDocument) error {
	filepath := path.Join(fs.Cfg.DataDir, fmt.Sprintf("%s.metadata", r.Id))

	js, err := json.Marshal(r)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath, js, 0700)
	return err

}

func (fs *FSStorage) DeleteFile(id string) error {
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

func (fs *FSStorage) GetStorageUrl(id string) string {
	uploadUrl := fs.Cfg.StorageUrl
	fmt.Println("url", uploadUrl)
	return fmt.Sprintf("%s/storage?id=%s", uploadUrl, id)
}

func (fs *FSStorage) SaveUpload(stream io.ReadCloser, id string) error {
	dataDir := fs.Cfg.DataDir
	fullPath := path.Join(dataDir, fmt.Sprintf("%s.zip", id))
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()
	io.Copy(file, stream)
	return nil
}

func (fs *FSStorage) LoadAll(withBlob bool) (result []*messages.RawDocument, err error) {
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
		doc, err := fs.LoadMetadata(id, withBlob)
		if err != nil {
			log.Println(err)
			continue
		}

		result = append(result, doc)
	}
	return
}

func (fs *FSStorage) LoadMetadata(id string, withBlob bool) (*messages.RawDocument, error) {
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
		response.BlobURLGet = fs.GetStorageUrl(response.Id)
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
