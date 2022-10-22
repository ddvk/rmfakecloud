package ui

import (
	"io"
	"time"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
	log "github.com/sirupsen/logrus"
)

type backend10 struct {
	documentHandler documentHandler
	h               *hub.Hub
}

func (d *backend10) Sync(uid string) {

}

func (d *backend10) CreateFolder(uid, filename, parent string) (doc *storage.Document, err error) {
	doc, err = d.documentHandler.CreateFolder(uid, filename, parent)
	if err != nil {
		return
	}

	ntf := hub.DocumentNotification{
		ID:      doc.ID,
		Type:    common.CollectionType,
		Version: 1,
		Parent:  parent,
		Name:    doc.Name,
	}
	log.Info(uiLogger, "created folder", doc.ID)
	d.h.Notify(uid, "web", ntf, hub.DocAddedEvent)
	return
}
func (d *backend10) CreateDocument(uid, filename, parent string, stream io.Reader) (doc *storage.Document, err error) {
	doc, err = d.documentHandler.CreateDocument(uid, filename, parent, stream)
	if err != nil {
		return
	}

	ntf := hub.DocumentNotification{
		ID:      doc.ID,
		Type:    common.DocumentType,
		Version: 1,
		Parent:  parent,
		Name:    doc.Name,
	}
	log.Info(uiLogger, "Uploaded document id", doc.ID)
	d.h.Notify(uid, "web", ntf, hub.DocAddedEvent)
	return
}

func (d *backend10) GetDocumentTree(uid string) (tree *viewmodel.DocumentTree, err error) {
	documents, err := d.documentHandler.GetAllMetadata(uid)
	if err != nil {
		return nil, err
	}
	docs := make([]*viewmodel.Doc, 0)
	for _, d := range documents {
		lastMod, _ := time.Parse(time.RFC3339Nano, d.ModifiedClient)
		docs = append(docs, &viewmodel.Doc{
			ID:           d.ID,
			Parent:       d.Parent,
			Name:         d.VissibleName,
			Type:         d.Type,
			FileType:     "TODO",
			LastModified: lastMod,
		})

	}
	return viewmodel.DocTreeFromRawMetadata(docs), nil
}
func (d *backend10) Export(uid, doc, exporttype string, opt storage.ExportOption) (stream io.ReadCloser, err error) {
	return d.documentHandler.ExportDocument(uid, doc, exporttype, opt)
}

func (d *backend10) UpdateDocument(uid, docID, name, parent string) (err error) {
	metadata, err := d.documentHandler.GetMetadata(uid, docID)
	if err != nil {
		return err
	}
	metadata.VissibleName = parent
	metadata.Parent = parent
	return d.documentHandler.UpdateMetadata(uid, metadata)
}
func (d *backend10) DeleteDocument(uid, docID string) (err error) {
	return d.documentHandler.RemoveDocument(uid, docID)
}
