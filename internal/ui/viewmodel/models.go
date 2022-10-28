package viewmodel

import (
	"sort"
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/messages"
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

func makeFolder(d *messages.RawMetadata) (entry *Directory) {
	entry = &Directory{
		ID:   d.ID,
		Name: d.VissibleName,
		// LastModified: d.ModifiedClient,
		Entries: make([]Entry, 0),
	}
	return
}
func makeDocument(d *messages.RawMetadata) (entry Entry) {
	entry = &Document{
		ID:   d.ID,
		Name: d.VissibleName,
		// LastModified: d.ModifiedClient,
		DocumentType: d.Type,
		Extension:    d.Extension,
	}
	return
}

const TrashID = "trash"

// DocTreeFromHashTree from hash tree
func DocTreeFromHashTree(tree *models.HashTree) *DocumentTree {
	docs := make([]*InternalDoc, 0)
	for _, d := range tree.Docs {
		md := &messages.RawMetadata{
			ID:     d.EntryName,
			Parent: d.MetadataFile.Parent,
			Name:   d.MetadataFile.DocumentName,
			Type:   d.MetadataFile.CollectionType,
		}

		isEpub := false
		isPDF := false

		for _, entry := range d.Files {
			if entry.EntryName == md.ID+models.EpubFileExt {
				isEpub = true
				break
			}
		}

		for _, entry := range d.Files {
			if !isEpub {
				if entry.EntryName == md.ID+models.PdfFileExt {
					isPDF = true
					break
				}
			}

		}

		if isEpub {
			md.Extension = models.EpubFileExt
		}

		if isPDF {
			md.Extension = models.PdfFileExt
		}

		docs = append(docs, md)

	}

	return DocTreeFromRawMetadata(docs)
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

		if parent == TrashID {
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
	ID           string `json:"id"`
	Name         string `json:"name"`
	DocumentType string `json:"type"` //notebook, pdf, epub
	Extension    string `json:"extension"`
	LastModified time.Time
	Size         int
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
	SetParentToRoot bool   `json:"setParentToRoot"`
	ParentID        string `json:"parentId"`
	Name            string `json:"name"`
}

type NewFolder struct {
	Name     string `json:"name" binding:"required"`
	ParentID string `json:"parentId"`
}
