package viewmodel

import (
	"time"

	"github.com/ddvk/rmfakecloud/internal/messages"
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
	Entries   []Entry
	folderMap map[string]*Directory
}

// Add adds and entry to the tree
func (tree *DocumentTree) Add(d *messages.RawDocument) {
	var entry Entry
	switch d.Type {
	case "CollectionType":
		dir := &Directory{
			ID:   d.ID,
			Name: d.VissibleName,
		}
		entry = dir
		tree.folderMap[d.ID] = dir
	default:
		entry = &Document{
			ID:           d.ID,
			Name:         d.VissibleName,
			DocumentType: d.Type,
		}
	}

	if d.Parent == "" {
		tree.Entries = append(tree.Entries, entry)
	} else {
		folder, ok := tree.folderMap[d.Parent]
		if ok {
			folder.Entries = append(folder.Entries, entry)
		} else {
			//missing parent

		}

	}

}

// Entry just an entry
type Entry interface {
}

// Directory entry
type Directory struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Entries      []Entry `json:"entries"`
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
