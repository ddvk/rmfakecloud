package integrations

import (
	"io"
	"io/fs"
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

func (d *localFS) GetMetadata(fileID string) (*messages.IntegrationMetadata, error) {
	decoded, err := decodeName(fileID)
	if err != nil {
		return nil, err
	}

	ext := path.Ext(decoded)
	contentType := contentTypeFromExt(ext)

	return &messages.IntegrationMetadata{
		ID:               fileID,
		Name:             path.Base(decoded),
		Thumbnail:        []byte{},
		SourceFileType:   contentType,
		ProvidedFileType: contentType,
		FileType:         ext,
	}, nil
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

	err := visitDir(d.rootPath, startPath, depth, response, func(s string) ([]fs.FileInfo, error) {
		di, err := os.ReadDir(s)
		if err != nil {
			return nil, err
		}
		result := make([]fs.FileInfo, 0, len(di))
		for _, d := range di {
			fi, err := d.Info()
			if err != nil {
				logrus.Warnf("[localfs] cant get fileinfo %v", err)
				continue
			}
			result = append(result, fi)
		}
		return result, nil
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (d *localFS) Download(fileID string) (io.ReadCloser, int64, error) {
	decoded, err := decodeName(fileID)
	if err != nil {
		return nil, 0, err
	}

	localPath := path.Join(d.rootPath, path.Clean(decoded))

	st, err := os.Stat(localPath)
	if err != nil {
		return nil, 0, err
	}

	res, err := os.Open(localPath)
	return res, st.Size(), err
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
