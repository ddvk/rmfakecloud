package integrations

import (
	"fmt"
	"io"

	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/storage"
)

const (
	webdavProvider  = "webdav"
	dropboxProvider = "dropbox"
	googleProvider  = "google"
	localfsProvider = "localfs"
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
		if intg.ID != integrationid {
			continue
		}
		switch intg.Provider {
		case webdavProvider:
			return NewWebDav(intg), nil
		case dropboxProvider:
			return NewDropbox(intg), nil
		case localfsProvider:
			return NewLocalFS(intg), nil
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

func List(userstorer storage.UserStorer, uid string, res *messages.IntegrationsResponse) error {
	user, err := userstorer.GetUser(uid)
	if err != nil {
		return err
	}

	for _, userIntg := range user.Integrations {
		resIntg := messages.Integration{
			ID:       userIntg.ID,
			Name:     userIntg.Name,
			Provider: fixProviderName(userIntg.Provider),
			UserID:   uid,
		}

		res.Integrations = append(res.Integrations, resIntg)
	}

	return nil
}
