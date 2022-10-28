package ui

import (
	"errors"
	"io"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type backend15 struct {
	blobHandler blobHandler
	h           *hub.Hub
}

func (b *backend15) GetDocumentTree(uid string) (tree *viewmodel.DocumentTree, err error) {
	hashTree, err := b.blobHandler.GetTree(uid)
	if err != nil {
		return nil, err
	}

	return viewmodel.DocTreeFromHashTree(hashTree), nil
}
func (b *backend15) Export(uid, docid, exporttype string, opt storage.ExportOption) (r io.ReadCloser, err error) {
	r, err = b.blobHandler.Export(uid, docid)
	return
}

func (b *backend15) CreateDocument(uid, filename, parent string, stream io.Reader) (doc *storage.Document, err error) {
	doc, err = b.blobHandler.CreateBlobDocument(uid, filename, parent, stream)
	return
}

func (d *backend15) DeleteDocument(uid, docid string) error {
	tree, err := d.blobHandler.GetTree(uid)

	if err != nil {
		return err
	}

	hashDoc, err := tree.FindDoc(docid)

	if err != nil {
		return err
	}

	md := hashDoc.MetadataFile

	// Confirm no child before remove a folder
	if md.CollectionType == models.CollectionType {
		docTree := viewmodel.DocTreeFromHashTree(tree)

		// O(n)
		for _, entry := range docTree.Entries {
			dir, ok := entry.(*viewmodel.Directory)
			if !ok {
				continue
			}
			if dir.ID == docid {
				if len(dir.Entries) > 0 {
					return errors.New("Can't remove non-empty folder")
				}
			}
		}
	}

	md.Deleted = true

	if err := d.blobHandler.UpdateBlobMetadata(uid, docid, &md); err != nil {
		return err
	}

	return nil
}

func (d *backend15) CreateFolder(uid, name, parent string) (*storage.Document, error) {
	return d.blobHandler.CreateBlobFolder(uid, name, parent)
}

func (b *backend15) Sync(uid string) {
	logrus.Info("notifying")
	b.h.NotifySync(uid, uuid.NewString())
}

// RenameDocument rename file and folder, the bool type returns value indicates
// whether updated or not
func (d *backend15) RenameDocument(uid, docId, newName string) (bool, error) {
	metadata, err := d.blobHandler.GetBlobMetadata(uid, docId)

	if err != nil {
		return false, err
	}

	if newName == metadata.DocumentName {
		return false, nil
	}

	metadata.DocumentName = newName

	if err = d.blobHandler.UpdateBlobMetadata(uid, docId, metadata); err != nil {
		return false, err
	}

	return true, nil
}

// MoveDocument move document to a new parent
func (d *backend15) MoveDocument(uid, docId, newParent string) (bool, error) {
	// Check parent
	parentMD, err := d.blobHandler.GetBlobMetadata(uid, newParent)

	if err != nil {
		return false, err
	}

	if parentMD.CollectionType != models.CollectionType {
		return false, errors.New("Parent is not a folder")
	}

	metadata, err := d.blobHandler.GetBlobMetadata(uid, docId)

	if err != nil {
		return false, err
	}

	if metadata.Parent == newParent {
		return false, nil
	}

	metadata.Parent = newParent

	if err = d.blobHandler.UpdateBlobMetadata(uid, docId, metadata); err != nil {
		return false, err
	}

	return true, nil
}
