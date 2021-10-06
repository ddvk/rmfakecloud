package models

import (
	"io"
)

// RemoteStorage remote access
type RemoteStorage interface {
	GetRootIndex() (hash string, generation int64, err error)
	GetReader(hash string) (io.ReadCloser, error)
}

// RemoteStorageWriter some writer
type RemoteStorageWriter interface {
	WriteRootIndex(hash string, generation int64) (gen int64, err error)
	Write(hash string, reader io.Reader) error
}
