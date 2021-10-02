package fs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
)

type LocalStore struct {
	Folder string
}

func (p *LocalStore) GetRootIndex() (string, int64, error) {
	rootPath := path.Join(p.Folder, rootFile)
	root_hash, err := ioutil.ReadFile(rootPath)
	if err != nil {
		return "", 0, err
	}
	strRootHash := string(root_hash)
	rootGenPath := path.Join(p.Folder, historyFile)
	var gen int64

	if fi, err := os.Stat(rootGenPath); err == nil {
		gen = int64(calcGen(fi.Size()))
	}
	fmt.Println("root ->", strRootHash)
	return strRootHash, gen, nil
}

func (p *LocalStore) GetReader(hash string) (io.ReadCloser, error) {
	rootIndexPath := path.Join(p.Folder, hash)
	return os.Open(rootIndexPath)
}
