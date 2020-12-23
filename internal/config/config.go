package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

const (
	defaultPort     = "3000"
	defaultDataDir  = "data"
	defaultTrashDir = "trash"

	// DefaultHost fake url
	DefaultHost = "local.appspot.com"

	envDataDir      = "DATADIR"
	envPort         = "PORT"
	envStorageURL   = "STORAGE_URL"
	envJWTSecretKey = "JWT_SECRET_KEY"
)

// Config config
type Config struct {
	Port         string
	StorageURL   string
	DataDir      string
	TrashDir     string
	JWTSecretKey []byte
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

	jwtSecretKey, err := hex.DecodeString(os.Getenv(envJWTSecretKey))
	if err != nil || len(jwtSecretKey) == 0 {
		jwtSecretKey = make([]byte, 32)
		_, err := rand.Read(jwtSecretKey)
		if err != nil {
			panic(err)
		}
		log.Warnf("You have to set %s with some content. Eg: %s='%X'", envJWTSecretKey, envJWTSecretKey, jwtSecretKey)
		log.Warn("  without this variable set, you'll be disconnected after this program restart")
	}

	cfg := Config{
		Port:         port,
		StorageURL:   uploadURL,
		DataDir:      dataDir,
		TrashDir:     trashDir,
		JWTSecretKey: jwtSecretKey,
	}
	return &cfg
}
