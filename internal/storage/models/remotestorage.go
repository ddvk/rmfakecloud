package models

import (
	"io"
)

// RemoteStorage abstraction for the blob storage
type RemoteStorage interface {
	// GetRootIndex returns the rootIndex
	GetRootIndex() (hash string, generation int64, err error)

	// GetReader returns a reader for the the blob with that hash
	GetReader(hash string) (io.ReadCloser, error)
}

// RemoteStorageWriter write abstraction
type RemoteStorageWriter interface {
	WriteRootIndex(generation int64, hash string) (gen int64, err error)
	Write(hash string, reader io.Reader) error
}
