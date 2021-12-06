package integrations

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/sirupsen/logrus"
)

const (
	loggerfs = "[localfs] "
)

type localFS struct {
	path string
}

// NewLocalFS localfs integration
func NewLocalFS(i model.IntegrationConfig) *localFS {
	return &localFS{
		path: i.LocalPath,
	}
}

// List populates the response
func (d *localFS) List(response *messages.IntegrationFolder, folder string, depth int) error {

	var visitDir func(curpath string, currentDepth, maxDepth int, e *messages.IntegrationFolder) error
	visitDir = func(currentPath string, currentDepth, maxDepth int, parentFolder *messages.IntegrationFolder) error {

		if currentDepth > maxDepth {
			return nil
		}

		logrus.Trace(loggerfs, "visiting: ", currentPath)
		fs, err := ioutil.ReadDir(currentPath)
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
				logrus.Trace(loggerfs, "dir added: ", fullPath)

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
				logrus.Trace(loggerfs, "file added: ", fullPath)
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
		response.Name = "LocalFS root"
	} else {
		decoded, err := decodeName(folder)
		if err != nil {
			return err
		}
		folder = decoded
		response.Name = path.Base(folder)
	}

	startPath := path.Join(d.path, path.Clean(folder))
	logrus.Info(loggerfs, "start path: ", startPath)
	logrus.Info("[localfs] query for: ", startPath, " depth: ", depth)

	err := visitDir(startPath, 0, depth, response)

	return err
}

func (d *localFS) Download(fileID string) (io.ReadCloser, error) {
	decoded, err := decodeName(fileID)
	if err != nil {
		return nil, err
	}

	localPath := path.Join(d.path, path.Clean(decoded))
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

	fullPath := path.Join(d.path, filePath)

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
