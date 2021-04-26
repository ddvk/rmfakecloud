package fs

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/gin-gonic/gin"
	rm2pdf2 "github.com/juruen/rmapi/annotations"
	rm2pdf "github.com/poundifdef/go-remarkable2pdf"
)

// DefaultTrashDir name of the trash dir
const (
	DefaultTrashDir = ".trash"
	CacheDir        = ".cache"
	Archive         = "archive"
)

// Storage file system document storage
type Storage struct {
	Cfg *config.Config
}

func (fs *Storage) getUserPath(uid string) string {

	return filepath.Join(fs.Cfg.DataDir, filepath.Base(userDir), filepath.Base(uid))
}
func (fs *Storage) getPathFromUser(uid, path string) string {
	return filepath.Join(fs.getUserPath(uid), filepath.Base(path))
}

const tokenParam = "token"

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
func (fs *Storage) GetStorageURL(uid string, id string) (docurl string, expiration time.Time, err error) {
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

	return fmt.Sprintf("%s/storage/%s", uploadRL, url.QueryEscape(signedToken)), exp, nil
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

func (fs *Storage) parseToken(token string) (*common.StorageClaim, error) {
	claim := &common.StorageClaim{}
	err := common.ClaimsFromToken(claim, token, fs.Cfg.JWTSecretKey)
	if err != nil {
		return nil, err
	}
	if claim.StandardClaims.Audience != common.StorageUsage {
		return nil, errors.New("not a storage token")
	}
	return claim, nil
}

func (fs *Storage) uploadDocument(c *gin.Context) {
	strToken := c.Param(tokenParam)
	log.Debug("[storage] uploading with token:", strToken)
	token, err := fs.parseToken(strToken)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	id := token.DocumentID
	log.Debug("[storage] uploading documentId: ", id)
	body := c.Request.Body
	defer body.Close()

	err = fs.StoreDocument(token.UserID, id, body)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
func (fs *Storage) downloadDocument(c *gin.Context) {
	strToken := c.Param(tokenParam)
	token, err := fs.parseToken(strToken)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	id := token.DocumentID

	//todo: storage provider
	log.Info("Requestng Id: ", id)

	reader, err := fs.GetDocument(token.UserID, id)

	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer reader.Close()
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

// RegisterRoutes blah
func (fs *Storage) RegisterRoutes(router *gin.Engine) {

	router.GET("/storage/:"+tokenParam, fs.downloadDocument)
	router.PUT("/storage/:"+tokenParam, fs.uploadDocument)
}
