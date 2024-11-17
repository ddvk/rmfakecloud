package ui

import (
	"io"
	"io/fs"
	"net/http"
	"path"
	"time"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
	webui "github.com/ddvk/rmfakecloud/ui"
	"github.com/gin-gonic/gin"
)

type backend interface {
	GetDocumentTree(uid string) (tree *viewmodel.DocumentTree, err error)
	Export(uid, doc, exporttype string, opt storage.ExportOption) (stream io.ReadCloser, err error)
	CreateDocument(uid, name, parent string, stream io.Reader) (doc *storage.Document, err error)
	CreateFolder(uid, name, parent string) (doc *storage.Document, err error)
	UpdateDocument(uid, docID, name, parent string) (err error)
	DeleteDocument(uid, docID string) (err error)
	Sync(uid string)
}
type codeGenerator interface {
	NewCode(string) (string, error)
}

type documentHandler interface {
	CreateDocument(uid, name, parent string, stream io.Reader) (doc *storage.Document, err error)
	CreateFolder(uid, name, parent string) (doc *storage.Document, err error)
	GetAllMetadata(uid string) (documents []*messages.RawMetadata, err error)
	ExportDocument(uid, id, format string, exportOption storage.ExportOption) (stream io.ReadCloser, err error)
	GetMetadata(uid, id string) (*messages.RawMetadata, error)
	UpdateMetadata(uid string, r *messages.RawMetadata) error
	RemoveDocument(uid, docid string) error
}

type blobHandler interface {
	GetCachedTree(uid string) (tree *models.HashTree, err error)
	CreateBlobDocument(uid, name, parent string, reader io.Reader) (doc *storage.Document, err error)
	UpdateBlobDocument(uid, docID, name, parent string) (err error)
	DeleteBlobDocument(uid, docID string) (err error)
	CreateBlobFolder(uid, name, parent string) (doc *storage.Document, err error)
	Export(uid, docid string) (io.ReadCloser, error)
}

type notificationHub interface {
	Deleted(uid, docID string) error
	Added(uid, docID string) error
	Updated(uid, docID string) error
	Sync(uid string) error
}

// ReactAppWrapper encapsulates an app
type ReactAppWrapper struct {
	fs            http.FileSystem
	prefix        string
	cfg           *config.Config
	userStorer    storage.UserStorer
	codeConnector codeGenerator
	h             *hub.Hub
	backends      map[common.SyncVersion]backend
}

// hack for serving index.html on /
const indexReplacement = "/default"



// New Create a React app
func New(cfg *config.Config,
	userStorer storage.UserStorer,
	codeConnector codeGenerator,
	h *hub.Hub,
	docHandler documentHandler,
	blobHandler blobHandler) *ReactAppWrapper {

	sub, err := fs.Sub(webui.Assets, "build")
	if err != nil {
		panic("not embedded?")
	}
	backend15 := &backend15{
		blobHandler: blobHandler,
		h:           h,
	}
	backend10 := &backend10{
		documentHandler: docHandler,
		hub:             h,
	}
	staticWrapper := ReactAppWrapper{
		fs:            common.NewLastModifiedFS(http.FS(sub), time.Now()),
		prefix:        "/static",
		cfg:           cfg,
		userStorer:    userStorer,
		codeConnector: codeConnector,
		h:             h,
		backends: map[common.SyncVersion]backend{
			common.Sync10: backend10,
			common.Sync15: backend15,
		},
	}
	return &staticWrapper
}

// Open opens a file from the fs (virtual)
func (w ReactAppWrapper) Open(filepath string) (http.File, error) {
	fullpath := filepath
	//index.html hack
	if filepath != indexReplacement {
		fullpath = path.Join(w.prefix, filepath)
	} else {
		fullpath = "/index.html"
	}
	f, err := w.fs.Open(fullpath)
	return f, err
}
func badReq(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": message})
}
