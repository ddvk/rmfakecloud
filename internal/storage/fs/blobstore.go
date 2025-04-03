package fs

import (
	"io"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/storage"
	log "github.com/sirupsen/logrus"
)

// GetBlobURL return a url for a file to store
func (fs *FileSystemStorage) GetBlobURL(uid, blobid string, write bool) (docurl string, exp time.Time, err error) {
	uploadRL := fs.Cfg.StorageURL
	exp = time.Now().Add(time.Minute * config.ReadStorageExpirationInMinutes)
	strExp := strconv.FormatInt(exp.Unix(), 10)

	scope := storage.ReadScope
	if write {
		scope = storage.WriteScope
	}

	signature, err := storage.SignURLParams([]string{uid, blobid, strExp, scope}, fs.Cfg.JWTSecretKey)
	if err != nil {
		return
	}

	params := url.Values{
		storage.ParamUID:       {uid},
		storage.ParamBlobID:    {blobid},
		storage.ParamExp:       {strExp},
		storage.ParamSignature: {signature},
		storage.ParamScope:     {scope},
	}

	blobURL := uploadRL + storage.RouteBlob + "?" + params.Encode()
	log.Debugln("blobUrl: ", blobURL)
	return blobURL, exp, nil
}

// LoadBlob Opens a blob by id
func (fs *FileSystemStorage) LoadBlob(uid, blobid string) (reader io.ReadCloser, size int64, hash string, err error) {
	blobPath := path.Join(fs.getUserBlobPath(uid), common.Sanitize(blobid))
	log.Debugln("Fullpath:", blobPath)

	fi, err := os.Stat(blobPath)
	if err != nil || fi.IsDir() {
		return nil, 0, "", storage.ErrorNotFound
	}

	osFile, err := os.Open(blobPath)
	if err != nil {
		log.Errorf("cannot open blob %v", err)
		return
	}
	//TODO: cache the crc32c
	hash, err = common.CRC32CFromReader(osFile)
	if err != nil {
		log.Errorf("cannot get crc32c hash %v", err)
		return
	}
	_, err = osFile.Seek(0, 0)
	if err != nil {
		log.Errorf("cannot rewind file %v", err)
		return
	}
	reader = osFile
	return reader, fi.Size(), "crc32c=" + hash, err
}

// StoreBlob stores a document
func (fs *FileSystemStorage) StoreBlob(uid, id string, fileName string, hash string, stream io.Reader) error {
	log.Debugf("TODO: check/save etc. write file '%s', hash '%s'", fileName, hash)

	blobPath := path.Join(fs.getUserBlobPath(uid), common.Sanitize(id))
	log.Info("Write: ", blobPath)
	file, err := os.Create(blobPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, stream)
	if err != nil {
		return err
	}

	return nil
}

// use file size as generation
func generationFromFileSize(size int64) int64 {
	//time + 1 space + 64 hash + 1 newline
	return size / 86
}
