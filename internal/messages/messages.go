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
	Method       string `json:"http_method"`
	RelativePath string `json:"relative_path"`
	Url          string `json:"url"`
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
	Subscription string              `json:"subscription"`
}

// NotificationMessage child object
type NotificationMessage struct {
	Attributes   Attributes `json:"attributes"`
	MessageID    string     `json:"messageId"`
	MessageID2   string     `json:"message_id"`
	PublishTime  string     `json:"publishTime"`
	PublishTime2 string     `json:"publish_time"`
}

// Attributes child object
type Attributes struct {
	Auth0UserID      string `json:"auth0UserID"`
	Bookmarked       bool   `json:"bookmarked"`
	Event            string `json:"event"`
	ID               string `json:"id"`
	Parent           string `json:"parent"`
	SourceDeviceDesc string `json:"sourceDeviceDesc"`
	SourceDeviceID   string `json:"sourceDeviceID"`
	Type             string `json:"type"`
	Version          string `json:"version"`
	VissibleName     string `json:"vissibleName"`
}

// RawDocument just a raw document
type RawDocument struct {
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
