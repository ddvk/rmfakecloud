package db

import (
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/model"
)

// MetadataStorer manages document metadata
type MetadataStorer interface {
	UpdateMetadata(r *messages.RawDocument) error
	GetAllMetadata(withBlob bool) ([]*messages.RawDocument, error)
	GetMetadata(string, bool) (*messages.RawDocument, error)
}

// UserStorer holds informations about users
type UserStorer interface {
	GetUsers() ([]*model.User, error)
	GetUser(string) (*model.User, error)
	RegisterUser(u *model.User) error
	UpdateUser(u *model.User) error
}
