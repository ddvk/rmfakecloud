package viewmodel

import (
	"sort"
	"time"

	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/fs/sync15"
	log "github.com/sirupsen/logrus"
)

// LoginForm the login form
type LoginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ResetPasswordForm reset password
type ResetPasswordForm struct {
	Email           string `json:"email"`
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// DocumentTree a tree of documents
type DocumentTree struct {
	Entries []Entry
	Trash   []Entry
}

func makeFolder(d *messages.RawDocument) (entry *Directory) {
	entry = &Directory{
		ID:   d.ID,
		Name: d.VissibleName,
		// LastModified: d.ModifiedClient,
		Entries: make([]Entry, 0),
	}
	return
}
func makeDocument(d *messages.RawDocument) (entry Entry) {
	entry = &Document{
		ID:   d.ID,
		Name: d.VissibleName,
		// LastModified: d.ModifiedClient,
		DocumentType: d.Type,
	}
	return
}

const trashId = "trash"

func NewTreeFromSync(tree *sync15.HashTree) *DocumentTree {
	docs := make([]*messages.RawDocument, 0)
	for _, d := range tree.Docs {
		docs = append(docs, &messages.RawDocument{
			ID:           d.DocumentID,
			Parent:       d.MetadataFile.Parent,
			VissibleName: d.MetadataFile.DocName,
			Type:         d.MetadataFile.CollectionType,
		})

	}

	return NewTree(docs)
}
func NewTree(documents []*messages.RawDocument) *DocumentTree {
	childParent := make(map[string]string)
	folders := make(map[string]*Directory)
	rootEntries := make([]Entry, 0)
	trashEntries := make([]Entry, 0)

	sort.Slice(documents, func(i, j int) bool {
		a, b := documents[i], documents[j]
		if a.Type != b.Type {
			return a.Type == storage.CollectionType
		}

		return a.VissibleName < b.VissibleName
	})

	// add all folders
	for _, d := range documents {
		switch d.Type {
		case storage.CollectionType:
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

		parentId := d.Parent

		if parentId == trashId {
			trashEntries = append(trashEntries, entry)
			continue
		}

		if parentId == "" {
			// empty parent = root
			rootEntries = append(rootEntries, entry)
			continue
		}

		if parent, ok := folders[parentId]; ok {

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

		log.Warn("parent not found: ", parentId)
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
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Entries      []Entry `json:"children"`
	LastModified time.Time
}

// Document is a single document
type Document struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DocumentType string `json:"type"` //notebook, pdf, epub
	LastModified time.Time
	Size         int
}

// DocumentList is a list of documents
type DocumentList struct {
	Documents []Document `json:"entries"`
}

// User user model
type User struct {
	ID        string `json:"userid"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt time.Time
}
