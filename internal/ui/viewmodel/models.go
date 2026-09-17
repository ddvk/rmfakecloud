package viewmodel

import (
	"sort"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	log "github.com/sirupsen/logrus"
)

func hashDocHasWritings(d *models.HashDoc) bool {
	if d == nil {
		return false
	}
	for _, f := range d.Files {
		if f != nil && strings.HasSuffix(strings.ToLower(f.EntryName), storage.RmFileExt) {
			return true
		}
	}
	return false
}

const trashID = "trash"

// LoginForm the login form
type LoginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ResetPasswordForm reset password
type ResetPasswordForm struct {
	UserID          string `json:"userid"`
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// ChangeEmail reset password
type ChangeEmailForm struct {
	UserID          string `json:"userid"`
	Email           string `json:"email"`
	CurrentPassword string `json:"currentPassword"`
}

// RegisteredDeviceEntry is a safe JSON view of a paired tablet (no secrets).
type RegisteredDeviceEntry struct {
	DeviceID     string `json:"deviceId"`
	DeviceDesc   string `json:"deviceDesc"`
	DeviceLink   string `json:"deviceLink,omitempty"`
	Make         string `json:"make,omitempty"`
	Model        string `json:"model,omitempty"`
	Year         string `json:"year,omitempty"`
	RegisteredAt string `json:"registeredAt,omitempty"`
	LastSeen     string `json:"lastSeen,omitempty"`
}

// RegisteredDevicesResponse lists devices for the logged-in user.
type RegisteredDevicesResponse struct {
	Devices []RegisteredDeviceEntry `json:"devices"`
}

// ReissueDeviceRequest asks for a new device JWT without a pairing code (web session only).
type ReissueDeviceRequest struct {
	DeviceID   string `json:"deviceId" binding:"required"`
	DeviceDesc string `json:"deviceDesc"`
	DeviceLink string `json:"deviceLink,omitempty"`
}

// ReissueDeviceResponse returns the raw device token for the tablet.
type ReissueDeviceResponse struct {
	Token string `json:"token"`
}

// ErrorResponse
type ErrorResponse struct {
	Error string `json:"error"`
}

func NewErrorResponse(errormsg string) ErrorResponse {
	return ErrorResponse{
		Error: errormsg,
	}
}

// DocumentTree a tree of documents
type DocumentTree struct {
	Entries   []Entry
	Trash     []Entry
	Templates []Entry // template documents (TemplateType)
	Methods   []Entry // rm Methods (source com.remarkable.methods)
}

type InternalDoc struct {
	ID           string
	Version      int
	LastModified time.Time
	Type         common.EntryType
	FileType     string
	FormatLabel  string
	HasWritings  bool
	Name         string
	CurrentPage  int
	Parent       string
	Size         int64
	PageCount    int
	Pinned       bool
}

func makeFolder(d *InternalDoc) (entry *Directory) {
	entry = &Directory{
		ID:           d.ID,
		Name:         d.Name,
		LastModified: d.LastModified,
		Entries:      make([]Entry, 0),
		IsFolder:     true,
		Pinned:       d.Pinned,
	}
	return
}
func makeDocument(d *InternalDoc) (entry Entry) {
	entry = &Document{
		ID:           d.ID,
		Name:         d.Name,
		LastModified: d.LastModified,
		DocumentType: d.FileType,
		FormatLabel:  d.FormatLabel,
		HasWritings:  d.HasWritings,
		Size:         d.Size,
		CurrentPage:  d.CurrentPage,
		PageCount:    d.PageCount,
		Pinned:       d.Pinned,
	}
	return
}

// DocTreeFromHashTree from hash tree. Templates and Methods are separated into their own sections.
// Methods: metadata type TemplateType + source com.remarkable.methods.
// Templates: metadata type TemplateType without that source.
func DocTreeFromHashTree(tree *models.HashTree) *DocumentTree {
	docs := make([]*InternalDoc, 0)
	templateDocs := make([]*InternalDoc, 0)
	methodDocs := make([]*InternalDoc, 0)
	for _, d := range tree.Docs {
		if d.Deleted {
			continue
		}

		lastModified, err := models.ToTime(d.LastModified)
		if err != nil {
			log.Warn("incorrect lastmodified for: ", d.DocumentName, " value: ", d.LastModified, " ", err)
		}
		fileType := d.LibraryFileType()
		internalDoc := &InternalDoc{
			ID:           d.EntryName,
			Parent:       d.MetadataFile.Parent,
			Name:         d.MetadataFile.DocumentName,
			Type:         d.MetadataFile.CollectionType,
			LastModified: lastModified,
			FileType:     fileType,
			FormatLabel:  d.FormatLabel,
			HasWritings:  hashDocHasWritings(d),
			Size:         d.Size,
			CurrentPage:  d.LastOpenedPage,
			PageCount:    d.PageCount,
			Pinned:       d.Pinned,
		}
		switch {
		case d.MetadataFile.IsMethod():
			methodDocs = append(methodDocs, internalDoc)
		case d.MetadataFile.IsTemplate():
			templateDocs = append(templateDocs, internalDoc)
		default:
			docs = append(docs, internalDoc)
		}
	}

	dt := DocTreeFromRawMetadata(docs)
	dt.Templates = flatDocsAsEntries(templateDocs)
	dt.Methods = flatDocsAsEntries(methodDocs)
	return dt
}

func flatDocsAsEntries(docs []*InternalDoc) []Entry {
	out := make([]Entry, 0, len(docs))
	for _, d := range docs {
		out = append(out, makeDocument(d))
	}
	return out
}

// DocTreeFromRawMetadata from raw metadata
func DocTreeFromRawMetadata(documents []*InternalDoc) *DocumentTree {
	childParent := map[string]string{}
	folders := map[string]*Directory{}
	rootEntries := make([]Entry, 0)
	trashEntries := make([]Entry, 0)

	sort.Slice(documents, func(i, j int) bool {
		a, b := documents[i], documents[j]
		if a.Type != b.Type {
			return a.Type == common.CollectionType
		}

		return a.Name < b.Name
	})

	// add all folders
	for _, d := range documents {
		switch d.Type {
		case common.CollectionType:
			folders[d.ID] = makeFolder(d)
		}
	}

	// create parent child relationships
	for _, d := range documents {
		var entry Entry
		var ok bool

		// look it up in folders fist
		if entry, ok = folders[d.ID]; !ok {
			entry = makeDocument(d)
		}

		parent := d.Parent

		if parent == trashID {
			trashEntries = append(trashEntries, entry)
			continue
		}

		if parent == "" {
			// empty parent = root
			rootEntries = append(rootEntries, entry)
			continue
		}

		if parent, ok := folders[parent]; ok {

			//check for  loops and cross adds (a->b->c  c->a)
			// if parentId, ok := childParent[parentId]; ok {
			// 	//todo forloop
			// 	if parentId == d.ID {
			// 		log.Warn("loop detected: ", parentId, " -> ", d.ID)
			// 		rootEntries = append(rootEntries, entry)
			// 		continue
			// 	}
			// } else {
			// }

			parent.Entries = append(parent.Entries, entry)
			childParent[d.ID] = d.Parent
			continue
		}

		log.Warn(d.Name, " parent not found: ", parent)
		rootEntries = append(rootEntries, entry)
	}

	tree := DocumentTree{
		Entries:   rootEntries,
		Trash:     trashEntries,
		Templates: nil,
		Methods:   nil,
	}

	return &tree
}

// Entry just an entry
type Entry interface {
}

// Directory entry
type Directory struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Entries      []Entry   `json:"children"`
	LastModified time.Time `json:"lastModified"`
	IsFolder     bool      `json:"isFolder"`
	Pinned       bool      `json:"pinned"`
}

