package storage

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// DocumentStorer stores documents
type DocumentStorer interface {
	StoreDocument(string, io.ReadCloser, string) error
	RemoveDocument(string, string) error
	GetDocument(string, string) (io.ReadCloser, error)

	// GetStorageURL creates a short lived url
	GetStorageURL(string, time.Time, string) (string, error)

	RegisterRoutes(*gin.Engine)
}
