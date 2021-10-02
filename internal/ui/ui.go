package ui

import (
	"io"
	"net/http"
	"path"

	"github.com/ddvk/rmfakecloud/internal/app/hub"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/fs/sync15"
	"github.com/ddvk/rmfakecloud/internal/webassets"
	"github.com/gin-gonic/gin"
)

type codeGenerator interface {
	NewCode(string) (string, error)
}

type documentHandler interface {
	CreateDocument(uid, filename string, stream io.ReadCloser) (doc *storage.Document, err error)
	GetAllMetadata(uid string) (do []*messages.RawDocument, err error)
	ExportDocument(uid, id, format string, exportOption storage.ExportOption) (stream io.ReadCloser, err error)

	GetTree(uid string) (tree *sync15.HashTree, err error)
	// CreateDocument15(uid, filename string, stream io.ReadCloser) (doc *storage.Document, err error)
}

// ReactAppWrapper wrap some stuff
type ReactAppWrapper struct {
	fs              http.FileSystem
	prefix          string
	cfg             *config.Config
	userStorer      storage.UserStorer
	codeConnector   codeGenerator
	h               *hub.Hub
	documentHandler documentHandler
	blobbackend     backend
	oldbbackend     backend
}

const indexReplacement = "/default"

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

// New Create a React app
func New(cfg *config.Config, userStorer storage.UserStorer,
	codeConnector codeGenerator, h *hub.Hub,
	docHandler documentHandler) *ReactAppWrapper {
	staticWrapper := ReactAppWrapper{
		fs:              webassets.Assets,
		prefix:          "/static",
		cfg:             cfg,
		userStorer:      userStorer,
		codeConnector:   codeConnector,
		h:               h,
		documentHandler: docHandler,
		blobbackend: &blobBackend{
			documentHandler: docHandler,
			h:               h,
		},
		oldbbackend: &oldhandler{
			documentHandler: docHandler,
			h:               h,
		},
	}
	return &staticWrapper
}
