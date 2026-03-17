package fs

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/danjacques/gofslock/fslock"
	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/epub"
	"github.com/ddvk/rmfakecloud/internal/storage/exporter"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	"github.com/google/uuid"
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
	tree.SchemaVersion = fs.Cfg.HashSchemaVersion
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

// ExportRmDoc exports a document as a zip of all blobs
func (fs *FileSystemStorage) ExportRmDoc(uid, docid string) (io.ReadCloser, error) {
	tree, err := fs.GetCachedTree(uid)
	if err != nil {
		return nil, err
	}
	doc, err := tree.FindDoc(docid)
	if err != nil {
		return nil, err
	}
	ls := fs.BlobStorage(uid)

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

// Export exports a document (as PDF)
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

// GetMethodSVG returns the SVG for a synced Method document by loading its .template blob and decoding iconData.
func (fs *FileSystemStorage) GetMethodSVG(uid, docid string) (string, error) {
	tree, err := fs.GetCachedTree(uid)
	if err != nil {
		return "", err
	}
	doc, err := tree.FindDoc(docid)
	if err != nil {
		return "", err
	}
	var templateHash string
	for _, f := range doc.Files {
		if strings.HasSuffix(strings.ToLower(f.EntryName), storage.TemplateFileExt) {
			templateHash = f.Hash
			break
		}
	}
	if templateHash == "" {
		return "", errors.New("no .template file")
	}
	ls := fs.BlobStorage(uid)
	r, err := ls.GetReader(templateHash)
	if err != nil {
		return "", err
	}
	defer r.Close()
	raw, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	var data struct {
		IconData string `json:"iconData"`
	}
	if err := json.Unmarshal(raw, &data); err != nil || data.IconData == "" {
		return "", errors.New("template has no iconData")
	}
	svg, err := base64.StdEncoding.DecodeString(data.IconData)
	if err != nil {
		return "", err
	}
	return string(svg), nil
}

// ExportEpub returns the raw EPUB payload for a document that has an .epub in its archive.
func (fs *FileSystemStorage) ExportEpub(uid, docid string) (io.ReadCloser, error) {
	tree, err := fs.GetCachedTree(uid)
	if err != nil {
		return nil, err
	}
	doc, err := tree.FindDoc(docid)
	if err != nil {
		return nil, err
	}
	hasEpub := false
	for _, f := range doc.Files {
		if strings.HasSuffix(strings.ToLower(f.EntryName), storage.EpubFileExt) {
			hasEpub = true
			break
		}
	}
	if !hasEpub {
		return nil, errors.New("document has no epub payload")
	}
	ls := fs.BlobStorage(uid)
	archive, err := models.ArchiveFromHashDoc(doc, ls)
	if err != nil {
		return nil, err
	}
	if archive.PayloadReader == nil {
		return nil, errors.New("epub payload not available")
	}
	return archive.PayloadReader, nil
}

// GetTemplate returns the raw .template file for a given entry (if present).
func (fs *FileSystemStorage) GetTemplate(uid, docid string) (r io.ReadCloser, err error) {
	tree, err := fs.GetCachedTree(uid)
	if err != nil {
		return nil, err
	}
	doc, err := tree.FindDoc(docid)
	if err != nil {
		return nil, err
	}
	ls := fs.BlobStorage(uid)
	for _, f := range doc.Files {
		if strings.HasSuffix(strings.ToLower(f.EntryName), storage.TemplateFileExt) {
			return ls.GetReader(f.Hash)
		}
	}
	return nil, errors.New("template not found")
}

// GetEpub returns the raw .epub file for a document (if present).
func (fs *FileSystemStorage) GetEpub(uid, docid string) (io.ReadCloser, error) {
	tree, err := fs.GetCachedTree(uid)
	if err != nil {
		return nil, err
	}
	doc, err := tree.FindDoc(docid)
	if err != nil {
		return nil, err
	}
	ls := fs.BlobStorage(uid)
	for _, f := range doc.Files {
		if strings.HasSuffix(strings.ToLower(f.EntryName), storage.EpubFileExt) {
			return ls.GetReader(f.Hash)
		}
	}
	return nil, errors.New("epub not found")
}

// GetEpubManifest parses the EPUB and returns spine and base path as JSON.
func (fs *FileSystemStorage) GetEpubManifest(uid, docid string) (*epub.Manifest, error) {
	rc, err := fs.GetEpub(uid, docid)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, err
	}
	return epub.ReadManifest(zr)
}

