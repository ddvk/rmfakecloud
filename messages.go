package main

type statusResponse struct {
	Id      string `json:"ID"`
	Message string `json:"Message"`
	Success bool   `json:"Success"`
	Version int    `json:"Version"`
}
type wsMessage struct {
	Message      notificationMessage `json:"message"`
	Subscription string              `json:"subscription"`
}
type notificationMessage struct {
	Attributes   attributes `json:"attributes"`
	MessageId    string     `json:"messageId"`
	MessageId2   string     `json:"message_id"`
	PublishTime  string     `json:"publishTime"`
	PublishTime2 string     `json:"publish_time"`
}
type attributes struct {
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

// type updateStatusRequest struct {
// 	Id             string `json:"ID"`
// 	Parent         string `json:"Parent"`
// 	Version        int    `json:"Version"`
// 	Message        string `json:"Message"`
// 	Success        bool   `json:"Success"`
// 	ModifiedClient string `json:"ModifiedClient"`
// 	Type           string `json:"Type"`
// 	VissibleName   string `json:"VissibleName"`
// 	CurrentPage    int    `json:"CurrentPage"`
// 	Bookmarked     bool   `json:"Bookmarked"`
// }
type rawDocument struct {
	Id                string `json:"ID"`
	Version           int    `json:"Version"`
	Message           string `json:"Message"`
	Success           bool   `json:"Success"`
	BlobURLGet        string `json:"BlobURLGet"`
	BlobURLGetExpires string `json:"BlobURLGetExpires"`
	BlobURLPut        string `json:"BlobURLPut"`
	BlobURLPutExpires string `json:"BlobURLPutExpires"`
	ModifiedClient    string `json:"ModifiedClient"`
	Type              string `json:"Type"`
	VissibleName      string `json:"VissibleName"`
	CurrentPage       int    `json:"CurrentPage"`
	Bookmarked        bool   `json:"Bookmarked"`
	Parent            string `json:"Parent"`
}

// request with id
type idRequest struct {
	Id string `json:"ID"`
}
type documentRequest struct {
	Id         string `json:"ID"`
	Message    string `json:"Mesasge"`
	Success    bool   `json:"Success"`
	BlobUrlPut string `json:"BlobURLPut"`
	Version    int    `json:"Version"`
}
type hostResponse struct {
	Host   string `json:"Host"`
	Status string `json:"Status"`
}

type deviceTokenRequest struct {
	Code       string `json:"code"`
	DeviceDesc string `json:"deviceDesc"`
	DeviceId   string `json:"deviceID"`
}
