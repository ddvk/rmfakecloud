package integrations

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/sirupsen/logrus"
	"github.com/studio-b12/gowebdav"
)

const (
	rootFolder     = "root"
	logger         = "[webdav] "
	webdavProvider = "webdav"
)

// IntegrationProvider abstracts 3rd party integrations
type IntegrationProvider interface {
	List(response *messages.IntegrationFolder, folderID string, depth int) error
	Download(fileID string) (io.ReadCloser, error)
	Upload(folderID, name, fileType string, reader io.ReadCloser) (string, error)
}

// GetIntegrationProvider finds the integration provider for the user
func GetIntegrationProvider(storer storage.UserStorer, uid, integrationid string) (IntegrationProvider, error) {
	usr, err := storer.GetUser(uid)
	if err != nil {
		return nil, err
	}
	for _, intg := range usr.Integrations {
		if intg.Provider == webdavProvider {
			return NewWebDav(intg), nil
		}
	}
	return nil, fmt.Errorf("integration not found or no implmentation (only webdav) %s", integrationid)

}

func encodeName(n string) string {
	return base64.URLEncoding.EncodeToString([]byte(n))
}

func decodeName(n string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(n)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// NewWebDav new webdav integration
func NewWebDav(in model.IntegrationConfig) *WebDavIntegration {
	logrus.Tracef("new client, address: %s, user: %s pass: %s, insecure: %t ", in.Address, in.Username, in.Password, in.Insecure)
	c := gowebdav.NewClient(in.Address, in.Username, in.Password)

	if in.Insecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			WriteBufferSize: 4096,
		}
		c.SetTransport(tr)
		c.SetInterceptor(func(method string, rq *http.Request) {
			logrus.Trace(method, " ", rq.URL, " ")
			if rq.Response != nil {
				logrus.Trace("RSP ", rq.Response.Status, " ")

			}
		})

	}
	return &WebDavIntegration{
		c,
	}
}

// WebDavIntegration webdav support
type WebDavIntegration struct {
	c *gowebdav.Client
}

// Upload uploads a file
func (w *WebDavIntegration) Upload(folderID, name, fileType string, reader io.ReadCloser) (id string, err error) {
	folder := "/"
	if folderID != rootFolder {
		folder, err = decodeName(folderID)
		if err != nil {
			return
		}
	}
	fullpath := path.Join(folder, name+"."+fileType)
	logrus.Trace(logger, "Uploading: ", fullpath)
	err = w.c.Connect()
	if err != nil {
		return
	}
	// This fails with 400 without the connect because the reader gets closed, also it may cause high memory usage
	// https://github.com/studio-b12/gowebdav/issues/24#issuecomment-5766910111
	err = w.c.WriteStream(fullpath, reader, 0644)

	if err != nil {
		return
	}
	id = encodeName(fullpath)
	return
}

// Download downloads
func (w *WebDavIntegration) Download(fileID string) (io.ReadCloser, error) {
	decoded, err := decodeName(fileID)
	if err != nil {
		return nil, err
	}
	return w.c.ReadStream(decoded)
}

// List populates the response
func (w *WebDavIntegration) List(response *messages.IntegrationFolder, folder string, depth int) error {

	var visitDir func(curpath string, currentDepth, maxDepth int, e *messages.IntegrationFolder) error
	visitDir = func(currentPath string, currentDepth, maxDepth int, parentFolder *messages.IntegrationFolder) error {

		if currentDepth > maxDepth {
			return nil
		}

		logrus.Trace(logger, "visiting: ", currentPath)
		fs, err := w.c.ReadDir(currentPath)
		if err != nil {
			return err
		}

		hasDirs := false

		for _, d := range fs {
			entryName := d.Name()
			fullPath := path.Join(currentPath, entryName)
			escaped := encodeName(fullPath)
			if d.IsDir() {
				hasDirs = true
				folder := messages.IntegrationFolder{}
				folder.FolderID = escaped
				folder.ID = escaped
				folder.Name = entryName

				err = visitDir(fullPath, currentDepth+1, maxDepth, &folder)
				if err != nil {
					return err
				}

				parentFolder.SubFolders = append(parentFolder.SubFolders, folder)
				logrus.Trace(logger, "dir added: ", fullPath)

			} else {
				ext := path.Ext(entryName)
				contentType := contentTypeFromExt(ext)
				if contentType == "" {
					continue
				}

				file := messages.IntegrationFile{}
				file.ProvidedFileType = contentType
				file.DateChanged = d.ModTime()
				docName := strings.TrimSuffix(entryName, ext)
				file.FileExtension = strings.TrimPrefix(ext, ".")
				file.FileType = file.FileExtension
				file.ID = escaped
				file.FileID = escaped
				file.Name = docName
				file.Size = int(d.Size())
				file.SourceFileType = file.ProvidedFileType

				parentFolder.Files = append(parentFolder.Files, file)
				logrus.Trace(logger, "file added: ", fullPath)
			}
		}
		if !hasDirs {
			parentFolder.SubFolders = make([]messages.IntegrationFolder, 0)
		}

		return nil
	}

	response.FolderID = folder
	response.ID = folder

	if folder == rootFolder {
		folder = "/"
		response.Name = "WebDav root"
	} else {
		decoded, err := decodeName(folder)
		if err != nil {
			return err
		}
		folder = decoded
		response.Name = path.Base(folder)
	}
	logrus.Info("[webdav] query for: ", folder, " depth: ", depth)

	err := visitDir(folder, 0, depth, response)

	return err

}

func contentTypeFromExt(ext string) string {
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".epub":
		return "application/epub+zip"
	}
	return ""
}