// GetEpubFile returns a reader for a file inside the document's EPUB (path relative to zip root).
func (fs *FileSystemStorage) GetEpubFile(uid, docid, filePath string) (io.ReadCloser, string, error) {
	rc, err := fs.GetEpub(uid, docid)
	if err != nil {
		return nil, "", err
	}
	defer rc.Close()
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, "", err
	}
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, "", err
	}
	f, err := epub.OpenZipFile(zr, filePath)
	if err != nil {
		return nil, "", err
	}
	contentType := epub.ContentType(filePath)
	return f, contentType, nil
}

// GetDocumentMetadata returns document type, hasWritings, and page count for the given doc.
func (fs *FileSystemStorage) GetDocumentMetadata(uid, docid string) (docType string, hasWritings bool, pageCount int, err error) {
	tree, err := fs.GetCachedTree(uid)
	if err != nil {
		return "", false, 0, err
	}
	doc, err := tree.FindDoc(docid)
	if err != nil {
		return "", false, 0, err
	}
	docType = doc.PayloadType
	hasWritings = doc.HasWritings()
	for _, f := range doc.Files {
		if strings.HasSuffix(strings.ToLower(f.EntryName), storage.ContentFileExt) {
			rc, err := fs.BlobStorage(uid).GetReader(f.Hash)
			if err != nil {
				return docType, hasWritings, 0, err
			}
			defer rc.Close()
			var content models.ContentFile
			if err := json.NewDecoder(rc).Decode(&content); err != nil {
				return docType, hasWritings, 0, err
			}
			if content.PageCount > 0 {
				pageCount = content.PageCount
			} else {
				pageCount = len(content.Pages)
			}
			return docType, hasWritings, pageCount, nil
		}
	}
	return docType, hasWritings, pageCount, nil
}

