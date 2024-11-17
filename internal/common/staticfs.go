package common

import (
	"fmt"
	"net/http"
	"os"
	"time"
)
type staticFS struct {
	http.FileSystem
	LastModified time.Time
}
type StaticFileWrapper struct {
	http.File
	LastModified time.Time
}
type StaticFileInfoWrapper struct {
	os.FileInfo
	LastModfied time.Time
}
func (f *StaticFileInfoWrapper) ModTime() time.Time {
	return f.LastModfied
}
func (f *StaticFileWrapper) Stat() (os.FileInfo, error) {
	fmt.Println("calling stat")
	fi, err := f.File.Stat()
	return &StaticFileInfoWrapper {FileInfo: fi, LastModfied: f.LastModified}, err
}
func (f *staticFS) Open(name string) (http.File, error) {
	file, err := f.FileSystem.Open(name)
	return &StaticFileWrapper {File: file, LastModified: f.LastModified}, err
}

// NewLastModifiedFS returns a filesystem with a fixed LastModified 
// to enable caching of embedded assets
func NewLastModifiedFS(fs http.FileSystem, lastModified time.Time) http.FileSystem {
	return &staticFS {
		FileSystem: fs,
		LastModified: lastModified,
	}
}
