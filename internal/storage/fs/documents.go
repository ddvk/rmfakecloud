package fs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
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
	if outputType != "pdf" {
		return nil, errors.New("todo: only pdfs supported")
	}

	cacheDirPath := fs.getPathFromUser(uid, CacheDir)
	err := os.MkdirAll(cacheDirPath, 0700)
	if err != nil {
		return nil, err
	}
	sanitizedID := common.Sanitize(id)

	zipFilePath := fs.getPathFromUser(uid, sanitizedID+storage.ZipFileExt)
	log.Debugln("Fullpath:", zipFilePath)
	rawStat, err := os.Stat(zipFilePath)
	if err != nil {
		return nil, fmt.Errorf("cant find raw document %v", err)
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

	if arch.Payload != nil {
		arch.PayloadReader = exporter.NewSeekCloser(arch.Payload)
	}

	// Detect version from first .rm file in archive
	version := exporter.VersionUnknown
	if len(arch.Pages) > 0 {
		version, err = detectArchiveVersion(arch)
		if err != nil {
			log.Warnf("Could not detect version for doc %s: %v, assuming v5", sanitizedID, err)
			version = exporter.VersionV5
		}
	}

	log.Debugf("Detected format %s for doc %s", version.String(), sanitizedID)

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return nil, err
	}
	defer outputFile.Close()

	// Route to appropriate renderer based on version
	if version == exporter.VersionV6 {
		log.Infof("Using rmc for v6 format doc %s", sanitizedID)

		// Use RMC for v6 files
		cfg := exporter.RmcConfig{
			RmcPath: fs.Cfg.RmcPath,
			TempDir: cacheDirPath,
			Timeout: time.Duration(fs.Cfg.RmcTimeout) * time.Second,
			InkscapePath: fs.Cfg.InkscapePath,
		}

		err = exporter.ExportV6ArchiveToPdf(arch, outputFilePath, cfg)
		if err != nil {
			return nil, fmt.Errorf("v6 export failed: %w", err)
		}
	} else {
		// Use existing v5 rendering
		log.Debugf("Using rmapi for v5 format doc %s", sanitizedID)
		err = exporter.RenderRmapi(arch, outputFile)
		if err != nil {
			return nil, fmt.Errorf("v5 export failed: %w", err)
		}
	}

	_, err = outputFile.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	return outputFile, nil

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

// detectArchiveVersion detects the .rm file version from an archive
func detectArchiveVersion(arch *exporter.MyArchive) (exporter.RmVersion, error) {
	if len(arch.Pages) == 0 {
		return exporter.VersionUnknown, fmt.Errorf("no pages in archive")
	}

	// Try to marshal first page and detect from header
	if arch.Pages[0].Data != nil {
		data, err := arch.Pages[0].Data.MarshalBinary()
		if err != nil {
			return exporter.VersionUnknown, fmt.Errorf("failed to marshal page data: %w", err)
		}
		return exporter.DetectRmVersion(bytes.NewReader(data))
	}

	return exporter.VersionUnknown, fmt.Errorf("no page data available")
}
