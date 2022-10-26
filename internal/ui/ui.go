package ui

import (
	"io"
	"io/fs"
	"log"
	"net/http"
	"path"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
	webui "github.com/ddvk/rmfakecloud/new-ui"
	"github.com/gin-gonic/gin"
)

type backend interface {
	GetDocumentTree(uid string) (tree *viewmodel.DocumentTree, err error)
	Export(uid, doc, exporttype string, opt storage.ExportOption) (stream io.ReadCloser, err error)
	CreateDocument(uid, name, parent string, stream io.Reader) (doc *storage.Document, err error)
	DeleteDocument(uid, docid string) error
	CreateFolder(uid, name, parent string) (*storage.Document, error)
	Sync(uid string)
}
type codeGenerator interface {
	NewCode(string) (string, error)
}

type documentHandler interface {
	CreateDocument(uid, name, parent string, stream io.Reader) (doc *storage.Document, err error)
	CreateFolder(uid, name, parent string) (doc *storage.Document, err error)
	RemoveDocument(uid, docid string) error
	GetAllMetadata(uid string) (do []*messages.RawMetadata, err error)
	ExportDocument(uid, id, format string, exportOption storage.ExportOption) (stream io.ReadCloser, err error)
}

type blobHandler interface {
	GetTree(uid string) (tree *models.HashTree, err error)
	CreateBlobDocument(uid, name, parent string, reader io.Reader) (doc *storage.Document, err error)
	CreateBlobFolder(uid, name, parent string) (*storage.Document, error)
	Export(uid, docid string) (io.ReadCloser, error)
}

// ReactAppWrapper encapsulates an app
type ReactAppWrapper struct {
	fs              http.FileSystem
	imagesFS        http.FileSystem
	libFS           http.FileSystem
	prefix          string
	cfg             *config.Config
	userStorer      storage.UserStorer
	metadataStore   storage.MetadataStorer
	codeConnector   codeGenerator
	h               *hub.Hub
	documentHandler documentHandler
	backend15       backend
	backend10       backend
}

//hack for serving index.html on /
const indexReplacement = "/default"

// New Create a React app
func New(cfg *config.Config,
	userStorer storage.UserStorer,
	codeConnector codeGenerator,
	h *hub.Hub,
	docHandler documentHandler,
	blobHandler blobHandler,
	metadataStore storage.MetadataStorer,
) *ReactAppWrapper {

	sub, err := fs.Sub(webui.Assets, "dist")
	if err != nil {
		panic("not embedded?")
	}
	imagesSub, err := fs.Sub(webui.Assets, "dist/images")
	if err != nil {
		panic("not embedded images?")
	}
	libSub, err := fs.Sub(webui.Assets, "dist/lib")
	if err != nil {
		panic("not embedded lib?")
	}
	staticWrapper := ReactAppWrapper{
		fs:              http.FS(sub),
		imagesFS:        http.FS(imagesSub),
		libFS:           http.FS(libSub),
		prefix:          "/assets",
		cfg:             cfg,
		userStorer:      userStorer,
		metadataStore:   metadataStore,
		codeConnector:   codeConnector,
		h:               h,
		documentHandler: docHandler,
		backend15: &backend15{
			blobHandler: blobHandler,
			h:           h,
		},
		backend10: &backend10{
			documentHandler: docHandler,
			metadataStore:   metadataStore,
			h:               h,
		},
	}
	return &staticWrapper
}

// Open opens a file from the fs (virtual)
func (w ReactAppWrapper) Open(filepath string) (http.File, error) {
	log.Println("Open: ", filepath)
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
