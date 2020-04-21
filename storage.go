package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
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
	return &response, nil

}
