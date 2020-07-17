package messages

type StatusResponse struct {
	Id      string `json:"ID"`
	Message string `json:"Message"`
	Success bool   `json:"Success"`
	Version int    `json:"Version"`
}
type WsMessage struct {
	Message      NotificationMessage `json:"message"`
	Subscription string              `json:"subscription"`
}
type NotificationMessage struct {
	Attributes   Attributes `json:"attributes"`
	MessageId    string     `json:"messageId"`
	MessageId2   string     `json:"message_id"`
	PublishTime  string     `json:"publishTime"`
	PublishTime2 string     `json:"publish_time"`
}
type Attributes struct {
	Auth0UserID      string `json:"auth0UserID"`
	Bookmarked       bool   `json:"bookmarked"`
	Event            string `json:"event"`
	Id               string `json:"id"`
	Parent           string `json:"parent"`
	SourceDeviceDesc string `json:"sourceDeviceDesc"`
	SourceDeviceId   string `json:"sourceDeviceID"`
	Type             string `json:"type"`
	Version          string `json:"version"`
	VissibleName     string `json:"vissibleName"`
	SourceDeviceID   string `json:"sourceDeviceID"`
}

type RawDocument struct {
	Id                string `json:"ID"`
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

// request with id
type IdRequest struct {
	Id string `json:"ID"`
}
type UploadRequest struct {
	Id      string `json:"ID"`
	Parent  string `json:"Parent"`
	Type    string `json:"Type"`
	Version int    `json:"Version"`
}
type UploadResponse struct {
	Id                string `json:"ID"`
	Message           string `json:"Mesasge"`
	Success           bool   `json:"Success"`
	BlobUrlPut        string `json:"BlobURLPut"`
	BlobURLPutExpires string `json:"BlobURLPutExpires"`
	Version           int    `json:"Version"`
}
type HostResponse struct {
	Host   string `json:"Host"`
	Status string `json:"Status"`
}

type DeviceTokenRequest struct {
	Code       string `json:"code"`
	DeviceDesc string `json:"deviceDesc"`
	DeviceId   string `json:"deviceID"`
}
