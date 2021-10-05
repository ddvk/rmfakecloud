package storage

import (
	"errors"
	"io"
	"time"

	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/model"
)

// ExportOption type of export
type ExportOption int

const (
	ExportWithAnnotations ExportOption = iota
	ExportOnlyAnnotations
)

// ErrorNotFound not found
var ErrorNotFound = errors.New("not found")

// ErrorWrongGeneration the geration did not match
var ErrorWrongGeneration = errors.New("wrong generation")

// DocumentStorer stores documents
type DocumentStorer interface {
	StoreDocument(uid, docid string, s io.ReadCloser) error
	RemoveDocument(uid, docid string) error
	GetDocument(uid, docid string) (io.ReadCloser, error)
	ExportDocument(uid, docid, outputType string, exportOption ExportOption) (io.ReadCloser, error)

	GetStorageURL(uid, docid string) (string, time.Time, error)
}

// BlobStorage stuff for sync15
type BlobStorage interface {
	GetBlobURL(uid, docid string) (string, time.Time, error)

	StoreBlob(uid, blobID string, s io.Reader, matchGeneration int64) (int64, error)
	LoadBlob(uid, blobID string) (io.ReadCloser, int64, error)
}

// MetadataStorer manages document metadata
type MetadataStorer interface {
	UpdateMetadata(uid string, r *messages.RawMetadata) error
	GetAllMetadata(uid string) ([]*messages.RawMetadata, error)
	GetMetadata(uid, docid string) (*messages.RawMetadata, error)
}

// UserStorer holds informations about users
type UserStorer interface {
	GetUsers() ([]*model.User, error)
	GetUser(string) (*model.User, error)
	RegisterUser(u *model.User) error
	UpdateUser(u *model.User) error
}

// Document represents a document in storage
type Document struct {
	ID      string
	Type    string
	Parent  string
	Name    string
	Version int
}
