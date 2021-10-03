package fs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/fs/sync15"
	"github.com/google/uuid"
)

const cachedTreeName = ".tree"

func (fs *Storage) GetTree(uid string) (t *sync15.HashTree, err error) {
	ls := &LocalBlobStorage{
		uid: uid,
		fs:  fs,
	}

	cachePath := path.Join(fs.getUserPath(uid), cachedTreeName)

	tree, err := sync15.LoadTree(cachePath)
	if err != nil {
		return nil, err
	}
	changed, err := tree.Mirror(ls)
	if err != nil {
		return nil, err
	}
	if changed {
		err = tree.Save(cachePath)
		if err != nil {
			return nil, err
		}
	}
	return tree, nil
}

func (fs *Storage) SaveTree(uid string, t *sync15.HashTree) error {
	cachePath := path.Join(fs.getUserPath(uid), cachedTreeName)
	return t.Save(cachePath)
}

// CreateDocument creates a new document
func (fs *Storage) CreateBlobDocument(uid, filename string, stream io.Reader) (doc *storage.Document, err error) {
	ext := path.Ext(filename)
	switch ext {
	case ".pdf":
	case ".epub":
	default:
		return nil, errors.New("unsupported extension: " + ext)
	}

	docid := uuid.New().String()
	//create metadata
	name := strings.TrimSuffix(filename, ext)
	syncpath := fs.getUserSyncPath(uid)

	tree, err := fs.GetTree(uid)
	if err != nil {
		return nil, err
	}

	blobDoc := sync15.NewBlobDoc(name, docid, storage.DocumentType)

	metahash, err := createMetadataFile(name, docid, syncpath)
	fi := &sync15.Entry{
		Hash:       metahash,
		Type:       sync15.FileType,
		DocumentID: uid + storage.MetadataFileExt,
	}
	if err != nil {
		return
	}

	err = blobDoc.AddFile(fi)
	if err != nil {
		return
	}

	content := createContent(ext)
	contentHash, err := sync15.Hash(strings.NewReader(content))
	saveTo(strings.NewReader(content), contentHash, syncpath)
	fi = &sync15.Entry{
		Hash:       contentHash,
		Type:       sync15.FileType,
		DocumentID: docid + contentFileExt,
	}

	blobDoc.AddFile(fi)

	tmpdoc, err := ioutil.TempFile(syncpath, ".tmp")
	if err != nil {
		return
	}
	defer tmpdoc.Close()
	defer os.Remove(tmpdoc.Name())

	tee := io.TeeReader(stream, tmpdoc)
	payloadHash, err := sync15.Hash(tee)
	if err != nil {
		return nil, err
	}
	tmpdoc.Close()
	payloadFilename := path.Join(syncpath, payloadHash)
	err = os.Rename(tmpdoc.Name(), payloadFilename)
	if err != nil {
		return nil, err
	}
	fi = &sync15.Entry{
		Hash:       payloadHash,
		Type:       sync15.FileType,
		DocumentID: fmt.Sprintf("%s.%s", docid, ext),
	}
	err = blobDoc.AddFile(fi)
	if err != nil {
		return nil, err
	}

	err = tree.Add(blobDoc)
	if err != nil {
		return
	}

	docIndexReader, err := blobDoc.IndexReader()
	err = saveTo(docIndexReader, blobDoc.Hash, syncpath)
	if err != nil {
		return
	}

	//loop
	rootIndexReader, err := tree.RootIndex()
	err = saveTo(rootIndexReader, tree.Hash, syncpath)
	if err != nil {
		return
	}
	blobStorage := &LocalBlobStorage{
		fs:  fs,
		uid: uid,
	}

	//todo:check gen
	gen, err := blobStorage.WriteRootIndex(tree.Generation, tree.Hash)
	tree.Generation = gen
	err = fs.SaveTree(uid, tree)

	if err != nil {
		return
	}

	doc = &storage.Document{
		ID:     docid,
		Type:   storage.DocumentType,
		Parent: "",
		Name:   name,
	}
	return
}

func saveTo(r io.Reader, hash, syncpath string) (err error) {
	rootIndexFilePath := path.Join(syncpath, hash)
	rootIndex, err := os.Create(rootIndexFilePath)
	if err != nil {
		return
	}
	_, err = io.Copy(rootIndex, r)
	if err != nil {
		return
	}
	return nil
}

func createMetadataFile(name, docid, spath string) (filehash string, err error) {
	metadata := model.MetadataFile{
		DocName:        name,
		CollectionType: storage.DocumentType,
		Parent:         "",
		Version:        0,
		LastModified:   strconv.FormatInt(time.Now().Unix(), 10),
	}

	jsn, err := json.Marshal(metadata)
	if err != nil {
		return
	}
	filehash, err = sync15.Hash(bytes.NewReader(jsn))
	if err != nil {
		return
	}
	filePath := path.Join(spath, filehash)
	err = ioutil.WriteFile(filePath, jsn, 0600)
	if err != nil {
		return
	}
	return
}
