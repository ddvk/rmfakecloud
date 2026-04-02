package integrations

import (
	"fmt"
	"image"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/sirupsen/logrus"
)

const (
	WebhookProvider = "webhook"
	SlackProvider   = "slack"
	FtpProvider     = "ftp"
	WebdavProvider  = "webdav"
	DropboxProvider = "dropbox"
	GoogleProvider  = "google"
	LocalfsProvider = "localfs"
)

type IntegrationProvider interface{}

type scopedIntegration struct {
	cfg       model.IntegrationConfig
	shared    bool
	readOnly  bool
	ownerUser string
}

// StorageIntegrationProvider abstracts 3rd party integrations
type StorageIntegrationProvider interface {
	IntegrationProvider
	GetMetadata(fileID string) (result *messages.IntegrationMetadata, err error)
	List(folderID string, depth int) (result *messages.IntegrationFolder, err error)
	Download(fileID string) (io.ReadCloser, int64, error)
	Upload(folderID, name, fileType string, reader io.ReadCloser) (string, error)
}

// MessagingIntegrationProvider abstracts 3rd party integrations
type MessagingIntegrationProvider interface {
	IntegrationProvider
	SendMessage(data messages.IntegrationMessageData, img image.Image) (string, error)
}

// getIntegrationProvider finds the integration provider for the user
func getIntegrationProvider(storer storage.UserStorer, uid, integrationid string) (IntegrationProvider, error) {
	effective, err := effectiveIntegrations(storer, uid)
	if err != nil {
		return nil, err
	}
	for _, intg := range effective {
		if intg.cfg.ID != integrationid {
			continue
		}
		switch intg.cfg.Provider {
		case WebhookProvider:
			return newWebhook(intg.cfg), nil
		case DropboxProvider:
			return newDropbox(intg.cfg), nil
		case FtpProvider:
			return newFTP(intg.cfg), nil
		case LocalfsProvider:
			return newLocalFS(intg.cfg), nil
		case WebdavProvider:
			return newWebDav(intg.cfg), nil
		}
	}
	return nil, fmt.Errorf("integration not found or no implmentation (only webdav) %s", integrationid)

}

func GetStorageIntegrationProvider(storer storage.UserStorer, uid, integrationid string) (StorageIntegrationProvider, error) {
	provider, err := getIntegrationProvider(storer, uid, integrationid)
	if err != nil {
		return nil, err
	}

	sip, ok := provider.(StorageIntegrationProvider)
	if !ok {
		return nil, fmt.Errorf("provider %q is not a storage provider", integrationid)
	}

	return sip, nil
}

func GetMessagingIntegrationProvider(storer storage.UserStorer, uid, integrationid string) (MessagingIntegrationProvider, error) {
	provider, err := getIntegrationProvider(storer, uid, integrationid)
	if err != nil {
		return nil, err
	}

	sip, ok := provider.(MessagingIntegrationProvider)
	if !ok {
		return nil, fmt.Errorf("provider %q is not a messaging provider", integrationid)
	}

	return sip, nil
}

// fix the name
func fixProviderName(n string) string {
	switch n {
	case WebhookProvider:
		fallthrough
	case SlackProvider:
		return "Slack"
	case FtpProvider:
		fallthrough
	case DropboxProvider:
		return "Dropbox"
	case GoogleProvider:
		fallthrough
	case WebdavProvider:
		return "GoogleDrive"
	default:
		return n
	}
}

func ProviderType(n string) string {
	switch n {
	case WebhookProvider:
		fallthrough
	case SlackProvider:
		return "Messaging"
	case FtpProvider:
		fallthrough
	case DropboxProvider:
		fallthrough
	case WebdavProvider:
		return "Storage"
	default:
		return n
	}
}

// List lists the integrations
func List(userstorer storage.UserStorer, uid string) (*messages.IntegrationsResponse, error) {
	effective, err := effectiveIntegrations(userstorer, uid)
	if err != nil {
		return nil, err
	}

	res := &messages.IntegrationsResponse{}
	for _, entry := range effective {
		userIntg := entry.cfg
		resIntg := messages.Integration{
			ID:           userIntg.ID,
			Name:         userIntg.Name,
			Provider:     fixProviderName(userIntg.Provider),
			ProviderType: ProviderType(userIntg.Provider),
			UserID:       entry.ownerUser,
			Shared:       entry.shared,
			ReadOnly:     entry.readOnly,
		}

		res.Integrations = append(res.Integrations, resIntg)
	}

	return res, nil
}

func effectiveIntegrations(storer storage.UserStorer, uid string) ([]scopedIntegration, error) {
	user, err := storer.GetUser(uid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found: %s", uid)
	}
	out := make([]scopedIntegration, 0, len(user.Integrations)+len(user.SharedIntegrations))
	for _, cfg := range user.Integrations {
		cfg.Shared = false
		out = append(out, scopedIntegration{
			cfg:       cfg,
			shared:    false,
			readOnly:  cfg.ReadOnly,
			ownerUser: user.ID,
		})
	}
	users, err := storer.GetUsers()
	if err != nil {
		return out, nil
	}
	for _, u := range users {
		if u == nil || !u.IsAdmin {
			continue
		}
		for _, cfg := range u.SharedIntegrations {
			readOnly := cfg.ReadOnly
			if user.IsAdmin && u.ID == user.ID {
				readOnly = cfg.ReadOnly
			} else {
				// Shared integrations are always non-writable for other users.
				readOnly = true
			}
			cfg.Shared = true
			cfg.ReadOnly = readOnly
			out = append(out, scopedIntegration{
				cfg:       cfg,
				shared:    true,
				readOnly:  readOnly,
				ownerUser: u.ID,
			})
		}
	}
	return out, nil
}

// IsReadOnly reports whether the integration should be treated as read-only for this user.
func IsReadOnly(storer storage.UserStorer, uid, integrationID string) (bool, error) {
	effective, err := effectiveIntegrations(storer, uid)
	if err != nil {
		return false, err
	}
	for _, intg := range effective {
		if intg.cfg.ID == integrationID {
			return intg.readOnly, nil
		}
	}
	return false, fmt.Errorf("integration not found: %s", integrationID)
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
				logrus.Tracef("[localfs] skipping unsupported content type for: %s", entryPath)
				continue
			}

			docName := strings.TrimSuffix(entryName, ext)
			extension := strings.TrimPrefix(ext, ".")

			file := &messages.IntegrationFile{
				ProvidedFileType:   contentType,
				SupportedFileTypes: []string{contentType},
				DateChanged:        d.ModTime(),
				FileExtension:      extension,
				FileType:           extension,
				ID:                 encodedPath,
				FileID:             encodedPath,
				Name:               docName,
				Size:               d.Size(),
				SourceFileType:     contentType,
			}

			parentFolder.Files = append(parentFolder.Files, file)
			logrus.Trace(loggerfs, "file added: ", entryPath)
		}
	}

	return nil
}
