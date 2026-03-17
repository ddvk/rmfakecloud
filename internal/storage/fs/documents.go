package fs

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/exporter"
)

// DefaultTrashDir name of the trash dir
const (
	DefaultTrashDir = ".trash"
	CacheDir        = ".cache"
	Archive         = "archive"
	SyncFolder      = "sync"
)

// FileSystemStorage store everything to disk
type FileSystemStorage struct {
	Cfg *config.Config
}

func sanitizeFileName(fileName string) string {
	return filepath.Clean(filepath.Base(fileName))
}

func (fs *FileSystemStorage) getUserPath(uid string) string {
	return filepath.Join(fs.Cfg.DataDir, filepath.Base(userDir), common.SanitizeUid(uid))
}

// gets the blobstorage path
func (fs *FileSystemStorage) getUserBlobPath(uid string) string {
	return filepath.Join(fs.getUserPath(uid), SyncFolder)
}

func (fs *FileSystemStorage) getPathFromUser(uid, path string) string {
	return filepath.Join(fs.getUserPath(uid), sanitizeFileName(path))
}

// ExportDocument Exports a document to the outputType
func (fs *FileSystemStorage) ExportDocument(uid, id, outputType string, exportOption storage.ExportOption) (io.ReadCloser, error) {
	sanitizedID := common.Sanitize(id)
	zipFilePath := fs.getPathFromUser(uid, sanitizedID+storage.ZipFileExt)
	log.Debugln("Fullpath:", zipFilePath)
	rawStat, err := os.Stat(zipFilePath)
	if err != nil {
		return nil, fmt.Errorf("cant find raw document %v", err)
	}

	if outputType == "epub" {
		return fs.exportEpubFromZip(zipFilePath, rawStat.Size())
	}

	if outputType != "pdf" {
		return nil, errors.New("todo: only pdf and epub supported")
	}

	cacheDirPath := fs.getPathFromUser(uid, CacheDir)
	err = os.MkdirAll(cacheDirPath, 0700)
	if err != nil {
		return nil, err
	}

	outputFilePath := path.Join(cacheDirPath, sanitizedID+"-annotated.pdf")
	outStat, err := os.Stat(outputFilePath)

	// exists and not older
	if err == nil && !rawStat.ModTime().After(outStat.ModTime()) {
		return os.Open(outputFilePath)
	}

	size := rawStat.Size()
	arch := &exporter.MyArchive{}
	zipFile, err := os.Open(zipFilePath)
	if err != nil {
		return nil, err
	}
	defer zipFile.Close()
	err = arch.Read(zipFile, size)
	if err != nil {
		return nil, err
	}

	// Export raw PDF only (no annotations/drawings) when requested
	if exportOption == storage.ExportPayload && len(arch.Payload) > 0 {
		return io.NopCloser(bytes.NewReader(arch.Payload)), nil
	}

	if arch.Payload != nil {
		arch.PayloadReader = exporter.NewSeekCloser(arch.Payload)
	}

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return nil, err
	}

	err = exporter.RenderRmapi(arch, outputFile)
	if err != nil {
		return nil, err
	}

	_, err = outputFile.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	return outputFile, nil

}

// exportEpubFromZip opens the document zip and returns a reader for the .epub entry.
func (fs *FileSystemStorage) exportEpubFromZip(zipFilePath string, size int64) (io.ReadCloser, error) {
	zipFile, err := os.Open(zipFilePath)
	if err != nil {
		return nil, err
	}
	zr, err := zip.NewReader(zipFile, size)
	if err != nil {
		zipFile.Close()
		return nil, err
	}
	for _, f := range zr.File {
		if strings.HasSuffix(strings.ToLower(f.Name), storage.EpubFileExt) {
			rc, err := f.Open()
			if err != nil {
				zipFile.Close()
				return nil, err
			}
			return &zipEntryReadCloser{ReadCloser: rc, close: zipFile.Close}, nil
		}
	}
	zipFile.Close()
	return nil, errors.New("no epub entry found in document")
}

type zipEntryReadCloser struct {
	io.ReadCloser
	close func() error
}

func (z *zipEntryReadCloser) Close() error {
	_ = z.ReadCloser.Close()
	return z.close()
}

// GetDocument Opens a document by id
func (fs *FileSystemStorage) GetDocument(uid, id string) (io.ReadCloser, error) {
	fullPath := fs.getPathFromUser(uid, id+storage.ZipFileExt)
	log.Debugln("Fullpath:", fullPath)
	reader, err := os.Open(fullPath)
	return reader, err
}

// RemoveDocument removes document (moves it to trash)
func (fs *FileSystemStorage) RemoveDocument(uid, id string) error {

	trashDir := fs.getPathFromUser(uid, DefaultTrashDir)
	err := os.MkdirAll(trashDir, 0700)
	if err != nil {
		return err
	}
	//do not delete, move to trash
	log.Info(trashDir)
	meta := filepath.Base(id + storage.MetadataFileExt)
	fullPath := fs.getPathFromUser(uid, meta)
	err = os.Rename(fullPath, path.Join(trashDir, meta))
	if err != nil {
		return err
	}

	zipfile := filepath.Base(id + storage.ZipFileExt)
	fullPath = fs.getPathFromUser(uid, zipfile)
	err = os.Rename(fullPath, path.Join(trashDir, zipfile))
	if err != nil {
		return err
	}
	return nil
}

// StoreDocument stores a document
func (fs *FileSystemStorage) StoreDocument(uid, id string, stream io.ReadCloser) error {
	fullPath := fs.getPathFromUser(uid, id+storage.ZipFileExt)
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, stream)
	return err
}

// GetStorageURL the storage url
func (fs *FileSystemStorage) GetStorageURL(uid, id string) (docurl string, expiration time.Time, err error) {
	uploadRL := fs.Cfg.StorageURL
	exp := time.Now().Add(time.Minute * config.ReadStorageExpirationInMinutes)

	log.Debugln("uploadUrl: ", uploadRL)
	claim := &StorageClaim{
		DocumentID: id,
		UserID:     uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			Audience:  []string{storageUsage},
		},
	}
	signedToken, err := common.SignClaims(claim, fs.Cfg.JWTSecretKey)
	if err != nil {
		return "", exp, err
	}

	return fmt.Sprintf("%s%s/%s", uploadRL, routeStorage, url.QueryEscape(signedToken)), exp, nil
}
