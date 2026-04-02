package storage

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/storage/exporter"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type BlobStorer struct {
	impl BlobStorage
	root RootStorer
}

func NewBlobStorer(impl BlobStorage, rootStorer RootStorer) *BlobStorer {
	return &BlobStorer{impl, rootStorer}
}

func (bs *BlobStorer) RemoteStorage(uid string) models.RemoteStorage {
	return &BlobRemoteStorage{
		bs:  bs,
		uid: uid,
	}
}

// GetBlobURL return a url for a file to store
func (bs *BlobStorer) GetBlobURL(uid, blobid string, write bool) (string, time.Time, error) {
	return bs.impl.GetBlobURL(uid, blobid, write)
}

// LoadBlob Opens a blob by id
func (bs *BlobStorer) LoadBlob(uid, blobid string) (io.ReadCloser, int64, string, error) {
	return bs.impl.LoadBlob(uid, blobid)
}

// StoreBlob stores a document
func (bs *BlobStorer) StoreBlob(uid, id string, fileName string, hash string, stream io.Reader) error {
	return bs.impl.StoreBlob(uid, id, fileName, hash, stream)
}

func (bs *BlobStorer) GetCachedTree(uid string) (tree *models.HashTree, err error) {
	return bs.root.GetCachedTree(uid, bs.RemoteStorage(uid))
}

// ExportRmDoc exports a document as a zip of all blobs
func (bs *BlobStorer) ExportRmDoc(uid, docid string) (io.ReadCloser, error) {
	tree, err := bs.root.GetCachedTree(uid, bs.RemoteStorage(uid))
	if err != nil {
		return nil, err
	}
	doc, err := tree.FindDoc(docid)
	if err != nil {
		return nil, err
	}
	ls := bs.RemoteStorage(uid)

	reader, writer := io.Pipe()
	go func() {
		zw := zip.NewWriter(writer)
		var writeErr error
		for _, entry := range doc.Files {
			blob, err := ls.GetReader(entry.Hash)
			if err != nil {
				writeErr = err
				break
			}
			fw, err := zw.Create(entry.EntryName)
			if err != nil {
				blob.Close()
				writeErr = err
				break
			}
			_, err = io.Copy(fw, blob)
			blob.Close()
			if err != nil {
				writeErr = err
				break
			}
		}
		if writeErr != nil {
			log.Error(writeErr)
			zw.Close()
			writer.CloseWithError(writeErr)
			return
		}
		zw.Close()
		writer.Close()
	}()
	return reader, nil
}

