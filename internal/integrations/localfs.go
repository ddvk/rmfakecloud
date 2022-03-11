package integrations

import (
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/sirupsen/logrus"
)

const (
	loggerfs = "[localfs] "
)

type localFS struct {
	rootPath string
}

// newLocalFS localfs integration
func newLocalFS(i model.IntegrationConfig) *localFS {
	return &localFS{
		rootPath: i.Path,
	}
}

// List populates the response
func (d *localFS) List(folder string, depth int) (*messages.IntegrationFolder, error) {
	response := messages.NewIntegrationFolder(folder, "")

	if folder == rootFolder {
		folder = "/"
		response.Name = "LocalFS root"
	} else {
		decoded, err := decodeName(folder)
		if err != nil {
			return nil, err
		}
		folder = decoded
		response.Name = path.Base(folder)
	}

	startPath := path.Clean(folder)

	logrus.Info("[localfs] query for: ", startPath, " depth: ", depth)

	err := visitDir(d.rootPath, startPath, depth, response, ioutil.ReadDir)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (d *localFS) Download(fileID string) (io.ReadCloser, error) {
	decoded, err := decodeName(fileID)
	if err != nil {
		return nil, err
	}

	localPath := path.Join(d.rootPath, path.Clean(decoded))
	return os.Open(localPath)

}
func (d *localFS) Upload(folderID, name, fileType string, reader io.ReadCloser) (id string, err error) {
	folder := "/"
	if folderID != rootFolder {
		folder, err = decodeName(folderID)
		if err != nil {
			return
		}
	}
	//TODO: more cleanup and checks
	filePath := path.Clean(path.Join(folder, name+"."+fileType))
	logrus.Trace(loggerfs, "Cleaned: ", filePath)

	fullPath := path.Join(d.rootPath, filePath)

	logrus.Trace(loggerfs, "Uploading to: ", fullPath)
	writer, err := os.Create(fullPath)
	if err != nil {
		return
	}
	defer writer.Close()

	_, err = io.Copy(writer, reader)

	if err != nil {
		return
	}
	id = encodeName(filePath)
	return
}
