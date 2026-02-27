// Package archive contains types for parsing reMarkable archive files
package archive

import (
	"github.com/ddvk/rmfakecloud/internal/encoding/rm"
)

// Set the default pagedata template to Blank
const defaultPagadata string = "Blank"

// Zip represents an entire Remarkable archive file.
type Zip struct {
	Content Content
	Pages   []Page
	Payload []byte
	UUID    string
	pageMap map[string]int
}

// NewZip creates a File with sane defaults.
func NewZip() *Zip {
	content := Content{
		DummyDocument: false,
		ExtraMetadata: ExtraMetadata{
			LastBrushColor:           "Black",
			LastBrushThicknessScale:  "2",
			LastColor:                "Black",
			LastEraserThicknessScale: "2",
			LastEraserTool:           "Eraser",
			LastPen:                  "Ballpoint",
			LastPenColor:             "Black",
			LastPenThicknessScale:    "2",
			LastPencil:               "SharpPencil",
			LastPencilColor:          "Black",
			LastPencilThicknessScale: "2",
			LastTool:                 "SharpPencil",
			ThicknessScale:           "2",
			LastFinelinerv2Size:      "1",
		},
		FileType:       "",
		FontName:       "",
		LastOpenedPage: 0,
		LineHeight:     -1,
		Margins:        100,
		Orientation:    "portrait",
		PageCount:      0,
		Pages:          []string{},
		TextScale:      1,
		Transform: Transform{
			M11: 1,
			M12: 0,
			M13: 0,
			M21: 0,
			M22: 1,
			M23: 0,
			M31: 0,
			M32: 0,
			M33: 1,
		},
	}

	return &Zip{
		Content: content,
	}
}

// A Page represents a note page.
type Page struct {
	// Data is the rm binary encoded file representing the drawn content
	Data *rm.Rm
	// Metadata is a json file containing information about layers
	Metadata Metadata
	// Thumbnail is a small image of the overall page
	Thumbnail []byte
	// Pagedata contains the name of the selected background template
	Pagedata string
	// page number of the underlying document
	DocPage int
}

// Metadata represents the structure of a .metadata json file associated to a page.
type Metadata struct {
	Layers []Layer `json:"layers"`
}

// Layers is a struct contained into a Metadata struct.
type Layer struct {
	Name string `json:"name"`
}

// Content represents the structure of a .content json file.
type Content struct {
	DummyDocument bool          `json:"dummyDocument"`
	ExtraMetadata ExtraMetadata `json:"extraMetadata"`

	// FileType is "pdf", "epub" or empty for a simple note
	FileType       string `json:"fileType"`
	FontName       string `json:"fontName"`
	LastOpenedPage int    `json:"lastOpenedPage"`
	LineHeight     int    `json:"lineHeight"`
	Margins        int    `json:"margins"`
	// Orientation can take "portrait" or "landscape".
	Orientation string `json:"orientation"`
	PageCount   int    `json:"pageCount"`
	// Pages is a list of page IDs
	Pages          []string `json:"pages"`
	Tags           []string `json:"pageTags"`
	RedirectionMap []int    `json:"redirectionPageMap"`
	TextScale      int      `json:"textScale"`

	Transform Transform `json:"transform"`
}

// ExtraMetadata is a struct contained into a Content struct.
type ExtraMetadata struct {
	LastBrushColor           string `json:"LastBrushColor"`
	LastBrushThicknessScale  string `json:"LastBrushThicknessScale"`
	LastColor                string `json:"LastColor"`
	LastEraserThicknessScale string `json:"LastEraserThicknessScale"`
	LastEraserTool           string `json:"LastEraserTool"`
	LastPen                  string `json:"LastPen"`
	LastPenColor             string `json:"LastPenColor"`
	LastPenThicknessScale    string `json:"LastPenThicknessScale"`
	LastPencil               string `json:"LastPencil"`
	LastPencilColor          string `json:"LastPencilColor"`
	LastPencilThicknessScale string `json:"LastPencilThicknessScale"`
	LastTool                 string `json:"LastTool"`
	ThicknessScale           string `json:"ThicknessScale"`
	LastFinelinerv2Size      string `json:"LastFinelinerv2Size"`
}

// Transform is a struct contained into a Content struct.
type Transform struct {
	M11 float32 `json:"m11"`
	M12 float32 `json:"m12"`
	M13 float32 `json:"m13"`
	M21 float32 `json:"m21"`
	M22 float32 `json:"m22"`
	M23 float32 `json:"m23"`
	M31 float32 `json:"m31"`
	M32 float32 `json:"m32"`
	M33 float32 `json:"m33"`
}

// MetadataFile content
type MetadataFile struct {
	DocName        string `json:"visibleName"`
	CollectionType string `json:"type"`
	Parent         string `json:"parent"`
	//LastModified in milliseconds
	LastModified     string `json:"lastModified"`
	LastOpened       string `json:"lastOpened"`
	LastOpenedPage   int    `json:"lastOpenedPage"`
	Version          int    `json:"version"`
	Pinned           bool   `json:"pinned"`
	Synced           bool   `json:"synced"`
	Modified         bool   `json:"modified"`
	Deleted          bool   `json:"deleted"`
	MetadataModified bool   `json:"metadatamodified"`
}
