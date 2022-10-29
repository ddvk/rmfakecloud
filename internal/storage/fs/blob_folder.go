package fs

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	"github.com/google/uuid"
)

// CreateBlobFolder create a sync15 folder
func (fs *FileSystemStorage) CreateBlobFolder(uid, name, parent string) (*storage.Document, error) {
	name = strings.TrimSpace(name)
	parent = strings.TrimSpace(parent)

	tree, err := fs.GetTree(uid)

	if err != nil {
		return nil, err
	}

	// Check parent
	if len(parent) != 0 {
		parentDoc, err := tree.FindDoc(parent)

		if err != nil {
			return nil, err
		}

		if parentDoc.MetadataFile.CollectionType != models.CollectionType {
			return nil, errors.New("Parent is not a folder")
		}
	}

	docId := uuid.New().String()
	blobPath := fs.getUserBlobPath(uid)

	// Create metadata
	metadata := models.MetadataFile{
		DocumentName:     name,
		CollectionType:   models.CollectionType,
		Parent:           parent,
		Version:          1,
		LastModified:     strconv.FormatInt(time.Now().UnixMilli(), 10),
		Synced:           true,
		MetadataModified: true,
	}

	mdHash, mdSize, err := createMetadataFile(metadata, blobPath)

	if err != nil {
		return nil, err
	}

	mdHashEntry := models.NewFileHashEntry(mdHash, docId+models.MetadataFileExt)
	mdHashEntry.Size = mdSize

	// New hash doc
	hashDoc := models.NewHashDocMeta(docId, metadata)

	if err = hashDoc.AddFile(mdHashEntry); err != nil {
		return nil, err
	}

	// Create content
	content := `{"tags":[]}`
	contentHash, contentSize, err := models.Hash(strings.NewReader(content))
	if err != nil {
		return nil, err
	}
	if err = saveTo(strings.NewReader(content), contentHash, blobPath); err != nil {
		return nil, err
	}
	contentHashEntry := models.NewFileHashEntry(contentHash, docId+models.ContentFileExt)
	contentHashEntry.Size = contentSize
	if err = hashDoc.AddFile(contentHashEntry); err != nil {
		return nil, err
	}

	if err = tree.Add(hashDoc); err != nil {
		return nil, err
	}

	hashDocIndexReader, err := hashDoc.IndexReader()
	if err != nil {
		return nil, err
	}
	if err = saveTo(hashDocIndexReader, hashDoc.Hash, blobPath); err != nil {
		return nil, err
	}

	rootIndexReader, err := tree.RootIndex()
	if err != nil {
		return nil, err
	}
	if err = saveTo(rootIndexReader, tree.Hash, blobPath); err != nil {
		return nil, err
	}

	blobStorage := &LocalBlobStorage{
		fs:  fs,
		uid: uid,
	}

	gen, err := blobStorage.WriteRootIndex(tree.Generation, tree.Hash)
	if err != nil {
		return nil, err
	}
	tree.Generation = gen

	if err = fs.SaveTree(uid, tree); err != nil {
		return nil, err
	}

	return &storage.Document{
		ID:      docId,
		Type:    models.CollectionType,
		Name:    name,
		Version: 1,
	}, nil
}
