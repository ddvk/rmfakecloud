package main

import (
	"encoding/json"
	"fmt"
	"io"
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
	fmt.Println("url", uploadUrl)
	return fmt.Sprintf("%s/storage?id=%s", uploadUrl, id)
}

func saveUpload(stream io.ReadCloser, id string) error {
	fullPath := path.Join(dataDir, fmt.Sprintf("%s.zip", id))
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()
	io.Copy(file, stream)
	return nil
}

func loadMetadata(filePath string, withBlob bool) (*rawDocument, error) {
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

	response.Success = true

	if withBlob {
		exp := time.Now().Add(time.Minute * 5)
		response.BlobURLGet = formatStorageUrl(response.Id)
		response.BlobURLGetExpires = exp.UTC().Format(time.RFC3339Nano)
	} else {
		response.BlobURLGetExpires = time.Time{}.Format(time.RFC3339Nano)

	}

	//fix time to utc because the tablet chokes
	tt, err := time.Parse(response.ModifiedClient, time.RFC3339Nano)
	if err != nil {
		tt = time.Now()
	}
	response.ModifiedClient = tt.UTC().Format(time.RFC3339)

	return &response, nil

}
