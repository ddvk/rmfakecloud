package storage

import (
	"io"
	"time"

	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/gin-gonic/gin"
)

type ExportOption int

const (
	ExportWithAnnotations ExportOption = iota
	ExportOnlyAnnotations
)

// DocumentStorer stores documents
type DocumentStorer interface {
	StoreDocument(uid, docid string, s io.ReadCloser) error
	RemoveDocument(uid, docid string) error
	GetDocument(uid, docid string) (io.ReadCloser, error)
	ExportDocument(uid, docid, outputType string, exportOption ExportOption) (io.ReadCloser, error)

	// GetStorageURL creates a short lived url
	GetStorageURL(uid, docid string) (string, time.Time, error)

	RegisterRoutes(*gin.Engine)
}

// MetadataStorer manages document metadata
type MetadataStorer interface {
	UpdateMetadata(uid string, r *messages.RawDocument) error
	GetAllMetadata(uid string) ([]*messages.RawDocument, error)
	GetMetadata(uid, docid string) (*messages.RawDocument, error)
}

// UserStorer holds informations about users
type UserStorer interface {
	GetUsers() ([]*model.User, error)
	GetUser(string) (*model.User, error)
	RegisterUser(u *model.User) error
	UpdateUser(u *model.User) error
}
