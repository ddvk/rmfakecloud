package fs

import (
	"archive/zip"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	"github.com/google/uuid"
)

func (fs *FileSystemStorage) CreateFolder(uid, name string) (*storage.Document, error) {
	docID := uuid.New().String()

	// Create zip file
	zipFilePath := fs.getPathFromUser(uid, docID+models.ZipFileExt)

	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return nil, err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	contentEntry, err := zipWriter.Create(docID + models.ContentFileExt)
	if err != nil {
		return nil, err
	}
	contentEntry.Write([]byte(`{"tags":[]}`))

	// Create metadata file
	mdFilePath := fs.getPathFromUser(uid, docID+models.MetadataFileExt)

	mdFile, err := os.Create(mdFilePath)
	if err != nil {
		return nil, err
	}
	defer mdFile.Close()

	md := messages.RawMetadata{
		ID:             docID,
		VissibleName:   strings.TrimSpace(name),
		Version:        1,
		ModifiedClient: time.Now().UTC().Format(time.RFC3339Nano),
		Type:           models.CollectionType,
	}

	if err = json.NewEncoder(mdFile).Encode(md); err != nil {
		return nil, err
	}

	doc := &storage.Document{
		ID:      md.ID,
		Type:    md.Type,
		Name:    md.VissibleName,
		Version: md.Version,
	}

	return doc, nil
}
