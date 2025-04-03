package fs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/danjacques/gofslock/fslock"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	log "github.com/sirupsen/logrus"
)

func (fs *FileSystemStorage) GetRootIndex(uid string) (string, int64, error) {
	historyPath := fs.getPathFromUser(uid, historyFile)

	lock, err := fslock.Lock(historyPath)
	if err != nil {
		log.Error("cannot obtain lock")
		return "", 0, err
	}
	defer lock.Unlock()

	fi, err := os.Stat(historyPath)
	if err != nil {
		return "", 0, err
	}

	if fi.Size() == 0 {
		return "", 0, storage.ErrorNotFound
	}

	fd, err := os.Open(historyPath)
	if err != nil {
		return "", 0, err
	}
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	lastline := ""
	var generation int64
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			generation += 1
			lastline = line
		}
	}

	fields := strings.Fields(lastline)
	if len(fields) != 2 {
		return "", 0, fmt.Errorf(".root.history corrupted")
	}

	return fields[1], generation, nil
}

func (fs *FileSystemStorage) UpdateRoot(uid string, stream io.Reader, lastGen int64) (int64, error) {
	historyPath := fs.getPathFromUser(uid, historyFile)

	lock, err := fslock.Lock(historyPath)
	if err != nil {
		log.Error("cannot obtain lock")
		return 0, err
	}
	defer lock.Unlock()

	currentGen := int64(0)
	fi, err := os.Stat(historyPath)
	if err == nil {
		currentGen = generationFromFileSize(fi.Size())
	}

	if currentGen != lastGen && currentGen > 0 {
		log.Warnf("wrong generation, currentGen %d, lastGen %d", currentGen, lastGen)
		return currentGen, storage.ErrorWrongGeneration
	}

	hist, err := os.OpenFile(historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return currentGen, err
	}
	defer hist.Close()

	hist.WriteString(time.Now().UTC().Format(time.RFC3339) + " ")
	_, err = io.Copy(hist, stream)
	if err != nil {
		return currentGen, err
	}
	hist.WriteString("\n")

	return currentGen + 1, nil
}

// GetCachedTree returns the cached blob tree for the user
func (fs *FileSystemStorage) GetCachedTree(uid string, blobStorage models.RemoteStorage) (t *models.HashTree, err error) {
	cachePath := path.Join(fs.getUserPath(uid), cachedTreeName)

	tree, err := models.LoadTree(cachePath)
	if err != nil {
		return nil, err
	}
	tree.SchemaVersion = fs.Cfg.HashSchemaVersion
	changed, err := tree.Mirror(blobStorage)
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

// SaveCachedTree saves the cached tree
func (fs *FileSystemStorage) SaveCachedTree(uid string, t *models.HashTree) error {
	cachePath := path.Join(fs.getUserPath(uid), cachedTreeName)
	return t.Save(cachePath)
}