// Export exports a document
func (bs *BlobStorer) Export(uid, docid string) (r io.ReadCloser, err error) {
	tree, err := bs.root.GetCachedTree(uid, bs.RemoteStorage(uid))
	if err != nil {
		return nil, err
	}
	doc, err := tree.FindDoc(docid)
	if err != nil {
		return nil, err
	}
	ls := bs.RemoteStorage(uid)

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
func (bs *BlobStorer) UpdateBlobDocument(uid, docID, name, parent string) (err error) {
	tree, err := bs.root.GetCachedTree(uid, bs.RemoteStorage(uid))
	if err != nil {
		return err
	}

	log.Info("updateBlobDocument: ", docID, "new name:", name)

	err = UpdateTree(tree, bs, uid, func(t *models.HashTree) error {
		hashDoc, err := tree.FindDoc(docID)
		if err != nil {
			return err
		}
		log.Info("updateBlobDocument: ", hashDoc.DocumentName)

		hashDoc.DocumentName = name
		hashDoc.Parent = parent
		hashDoc.Version++

		metadataHash, metadataCRC32C, metadataReader, err := hashDoc.MetadataReader()
		if err != nil {
			return err
		}

		err = bs.impl.StoreBlob(uid, metadataHash, hashDoc.DocumentName+models.MetadataFileExt, "crc32c="+metadataCRC32C, metadataReader)
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
		hashDocContent, err := io.ReadAll(hashDocReader)
		if err != nil {
			return err
		}

		crc32cIndex, err := common.CRC32CFromReader(bytes.NewReader(hashDocContent))
		if err != nil {
			return err
		}

		err = bs.impl.StoreBlob(uid, hashDoc.Hash, hashDoc.DocumentName, "crc32c="+crc32cIndex, bytes.NewReader(hashDocContent))
		if err != nil {
			return err
		}

		t.Rehash()
		return nil
	})

	return err
}

// DeleteBlobDocument deletes blob document
func (bs *BlobStorer) DeleteBlobDocument(uid, docID string) (err error) {
	tree, err := bs.root.GetCachedTree(uid, bs.RemoteStorage(uid))
	if err != nil {
		return err
	}

	return UpdateTree(tree, bs, uid, func(t *models.HashTree) error {
		return tree.Remove(docID)
	})
}

// CreateBlobFolder creates a new folder
func (bs *BlobStorer) CreateBlobFolder(uid, foldername, parent string) (doc *Document, err error) {
	docID := uuid.New().String()
	tree, err := bs.root.GetCachedTree(uid, bs.RemoteStorage(uid))
	if err != nil {
		return nil, err
	}

	log.Info("Creating blob folder ", foldername, " parent: ", parent)

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

	metadataReader, metahash, crc32c, size, err := createMetadataFile(metadata)
	if err != nil {
		return
	}

	log.Info("meta hash: ", metahash)
	err = bs.impl.StoreBlob(uid, metahash, "", "crc32c="+crc32c, metadataReader)
	if err != nil {
		return nil, err
	}

	metadataEntry := models.NewHashEntry(metahash, docID+models.MetadataFileExt, size)
	hashDoc := models.NewHashDocWithMeta(docID, metadata)
	err = hashDoc.AddFile(metadataEntry)

	if err != nil {
		return nil, err
	}
	hashDocReader, err := hashDoc.IndexReader()
	if err != nil {
		return nil, err
	}

	hashDocContent, err := io.ReadAll(hashDocReader)
	if err != nil {
		return nil, err
	}

	crc32c, err = common.CRC32CFromReader(bytes.NewReader(hashDocContent))
	if err != nil {
		return nil, err
	}

	err = bs.impl.StoreBlob(uid, hashDoc.Hash, "", "crc32c="+crc32c, bytes.NewReader(hashDocContent))
	if err != nil {
		return nil, err
	}

	err = UpdateTree(tree, bs, uid, func(t *models.HashTree) error {
		return t.Add(hashDoc)
	})

	if err != nil {
		return nil, err
	}

	doc = &Document{
		ID:     docID,
		Type:   common.CollectionType,
		Parent: parent,
		Name:   foldername,
	}
	return doc, nil
}

// updates the tree and saves the new root
func UpdateTree(tree *models.HashTree, bs *BlobStorer, uid string, treeMutation func(t *models.HashTree) error) error {
	for i := 0; i < 3; i++ {
		err := treeMutation(tree)
		if err != nil {
			return err
		}

		rootIndexReader, err := tree.RootIndex()
		if err != nil {
			return err
		}

		rootIndexContent, err := io.ReadAll(rootIndexReader)
		if err != nil {
			return err
		}

		crc32cIndex, err := common.CRC32CFromReader(bytes.NewReader(rootIndexContent))
		if err != nil {
			return err
		}

		err = bs.impl.StoreBlob(uid, tree.Hash, "roothash", "crc32c="+crc32cIndex, bytes.NewReader(rootIndexContent))
		if err != nil {
			return err
		}

		gen, err := bs.root.UpdateRoot(uid, bytes.NewBufferString(tree.Hash), tree.Generation)
		//the tree has been updated
		if err == ErrorWrongGeneration {
			tree.Mirror(bs.RemoteStorage(uid))
			continue
		}
		if err != nil {
			return err
		}
		log.Info("got new root gen ", gen)
		tree.Generation = gen
		//TODO: concurrency
		err = bs.root.SaveCachedTree(uid, tree)

		if err != nil {
			return err
		}

		return nil
	}
	return errors.New("could not update")
}

// CreateBlobDocument creates a new document
func (bs *BlobStorer) CreateBlobDocument(uid, filename, parent string, stream io.Reader) (doc *Document, err error) {
	ext := path.Ext(filename)
	switch ext {
	case models.EpubFileExt, models.PdfFileExt, models.RmDocFileExt:
	default:
		return nil, errors.New("unsupported extension: " + ext)
	}

	if ext == models.RmDocFileExt {
		return bs.createFromRmDoc(uid, parent, stream)
	}

	docid := uuid.New().String()
	docName := strings.TrimSuffix(filename, ext)

	tree, err := bs.root.GetCachedTree(uid, bs.RemoteStorage(uid))
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

	r, metahash, crc32c, size, err := createMetadataFile(metadata)
	if err != nil {
		return nil, err
	}

	err = bs.impl.StoreBlob(uid, metahash, "", "crc32c="+crc32c, r)
	if err != nil {
		return nil, err
	}

	payloadEntry := models.NewHashEntry(metahash, docid+models.MetadataFileExt, size)

	hashDoc := models.NewHashDocWithMeta(docid, metadata)
	hashDoc.PayloadType = docName

	err = hashDoc.AddFile(payloadEntry)
	if err != nil {
		return
	}

	content := models.CreateContent(ext)

	contentReader := strings.NewReader(content)
	contentHash, crc32c, size, err := models.Hash(contentReader)
	if err != nil {
		return
	}
	_, err = contentReader.Seek(0, io.SeekStart)
	if err != nil {
		return
	}
	err = bs.impl.StoreBlob(uid, contentHash, "", "crc32c="+crc32c, contentReader)
	if err != nil {
		return
	}
	payloadEntry = models.NewHashEntry(contentHash, docid+models.ContentFileExt, size)

	err = hashDoc.AddFile(payloadEntry)
	if err != nil {
		return
	}

	// given that the payload can be huge
	// calculate the hash while streaming the payload to the storage
	// then rename it
	tmpdoc, err := os.CreateTemp("", "rmfakecloud-upload")
	if err != nil {
		return
	}
	defer tmpdoc.Close()
	defer os.Remove(tmpdoc.Name())

	tee := io.TeeReader(stream, tmpdoc)
	payloadHash, crc32c, size, err := models.Hash(tee)
	if err != nil {
		return nil, err
	}
	tmpdoc.Close()

	// Save payload
	fd, err := os.Open(tmpdoc.Name())
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	err = bs.impl.StoreBlob(uid, payloadHash, "", "crc32c="+crc32c, fd)
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

	indexContent, err := io.ReadAll(indexReader)
	if err != nil {
		return nil, err
	}

	crc32cIndex, err := common.CRC32CFromReader(bytes.NewReader(indexContent))
	if err != nil {
		return nil, err
	}

	err = bs.impl.StoreBlob(uid, hashDoc.Hash, "", "crc32c="+crc32cIndex, bytes.NewReader(indexContent))
	if err != nil {
		return nil, err
	}

	err = UpdateTree(tree, bs, uid, func(t *models.HashTree) error {
		return tree.Add(hashDoc)
	})

	if err != nil {
		return
	}

	doc = &Document{
		ID:     docid,
		Type:   common.DocumentType,
		Parent: "",
		Name:   docName,
	}
	return
}

func (bs *BlobStorer) createFromRmDoc(uid, parent string, stream io.Reader) (*Document, error) {
	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, err
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	var metadataEntry *zip.File
	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, models.MetadataFileExt) {
			metadataEntry = f
			break
		}
	}
	if metadataEntry == nil {
		return nil, errors.New("rmdoc: no .metadata file found in archive")
	}

	docid := strings.TrimSuffix(metadataEntry.Name, models.MetadataFileExt)

	mr, err := metadataEntry.Open()
	if err != nil {
		return nil, err
	}
	metaBytes, err := io.ReadAll(mr)
	mr.Close()
	if err != nil {
		return nil, err
	}

	var metadata models.MetadataFile
	if err := json.Unmarshal(metaBytes, &metadata); err != nil {
		return nil, err
	}

	if parent != "" {
		metadata.Parent = parent
	}
	metadata.Synced = true
	metadata.MetadataModified = true

	metaReader := bytes.NewReader(metaBytes)
	metaHash, metaCRC32C, metaSize, err := models.Hash(metaReader)
	if err != nil {
		return nil, err
	}
	metaReader.Seek(0, io.SeekStart)
	err = bs.impl.StoreBlob(uid, metaHash, "", "crc32c="+metaCRC32C, metaReader)
	if err != nil {
		return nil, err
	}

	hashDoc := models.NewHashDocWithMeta(docid, metadata)
	hashDoc.PayloadType = metadata.DocumentName

	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, models.ContentFileExt) {
			cr, err := f.Open()
			if err == nil {
				var contentFile models.ContentFile
				contentBytes, err := io.ReadAll(cr)
				cr.Close()
				if err == nil {
					if json.Unmarshal(contentBytes, &contentFile) == nil && contentFile.FileType != "" {
						hashDoc.PayloadType = contentFile.FileType
					}
				}
			}
			break
		}
	}

	entry := models.NewHashEntry(metaHash, metadataEntry.Name, metaSize)
	if err := hashDoc.AddFile(entry); err != nil {
		return nil, err
	}

	for _, f := range zr.File {
		if f.Name == metadataEntry.Name {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		fileData, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}

		reader := bytes.NewReader(fileData)
		fileHash, fileCRC32C, fileSize, err := models.Hash(reader)
		if err != nil {
			return nil, err
		}
		reader.Seek(0, io.SeekStart)
		err = bs.impl.StoreBlob(uid, fileHash, "", "crc32c="+fileCRC32C, reader)
		if err != nil {
			return nil, err
		}

		entry := models.NewHashEntry(fileHash, f.Name, fileSize)
		if err := hashDoc.AddFile(entry); err != nil {
			return nil, err
		}
	}

	indexReader, err := hashDoc.IndexReader()
	if err != nil {
		return nil, err
	}

	indexContent, err := io.ReadAll(indexReader)
	if err != nil {
		return nil, err
	}

	crc32cIndex, err := common.CRC32CFromReader(bytes.NewReader(indexContent))
	if err != nil {
		return nil, err
	}

	err = bs.impl.StoreBlob(uid, hashDoc.Hash, "", "crc32c="+crc32cIndex, bytes.NewReader(indexContent))
	if err != nil {
		return nil, err
	}

	tree, err := bs.root.GetCachedTree(uid, bs.RemoteStorage(uid))
	if err != nil {
		return nil, err
	}
	err = UpdateTree(tree, bs, uid, func(t *models.HashTree) error {
		return tree.Add(hashDoc)
	})
	if err != nil {
		return nil, err
	}

	return &Document{
		ID:     docid,
		Type:   metadata.CollectionType,
		Parent: metadata.Parent,
		Name:   metadata.DocumentName,
	}, nil
}

func createMetadataFile(metadata models.MetadataFile) (r io.Reader, filehash string, crc32c string, size int64, err error) {
	jsn, err := json.Marshal(metadata)
	if err != nil {
		return
	}
	reader := bytes.NewReader(jsn)
	filehash, crc32c, size, err = models.Hash(reader)
	if err != nil {
		return
	}
	reader.Seek(0, io.SeekStart)
	r = reader
	return
}

type BlobRemoteStorage struct {
	bs  *BlobStorer
	uid string
}

func (brs *BlobRemoteStorage) GetRootIndex() (hash string, generation int64, err error) {
	return brs.bs.root.GetRootIndex(brs.uid)
}

func (brs *BlobRemoteStorage) GetReader(hash string) (io.ReadCloser, error) {
	rc, _, _, err := brs.bs.LoadBlob(brs.uid, hash)
	return rc, err
}
