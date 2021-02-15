package ui

import (
	"time"

	"github.com/ddvk/rmfakecloud/internal/messages"
)

type loginForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type resetPasswordForm struct {
	Email    string `json:"email"`
	CurrentPassword string `json:"currentPassword"`
	NewPassword string `json:"newPassword"`
}

type DocumentTree struct {
	Entries   []Entry
	folderMap map[string]*Directory
}

func (tree *DocumentTree) Add(d *messages.RawDocument) {
	var entry Entry
	switch d.Type {
	case "CollectionType":
		dir := &Directory{
			ID:   d.Id,
			Name: d.VissibleName,
		}
		entry = dir
		tree.folderMap[d.Id] = dir
	default:
		entry = &Document{
			ID:           d.Id,
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

type Entry interface {
}
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
type user struct {
	ID        string `json:"userid"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt time.Time
}
