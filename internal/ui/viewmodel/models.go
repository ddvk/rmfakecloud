package viewmodel

import (
	"sort"
	"strconv"
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

// DocumentTree a tree of documents
type DocumentTree struct {
	Entries []Entry
	Trash   []Entry
}

type Doc struct {
	ID           string           `json:"documentId"`
	Version      int              `json:"version"`
	LastModified time.Time        `json:"lastModified"`
	Type         common.EntryType `json:"Type"`
	FileType     string           `json:"fileType"`
	Name         string           `json:"Name"`
	CurrentPage  int              `json:"CurrentPage"`
	Parent       string           `json:"Parent"`
	Size         int
}

func makeFolder(d *Doc) (entry *Directory) {
	entry = &Directory{
		ID:           d.ID,
		Name:         d.Name,
		LastModified: d.LastModified,
		Entries:      make([]Entry, 0),
		IsFolder:     true,
	}
	return
}
func makeDocument(d *Doc) (entry Entry) {
	entry = &Document{
		ID:           d.ID,
		Name:         d.Name,
		LastModified: d.LastModified,
		DocumentType: d.FileType,
		Size:         d.Size,
	}
	return
}

// DocTreeFromHashTree from hash tree
func DocTreeFromHashTree(tree *models.HashTree) *DocumentTree {
	docs := make([]*Doc, 0)
	for _, d := range tree.Docs {
		lastMod, _ := strconv.ParseInt(d.LastModified, 10, 64)
		unixTime := time.UnixMilli(lastMod)
		docs = append(docs, &Doc{
			ID:           d.EntryName,
			Parent:       d.MetadataFile.Parent,
			Name:         d.MetadataFile.DocumentName,
			Type:         d.MetadataFile.CollectionType,
			LastModified: unixTime,
		})
	}

	return DocTreeFromRawMetadata(docs)
}

// DocTreeFromRawMetadata from raw metadata
func DocTreeFromRawMetadata(documents []*Doc) *DocumentTree {
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
		Entries: rootEntries,
		Trash:   trashEntries,
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
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	DocumentType string    `json:"type"` //notebook, pdf, epub
	LastModified time.Time `json:"lastModified"`
	Size         int       `json:"size"`
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