// ExportPagePNG exports a single page of the document as PNG (1-based page number).
func (fs *FileSystemStorage) ExportPagePNG(uid, docid string, pageNum int) (io.ReadCloser, error) {
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
	return exporter.RenderPagePNGReader(archive, pageNum)
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
	origExt := path.Ext(filename)
	ext := strings.ToLower(origExt)
	switch ext {
	case storage.EpubFileExt, storage.PdfFileExt, storage.RmDocFileExt, storage.TemplateFileExt:
	default:
		return nil, errors.New("unsupported extension: " + ext)
	}

	if ext == storage.RmDocFileExt {
		// Decode .rmdoc container (from mymod: full inline with template support and corrected metadata).
		tmpFile, err := os.CreateTemp("", "rmdoc-*")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tmpFile.Name())

		if _, err = io.Copy(tmpFile, stream); err != nil {
			tmpFile.Close()
			return nil, err
		}
		if err = tmpFile.Close(); err != nil {
			return nil, err
		}

		zr, err := zip.OpenReader(tmpFile.Name())
		if err != nil {
			return nil, err
		}
		defer zr.Close()

		var docid string
		for _, f := range zr.File {
			base := path.Base(f.Name)
			low := strings.ToLower(base)
			switch {
			case strings.HasSuffix(low, storage.MetadataFileExt),
				strings.HasSuffix(low, storage.ContentFileExt),
				strings.HasSuffix(low, storage.TemplateFileExt),
				strings.HasSuffix(low, storage.PdfFileExt),
				strings.HasSuffix(low, storage.EpubFileExt):
				docid = strings.TrimSuffix(base, path.Ext(base))
				break
			}
			if docid != "" {
				break
			}
		}
		if docid == "" {
			return nil, errors.New("rmdoc: could not determine document id")
		}

		var metaFile, contentFile, templateFile, pdfFile, epubFile *zip.File
		rmFiles := []*zip.File{}
		pagedataFiles := []*zip.File{}
		for _, f := range zr.File {
			low := strings.ToLower(f.Name)
			switch {
			case f.Name == docid+storage.MetadataFileExt:
				metaFile = f
			case f.Name == docid+storage.ContentFileExt:
				contentFile = f
			case f.Name == docid+storage.TemplateFileExt:
				templateFile = f
			case f.Name == docid+storage.PdfFileExt:
				pdfFile = f
			case f.Name == docid+storage.EpubFileExt:
				epubFile = f
			case strings.HasSuffix(low, storage.PageFileExt):
				pagedataFiles = append(pagedataFiles, f)
			case strings.HasSuffix(low, storage.RmFileExt):
				rmFiles = append(rmFiles, f)
			}
		}

		payloadFile := pdfFile
		payloadExt := storage.PdfFileExt
		if payloadFile == nil && epubFile != nil {
			payloadFile = epubFile
			payloadExt = storage.EpubFileExt
		}
		isTemplate := templateFile != nil && payloadFile == nil

		embeddedMeta := models.MetadataFile{}
		metaOK := false
		if metaFile != nil {
			rc, err := metaFile.Open()
			if err != nil {
				return nil, err
			}
			metaBytes, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(metaBytes, &embeddedMeta); err == nil {
				metaOK = true
			}
		}

		embeddedContent := models.ContentFile{}
		contentOK := false
		if contentFile != nil {
			rc, err := contentFile.Open()
			if err != nil {
				return nil, err
			}
			contentBytes, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(contentBytes, &embeddedContent); err == nil {
				contentOK = true
			}
		}

		now := models.FromTime(time.Now())
		docName := ""
		if metaOK && embeddedMeta.DocumentName != "" {
			docName = embeddedMeta.DocumentName
		}
		if docName == "" {
			docName = strings.TrimSuffix(filename, origExt)
		}

		correctedMeta := embeddedMeta
		if !metaOK {
			correctedMeta = models.MetadataFile{}
			correctedMeta.CreatedTime = now
			correctedMeta.LastModified = now
		}
		correctedMeta.DocumentName = docName
		correctedMeta.Parent = parent
		correctedMeta.Version = 1
		correctedMeta.Synced = true
		correctedMeta.MetadataModified = true
		if correctedMeta.CreatedTime == "" {
			correctedMeta.CreatedTime = now
		}
		if correctedMeta.LastModified == "" {
			correctedMeta.LastModified = now
		}
		if isTemplate {
			correctedMeta.CollectionType = common.EntryType("TemplateType")
		} else {
			correctedMeta.CollectionType = common.DocumentType
		}

		correctedContent := embeddedContent
		if !contentOK {
			correctedContent = models.ContentFile{}
		}
		if isTemplate {
			correctedContent.FileType = "template"
		} else {
			correctedContent.FileType = strings.TrimPrefix(payloadExt, ".")
		}

		blobStorage := fs.BlobStorage(uid)
		tree, err := fs.GetCachedTree(uid)
		if err != nil {
			return nil, err
		}

		hashDoc := models.NewHashDocWithMeta(docid, correctedMeta)
		hashDoc.PayloadType = correctedContent.FileType

		writeZipEntry := func(zf *zip.File) (hash string, size int64, err error) {
			tf, err := os.CreateTemp(fs.getUserBlobPath(uid), "rmdoc-entry-*")
			if err != nil {
				return "", 0, err
			}
			defer os.Remove(tf.Name())
			defer tf.Close()

			rc, err := zf.Open()
			if err != nil {
				return "", 0, err
			}
			defer rc.Close()

			tee := io.TeeReader(rc, tf)
			hash, size, err = models.Hash(tee)
			if err != nil {
				return "", size, err
			}
			if _, err := tf.Seek(0, io.SeekStart); err != nil {
				return "", size, err
			}
			if err := blobStorage.Write(hash, tf); err != nil {
				return "", size, err
			}
			return hash, size, nil
		}

		metaJSON, err := json.Marshal(correctedMeta)
		if err != nil {
			return nil, err
		}
		metaReader := bytes.NewReader(metaJSON)
		metaHash, metaSize, err := models.Hash(metaReader)
		if err != nil {
			return nil, err
		}
		if _, err := metaReader.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		if err := blobStorage.Write(metaHash, metaReader); err != nil {
			return nil, err
		}
		if err := hashDoc.AddFile(models.NewHashEntry(metaHash, docid+storage.MetadataFileExt, metaSize)); err != nil {
			return nil, err
		}

		var payloadSize int64
		if isTemplate && templateFile != nil {
			payloadSize = int64(templateFile.UncompressedSize64)
		} else if payloadFile != nil {
			payloadSize = int64(payloadFile.UncompressedSize64)
		}
		if payloadSize > 0 {
			correctedContent.SizeInBytes = fmt.Sprintf("%d", payloadSize)
		}
		contentJSON, err := json.Marshal(correctedContent)
		if err != nil {
			return nil, err
		}
		contentReader := bytes.NewReader(contentJSON)
		contentHash, contentSize, err := models.Hash(contentReader)
		if err != nil {
			return nil, err
		}
		if _, err := contentReader.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		if err := blobStorage.Write(contentHash, contentReader); err != nil {
			return nil, err
		}
		if err := hashDoc.AddFile(models.NewHashEntry(contentHash, docid+storage.ContentFileExt, contentSize)); err != nil {
			return nil, err
		}

		for _, pf := range pagedataFiles {
			h, s, err := writeZipEntry(pf)
			if err != nil {
				return nil, err
			}
			if err := hashDoc.AddFile(models.NewHashEntry(h, pf.Name, s)); err != nil {
				return nil, err
			}
		}

		for _, rf := range rmFiles {
			h, s, err := writeZipEntry(rf)
			if err != nil {
				return nil, err
			}
			if err := hashDoc.AddFile(models.NewHashEntry(h, rf.Name, s)); err != nil {
				return nil, err
			}
		}

		if isTemplate {
			if templateFile == nil {
				return nil, errors.New("rmdoc: template type but no .template found")
			}
			h, s, err := writeZipEntry(templateFile)
			if err != nil {
				return nil, err
			}
			if err := hashDoc.AddFile(models.NewHashEntry(h, docid+storage.TemplateFileExt, s)); err != nil {
				return nil, err
			}
			ext = storage.TemplateFileExt
		} else {
			if payloadFile == nil {
				return nil, errors.New("rmdoc: no supported document payload (pdf/epub) found")
			}
			h, s, err := writeZipEntry(payloadFile)
			if err != nil {
				return nil, err
			}
			if err := hashDoc.AddFile(models.NewHashEntry(h, docid+payloadExt, s)); err != nil {
				return nil, err
			}
			ext = payloadExt
		}

		indexReader, err := hashDoc.IndexReader()
		if err != nil {
			return nil, err
		}
		if err := blobStorage.Write(hashDoc.Hash, indexReader); err != nil {
			return nil, err
		}

		if err := updateTree(tree, blobStorage, func(t *models.HashTree) error {
			return tree.Add(hashDoc)
		}); err != nil {
			return nil, err
		}

		return &storage.Document{
			ID:     docid,
			Type:   correctedMeta.CollectionType,
			Parent: "",
			Name:   docName,
		}, nil
	}

	blobPath := fs.getUserBlobPath(uid)
	docid := uuid.New().String()
	docName := strings.TrimSuffix(filename, origExt)

	tree, err := fs.GetCachedTree(uid)
	if err != nil {
		return nil, err
	}

	log.Info("Creating metadata... parent: ", parent)

	collectionType := common.DocumentType
	if ext == storage.TemplateFileExt {
		collectionType = common.EntryType("TemplateType")
	}
	metadata := models.MetadataFile{
		DocumentName:     docName,
		CollectionType:   collectionType,
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

	content := "{}"
	if ext != storage.TemplateFileExt {
		content = createContent(ext)
	}

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
		Type:   collectionType,
		Parent: "",
		Name:   docName,
	}
	return
}

