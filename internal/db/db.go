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

// UserStorer holds informations about users
type UserStorer interface {
	GetUsers() ([]*messages.User, error)
	GetUser(string) (*messages.User, error)
	RegisterUser(u *messages.User) error
	UpdateUser(u *messages.User) error
}
