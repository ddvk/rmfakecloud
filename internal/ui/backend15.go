package ui

import (
	"io"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
)

type backend15 struct {
	blobHandler blobHandler
	h           *hub.Hub
}

func (d *backend15) GetDocumentTree(uid string) (tree *viewmodel.DocumentTree, err error) {
	syncTree, err := d.blobHandler.GetTree(uid)
	if err != nil {
		return nil, err
	}

	return viewmodel.NewTreeFromSync(syncTree), nil
}
func (*backend15) Export(uid, doc, exporttype string, opt storage.ExportOption) (stream io.ReadCloser, err error) {
	return nil, nil
}

func (d *backend15) CreateDocument(uid, filename string, stream io.Reader) (doc *storage.Document, err error) {
	doc, err = d.blobHandler.CreateBlobDocument(uid, filename, stream)
	return
}

func (d *backend15) Sync(uid string) {
	d.h.NotifySync(uid, "web")
}
