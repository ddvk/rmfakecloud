package ui

import (
	"io"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/epub"
	"github.com/ddvk/rmfakecloud/internal/ui/methods"
	"github.com/ddvk/rmfakecloud/internal/ui/templates"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
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
	tree = viewmodel.DocTreeFromHashTree(hashTree)
	// Templates: merge builtin with synced (separate section in tree)
	tDir := templates.BuiltinTemplatesDirectory()
	if len(tree.Templates) > 0 {
		if d, ok := tree.Templates[0].(*viewmodel.Directory); ok {
			tDir.Entries = append(tDir.Entries, d.Entries...)
		}
	}
	tree.Templates = []viewmodel.Entry{tDir}
	// Methods: merge builtin with synced (separate section in tree)
	mDir := methods.BuiltinMethodsDirectory()
	if len(tree.Methods) > 0 {
		if d, ok := tree.Methods[0].(*viewmodel.Directory); ok {
			mDir.Entries = append(mDir.Entries, d.Entries...)
		}
	}
	tree.Methods = []viewmodel.Entry{mDir}
	return tree, nil
}
func (b *backend15) Export(uid, docid, exporttype string, opt storage.ExportOption) (r io.ReadCloser, err error) {
	r, err = b.blobHandler.Export(uid, docid)
	return
}

func (b *backend15) GetTemplate(uid, docid string) (r io.ReadCloser, err error) {
	return b.blobHandler.GetTemplate(uid, docid)
}

func (b *backend15) GetDocumentMetadata(uid, docid string) (docType string, hasWritings bool, pageCount int, err error) {
	return b.blobHandler.GetDocumentMetadata(uid, docid)
}

func (b *backend15) ExportPagePNG(uid, docid string, pageNum int) (io.ReadCloser, error) {
	return b.blobHandler.ExportPagePNG(uid, docid, pageNum)
}

func (b *backend15) GetEpubManifest(uid, docid string) (*epub.Manifest, error) {
	return b.blobHandler.GetEpubManifest(uid, docid)
}

func (b *backend15) GetEpubFile(uid, docid, filePath string) (io.ReadCloser, string, error) {
	return b.blobHandler.GetEpubFile(uid, docid, filePath)
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
