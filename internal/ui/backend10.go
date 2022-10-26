package ui

import (
	"errors"
	"io"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
	log "github.com/sirupsen/logrus"
)

type backend10 struct {
	documentHandler documentHandler
	metadataStore   storage.MetadataStorer
	h               *hub.Hub
}

func (d *backend10) Sync(uid string) {

}

func (d *backend10) CreateDocument(uid, filename, parent string, stream io.Reader) (doc *storage.Document, err error) {
	doc, err = d.documentHandler.CreateDocument(uid, filename, parent, stream)
	if err != nil {
		return
	}

	ntf := hub.DocumentNotification{
		ID:      doc.ID,
		Type:    models.DocumentType,
		Version: 1,
		Parent:  parent,
		Name:    doc.Name,
	}
	log.Info(uiLogger, "Uploaded document id", doc.ID)
	d.h.Notify(uid, "web", ntf, hub.DocAddedEvent)
	return
}

func (d *backend10) DeleteDocument(uid, docid string) error {
	meta, err := d.metadataStore.GetMetadata(uid, docid)

	if err != nil {
		return err
	}

	// Check if document is folder, it must be empty
	if meta.Type == models.CollectionType {
		tree, err := d.GetDocumentTree(uid)
		if err != nil {
			return err
		}

		for _, entry := range tree.Entries {
			dir, ok := entry.(*viewmodel.Directory)
			if !ok {
				continue
			}
			if dir.ID == meta.ID {
				if len(dir.Entries) > 0 {
					return errors.New("can't remove non-empty folder")
				}
			}
		}
	}

	if err = d.documentHandler.RemoveDocument(uid, docid); err != nil {
		return err
	}

	ntf := hub.DocumentNotification{
		ID:      meta.ID,
		Type:    meta.Type,
		Version: meta.Version,
		Parent:  meta.Parent,
		Name:    meta.VissibleName,
	}
	log.Info(uiLogger, "Document deleted: id=", meta.ID, " name=", meta.VissibleName)
	d.h.Notify(uid, "web", ntf, hub.DocDeletedEvent)
	return nil
}

func (d *backend10) GetDocumentTree(uid string) (tree *viewmodel.DocumentTree, err error) {
	documents, err := d.documentHandler.GetAllMetadata(uid)
	if err != nil {
		return nil, err
	}

	return viewmodel.DocTreeFromRawMetadata(documents), nil
}
func (d *backend10) Export(uid, doc, exporttype string, opt storage.ExportOption) (stream io.ReadCloser, err error) {
	return d.documentHandler.ExportDocument(uid, doc, exporttype, opt)
}

func (d *backend10) CreateFolder(uid, name, parent string) (*storage.Document, error) {
	if len(parent) > 0 {
		md, err := d.metadataStore.GetMetadata(uid, parent)
		if err != nil {
			return nil, err
		}
		if md.Type != models.CollectionType {
			return nil, errors.New("Parent is not a folder")
		}
	}

	return d.documentHandler.CreateFolder(uid, name, parent)
}
