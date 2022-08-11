package integrations

import (
	"crypto/tls"
	"encoding/base64"
	"io"
	"net/http"
	"path"

	"github.com/zgs225/rmfakecloud/internal/messages"
	"github.com/zgs225/rmfakecloud/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/studio-b12/gowebdav"
)

const (
	rootFolder = "root"
	logger     = "[webdav] "
)

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

// newWebDav new webdav integration
func newWebDav(in model.IntegrationConfig) *WebDavIntegration {
	logrus.Tracef("new client, address: %s, user: %s pass: %s, insecure: %t ", in.Address, in.Username, in.Password, in.Insecure)
	c := gowebdav.NewClient(in.Address, in.Username, in.Password)

	if in.Insecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			WriteBufferSize: 4096,
		}
		c.SetTransport(tr)
		if logrus.IsLevelEnabled(logrus.TraceLevel) {
			c.SetInterceptor(func(method string, rq *http.Request) {
				logrus.Trace(method, " ", rq.URL, " ")
				if rq.Response != nil {
					logrus.Trace("RSP ", rq.Response.Status, " ")

				}
			})
		}

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
	// This fails with 400 without the connect, also it may cause high memory usage
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
func (w *WebDavIntegration) List(folder string, depth int) (*messages.IntegrationFolder, error) {
	response := messages.NewIntegrationFolder(folder, "")

	if folder == rootFolder {
		folder = "/"
		response.Name = "WebDav root"
	} else {
		decoded, err := decodeName(folder)
		if err != nil {
			return nil, err
		}
		folder = decoded
		response.Name = path.Base(folder)
	}
	logrus.Info("[webdav] query for: ", folder, " depth: ", depth)

	err := visitDir("", folder, depth, response, w.c.ReadDir)
	if err != nil {
		return nil, err
	}

	return response, nil

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
