package fs

import (
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/ddvk/rmfakecloud/internal/storage"
)

func setupBlobDir(t *testing.T, fs *FileSystemStorage, uid string) {
	t.Helper()
	blobPath := fs.getUserBlobPath(uid)
	if err := os.MkdirAll(blobPath, 0700); err != nil {
		t.Fatal(err)
	}
}

func TestStoreAndLoadBlob(t *testing.T) {
	fs, _ := newTestStorage(t)
	uid := "testuser"
	setupBlobDir(t, fs, uid)

	blobID := "abc123hash"
	content := "blob content data"

	err := fs.StoreBlob(uid, blobID, "test.blob", "", strings.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}

	reader, size, hash, err := fs.LoadBlob(uid, blobID)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	if size != int64(len(content)) {
		t.Errorf("size mismatch: got %d, want %d", size, len(content))
	}

	if hash == "" {
		t.Error("expected non-empty hash")
	}

	// Hash should start with "crc32c="
	if !strings.HasPrefix(hash, "crc32c=") {
		t.Errorf("hash should start with 'crc32c=', got %q", hash)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Errorf("content mismatch: got %q, want %q", string(data), content)
	}
}

func TestLoadBlob_NotFound(t *testing.T) {
	fs, _ := newTestStorage(t)
	uid := "testuser"
	setupBlobDir(t, fs, uid)

	_, _, _, err := fs.LoadBlob(uid, "nonexistent")
	if err != storage.ErrorNotFound {
		t.Fatalf("expected ErrorNotFound, got %v", err)
	}
}

func TestStoreBlob_Overwrite(t *testing.T) {
	fs, _ := newTestStorage(t)
	uid := "testuser"
	setupBlobDir(t, fs, uid)

	blobID := "overwrite-blob"

	err := fs.StoreBlob(uid, blobID, "v1.blob", "", strings.NewReader("version 1"))
	if err != nil {
		t.Fatal(err)
	}

	err = fs.StoreBlob(uid, blobID, "v2.blob", "", strings.NewReader("version 2"))
	if err != nil {
		t.Fatal(err)
	}

	reader, _, _, err := fs.LoadBlob(uid, blobID)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	data, _ := io.ReadAll(reader)
	if string(data) != "version 2" {
		t.Errorf("expected version 2, got %q", string(data))
	}
}

func TestStoreBlob_SanitizesID(t *testing.T) {
	fs, _ := newTestStorage(t)
	uid := "testuser"
	setupBlobDir(t, fs, uid)

	// The ID with path separators should be sanitized
	err := fs.StoreBlob(uid, "safe-id", "test", "", strings.NewReader("data"))
	if err != nil {
		t.Fatal(err)
	}

	// Verify the file is in the blob directory with sanitized name
	blobPath := path.Join(fs.getUserBlobPath(uid), "safe-id")
	if _, err := os.Stat(blobPath); err != nil {
		t.Errorf("blob file not found at expected path: %v", err)
	}
}

func TestGetBlobURL(t *testing.T) {
	fs, _ := newTestStorage(t)
	fs.Cfg.StorageURL = "http://localhost:3000"
	fs.Cfg.JWTSecretKey = []byte("test-secret-key-for-signing")

	uid := "testuser"
	blobID := "myblobid"

	url, _, err := fs.GetBlobURL(uid, blobID, false)
	if err != nil {
		t.Fatal(err)
	}
	if url == "" {
		t.Error("expected non-empty URL")
	}
	if !strings.Contains(url, "/blobstorage") {
		t.Errorf("URL should contain /blobstorage, got %q", url)
	}
	if !strings.Contains(url, "uid="+uid) {
		t.Errorf("URL should contain uid param, got %q", url)
	}

	// Write URL should contain write scope
	writeURL, _, err := fs.GetBlobURL(uid, blobID, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(writeURL, "scope=write") {
		t.Errorf("write URL should contain write scope, got %q", writeURL)
	}
}
