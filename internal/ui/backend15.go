package ui

import (
	"io"

	"github.com/zgs225/rmfakecloud/internal/app/hub"
	"github.com/zgs225/rmfakecloud/internal/storage"
	"github.com/zgs225/rmfakecloud/internal/ui/viewmodel"
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

func (b *backend15) Sync(uid string) {
	logrus.Info("notifying")
	b.h.NotifySync(uid, uuid.NewString())
}
