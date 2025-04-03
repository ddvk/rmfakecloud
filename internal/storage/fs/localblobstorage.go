package fs

import (
	"io"
	"strings"
)

// LocalBlobStorage local file system storage
type LocalBlobStorage struct {
	fs  *FileSystemStorage
	uid string
}

// GetRootIndex the hash of the root index
func (p *LocalBlobStorage) GetRootIndex() (hash string, gen int64, err error) {
	return p.fs.GetRoot(p.uid)
}

// WriteRootIndex writes the root index
func (p *LocalBlobStorage) WriteRootIndex(generation int64, roothash string) (int64, error) {
	r := strings.NewReader(roothash)
	return p.fs.UpdateRoot(p.uid, r, generation)
}

// GetReader reader for a given hash
func (p *LocalBlobStorage) GetReader(hash string) (io.ReadCloser, error) {
	r, _, _, err := p.fs.LoadBlob(p.uid, hash)
	return r, err
}

// Write stores the reader in the hash
func (p *LocalBlobStorage) Write(hash string, r io.Reader) error {
	return p.fs.StoreBlob(p.uid, hash, "", "", r)
}
