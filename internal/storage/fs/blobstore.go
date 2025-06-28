package fs

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/exporter"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const cachedTreeName = ".tree"

// serves as root modification log and generation number source
const historyFile = ".root.history"

// GetCachedTree returns the cached blob tree for the user
func (fs *FileSystemStorage) GetCachedTree(uid string) (t *models.HashTree, err error) {
	blobStorage := &LocalBlobStorage{
		uid: uid,
		fs:  fs,
	}

	cachePath := path.Join(fs.getUserPath(uid), cachedTreeName)

	tree, err := models.LoadTree(cachePath)
	if err != nil {
		return nil, err
	}
	changed, err := tree.Mirror(blobStorage)
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

// SaveCachedTree saves the cached tree
func (fs *FileSystemStorage) SaveCachedTree(uid string, t *models.HashTree) error {
	cachePath := path.Join(fs.getUserPath(uid), cachedTreeName)
	return t.Save(cachePath)
}

func (fs *FileSystemStorage) BlobStorage(uid string) *LocalBlobStorage {
	return &LocalBlobStorage{
		fs:  fs,
		uid: uid,
	}
}

// Export exports a document
func (fs *FileSystemStorage) Export(uid, docid string) (r io.ReadCloser, err error) {
	tree, err := fs.GetCachedTree(uid)
	if err != nil {
		return nil, err
	}
	doc, err := tree.FindDoc(docid)
	if err != nil {
		return nil, err
	}
	ls := fs.BlobStorage(uid)

	archive, err := models.ArchiveFromHashDoc(doc, ls)
	if err != nil {
		return nil, err
	}
	reader, writer := io.Pipe()
	go func() {
		err = exporter.RenderRmapi(archive, writer)
		if err != nil {
			log.Error(err)
			writer.Close()
			return
		}
		writer.Close()
	}()
	return reader, err
}

// UpdateBlobDocument updates metadata
func (fs *FileSystemStorage) UpdateBlobDocument(uid, docID, name, parent string) (err error) {
	tree, err := fs.GetCachedTree(uid)
	if err != nil {
		return nil
	}

	log.Info("updateBlobDocument: ", docID, "new name:", name)
	blobStorage := fs.BlobStorage(uid)

	err = updateTree(tree, blobStorage, func(t *models.HashTree) error {
		hashDoc, err := tree.FindDoc(docID)
		if err != nil {
			return err
		}
		log.Info("updateBlobDocument: ", hashDoc.DocumentName)

		hashDoc.DocumentName = name
		hashDoc.Parent = parent
		hashDoc.Version++

		metadataHash, metadataReader, err := hashDoc.MetadataReader()
		if err != nil {
			return err
		}

		err = blobStorage.Write(metadataHash, metadataReader)
		if err != nil {
			return err
		}

		//update the metadata hash
		for _, hashEntry := range hashDoc.Files {
			if hashEntry.IsMetadata() {
				hashEntry.Hash = metadataHash
				break
			}
		}
		hashDoc.Rehash()
		hashDocReader, err := hashDoc.IndexReader()
		if err != nil {
			return err
		}

		err = blobStorage.Write(hashDoc.Hash, hashDocReader)
		if err != nil {
			return err
		}

		t.Rehash()
		return nil
	})

	return err
}

// DeleteBlobDocument deletes blob document
func (fs *FileSystemStorage) DeleteBlobDocument(uid, docID string) (err error) {
	tree, err := fs.GetCachedTree(uid)
	if err != nil {
		return nil
	}

	blobStorage := fs.BlobStorage(uid)

	return updateTree(tree, blobStorage, func(t *models.HashTree) error {
		return tree.Remove(docID)
	})
}

// CreateBlobFolder creates a new folder
func (fs *FileSystemStorage) CreateBlobFolder(uid, foldername, parent string) (doc *storage.Document, err error) {
	docID := uuid.New().String()
	tree, err := fs.GetCachedTree(uid)
	if err != nil {
		return nil, err
	}

	log.Info("Creating blob folder ", foldername, " parent: ", parent)

	blobStorage := fs.BlobStorage(uid)

	metadata := models.MetadataFile{
		DocumentName:     foldername,
		CollectionType:   common.CollectionType,
		Parent:           parent,
		Version:          1,
		CreatedTime:      models.FromTime(time.Now()),
		LastModified:     models.FromTime(time.Now()),
		Synced:           true,
		MetadataModified: true,
	}

	metadataReader, metahash, size, err := createMetadataFile(metadata)
	log.Info("meta hash: ", metahash)
	err = blobStorage.Write(metahash, metadataReader)
	if err != nil {
		return nil, err
	}

	metadataEntry := models.NewHashEntry(metahash, docID+storage.MetadataFileExt, size)
	hashDoc := models.NewHashDocWithMeta(docID, metadata)
	err = hashDoc.AddFile(metadataEntry)

	if err != nil {
		return nil, err
	}
	hashDocReader, err := hashDoc.IndexReader()
	if err != nil {
		return nil, err
	}

	err = blobStorage.Write(hashDoc.Hash, hashDocReader)
	if err != nil {
		return nil, err
	}

	err = updateTree(tree, blobStorage, func(t *models.HashTree) error {
		return t.Add(hashDoc)
	})

	if err != nil {
		return nil, err
	}

	doc = &storage.Document{
		ID:     docID,
		Type:   common.CollectionType,
		Parent: parent,
		Name:   foldername,
	}
	return doc, nil
}

func UpdateTree(tree *models.HashTree, storage *LocalBlobStorage, treeMutation func(t *models.HashTree) error) error {
	return updateTree(tree, storage, treeMutation)
}

// updates the tree and saves the new root
func updateTree(tree *models.HashTree, blobStorage *LocalBlobStorage, treeMutation func(t *models.HashTree) error) error {
	for i := 0; i < 3; i++ {
		err := treeMutation(tree)
		if err != nil {
			return err
		}

		rootIndexReader, err := tree.RootIndex()
		if err != nil {
			return err
		}
		err = blobStorage.Write(tree.Hash, rootIndexReader)
		if err != nil {
			return err
		}

		gen, err := blobStorage.WriteRootIndex(tree.Generation, tree.Hash)
		//the tree has been updated
		if err == storage.ErrorWrongGeneration {
			tree.Mirror(blobStorage)
			continue
		}
		if err != nil {
			return err
		}
		log.Info("got new root gen ", gen)
		tree.Generation = gen
		//TODO: concurrency
		err = blobStorage.fs.SaveCachedTree(blobStorage.uid, tree)

		if err != nil {
			return err
		}

		return nil
	}
	return errors.New("could not update")
}

// CreateBlobDocument creates a new document
func (fs *FileSystemStorage) CreateBlobDocument(uid, filename, parent string, stream io.Reader) (doc *storage.Document, err error) {
	ext := path.Ext(filename)
	switch ext {
	case storage.EpubFileExt, storage.PdfFileExt, storage.RmDocFileExt:
	default:
		return nil, errors.New("unsupported extension: " + ext)
	}

	if ext == storage.RmDocFileExt {
		return nil, errors.New("TODO: not implemented yet")
	}

	//TODO: zips and rm
	blobPath := fs.getUserBlobPath(uid)
	docid := uuid.New().String()
	//create metadata
	docName := strings.TrimSuffix(filename, ext)

	tree, err := fs.GetCachedTree(uid)
	if err != nil {
		return nil, err
	}

	log.Info("Creating metadata... parent: ", parent)

	metadata := models.MetadataFile{
		DocumentName:     docName,
		CollectionType:   common.DocumentType,
		Parent:           parent,
		Version:          1,
		CreatedTime:      models.FromTime(time.Now()),
		LastModified:     models.FromTime(time.Now()),
		Synced:           true,
		MetadataModified: true,
	}

	blobStorage := fs.BlobStorage(uid)
	r, metahash, size, err := createMetadataFile(metadata)
	blobStorage.Write(metahash, r)
	if err != nil {
		return nil, err
	}

	payloadEntry := models.NewHashEntry(metahash, docid+storage.MetadataFileExt, size)
	if err != nil {
		return
	}

	hashDoc := models.NewHashDocWithMeta(docid, metadata)
	hashDoc.PayloadType = docName

	err = hashDoc.AddFile(payloadEntry)
	if err != nil {
		return
	}

	content := createContent(ext)

	contentReader := strings.NewReader(content)
	contentHash, size, err := models.Hash(contentReader)
	if err != nil {
		return
	}
	_, err = contentReader.Seek(0, io.SeekStart)
	if err != nil {
		return
	}
	err = blobStorage.Write(contentHash, contentReader)
	if err != nil {
		return
	}
	payloadEntry = models.NewHashEntry(contentHash, docid+storage.ContentFileExt, size)

	err = hashDoc.AddFile(payloadEntry)
	if err != nil {
		return
	}

	// given that the payload can be huge
	// calculate the hash while streaming the payload to the storage
	// then rename it
	tmpdoc, err := os.CreateTemp(blobPath, "blob-upload")
	if err != nil {
		return
	}
	defer tmpdoc.Close()
	defer os.Remove(tmpdoc.Name())

	tee := io.TeeReader(stream, tmpdoc)
	payloadHash, size, err := models.Hash(tee)
	if err != nil {
		return nil, err
	}
	tmpdoc.Close()
	payloadFilename := path.Join(blobPath, payloadHash)
	log.Debug("new payload name: ", payloadFilename)
	err = os.Rename(tmpdoc.Name(), payloadFilename)
	if err != nil {
		return nil, err
	}
	payloadEntry = models.NewHashEntry(payloadHash, docid+ext, size)
	err = hashDoc.AddFile(payloadEntry)

	if err != nil {
		return nil, err
	}

	indexReader, err := hashDoc.IndexReader()
	if err != nil {
		return nil, err
	}
	err = blobStorage.Write(hashDoc.Hash, indexReader)
	if err != nil {
		return nil, err
	}

	err = updateTree(tree, blobStorage, func(t *models.HashTree) error {
		return tree.Add(hashDoc)
	})

	if err != nil {
		return
	}

	doc = &storage.Document{
		ID:     docid,
		Type:   common.DocumentType,
		Parent: "",
		Name:   docName,
	}
	return
}

func createMetadataFile(metadata models.MetadataFile) (r io.Reader, filehash string, size int64, err error) {
	jsn, err := json.Marshal(metadata)
	if err != nil {
		return
	}
	reader := bytes.NewReader(jsn)
	filehash, size, err = models.Hash(reader)
	if err != nil {
		return
	}
	reader.Seek(0, io.SeekStart)
	r = reader
	return
}

// GetBlobURL return a url for a file to store
func (fs *FileSystemStorage) GetBlobURL(uid, blobid string, write bool) (docurl string, exp time.Time, err error) {
	uploadRL := fs.Cfg.StorageURL
	exp = time.Now().Add(time.Minute * config.ReadStorageExpirationInMinutes)
	strExp := strconv.FormatInt(exp.Unix(), 10)

	scope := storage.ReadScope
	if write {
		scope = storage.WriteScope
	}

	signature, err := storage.SignURLParams([]string{uid, blobid, strExp, scope}, fs.Cfg.JWTSecretKey)
	if err != nil {
		return
	}

	params := url.Values{
		storage.ParamUID:       {uid},
		storage.ParamBlobID:    {blobid},
		storage.ParamExp:       {strExp},
		storage.ParamSignature: {signature},
		storage.ParamScope:     {scope},
	}

	blobURL := uploadRL + storage.RouteBlob + "?" + params.Encode()
	log.Debugln("blobUrl: ", blobURL)
	return blobURL, exp, nil
}

// LoadBlob Opens a blob by id
func (fs *FileSystemStorage) LoadBlob(uid, blobid string) (reader io.ReadCloser, size int64, hash string, err error) {
	blobPath := path.Join(fs.getUserBlobPath(uid), common.Sanitize(blobid))
	log.Debugln("Fullpath:", blobPath)

	fi, err := os.Stat(blobPath)
	if err != nil || fi.IsDir() {
		return nil, 0, "", storage.ErrorNotFound
	}

	osFile, err := os.Open(blobPath)
	if err != nil {
		log.Errorf("cannot open blob %v", err)
		return
	}
	//TODO: cache the crc32c
	hash, err = common.CRC32CFromReader(osFile)
	if err != nil {
		log.Errorf("cannot get crc32c hash %v", err)
		return
	}
	_, err = osFile.Seek(0, 0)
	if err != nil {
		log.Errorf("cannot rewind file %v", err)
		return
	}
	reader = osFile
	return reader, fi.Size(), "crc32c=" + hash, err
}

// StoreBlob stores a document
func (fs *FileSystemStorage) StoreBlob(uid, id string, fileName string, hash string, stream io.Reader) error {
	log.Debugf("TODO: check/save etc. write file '%s', hash '%s'", fileName, hash)

	blobPath := path.Join(fs.getUserBlobPath(uid), common.Sanitize(id))
	log.Info("Write: ", blobPath)
	file, err := os.Create(blobPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, stream)
	if err != nil {
		return err
	}

	return nil
}

// use file size as generation
func generationFromFileSize(size int64) int64 {
	//time + 1 space + 64 hash + 1 newline
	return size / 86
}