func (fs *FileSystemStorage) createFromRmDoc(uid, parent string, stream io.Reader) (*storage.Document, error) {
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
		if strings.HasSuffix(f.Name, storage.MetadataFileExt) {
			metadataEntry = f
			break
		}
	}
	if metadataEntry == nil {
		return nil, errors.New("rmdoc: no .metadata file found in archive")
	}

	docid := strings.TrimSuffix(metadataEntry.Name, storage.MetadataFileExt)

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

	blobStorage := fs.BlobStorage(uid)

	metaReader := bytes.NewReader(metaBytes)
	metaHash, metaSize, err := models.Hash(metaReader)
	if err != nil {
		return nil, err
	}
	metaReader.Seek(0, io.SeekStart)
	if err := blobStorage.Write(metaHash, metaReader); err != nil {
		return nil, err
	}

	hashDoc := models.NewHashDocWithMeta(docid, metadata)
	hashDoc.PayloadType = metadata.DocumentName

	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, storage.ContentFileExt) {
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
		fileHash, fileSize, err := models.Hash(reader)
		if err != nil {
			return nil, err
		}
		reader.Seek(0, io.SeekStart)
		if err := blobStorage.Write(fileHash, reader); err != nil {
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
	if err := blobStorage.Write(hashDoc.Hash, indexReader); err != nil {
		return nil, err
	}

	tree, err := fs.GetCachedTree(uid)
	if err != nil {
		return nil, err
	}
	err = updateTree(tree, blobStorage, func(t *models.HashTree) error {
		return tree.Add(hashDoc)
	})
	if err != nil {
		return nil, err
	}

	return &storage.Document{
		ID:     docid,
		Type:   metadata.CollectionType,
		Parent: metadata.Parent,
		Name:   metadata.DocumentName,
	}, nil
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
		lock, err := fslock.Lock(historyPath)
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
		var lock fslock.Handle
		lock, err = fslock.Lock(historyPath)
		if err != nil {
			log.Error("cannot obtain lock")
			return 0, err
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
