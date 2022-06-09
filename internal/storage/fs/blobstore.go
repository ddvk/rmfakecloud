package fs

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
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
	"github.com/juju/fslock"
	log "github.com/sirupsen/logrus"
)

const cachedTreeName = ".tree"

// GetTree returns the cached blob tree for the user
func (fs *FileSystemStorage) GetTree(uid string) (t *models.HashTree, err error) {
	ls := &LocalBlobStorage{
		uid: uid,
		fs:  fs,
	}

	cachePath := path.Join(fs.getUserPath(uid), cachedTreeName)

	tree, err := models.LoadTree(cachePath)
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

// SaveTree saves the cached tree
func (fs *FileSystemStorage) SaveTree(uid string, t *models.HashTree) error {
	cachePath := path.Join(fs.getUserPath(uid), cachedTreeName)
	return t.Save(cachePath)
}

// Export exports a document
func (fs *FileSystemStorage) Export(uid, docid string) (r io.ReadCloser, err error) {
	tree, err := fs.GetTree(uid)
	if err != nil {
		return nil, err
	}
	doc, err := tree.FindDoc(docid)
	if err != nil {
		return nil, err
	}
	ls := &LocalBlobStorage{
		fs:  fs,
		uid: uid,
	}

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

// CreateBlobDocument creates a new document
func (fs *FileSystemStorage) CreateBlobDocument(uid, filename, parent string, stream io.Reader) (doc *storage.Document, err error) {
	ext := path.Ext(filename)
	switch ext {
	case models.PdfFileExt:
		fallthrough
	case models.EpubFileExt:
	default:
		return nil, errors.New("unsupported extension: " + ext)
	}
	//TODO: zips and rm

	docid := uuid.New().String()
	//create metadata
	docName := strings.TrimSuffix(filename, ext)
	blobPath := fs.getUserBlobPath(uid)

	tree, err := fs.GetTree(uid)
	if err != nil {
		return nil, err
	}

	log.Info("Creating metadata... parent: ", parent)

	metadata := models.MetadataFile{
		DocumentName:     docName,
		CollectionType:   models.DocumentType,
		Parent:           parent,
		Version:          1,
		LastModified:     strconv.FormatInt(time.Now().UnixMilli(), 10),
		Synced:           true,
		MetadataModified: true,
	}
	metahash, size, err := createMetadataFile(metadata, blobPath)
	fi := models.NewFileHashEntry(metahash, docid+models.MetadataFileExt)
	fi.Size = size
	if err != nil {
		return
	}

	hashDoc := models.NewHashDocMeta(docid, metadata)

	err = hashDoc.AddFile(fi)
	if err != nil {
		return
	}

	content := createContent(ext)
	contentHash, size, err := models.Hash(strings.NewReader(content))
	if err != nil {
		return
	}
	err = saveTo(strings.NewReader(content), contentHash, blobPath)
	if err != nil {
		return
	}
	fi = models.NewFileHashEntry(contentHash, docid+models.ContentFileExt)
	fi.Size = size

	hashDoc.AddFile(fi)

	tmpdoc, err := ioutil.TempFile(blobPath, ".tmp")
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
	err = os.Rename(tmpdoc.Name(), payloadFilename)
	if err != nil {
		return nil, err
	}
	fi = models.NewFileHashEntry(payloadHash, docid+ext)
	fi.Size = size
	err = hashDoc.AddFile(fi)

	if err != nil {
		return nil, err
	}

	//TODO: loop
	err = tree.Add(hashDoc)
	if err != nil {
		return
	}

	docIndexReader, err := hashDoc.IndexReader()
	err = saveTo(docIndexReader, hashDoc.Hash, blobPath)
	if err != nil {
		return
	}

	rootIndexReader, err := tree.RootIndex()
	err = saveTo(rootIndexReader, tree.Hash, blobPath)
	if err != nil {
		return
	}
	blobStorage := &LocalBlobStorage{
		fs:  fs,
		uid: uid,
	}

	//todo:check gen + locking
	gen, err := blobStorage.WriteRootIndex(tree.Generation, tree.Hash)
	if err != nil {
		return
	}
	log.Info("got gen ", gen)
	tree.Generation = gen
	err = fs.SaveTree(uid, tree)

	if err != nil {
		return
	}

	doc = &storage.Document{
		ID:     docid,
		Type:   models.DocumentType,
		Parent: "",
		Name:   docName,
	}
	return
}

func saveTo(r io.Reader, hash, blobPath string) (err error) {
	rootIndexFilePath := path.Join(blobPath, hash)
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

func createMetadataFile(metadata models.MetadataFile, spath string) (filehash string, size int64, err error) {

	jsn, err := json.Marshal(metadata)
	if err != nil {
		return
	}
	filehash, size, err = models.Hash(bytes.NewReader(jsn))
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

//serves as root modification log and generation number source
const historyFile = ".root.history"
const rootFile = "root"

// GetBlobURL return a url for a file to store
func (fs *FileSystemStorage) GetBlobURL(uid, blobid, scope string) (docurl string, exp time.Time, err error) {
	uploadRL := fs.Cfg.StorageURL
	exp = time.Now().Add(time.Minute * config.ReadStorageExpirationInMinutes)
	strExp := strconv.FormatInt(exp.Unix(), 10)

	signature, err := SignURLParams([]string{uid, blobid, strExp, scope}, fs.Cfg.JWTSecretKey)
	if err != nil {
		return
	}

	params := url.Values{
		paramUID:       {uid},
		paramBlobID:    {blobid},
		paramExp:       {strExp},
		paramSignature: {signature},
		paramScope:     {scope},
	}

	blobURL := uploadRL + routeBlob + "?" + params.Encode()
	log.Debugln("blobUrl: ", blobURL)
	return blobURL, exp, nil
}

// LoadBlob Opens a blob by id
func (fs *FileSystemStorage) LoadBlob(uid, blobid string) (reader io.ReadCloser, gen int64, size int64, err error) {
	generation := int64(0)
	blobPath := path.Join(fs.getUserBlobPath(uid), common.Sanitize(blobid))
	log.Debugln("Fullpath:", blobPath)
	if blobid == rootFile {
		historyPath := path.Join(fs.getUserBlobPath(uid), historyFile)
		lock := fslock.New(historyPath)
		err := lock.LockWithTimeout(time.Duration(time.Second * 5))
		if err != nil {
			log.Error("cannot obtain lock")
			return nil, 0, 0, err
		}
		defer lock.Unlock()

		fi, err1 := os.Stat(historyPath)
		if err1 == nil {
			generation = generationFromFileSize(fi.Size())
		}
	}

	fi, err := os.Stat(blobPath)
	if err != nil || fi.IsDir() {
		return nil, 0, 0, ErrorNotFound
	}

	reader, err = os.Open(blobPath)
	return reader, generation, fi.Size(), err
}

// StoreBlob stores a document
func (fs *FileSystemStorage) StoreBlob(uid, id string, stream io.Reader, lastGen int64) (generation int64, err error) {
	generation = 1

	reader := stream
	if id == rootFile {
		historyPath := path.Join(fs.getUserBlobPath(uid), historyFile)
		lock := fslock.New(historyPath)
		err = lock.LockWithTimeout(time.Duration(time.Second * 5))
		if err != nil {
			log.Error("cannot obtain lock")
		}
		defer lock.Unlock()

		currentGen := int64(0)
		fi, err1 := os.Stat(historyPath)
		if err1 == nil {
			currentGen = generationFromFileSize(fi.Size())
		}

		if currentGen != lastGen && currentGen > 0 {
			log.Warnf("wrong generation, server %d, client %d", currentGen, lastGen)
			return currentGen, ErrorWrongGeneration
		}

		var buf bytes.Buffer
		tee := io.TeeReader(stream, &buf)

		var hist *os.File
		hist, err = os.OpenFile(historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return
		}
		defer hist.Close()
		t := time.Now().UTC().Format(time.RFC3339) + " "
		hist.WriteString(t)
		_, err = io.Copy(hist, tee)
		if err != nil {
			return
		}
		hist.WriteString("\n")

		reader = ioutil.NopCloser(&buf)
		size, err1 := hist.Seek(0, os.SEEK_CUR)
		if err1 != nil {
			err = err1
			return
		}
		generation = generationFromFileSize(size)
	}

	blobPath := path.Join(fs.getUserBlobPath(uid), common.Sanitize(id))
	file, err := os.Create(blobPath)
	if err != nil {
		return
	}
	defer file.Close()
	_, err = io.Copy(file, reader)
	if err != nil {
		return
	}

	return
}

//use file size as generation
func generationFromFileSize(size int64) int64 {
	//time + 1 space + 64 hash + 1 newline
	return size / 86
}
