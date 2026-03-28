package viewmodel

import (
	"sort"
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	log "github.com/sirupsen/logrus"
)

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

// ErrorResponse
type ErrorResponse struct {
	Error string `json:"error"`
}
func NewErrorResponse(errormsg string) ErrorResponse {
	return ErrorResponse {
		Error: errormsg,
	}
}

// RegisteredDeviceEntry is a safe JSON view of a paired tablet (no secrets).
type RegisteredDeviceEntry struct {
	DeviceID     string `json:"deviceId"`
	DeviceDesc   string `json:"deviceDesc"`
	RegisteredAt string `json:"registeredAt,omitempty"`
	LastSeen     string `json:"lastSeen,omitempty"`
}

// RegisteredDevicesResponse lists devices for the logged-in user.
type RegisteredDevicesResponse struct {
	Devices []RegisteredDeviceEntry `json:"devices"`
}

// DocumentTree a tree of documents
type DocumentTree struct {
	Entries   []Entry
	Trash     []Entry
	Templates []Entry // [ Directory ] for frontend; optional
	Methods   []Entry // [ Directory ] for frontend; synced + builtin merged in backend
}

type InternalDoc struct {
	ID           string
	Version      int
	LastModified time.Time
	Type         common.EntryType
	FileType     string
	Name         string
	CurrentPage  int
	Parent       string
	Size         int64
	HasWritings  bool
	Orientation  string // from .content: "portrait", "landscape", or ""
	Pinned       bool   // starred in metadata
}

func makeFolder(d *InternalDoc) (entry *Directory) {
	entry = &Directory{
		ID:           d.ID,
		Name:         d.Name,
		LastModified: d.LastModified,
		Entries:      make([]Entry, 0),
		IsFolder:     true,
	}
	return
}
func makeDocument(d *InternalDoc) (entry Entry) {
	entry = &Document{
		ID:           d.ID,
		Name:         d.Name,
		LastModified: d.LastModified,
		DocumentType: d.FileType,
		Collection:   d.Type,
		Size:         d.Size,
		HasWritings:  d.HasWritings,
		Orientation:  d.Orientation,
		Pinned:       d.Pinned,
	}
	return
}

// methodsSource is the reMarkable metadata source for rm Methods.
const methodsSource = "com.remarkable.methods"

// templateType is the collection type for Templates (device/synced templates).
const templateType = common.EntryType("TemplateType")

// fileTypeFromDoc returns document type by file extension, .content fileType (when size > 4), or "notebook".
func fileTypeFromDoc(d *models.HashDoc) string {
	if t := d.PayloadTypeFromFiles(); t != "" {
		return t
	}
	if d.PayloadType != "" {
		return d.PayloadType
	}
	return "notebook"
}

// DocTreeFromHashTree from hash tree. Templates and Methods are separated into their own sections.
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
		ft := fileTypeFromDoc(d)
		internalDoc := &InternalDoc{
			ID:           d.EntryName,
			Parent:       d.MetadataFile.Parent,
			Name:         d.MetadataFile.DocumentName,
			Type:         d.MetadataFile.CollectionType,
			LastModified: lastModified,
			FileType:     ft,
			Size:         d.Size,
			HasWritings:  d.HasWritings(),
			Pinned:       d.MetadataFile.Pinned,
		}
		if d.MetadataFile.Source == methodsSource {
			methodDocs = append(methodDocs, internalDoc)
			continue
		}
		if d.MetadataFile.CollectionType == templateType {
			templateDocs = append(templateDocs, internalDoc)
			continue
		}
		docs = append(docs, internalDoc)
	}
	dt := DocTreeFromRawMetadata(docs)
	dt.Templates = templateEntriesToDirectory(templateDocs)
	dt.Methods = methodEntriesToDirectory(methodDocs)
	return dt
}

// templateEntriesToDirectory returns a single Directory (as []Entry) for synced template documents.
func templateEntriesToDirectory(templateDocs []*InternalDoc) []Entry {
	children := make([]Entry, 0, len(templateDocs))
	for _, d := range templateDocs {
		children = append(children, makeDocument(d))
	}
	dir := &Directory{
		ID:           "templates",
		Name:         "Templates",
		Entries:      children,
		LastModified: time.Time{},
		IsFolder:     true,
	}
	return []Entry{dir}
}

// methodEntriesToDirectory returns a single Directory (as []Entry) for method documents with type by extension.
func methodEntriesToDirectory(methodDocs []*InternalDoc) []Entry {
	children := make([]Entry, 0, len(methodDocs))
	for _, d := range methodDocs {
		children = append(children, makeDocument(d))
	}
	dir := &Directory{
		ID:           "methods",
		Name:         "rm Methods",
		Entries:      children,
		LastModified: time.Time{},
		IsFolder:     true,
	}
	return []Entry{dir}
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
}

// Document is a single document
type Document struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	DocumentType string             `json:"type"` // notebook, pdf, epub
	Collection   common.EntryType   `json:"collectionType"`
	LastModified time.Time          `json:"lastModified"`
	Size         int64              `json:"size"`
	HasWritings  bool               `json:"hasWritings"`
	Orientation  string             `json:"orientation,omitempty"` // from .content: "portrait", "landscape", or empty if both/unspecified
	Pinned       bool               `json:"pinned,omitempty"`
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
	IsAdmin 	 bool `json:"isAdmin"`
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
