package integrations

import (
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/sirupsen/logrus"
)

const (
	webdavProvider  = "webdav"
	dropboxProvider = "dropbox"
	googleProvider  = "google"
	localfsProvider = "localfs"
)

// IntegrationProvider abstracts 3rd party integrations
type IntegrationProvider interface {
	List(folderID string, depth int) (result *messages.IntegrationFolder, err error)
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
		if intg.ID != integrationid {
			continue
		}
		switch intg.Provider {
		case webdavProvider:
			return newWebDav(intg), nil
		case dropboxProvider:
			return newDropbox(intg), nil
		case localfsProvider:
			return newLocalFS(intg), nil
		}
	}
	return nil, fmt.Errorf("integration not found or no implmentation (only webdav) %s", integrationid)

}

// fix the name
func fixProviderName(n string) string {
	switch n {
	case dropboxProvider:
		return "Dropbox"
	case googleProvider:
		fallthrough
	case webdavProvider:
		return "GoogleDrive"
	default:
		return n
	}
}

// List lists the integrations
func List(userstorer storage.UserStorer, uid string) (*messages.IntegrationsResponse, error) {
	user, err := userstorer.GetUser(uid)
	if err != nil {
		return nil, err
	}

	res := &messages.IntegrationsResponse{}
	for _, userIntg := range user.Integrations {
		resIntg := messages.Integration{
			ID:       userIntg.ID,
			Name:     userIntg.Name,
			Provider: fixProviderName(userIntg.Provider),
			UserID:   uid,
		}

		res.Integrations = append(res.Integrations, resIntg)
	}

	return res, nil
}

func visitDir(root, currentPath string, depth int, parentFolder *messages.IntegrationFolder,
	readDir func(string) ([]fs.FileInfo, error)) error {
	if depth < 1 {
		return nil
	}

	fullPath := path.Join(root, currentPath)
	logrus.Trace(loggerfs, "visiting: ", currentPath)
	fs, err := readDir(fullPath)
	if err != nil {
		return err
	}

	for _, d := range fs {
		entryName := d.Name()
		entryPath := path.Join(currentPath, entryName)
		encodedPath := encodeName(entryPath)
		if d.IsDir() {

			folder := messages.NewIntegrationFolder(encodedPath, entryName)

			err = visitDir(root, entryPath, depth-1, folder, readDir)
			if err != nil {
				return err
			}

			parentFolder.SubFolders = append(parentFolder.SubFolders, folder)
			logrus.Trace(loggerfs, "dir added: ", entryPath)

		} else {
			ext := path.Ext(entryName)
			contentType := contentTypeFromExt(ext)
			if contentType == "" {
				continue
			}

			docName := strings.TrimSuffix(entryName, ext)
			extension := strings.TrimPrefix(ext, ".")

			file := &messages.IntegrationFile{
				ProvidedFileType: contentType,
				DateChanged:      d.ModTime(),
				FileExtension:    extension,
				FileType:         extension,
				ID:               encodedPath,
				FileID:           encodedPath,
				Name:             docName,
				Size:             int(d.Size()),
				SourceFileType:   contentType,
			}

			parentFolder.Files = append(parentFolder.Files, file)
			logrus.Trace(loggerfs, "file added: ", entryPath)
		}
	}

	return nil
}
