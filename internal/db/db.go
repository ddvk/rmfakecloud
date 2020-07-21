package db

import (
	"github.com/ddvk/rmfakecloud/internal/messages"
)

// MetadataStorer manages document metadata
type MetadataStorer interface {
	UpdateMetadata(r *messages.RawDocument) error
	GetAllMetadata(withBlob bool) ([]*messages.RawDocument, error)
	GetMetadata(string, bool) (*messages.RawDocument, error)
}
