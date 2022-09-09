package ui

import (
	"io"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
	log "github.com/sirupsen/logrus"
)

type backend10 struct {
	documentHandler documentHandler
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
