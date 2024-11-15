package fs

import (
	"io"
	"strings"

	log "github.com/sirupsen/logrus"
)

// LocalBlobStorage local file system storage
type LocalBlobStorage struct {
	fs  *FileSystemStorage
	uid string
}

// GetRootIndex the hash of the root index
func (p *LocalBlobStorage) GetRootIndex() (hash string, gen int64, err error) {
	r, gen, _, _, err := p.fs.LoadBlob(p.uid, rootBlob)
	if err == ErrorNotFound {
		log.Info("root not found")
		return "", gen, nil
	}
	if err != nil {
		return "", 0, err
	}
	defer r.Close()
	s, err := io.ReadAll(r)
	if err != nil {
		return "", 0, err
	}
	return string(s), int64(gen), nil

}

// WriteRootIndex writes the root index
func (p *LocalBlobStorage) WriteRootIndex(generation int64, roothash string) (int64, error) {
	r := strings.NewReader(roothash)
	return p.fs.StoreBlob(p.uid, rootBlob, r, generation)
}

// GetReader reader for a given hash
func (p *LocalBlobStorage) GetReader(hash string) (io.ReadCloser, error) {
	r, _, _, _, err := p.fs.LoadBlob(p.uid, hash)
	return r, err
}

// Write stores the reader in the hash
func (p *LocalBlobStorage) Write(hash string, r io.Reader) error {
	_, err := p.fs.StoreBlob(p.uid, hash, r, -1)

	return err
}
