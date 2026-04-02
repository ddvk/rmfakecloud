package fs

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/ddvk/rmfakecloud/internal/storage"
)

func TestGetRootIndex_NoHistory(t *testing.T) {
	fs, _ := newTestStorage(t)
	uid := "testuser"
	setupUserDir(t, fs, uid)

	_, _, err := fs.GetRootIndex(uid)
	if err == nil {
		t.Fatal("expected error for missing history file")
	}
}

func TestGetRootIndex_EmptyHistory(t *testing.T) {
	fs, _ := newTestStorage(t)
	uid := "testuser"
	setupUserDir(t, fs, uid)

	// Create empty history file
	histPath := path.Join(fs.getUserPath(uid), historyFile)
	if err := os.WriteFile(histPath, []byte(""), 0600); err != nil {
		t.Fatal(err)
	}

	_, _, err := fs.GetRootIndex(uid)
	if err != storage.ErrorNotFound {
		t.Fatalf("expected ErrorNotFound, got %v", err)
	}
}

func TestUpdateAndGetRoot(t *testing.T) {
	fs, _ := newTestStorage(t)
	uid := "testuser"
	setupUserDir(t, fs, uid)
	setupBlobDir(t, fs, uid)

	// Use 64-char hashes to match the expected line size (86 = timestamp + space + 64-char hash + newline)
	hash1 := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

	// First update (gen 0 -> 1)
	gen, err := fs.UpdateRoot(uid, strings.NewReader(hash1), 0)
	if err != nil {
		t.Fatal(err)
	}
	if gen != 1 {
		t.Errorf("expected generation 1, got %d", gen)
	}

	// Read back
	rootHash, rootGen, err := fs.GetRootIndex(uid)
	if err != nil {
		t.Fatal(err)
	}
	if rootHash != hash1 {
		t.Errorf("hash mismatch: got %q, want %q", rootHash, hash1)
	}
	if rootGen != 1 {
		t.Errorf("generation mismatch: got %d, want 1", rootGen)
	}

	// Create root blob so the generation check recognizes existing root
	blobPath := path.Join(fs.getUserBlobPath(uid), hash1)
	if err := os.WriteFile(blobPath, []byte("blob"), 0600); err != nil {
		t.Fatal(err)
	}

	// Second update (gen 1 -> 2)
	hash2 := "f1e2d3c4b5a6f1e2d3c4b5a6f1e2d3c4b5a6f1e2d3c4b5a6f1e2d3c4b5a6f1e2"
	gen, err = fs.UpdateRoot(uid, strings.NewReader(hash2), 1)
	if err != nil {
		t.Fatal(err)
	}
	if gen != 2 {
		t.Errorf("expected generation 2, got %d", gen)
	}
}

func TestUpdateRoot_WrongGeneration(t *testing.T) {
	fs, _ := newTestStorage(t)
	uid := "testuser"
	setupUserDir(t, fs, uid)
	setupBlobDir(t, fs, uid)

	// Use 64-char hash for correct generation calculation
	hash1 := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	_, err := fs.UpdateRoot(uid, strings.NewReader(hash1), 0)
	if err != nil {
		t.Fatal(err)
	}

	// Create root blob so generation check recognizes the root exists
	blobPath := path.Join(fs.getUserBlobPath(uid), hash1)
	if err := os.WriteFile(blobPath, []byte("blob"), 0600); err != nil {
		t.Fatal(err)
	}

	// Try to update with wrong generation (99 != 1)
	hash2 := "f1e2d3c4b5a6f1e2d3c4b5a6f1e2d3c4b5a6f1e2d3c4b5a6f1e2d3c4b5a6f1e2"
	_, err = fs.UpdateRoot(uid, strings.NewReader(hash2), 99)
	if err != storage.ErrorWrongGeneration {
		t.Fatalf("expected ErrorWrongGeneration, got %v", err)
	}
}

func TestReadRootIndex_CorruptedHistory(t *testing.T) {
	// Create a history file with bad format
	dir := t.TempDir()
	histPath := path.Join(dir, "corrupted.history")
	if err := os.WriteFile(histPath, []byte("badline\n"), 0600); err != nil {
		t.Fatal(err)
	}

	_, _, err := readRootIndex(histPath)
	if err == nil {
		t.Fatal("expected error for corrupted history")
	}
}
