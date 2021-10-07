package messages

// BlobStorageRequest else
type BlobStorageRequest struct {
	Method       string `json:"http_method"`
	Initial      bool   `json:"initial_sync"`
	RelativePath string `json:"relative_path"`
}

// BlobStorageResponse  what else
type BlobStorageResponse struct {
	Expires      string `json:"expires"`
	Method       string `json:"method"`
	RelativePath string `json:"relative_path"`
	URL          string `json:"url"`
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
	Auth0UserID      string `json:"auth0UserID"`
	Bookmarked       bool   `json:"bookmarked,omitempty"`
	Event            string `json:"event"`
	ID               string `json:"id,omitempty"`
	Parent           string `json:"parent,omitempty"`
	SourceDeviceDesc string `json:"sourceDeviceDesc"`
	SourceDeviceID   string `json:"sourceDeviceID"`
	Type             string `json:"type,omitempty"`
	Version          string `json:"version,omitempty"`
	VissibleName     string `json:"vissibleName,omitempty"`
}

// RawMetadata just a raw document
type RawMetadata struct {
	ID                string `json:"ID"`
	Version           int    `json:"Version"`
	Message           string `json:"Message"`
	Success           bool   `json:"Success"`
	BlobURLGet        string `json:"BlobURLGet"`
	BlobURLGetExpires string `json:"BlobURLGetExpires"`
	ModifiedClient    string `json:"ModifiedClient"`
	Type              string `json:"Type"`
	VissibleName      string `json:"VissibleName"`
	CurrentPage       int    `json:"CurrentPage"`
	Bookmarked        bool   `json:"Bookmarked"`
	Parent            string `json:"Parent"`
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

// SyncCompleted what else
type SyncCompleted struct {
	ID string `json:"id"`
}

// IntegrationsResponse what else
type IntegrationsResponse struct {
	//TODO:
	Integrations []string `json:"integrations"`
}
