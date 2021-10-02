package model

// MetadataFile content
type MetadataFile struct {
	DocName          string `json:"visibleName"`
	CollectionType   string `json:"type"`
	Parent           string `json:"parent"`
	LastModified     string `json:"lastModified"`
	LastOpened       string `json:"lastOpened"`
	Version          int    `json:"version"`
	Pinned           bool   `json:"pinned"`
	Synced           bool   `json:"synced"`
	Modified         bool   `json:"modified"`
	Deleted          bool   `json:"deleted"`
	MetadataModified bool   `json:"metadatamodified"`
}
