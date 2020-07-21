package config

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
)

const (
	defaultPort     = "3000"
	defaultDataDir  = "data"
	defaultTrashDir = "trash"

	// DefaultHost fake url
	DefaultHost = "local.appspot.com"

	envDataDir    = "DATADIR"
	envPort       = "PORT"
	envStorageURL = "STORAGE_URL"
)

// Config config
type Config struct {
	Port       string
	StorageURL string
	DataDir    string
	TrashDir   string
}

// FromEnv config from environment values
func FromEnv() *Config {
	var err error
	var dataDir string
	data := os.Getenv(envDataDir)
	if data != "" {
		dataDir = data
	} else {
		dataDir, err = filepath.Abs(defaultDataDir)
		if err != nil {
			panic(err)
		}
	}
	trashDir := path.Join(dataDir, defaultTrashDir)
	err = os.MkdirAll(trashDir, 0700)
	if err != nil {
		panic(err)
	}

	port := os.Getenv(envPort)
	if port == "" {
		port = defaultPort
	}

	uploadURL := os.Getenv(envStorageURL)
	if uploadURL == "" {
		host, err := os.Hostname()
		if err != nil {
			log.Println("cannot get hostname")
			host = DefaultHost
		}
		uploadURL = fmt.Sprintf("http://%s:%s", host, port)
	}

	if err != nil {
		panic(err)
	}

	cfg := Config{
		Port:       port,
		StorageURL: uploadURL,
		DataDir:    dataDir,
		TrashDir:   trashDir,
	}
	return &cfg
}
