package fs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
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
	"github.com/juruen/rmapi/archive"
	log "github.com/sirupsen/logrus"
)

const cachedTreeName = ".tree"

// serves as root modification log and generation number source
const historyFile = ".root.history"
const rootBlob = "root"

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

	// Detect version BEFORE trying to load archive
	// This is crucial because v6 files can't be unmarshaled by rmapi
	version := exporter.VersionUnknown
	var firstRmHash string

	// Find first .rm file in doc
	for _, f := range doc.Files {
		if filepath.Ext(f.EntryName) == storage.RmFileExt {
			firstRmHash = f.Hash
			break
		}
	}

	// Detect version from raw .rm blob
	if firstRmHash != "" {
		reader, err := ls.GetReader(firstRmHash)
		if err == nil {
			defer reader.Close()
			// Read just enough for version detection
			header := make([]byte, 43)
			n, err := reader.Read(header)
			if err == nil || err == io.EOF {
				version, _ = exporter.DetectRmVersionFromBytes(header[:n])
			}
		}
	}

	log.Debugf("Detected format %s for blob doc %s", version.String(), docid)

	reader, writer := io.Pipe()

	// Route to appropriate renderer
	if version == exporter.VersionV6 {
		log.Infof("Using native rmc-go for v6 format blob doc %s", docid)

		go func() {
			defer writer.Close()

			// Extract .rm files directly from blob storage without parsing
			// First, get content.json to know page order
			var contentData archive.Content
			for _, f := range doc.Files {
				if filepath.Ext(f.EntryName) == storage.ContentFileExt {
					blob, err := ls.GetReader(f.Hash)
					if err == nil {
						contentBytes, _ := io.ReadAll(blob)
						blob.Close()
						err = json.Unmarshal(contentBytes, &contentData)
						if err != nil {
							log.Warnf("Failed to unmarshal content.json: %v", err)
						}
					}
					break
				}
			}

			log.Debugf("Content has %d pages", len(contentData.Pages))

			// Build map of page names to hashes
			pageMap := make(map[string]string)
			for _, f := range doc.Files {
				if filepath.Ext(f.EntryName) == storage.RmFileExt {
					name := strings.TrimSuffix(filepath.Base(f.EntryName), storage.RmFileExt)
					pageMap[name] = f.Hash
					log.Debugf("Found .rm file: %s -> %s", name, f.Hash)
				}
			}

			log.Debugf("Built page map with %d entries", len(pageMap))

			// Extract all pages in order
			var pageHashes []string
			if len(contentData.Pages) > 0 {
				// Use pages from content.json in the correct order
				for _, pageName := range contentData.Pages {
					if hash, ok := pageMap[pageName]; ok {
						pageHashes = append(pageHashes, hash)
						log.Debugf("Found page %s -> %s", pageName, hash)
					} else {
						log.Warnf("Page %s not found in pageMap", pageName)
					}
				}
			} else {
				// No pages in content.json, use order from doc.Files (index file order)
				// doc.Files is sorted alphabetically which reverses page order, so we reverse it back
				log.Warn("content.json has no pages array, using .rm files in reversed index order")
				var tempHashes []string
				for _, f := range doc.Files {
					if filepath.Ext(f.EntryName) == storage.RmFileExt {
						tempHashes = append(tempHashes, f.Hash)
					}
				}
				// Reverse the order to get correct page sequence
				for i := len(tempHashes) - 1; i >= 0; i-- {
					pageHashes = append(pageHashes, tempHashes[i])
					log.Infof("Using .rm file in reversed order: page %d", len(tempHashes)-i)
				}
			}

			if len(pageHashes) == 0 {
				log.Error("No pages found in v6 document")
				log.Debugf("Doc files: %+v", doc.Files)
				writer.CloseWithError(fmt.Errorf("no pages found"))
				return
			}

			log.Infof("Exporting %d v6 pages", len(pageHashes))

			// Read all pages into memory
			var pages [][]byte
			for i, pageHash := range pageHashes {
				rmReader, err := ls.GetReader(pageHash)
				if err != nil {
					log.Errorf("Failed to get v6 page %d data: %v", i, err)
					writer.CloseWithError(err)
					return
				}

				rmData, err := io.ReadAll(rmReader)
				rmReader.Close()
				if err != nil {
					log.Errorf("Failed to read v6 page %d data: %v", i, err)
					writer.CloseWithError(err)
					return
				}

				pages = append(pages, rmData)
			}

			// Use rmc-go library for multipage export (in-process, Cairo renderer)
			err = exporter.ExportV6MultiPageToPdfNative(pages, writer)
			if err != nil {
				log.Errorf("Failed to export v6 multipage with rmc-go: %v", err)
				writer.CloseWithError(err)
				return
			}
		}()
	} else {
		// Use existing v5 rendering
		log.Debugf("Using rmapi for v5 format blob doc %s", docid)

		archive, err := models.ArchiveFromHashDoc(doc, ls)
		if err != nil {
			log.Error("Failed to load v5 archive:", err)
			return nil, err
		}

		go func() {
			err = exporter.RenderRmapi(archive, writer)
			if err != nil {
				log.Error(err)
				writer.Close()
				return
			}
			writer.Close()
		}()
	}

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
func updateTree(tree *models.HashTree, storage *LocalBlobStorage, treeMutation func(t *models.HashTree) error) error {
	for i := 0; i < 3; i++ {
		err := treeMutation(tree)
		if err != nil {
			return err
		}

		rootIndexReader, err := tree.RootIndex()
		if err != nil {
			return err
		}
		err = storage.Write(tree.Hash, rootIndexReader)
		if err != nil {
			return err
		}

		gen, err := storage.WriteRootIndex(tree.Generation, tree.Hash)
		//the tree has been updated
		if err == ErrorWrongGeneration {
			tree.Mirror(storage)
			continue
		}
		if err != nil {
			return err
		}
		log.Info("got new root gen ", gen)
		tree.Generation = gen
		//TODO: concurrency
		err = storage.fs.SaveCachedTree(storage.uid, tree)

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

	scope := ReadScope
	if write {
		scope = WriteScope
	}

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
func (fs *FileSystemStorage) LoadBlob(uid, blobid string) (reader io.ReadCloser, gen int64, size int64, hash string, err error) {
	generation := int64(0)
	blobPath := path.Join(fs.getUserBlobPath(uid), common.Sanitize(blobid))
	log.Debugln("Fullpath:", blobPath)
	if blobid == rootBlob {
		historyPath := path.Join(fs.getUserBlobPath(uid), historyFile)
		lock := fslock.New(historyPath)
		err := lock.LockWithTimeout(time.Duration(time.Second * 5))
		if err != nil {
			log.Error("cannot obtain lock")
			return nil, 0, 0, "", err
		}
		defer lock.Unlock()

		fi, err1 := os.Stat(historyPath)
		if err1 == nil {
			generation = generationFromFileSize(fi.Size())
		}
	}

	fi, err := os.Stat(blobPath)
	if err != nil || fi.IsDir() {
		return nil, generation, 0, "", ErrorNotFound
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
	return reader, generation, fi.Size(), "crc32c=" + hash, err
}

// StoreBlob stores a document
func (fs *FileSystemStorage) StoreBlob(uid, id string, stream io.Reader, lastGen int64) (generation int64, err error) {
	generation = 1

	reader := stream
	if id == rootBlob {
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
			log.Warnf("wrong generation, currentGen %d, lastGen %d", currentGen, lastGen)
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

		reader = io.NopCloser(&buf)
		size, err1 := hist.Seek(0, io.SeekCurrent)
		if err1 != nil {
			err = err1
			return
		}
		generation = generationFromFileSize(size)
	}

	blobPath := path.Join(fs.getUserBlobPath(uid), common.Sanitize(id))
	log.Info("Write: ", blobPath)
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

// use file size as generation
func generationFromFileSize(size int64) int64 {
	//time + 1 space + 64 hash + 1 newline
	return size / 86
}

// detectBlobArchiveVersion detects the .rm file version from a blob archive
func detectBlobArchiveVersion(arch *exporter.MyArchive) (exporter.RmVersion, error) {
	if len(arch.Pages) == 0 {
		return exporter.VersionUnknown, fmt.Errorf("no pages in archive")
	}

	// Try to marshal first page and detect from header
	if arch.Pages[0].Data != nil {
		data, err := arch.Pages[0].Data.MarshalBinary()
		if err != nil {
			return exporter.VersionUnknown, fmt.Errorf("failed to marshal page data: %w", err)
		}
		return exporter.DetectRmVersion(bytes.NewReader(data))
	}

	return exporter.VersionUnknown, fmt.Errorf("no page data available")
}
