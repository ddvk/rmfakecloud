package fs

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/juju/fslock"
	log "github.com/sirupsen/logrus"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/storage"
	rm2pdf2 "github.com/juruen/rmapi/annotations"
	rm2pdf "github.com/poundifdef/go-remarkable2pdf"
)

// DefaultTrashDir name of the trash dir
const (
	DefaultTrashDir = ".trash"
	CacheDir        = ".cache"
	Archive         = "archive"
	Sync            = "sync"
)

// Storage file system document storage
type Storage struct {
	Cfg *config.Config
}

func (fs *Storage) getUserSyncPath(uid string) string {
	return filepath.Join(fs.getUserPath(uid), Sync)
}

func (fs *Storage) getUserPath(uid string) string {

	return filepath.Join(fs.Cfg.DataDir, filepath.Base(userDir), filepath.Base(uid))
}
func (fs *Storage) getPathFromUser(uid, path string) string {
	return filepath.Join(fs.getUserPath(uid), filepath.Base(path))
}

func sanitize(id string) string {
	//TODO: more
	return path.Base(id)
}

// poundifdef caligraphy pen is nice
func render1(input, output string) (io.ReadCloser, error) {
	reader, err := zip.OpenReader(input)
	if err != nil {
		return nil, fmt.Errorf("can't open file %w", err)
	}
	defer reader.Close()

	writer, err := os.Create(output)
	if err != nil {
		return nil, fmt.Errorf("can't create outputfile %w", err)
	}
	//defer outputFile.Close()

	err = rm2pdf.RenderRmNotebookFromZip(&reader.Reader, writer)
	if err != nil {
		writer.Close()
		return nil, fmt.Errorf("can't render file %w", err)
	}

	_, err = writer.Seek(0, 0)
	if err != nil {
		writer.Close()
		return nil, fmt.Errorf("can't rewind file %w", err)
	}

	return writer, nil
}

//using rmapi (whole pdf)
func render2(input, output string) (io.ReadCloser, error) {
	options := rm2pdf2.PdfGeneratorOptions{
		AllPages: true,
	}
	gen := rm2pdf2.CreatePdfGenerator(input, output, options)
	err := gen.Generate()
	if err != nil {
		return nil, err
	}

	return os.Open(output)

}

// ExportDocument Exports a document to the outputType
func (fs *Storage) ExportDocument(uid, id, outputType string, exportOption storage.ExportOption) (io.ReadCloser, error) {
	if outputType != "pdf" {
		return nil, errors.New("todo: only pdfs supported")
	}

	cacheDirPath := fs.getPathFromUser(uid, CacheDir)
	err := os.MkdirAll(cacheDirPath, 0700)
	if err != nil {
		return nil, err
	}

	fullPath := fs.getPathFromUser(uid, id+zipExtension)
	log.Debugln("Fullpath:", fullPath)
	rawStat, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("cant find raw document %v", err)
	}

	sanitizedId := sanitize(id)
	outputname := path.Join(cacheDirPath, sanitizedId+"-annotated.pdf")
	outStat, err := os.Stat(outputname)

	// exists and not older
	if err == nil && !rawStat.ModTime().After(outStat.ModTime()) {
		return os.Open(outputname)
	}

	return render1(fullPath, outputname)

}

// GetDocument Opens a document by id
func (fs *Storage) GetDocument(uid, id string) (io.ReadCloser, error) {
	fullPath := fs.getPathFromUser(uid, id+zipExtension)
	log.Debugln("Fullpath:", fullPath)
	reader, err := os.Open(fullPath)
	return reader, err
}

// RemoveDocument removes document (moves it to trash)
func (fs *Storage) RemoveDocument(uid, id string) error {

	trashDir := fs.getPathFromUser(uid, DefaultTrashDir)
	err := os.MkdirAll(trashDir, 0700)
	if err != nil {
		return err
	}
	//do not delete, move to trash
	log.Info(trashDir)
	meta := filepath.Base(id + metadataExtension)
	fullPath := fs.getPathFromUser(uid, meta)
	err = os.Rename(fullPath, path.Join(trashDir, meta))
	if err != nil {
		return err
	}

	zipfile := filepath.Base(id + zipExtension)
	fullPath = fs.getPathFromUser(uid, zipfile)
	err = os.Rename(fullPath, path.Join(trashDir, zipfile))
	if err != nil {
		return err
	}
	return nil
}

// GetStorageURL return a url for a file to store
func (fs *Storage) GetStorageURL(uid, id, urltype string) (docurl string, expiration time.Time, err error) {
	uploadRL := fs.Cfg.StorageURL
	exp := time.Now().Add(time.Minute * config.ReadStorageExpirationInMinutes)

	log.Debugln("uploadUrl: ", uploadRL)
	claim := &common.StorageClaim{
		DocumentID: id,
		UserID:     uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: exp.Unix(),
			Audience:  common.StorageUsage,
		},
	}
	signedToken, err := common.SignClaims(claim, fs.Cfg.JWTSecretKey)
	if err != nil {
		return "", exp, err
	}

	return fmt.Sprintf("%s/%s/%s", uploadRL, urltype, url.QueryEscape(signedToken)), exp, nil
}

// GetDocument Opens a document by id
func (fs *Storage) LoadBlob(uid, id string) (io.ReadCloser, error) {
	fullPath := path.Join(fs.getUserSyncPath(uid), sanitize(id))
	log.Debugln("Fullpath:", fullPath)
	reader, err := os.Open(fullPath)
	return reader, err
}

// StoreDocument stores a document
func (fs *Storage) StoreBlob(uid, id string, stream io.ReadCloser) error {

	reader := stream
	//todo: locking
	if id == "root" {
		history := path.Join(fs.getUserSyncPath(uid), "root.history")

		lock := fslock.New(history)
		err := lock.LockWithTimeout(time.Duration(time.Second * 5))
		if err != nil {
			log.Error("cannot obtain lock")
		}
		defer lock.Unlock()

		var buf bytes.Buffer
		tee := io.TeeReader(stream, &buf)

		hist, err := os.OpenFile(history, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer hist.Close()
		t := time.Now().UTC().Format(time.RFC3339) + " "
		hist.WriteString(t)
		_, err = io.Copy(hist, tee)
		if err != nil {
			return err
		}
		hist.WriteString("\n")

		reader = ioutil.NopCloser(&buf)
	}

	fullPath := path.Join(fs.getUserSyncPath(uid), sanitize(id))
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, reader)
	if err != nil {
		return err
	}

	return nil
}

func (fs *Storage) RootGen(uid string) int {
	history := path.Join(fs.getUserSyncPath(uid), "root.history")
	lock := fslock.New(history)
	err := lock.LockWithTimeout(time.Duration(time.Second * 5))
	if err != nil {
		log.Error("cannot obtain lock")
		return 0
	}
	defer lock.Unlock()

	f, err := os.Stat(history)
	if err != nil {
		return 0
	}
	//time len + 1 + k64 bytes for the hash + newline
	return int(f.Size() / 86)

}

// StoreDocument stores a document
func (fs *Storage) StoreDocument(uid, id string, stream io.ReadCloser) error {
	fullPath := fs.getPathFromUser(uid, id+zipExtension)
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, stream)
	return err
}
