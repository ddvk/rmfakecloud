package fs

import (
	"io"
	"io/ioutil"
	"strings"
)

type LocalBlobStorage struct {
	fs  *Storage
	uid string
}

func (p *LocalBlobStorage) GetRootIndex() (string, int64, error) {
	r, gen, err := p.fs.LoadBlob(p.uid, rootFile)
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

func (p *LocalBlobStorage) WriteRootIndex(generation int64, rootindex string) (int64, error) {
	r := strings.NewReader(rootindex)
	newGen, err := p.fs.StoreBlob(p.uid, rootFile, r, generation)
	return int64(newGen), err
}

func (p *LocalBlobStorage) GetReader(hash string) (io.ReadCloser, error) {
	r, _, err := p.fs.LoadBlob(p.uid, hash)
	return r, err
}
