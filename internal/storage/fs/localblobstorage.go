package fs

import (
	"io"
	"io/ioutil"
	"strings"
)

// LocalBlobStorage local file system storage
type LocalBlobStorage struct {
	fs  *FileSystemStorage
	uid string
}

// GetRootIndex the hash of the root index
func (p *LocalBlobStorage) GetRootIndex() (string, int64, error) {
	r, gen, err := p.fs.LoadBlob(p.uid, rootFile)
	if err == ErrorNotFound {
		return "", 0, nil
	}
	if err != nil {
		return "", 0, err
	}
	defer r.Close()
	s, err := ioutil.ReadAll(r)
	if err != nil {
		return "", 0, err
	}
	return string(s), int64(gen), nil

}

// WriteRootIndex writes the root index
func (p *LocalBlobStorage) WriteRootIndex(generation int64, roothash string) (int64, error) {
	r := strings.NewReader(roothash)
	newGen, err := p.fs.StoreBlob(p.uid, rootFile, r, generation)
	return int64(newGen), err
}

// GetReader reader for a given hash
func (p *LocalBlobStorage) GetReader(hash string) (io.ReadCloser, error) {
	r, _, err := p.fs.LoadBlob(p.uid, hash)
	return r, err
}

// Write writes the hash from the reader
func (p *LocalBlobStorage) Write(hash string, r io.Reader) error {
	_, err := p.fs.StoreBlob(p.uid, hash, r, -1)

	return err
}
