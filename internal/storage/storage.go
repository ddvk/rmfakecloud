package storage

import (
	"io"

	"github.com/gin-gonic/gin"
)

// DocumentStorer stores documents
type DocumentStorer interface {
	StoreDocument(io.ReadCloser, string) error
	RemoveDocument(string) error
	GetDocument(string) (io.ReadCloser, error)
	GetStorageURL(string) string

	RegisterRoutes(*gin.Engine)
}
