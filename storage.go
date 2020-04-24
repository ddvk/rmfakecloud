package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"
)

func deleteFile(id string) error {
	//do not delete, move to trash
	meta := fmt.Sprintf("%s.metadata", id)
	fullPath := path.Join(dataDir, meta)
	err := os.Rename(fullPath, path.Join(dataDir, defaultTrashDir, meta))
	if err != nil {
		return err
	}
	meta = fmt.Sprintf("%s.zip", id)
	fullPath = path.Join(dataDir, meta)
	err = os.Rename(fullPath, path.Join(dataDir, defaultTrashDir, meta))
	if err != nil {
		return err
	}
	return nil
}

func formatStorageUrl(id string) string {
	return fmt.Sprintf("%s/storage?id=%s", uploadUrl, id)
}

func loadMetadata(filePath string) (*rawDocument, error) {
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

	response := rawDocument{}
	err = json.Unmarshal(content, &response)
	if err != nil {
		return nil, err
	}
	response.BlobURLGet = formatStorageUrl(response.Id)
	response.Success = true

	//fix time to utc because the tablet chokes
	tt, err := time.Parse(response.ModifiedClient, time.RFC3339Nano)
	if err != nil {
		tt = time.Now()
	}
	response.ModifiedClient = tt.UTC().Format(time.RFC3339Nano)

	return &response, nil

}
