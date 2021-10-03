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
	"strconv"
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
func render2(input, output string) (io.ReadCloser, error) {
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
func render1(input, output string) (io.ReadCloser, error) {
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

	fullPath := fs.getPathFromUser(uid, id+ZipFileExt)
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
	fullPath := fs.getPathFromUser(uid, id+ZipFileExt)
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
	meta := filepath.Base(id + storage.MetadataFileExt)
	fullPath := fs.getPathFromUser(uid, meta)
	err = os.Rename(fullPath, path.Join(trashDir, meta))
	if err != nil {
		return err
	}

	zipfile := filepath.Base(id + ZipFileExt)
	fullPath = fs.getPathFromUser(uid, zipfile)
	err = os.Rename(fullPath, path.Join(trashDir, zipfile))
	if err != nil {
		return err
	}
	return nil
}

// StoreDocument stores a document
func (fs *Storage) StoreDocument(uid, id string, stream io.ReadCloser) error {
	fullPath := fs.getPathFromUser(uid, id+ZipFileExt)
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, stream)
	return err
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

//severs as root modification log and generation number source
const historyFile = ".root.history"
const rootFile = "root"

// GetStorageURL return a url for a file to store
func (fs *Storage) GetBlobURL(uid, blobid string) (docurl string, exp time.Time, err error) {
	uploadRL := fs.Cfg.StorageURL
	exp = time.Now().Add(time.Minute * config.ReadStorageExpirationInMinutes)
	strExp := strconv.FormatInt(exp.Unix(), 10)

	signature, err := Sign([]string{uid, blobid, strExp}, fs.Cfg.JWTSecretKey)
	if err != nil {
		return
	}

	params := url.Values{
		paramUid:       {uid},
		paramBlobId:    {blobid},
		paramExp:       {strExp},
		paramSignature: {signature},
	}

	blobUrl := uploadRL + routeBlob + "?" + params.Encode()
	log.Debugln("blobUrl: ", blobUrl)
	return blobUrl, exp, nil
}

// GetDocument Opens a document by id
func (fs *Storage) LoadBlob(uid, id string) (io.ReadCloser, int64, error) {
	generation := int64(1)
	blobPath := path.Join(fs.getUserSyncPath(uid), sanitize(id))
	log.Debugln("Fullpath:", blobPath)
	if id == rootFile {
		historyPath := path.Join(fs.getUserSyncPath(uid), historyFile)
		lock := fslock.New(historyPath)
		err := lock.LockWithTimeout(time.Duration(time.Second * 5))
		if err != nil {
			log.Error("cannot obtain lock")
			return nil, 0, err
		}
		defer lock.Unlock()

		fi, err1 := os.Stat(historyPath)
		if err1 == nil {
			generation = calcGen(fi.Size())
		}
	}

	if fi, err := os.Stat(blobPath); err != nil || fi.IsDir() {
		return nil, 0, storage.ErrorNotFound
	}

	reader, err := os.Open(blobPath)
	return reader, generation, err
}

// StoreDocument stores a document
func (fs *Storage) StoreBlob(uid, id string, stream io.Reader, matchGen int64) (generation int64, err error) {
	generation = 1

	reader := stream
	if id == rootFile {
		historyPath := path.Join(fs.getUserSyncPath(uid), historyFile)
		lock := fslock.New(historyPath)
		err = lock.LockWithTimeout(time.Duration(time.Second * 5))
		if err != nil {
			log.Error("cannot obtain lock")
		}
		defer lock.Unlock()

		currentGen := int64(0)
		fi, err1 := os.Stat(historyPath)
		if err1 == nil {
			currentGen = calcGen(fi.Size())
		}

		if currentGen != matchGen && matchGen > 0 {
			log.Warnf("wrong gen, has %d but is %d", matchGen, currentGen)
			return currentGen, storage.ErrorWrongGeneration
		}

		var buf bytes.Buffer
		tee := io.TeeReader(stream, &buf)

		var hist *os.File
		hist, err = os.OpenFile(historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		defer hist.Close()
		t := time.Now().UTC().Format(time.RFC3339) + " "
		hist.WriteString(t)
		_, err = io.Copy(hist, tee)
		if err != nil {
			return
		}
		hist.WriteString("\n")

		reader = ioutil.NopCloser(&buf)
		size, err1 := hist.Seek(0, os.SEEK_CUR)
		if err1 != nil {
			err = err1
			return
		}
		generation = calcGen(size)
	}

	blobPath := path.Join(fs.getUserSyncPath(uid), sanitize(id))
	file, err := os.Create(blobPath)
	if err != nil {
		return
	}
	defer file.Close()
	_, err = io.Copy(file, reader)
	if err != nil {
		return
	}

	return
}

//use file size as generation
func calcGen(size int64) int64 {
	//time + 1 space + 64 hash + 1 newline
	return size / 86
}
