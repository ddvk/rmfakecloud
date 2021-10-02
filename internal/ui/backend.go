package ui

import (
	"io"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
	log "github.com/sirupsen/logrus"
)

type backend interface {
	GetDocumentTree(uid string) (tree *viewmodel.DocumentTree, err error)
	Export(uid, doc, exporttype string, opt storage.ExportOption) (stream io.ReadCloser, err error)
	CreateDocument(uid, filename string, stream io.ReadCloser) (doc *storage.Document, err error)
	Sync(uid string)
}
type oldhandler struct {
	documentHandler documentHandler
	h               *hub.Hub
}

func (d *oldhandler) Sync(uid string) {

}

func (d *oldhandler) CreateDocument(uid, filename string, stream io.ReadCloser) (doc *storage.Document, err error) {
	doc, err = d.documentHandler.CreateDocument(uid, filename, stream)
	if err != nil {
		return
	}

	ntf := hub.DocumentNotification{
		ID:      doc.ID,
		Type:    doc.Type,
		Version: -1,
		Parent:  doc.Parent,
		Name:    doc.Name,
	}
	log.Info(uiLogger, "Uploaded document id", doc.ID)
	d.h.Notify(uid, "web", ntf, hub.DocAddedEvent)
	return
}

func (d *oldhandler) GetDocumentTree(uid string) (tree *viewmodel.DocumentTree, err error) {
	documents, err := d.documentHandler.GetAllMetadata(uid)
	if err != nil {
		return nil, err
	}

	return viewmodel.NewTree(documents), nil
}
func (*oldhandler) Export(uid, doc, exporttype string, opt storage.ExportOption) (stream io.ReadCloser, err error) {
	return nil, nil
}

type blobBackend struct {
	documentHandler documentHandler
	h               *hub.Hub
}

func (d *blobBackend) GetDocumentTree(uid string) (tree *viewmodel.DocumentTree, err error) {
	syncTree, err := d.documentHandler.GetTree(uid)
	if err != nil {
		return nil, err
	}

	return viewmodel.NewTreeFromSync(syncTree), nil
}
func (*blobBackend) Export(uid, doc, exporttype string, opt storage.ExportOption) (stream io.ReadCloser, err error) {
	return nil, nil
}

func (d *blobBackend) CreateDocument(uid, filename string, stream io.ReadCloser) (doc *storage.Document, err error) {
	doc, err = d.documentHandler.CreateDocument(uid, filename, stream)
	return
}

func (d *blobBackend) Sync(uid string) {
	d.h.NotifySync(uid, "web")
}
