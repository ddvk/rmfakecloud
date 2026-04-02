package fs

import (
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/ddvk/rmfakecloud/internal/storage/models"
)

func setupUserDir(t *testing.T, fs *FileSystemStorage, uid string) {
	t.Helper()
	userPath := fs.getUserPath(uid)
	if err := os.MkdirAll(userPath, 0700); err != nil {
		t.Fatal(err)
	}
}

func TestStoreAndGetDocument(t *testing.T) {
	fs, _ := newTestStorage(t)
	uid := "testuser"
	setupUserDir(t, fs, uid)

	docID := "test-doc-id"
	content := "fake zip content"

	err := fs.StoreDocument(uid, docID, io.NopCloser(strings.NewReader(content)))
	if err != nil {
		t.Fatal(err)
	}

	reader, err := fs.GetDocument(uid, docID)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != content {
		t.Errorf("content mismatch: got %q, want %q", string(data), content)
	}
}

func TestGetDocument_NotFound(t *testing.T) {
	fs, _ := newTestStorage(t)
	uid := "testuser"
	setupUserDir(t, fs, uid)

	_, err := fs.GetDocument(uid, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent document")
	}
}

func TestRemoveDocument(t *testing.T) {
	fs, dir := newTestStorage(t)
	uid := "testuser"
	setupUserDir(t, fs, uid)

	docID := "doc-to-remove"
	userPath := path.Join(dir, userDir, uid)

	// Create both metadata and zip files
	metaPath := path.Join(userPath, docID+models.MetadataFileExt)
	zipPath := path.Join(userPath, docID+models.ZipFileExt)

	if err := os.WriteFile(metaPath, []byte(`{"ID":"doc-to-remove"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(zipPath, []byte("zip content"), 0600); err != nil {
		t.Fatal(err)
	}

	err := fs.RemoveDocument(uid, docID)
	if err != nil {
		t.Fatal(err)
	}

	// Files should be in trash, not in original location
	if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
		t.Error("metadata file should be moved from original location")
	}
	if _, err := os.Stat(zipPath); !os.IsNotExist(err) {
		t.Error("zip file should be moved from original location")
	}

	// Check trash directory
	trashDir := path.Join(userPath, DefaultTrashDir)
	if _, err := os.Stat(path.Join(trashDir, docID+models.MetadataFileExt)); err != nil {
		t.Error("metadata file should exist in trash")
	}
	if _, err := os.Stat(path.Join(trashDir, docID+models.ZipFileExt)); err != nil {
		t.Error("zip file should exist in trash")
	}
}

func TestCreateFolder(t *testing.T) {
	fs, dir := newTestStorage(t)
	uid := "testuser"
	setupUserDir(t, fs, uid)

	doc, err := fs.CreateFolder(uid, "My Folder", "")
	if err != nil {
		t.Fatal(err)
	}

	if doc.Name != "My Folder" {
		t.Errorf("name mismatch: got %q", doc.Name)
	}
	if doc.ID == "" {
		t.Error("expected non-empty document ID")
	}

	// Check metadata file exists
	userPath := path.Join(dir, userDir, uid)
	metaPath := path.Join(userPath, doc.ID+models.MetadataFileExt)
	if _, err := os.Stat(metaPath); err != nil {
		t.Errorf("metadata file not created: %v", err)
	}

	// Check zip file exists
	zipPath := path.Join(userPath, doc.ID+models.ZipFileExt)
	if _, err := os.Stat(zipPath); err != nil {
		t.Errorf("zip file not created: %v", err)
	}
}

func TestCreateDocument_PDF(t *testing.T) {
	fs, dir := newTestStorage(t)
	uid := "testuser"
	setupUserDir(t, fs, uid)

	content := strings.NewReader("fake pdf content")
	doc, err := fs.CreateDocument(uid, "test.pdf", "", content)
	if err != nil {
		t.Fatal(err)
	}

	if doc.Name != "test" {
		t.Errorf("name mismatch: got %q, want %q", doc.Name, "test")
	}
	if doc.ID == "" {
		t.Error("expected non-empty document ID")
	}

	// Verify files were created
	userPath := path.Join(dir, userDir, uid)
	if _, err := os.Stat(path.Join(userPath, doc.ID+models.MetadataFileExt)); err != nil {
		t.Error("metadata file not created")
	}
	if _, err := os.Stat(path.Join(userPath, doc.ID+models.ZipFileExt)); err != nil {
		t.Error("zip file not created")
	}
}

func TestCreateDocument_EPUB(t *testing.T) {
	fs, _ := newTestStorage(t)
	uid := "testuser"
	setupUserDir(t, fs, uid)

	content := strings.NewReader("fake epub content")
	doc, err := fs.CreateDocument(uid, "book.epub", "", content)
	if err != nil {
		t.Fatal(err)
	}

	if doc.Name != "book" {
		t.Errorf("name mismatch: got %q, want %q", doc.Name, "book")
	}
}

func TestCreateDocument_UnsupportedType(t *testing.T) {
	fs, _ := newTestStorage(t)
	uid := "testuser"
	setupUserDir(t, fs, uid)

	content := strings.NewReader("data")
	_, err := fs.CreateDocument(uid, "file.docx", "", content)
	if err == nil {
		t.Fatal("expected error for unsupported file type")
	}
}
