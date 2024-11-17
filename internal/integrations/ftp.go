package integrations

import (
	"bytes"
	"io"
	"os"
	"path"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/secsy/goftp"
	"github.com/sirupsen/logrus"
)

type FTPIntegration struct {
	client *goftp.Client
}

func newFTP(i model.IntegrationConfig) *FTPIntegration {
	config := goftp.Config{
		Logger:          os.Stderr,
		ActiveTransfers: i.ActiveTransfers,
	}

	if i.Username != "" {
		config.User = i.Username
	}
	if i.Password != "" {
		config.Password = i.Password
	}

	if strings.HasPrefix(i.Address, "ftps://") {
		config.TLSMode = goftp.TLSImplicit
		i.Address = strings.TrimPrefix(i.Address, "ftps://")
	} else if strings.HasPrefix(i.Address, "ftpes://") {
		config.TLSMode = goftp.TLSExplicit
		i.Address = strings.TrimPrefix(i.Address, "ftpes://")
	}

	client, err := goftp.DialConfig(config, strings.TrimPrefix(i.Address, "ftp://"))
	if err != nil {
		logrus.Errorf("An error occurred creating FTP client: %v\n", err)
		return nil
	}

	return &FTPIntegration{
		client,
	}
}

func (g *FTPIntegration) GetMetadata(fileID string) (*messages.IntegrationMetadata, error) {
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

func (g *FTPIntegration) List(folder string, depth int) (*messages.IntegrationFolder, error) {
	response := messages.NewIntegrationFolder(folder, "")

	if folder == rootFolder {
		folder = "/"
		response.Name = "FTP root"
	} else {
		decoded, err := decodeName(folder)
		if err != nil {
			return nil, err
		}
		folder = decoded
		response.Name = path.Base(folder)
	}
	logrus.Info("[ftp] query for: ", folder, " depth: ", depth)

	err := visitDir("", folder, depth, response, g.client.ReadDir)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (g *FTPIntegration) Download(fileID string) (io.ReadCloser, int64, error) {
	decoded, err := decodeName(fileID)
	if err != nil {
		return nil, 0, err
	}

	st, err := g.client.Stat(decoded)
	if err != nil {
		return nil, 0, err
	}

	var buf bytes.Buffer

	err = g.client.Retrieve(decoded, &buf)
	if err != nil {
		return nil, st.Size(), err
	}

	return io.NopCloser(&buf), st.Size(), err
}

func (g *FTPIntegration) Upload(folderID, name, fileType string, reader io.ReadCloser) (id string, err error) {
	folder := "/"
	if folderID != rootFolder {
		folder, err = decodeName(folderID)
		if err != nil {
			return
		}
	}
	fullpath := path.Join(folder, name+"."+fileType)
	logrus.Trace(logger, "Uploading: ", fullpath)

	err = g.client.Store(fullpath, reader)

	if err != nil {
		return
	}
	id = encodeName(fullpath)
	return
}
