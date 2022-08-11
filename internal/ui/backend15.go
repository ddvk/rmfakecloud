package ui

import (
	"io"

	"github.com/zgs225/rmfakecloud/internal/app/hub"
	"github.com/zgs225/rmfakecloud/internal/storage"
	"github.com/zgs225/rmfakecloud/internal/ui/viewmodel"
	"github.com/google/uuid"
)

type backend15 struct {
	blobHandler blobHandler
	h           *hub.Hub
}

func (b *backend15) GetDocumentTree(uid string) (tree *viewmodel.DocumentTree, err error) {
	hashTree, err := b.blobHandler.GetCachedTree(uid)
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

func (b *backend15) UpdateDocument(uid, docID, name, parent string) (err error) {
	return b.blobHandler.UpdateBlobDocument(uid, docID, name, parent)
}
func (b *backend15) CreateFolder(uid, name, parent string) (doc *storage.Document, err error) {
	return b.blobHandler.CreateBlobFolder(uid, name, parent)
}

func (b *backend15) DeleteDocument(uid, docID string) (err error) {
	return b.blobHandler.DeleteBlobDocument(uid, docID)
}

func (b *backend15) Sync(uid string) {
	b.h.NotifySync(uid, uuid.NewString())
}
