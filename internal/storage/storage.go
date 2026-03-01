package storage

import (
	"io"
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
)

// ExportOption type of export
type ExportOption int

const (
	ExportWithAnnotations ExportOption = iota
	ExportOnlyAnnotations
	ExportPayload
)

// DocumentStorer stores documents
type DocumentStorer interface {
	StoreDocument(uid, docid string, s io.ReadCloser) error
	RemoveDocument(uid, docid string) error
	GetDocument(uid, docid string) (io.ReadCloser, error)
	ExportDocument(uid, docid, outputType string, exportOption ExportOption) (io.ReadCloser, error)

	GetStorageURL(uid, docid string) (string, time.Time, error)
	CreateDocument(uid, name, parent string, stream io.Reader) (doc *Document, err error)
	CreateFolder(uid, name, parent string) (doc *Document, err error)
}

// BlobStorage stuff for sync15
type BlobStorage interface {
	GetBlobURL(uid, docid string, write bool) (string, time.Time, error)

	StoreBlob(uid, blobID string, filename string, hash string, s io.Reader) error
	LoadBlob(uid, blobID string) (reader io.ReadCloser, size int64, crc32c string, err error)
}

// MetadataStorer manages document metadata
type MetadataStorer interface {
	UpdateMetadata(uid string, r *messages.RawMetadata) error
	GetAllMetadata(uid string) ([]*messages.RawMetadata, error)
	GetMetadata(uid, docid string) (*messages.RawMetadata, error)
}

// RootStorer holds informations about users
type RootStorer interface {
	GetCachedTree(uid string, rs models.RemoteStorage) (t *models.HashTree, err error)
	GetRootIndex(uid string) (string, int64, error)
	SaveCachedTree(uid string, t *models.HashTree) error
	UpdateRoot(uid string, stream io.Reader, lastGen int64) (int64, error)
}

// UserStorer holds informations about users
type UserStorer interface {
	GetUsers() ([]*model.User, error)
	GetUser(string) (*model.User, error)
	RegisterUser(u *model.User) error
	UpdateUser(u *model.User) error
	RemoveUser(uid string) error
}

type UserRootStorer interface {
	GetCachedTree() (t *models.HashTree, err error)
	GetRootIndex() (string, int64, error)
	SaveCachedTree(t *models.HashTree) error
	UpdateRoot(stream io.Reader, lastGen int64) (int64, error)
}

// Document represents a document in storage
type Document struct {
	ID      string
	Type    common.EntryType
	Parent  string
	Name    string
	Version int
}
