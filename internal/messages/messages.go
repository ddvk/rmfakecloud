package messages

import (
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
)

// NotificationType type of the notifiction
type NotificationType string

const (
	//DocAddedEvent addded
	DocAddedEvent NotificationType = "DocAdded"
	//DocDeletedEvent deleted
	DocDeletedEvent NotificationType = "DocDeleted"

	//SyncCompletedEvent sync completed sync15
	SyncCompletedEvent NotificationType = "SyncComplete"
)

// BlobStorageRequest else
type BlobStorageRequest struct {
	Method       string `json:"http_method"`
	Initial      bool   `json:"initial_sync"`
	RelativePath string `json:"relative_path"`
}

// BlobStorageResponse  what else
type BlobStorageResponse struct {
	Expires        string `json:"expires"`
	Method         string `json:"method"`
	RelativePath   string `json:"relative_path"`
	URL            string `json:"url"`
	MaxRequestSize int64  `json:"maxuploadsize_bytes,omitempty"`
}

// StatusResponse what else
type StatusResponse struct {
	ID      string `json:"ID"`
	Message string `json:"Message"`
	Success bool   `json:"Success"`
	Version int    `json:"Version"`
}

// WsMessage websocket notification
type WsMessage struct {
	Message      NotificationMessage `json:"message"`
	Subscription string              `json:"subscription,omitempty"`
}

// NotificationMessage child object
type NotificationMessage struct {
	Attributes   Attributes `json:"attributes"`
	MessageID    string     `json:"messageId,omitempty"`
	MessageID2   string     `json:"message_id,omitempty"`
	MessageID3   string     `json:"messageid,omitempty"`
	PublishTime  string     `json:"publishTime,omitempty"`
	PublishTime2 string     `json:"publish_time,omitempty"`
}

// Attributes child object
type Attributes struct {
	Auth0UserID      string           `json:"auth0UserID"`
	Bookmarked       bool             `json:"bookmarked,omitempty"`
	Event            NotificationType `json:"event"`
	ID               string           `json:"id,omitempty"`
	Parent           string           `json:"parent,omitempty"`
	SourceDeviceDesc string           `json:"sourceDeviceDesc"`
	SourceDeviceID   string           `json:"sourceDeviceID"`
	Type             common.EntryType `json:"type,omitempty"`
	Version          string           `json:"version,omitempty"`
	VissibleName     string           `json:"vissibleName,omitempty"`
}

// RawMetadata just a raw document, used by the legacy api
type RawMetadata struct {
	ID                string           `json:"ID"`
	Version           int              `json:"Version"`
	Message           string           `json:"Message"`
	Success           bool             `json:"Success"`
	BlobURLGet        string           `json:"BlobURLGet"`
	BlobURLGetExpires string           `json:"BlobURLGetExpires"`
	ModifiedClient    string           `json:"ModifiedClient"`
	Type              common.EntryType `json:"Type"`
	VissibleName      string           `json:"VissibleName"`
	CurrentPage       int              `json:"CurrentPage"`
	Bookmarked        bool             `json:"Bookmarked"`
	Parent            string           `json:"Parent"`
}

// IDRequest request with only an id
type IDRequest struct {
	ID string `json:"ID"`
}

// UploadRequest upload reuquest
type UploadRequest struct {
	ID      string `json:"ID"`
	Parent  string `json:"Parent"`
	Type    string `json:"Type"`
	Version int    `json:"Version"`
}

// UploadResponse surprise
type UploadResponse struct {
	ID                string `json:"ID"`
	Message           string `json:"Mesasge"`
	Success           bool   `json:"Success"`
	BlobURLPut        string `json:"BlobURLPut"`
	BlobURLPutExpires string `json:"BlobURLPutExpires"`
	Version           int    `json:"Version"`
}

// HostResponse what the host responded
type HostResponse struct {
	Host   string `json:"Host"`
	Status string `json:"Status"`
}

// DeviceTokenRequest give me token
type DeviceTokenRequest struct {
	Code       string `json:"code"`
	DeviceDesc string `json:"deviceDesc"`
	DeviceID   string `json:"deviceID"`
}

// SyncCompleted sync ended
type SyncCompleted struct {
	ID         string `json:"id"`
	Generation int64  `json:"generation"`
}

// SyncCompleted sync ended
type SyncCompletedRequestV2 struct {
	Generation int64 `json:"generation"`
}

// SyncRootV3
type SyncRootV3 struct {
	Generation int64  `json:"generation"`
	Hash       string `json:"hash"`
}

// IntegrationsResponse integrations
type IntegrationsResponse struct {
	Integrations []Integration `json:"integrations"`
}

// Integration integrations (google,dropbox)
type Integration struct {
	Added    time.Time `json:"added"`
	ID       string    `json:"id"`
	Issues   string    `json:"issues"`
	Name     string    `json:"name"`
	Provider string    `json:"provider"`
	UserID   string    `json:"userID"`
}

type IntegrationFile struct {
	DateChanged      time.Time `json:"dateChanged"`
	FileExtension    string    `json:"fileExtension"`
	FileID           string    `json:"fileID"`
	FileType         string    `json:"fileType"`
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	ProvidedFileType string    `json:"providedFileType"`
	Size             int       `json:"size"`
	SourceFileType   string    `json:"sourceFileType"`
}

func NewIntegrationFolder(id, name string) *IntegrationFolder {
	return &IntegrationFolder{
		FolderID:   id,
		ID:         id,
		Name:       name,
		Path:       "",
		Files:      []*IntegrationFile{},
		SubFolders: []*IntegrationFolder{},
	}
}

type IntegrationFolder struct {
	FolderID   string               `json:"folderID"`
	ID         string               `json:"id"`
	Name       string               `json:"name"`
	Path       string               `json:"path"`
	Files      []*IntegrationFile   `json:"files"`
	SubFolders []*IntegrationFolder `json:"subFolders"`
}

type IntegrationMetadata struct {
	FileType  string `json:"fileType"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	Thumbnail string `json:"thumbnail"`
}
