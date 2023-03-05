package integrations

import (
	"io"

	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/sirupsen/logrus"
)

type DropBox struct {
	client files.Client
}

func newDropbox(i model.IntegrationConfig) *DropBox {
	cfg := dropbox.Config{
		Token: i.Accesstoken,
	}
	client := files.New(cfg)
	return &DropBox{
		client,
	}
}

func (d *DropBox) List(folderID string, depth int) (*messages.IntegrationFolder, error) {

	args := files.ListFolderArg{
		Recursive: true,
		Limit:     3,
		Path:      "",
	}
	res, err := d.client.ListFolder(&args)
	if err != nil {
		return nil, err
	}

	entries := res.Entries

	for res.HasMore {
		logrus.Info("has more")
		arg := files.NewListFolderContinueArg(res.Cursor)

		res, err = d.client.ListFolderContinue(arg)
		if err != nil {
			return nil, err
		}

		entries = append(entries, res.Entries...)
	}
	logrus.Info("Entries: ", len(entries))
	for _, entry := range entries {
		switch f := entry.(type) {
		case *files.FileMetadata:
			logrus.Info(f)
			logrus.Info("file: ", f.Id, f.Name, f.PathLower)

		case *files.FolderMetadata:
			logrus.Info(f)
			logrus.Info("folder:", f.Id)
		}
	}
	response := messages.NewIntegrationFolder(rootFolder, "DropBox Root")
	return response, nil
}

func (d *DropBox) Download(fileID string) (io.ReadCloser, error) {
	return nil, nil

}
func (d *DropBox) Upload(folderID, name, fileType string, reader io.ReadCloser) (string, error) {
	// commit := files.CommitInfo{
	// 	Path: "/" + name + "." + fileType,
	// 	Mode: &files.WriteMode{
	// 		Tagged: dropbox.Tagged{
	// 			Tag: "overwrite",
	// 		},
	// 	},
	// }
	// r, err := d.client.Upload()
	// if err != nil {
	// 	return "", err
	// }
	return "", nil
}
