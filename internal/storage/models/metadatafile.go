package models

import (
	"errors"
	"strconv"
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
)

// MetadataFile content
type MetadataFile struct {
	DocumentName     string           `json:"visibleName"`
	CollectionType   common.EntryType `json:"type"`
	Parent           string           `json:"parent"`
	LastModified     string           `json:"lastModified"`
	LastOpened       string           `json:"lastOpened"`
	Version          int              `json:"version"`
	Pinned           bool             `json:"pinned"`
	Synced           bool             `json:"synced"`
	Modified         bool             `json:"modified"`
	Deleted          bool             `json:"deleted"`
	MetadataModified bool             `json:"metadatamodified"`
}

func ToTime(timeStamp string) (t time.Time, err error) {
	t = time.Time{}
	lastMod, err := strconv.ParseInt(timeStamp, 10, 64)
	if err != nil {
		return t, err
	}
	fromUnix := time.UnixMilli(lastMod)
	if fromUnix.Year() > 9999 {
		//json panics
		return t, errors.New("incorrect time > 9999")
	}
	return fromUnix, nil
}

func FromTime(t time.Time) string {
	if t.Year() > 9999 {
		return "0"
	}
	return strconv.FormatInt(t.UnixMilli(), 10)
}

type ContentFile struct {
	DummyDocument  bool          `json:"dummyDocument"`
	ExtraMetadata  ExtraMetadata `json:"extraMetadata"`
	FileType       string        `json:"fileType"`
	FontName       string        `json:"fontName"`
	LastOpenedPage int           `json:"lastOpenedPage"`
	LineHeight     int           `json:"lineHeight"`
	Margins        int           `json:"margins"`
	Orientation    string        `json:"orientation"`
	PageCount      int           `json:"pageCount"`
	Pages          []interface{} `json:"pages"`
	TextScale      int           `json:"textScale"`
	Transform      Transform     `json:"transform"`
	SizeInBytes    string        `json:"sizeInBytes"`
}
type ExtraMetadata struct {
	LastPen             string `json:"LastPen"`
	LastTool            string `json:"LastTool"`
	ThicknessScale      string `json:"ThicknessScale"`
	LastFinelinerv2Size string `json:"LastFinelinerv2Size"`
}
type Transform struct {
	M11 int `json:"m11"`
	M12 int `json:"m12"`
	M13 int `json:"m13"`
	M21 int `json:"m21"`
	M22 int `json:"m22"`
	M23 int `json:"m23"`
	M31 int `json:"m31"`
	M32 int `json:"m32"`
	M33 int `json:"m33"`
}