// Document is a single document
type Document struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	DocumentType string    `json:"type"` //notebook, pdf, epub
	FormatLabel  string    `json:"formatLabel,omitempty"`
	HasWritings  bool      `json:"hasWritings,omitempty"`
	LastModified time.Time `json:"lastModified"`
	Size         int64     `json:"size"`
	CurrentPage  int       `json:"currentPage"`
	PageCount    int       `json:"pageCount"`
	Pinned       bool      `json:"pinned"`
}

// DocumentList is a list of documents
type DocumentList struct {
	Documents []Document `json:"entries"`
}

// User user model
type User struct {
	ID           string `json:"userid"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	NewPassword  string `json:"newpassword,omitempty"`
	IsAdmin      bool   `json:"isAdmin"`
	CreatedAt    time.Time
	Integrations []string `json:"integrations,omitempty"`
}

// NewUser new user creation
type NewUser struct {
	ID          string `json:"userid" binding:"required"`
	Email       string `json:"email" binding:"email"`
	NewPassword string `json:"newpassword" binding:"required"`
}

// UpdateDoc with somethin
type UpdateDoc struct {
	DocumentID string `json:"documentId" binding:"required"`
	ParentID   string `json:"parentId"`
	Name       string `json:"name"`
}

type NewFolder struct {
	ParentID string `json:"parentId"`
	Name     string `json:"name"`
}
