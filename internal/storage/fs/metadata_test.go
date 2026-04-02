package fs

import (
	"testing"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/messages"
)

func TestUpdateAndGetMetadata(t *testing.T) {
	fs, _ := newTestStorage(t)
	uid := "testuser"
	setupUserDir(t, fs, uid)

	meta := &messages.RawMetadata{
		ID:           "doc-123",
		VissibleName: "My Document",
		Version:      1,
		Type:         common.DocumentType,
		Parent:       "",
	}

	if err := fs.UpdateMetadata(uid, meta); err != nil {
		t.Fatal(err)
	}

	loaded, err := fs.GetMetadata(uid, "doc-123")
	if err != nil {
		t.Fatal(err)
	}

	if loaded.ID != meta.ID {
		t.Errorf("ID mismatch: got %q, want %q", loaded.ID, meta.ID)
	}
	if loaded.VissibleName != meta.VissibleName {
		t.Errorf("Name mismatch: got %q, want %q", loaded.VissibleName, meta.VissibleName)
	}
	if loaded.Type != meta.Type {
		t.Errorf("Type mismatch: got %q, want %q", loaded.Type, meta.Type)
	}
}

func TestGetMetadata_NotFound(t *testing.T) {
	fs, _ := newTestStorage(t)
	uid := "testuser"
	setupUserDir(t, fs, uid)

	_, err := fs.GetMetadata(uid, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent metadata")
	}
}

func TestGetAllMetadata(t *testing.T) {
	fs, _ := newTestStorage(t)
	uid := "testuser"
	setupUserDir(t, fs, uid)

	docs := []*messages.RawMetadata{
		{ID: "doc-1", VissibleName: "Doc 1", Version: 1, Type: common.DocumentType},
		{ID: "doc-2", VissibleName: "Doc 2", Version: 1, Type: common.DocumentType},
		{ID: "doc-3", VissibleName: "Doc 3", Version: 1, Type: common.CollectionType},
	}

	for _, meta := range docs {
		if err := fs.UpdateMetadata(uid, meta); err != nil {
			t.Fatal(err)
		}
	}

	all, err := fs.GetAllMetadata(uid)
	if err != nil {
		t.Fatal(err)
	}

	if len(all) != 3 {
		t.Fatalf("expected 3 metadata entries, got %d", len(all))
	}
}

func TestGetAllMetadata_Empty(t *testing.T) {
	fs, _ := newTestStorage(t)
	uid := "testuser"
	setupUserDir(t, fs, uid)

	all, err := fs.GetAllMetadata(uid)
	if err != nil {
		t.Fatal(err)
	}

	if len(all) != 0 {
		t.Fatalf("expected 0 metadata entries, got %d", len(all))
	}
}

func TestUpdateMetadata_Overwrite(t *testing.T) {
	fs, _ := newTestStorage(t)
	uid := "testuser"
	setupUserDir(t, fs, uid)

	meta := &messages.RawMetadata{
		ID:           "doc-overwrite",
		VissibleName: "Original",
		Version:      1,
		Type:         common.DocumentType,
	}
	if err := fs.UpdateMetadata(uid, meta); err != nil {
		t.Fatal(err)
	}

	meta.VissibleName = "Updated"
	meta.Version = 2
	if err := fs.UpdateMetadata(uid, meta); err != nil {
		t.Fatal(err)
	}

	loaded, err := fs.GetMetadata(uid, "doc-overwrite")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.VissibleName != "Updated" {
		t.Errorf("expected updated name, got %q", loaded.VissibleName)
	}
	if loaded.Version != 2 {
		t.Errorf("expected version 2, got %d", loaded.Version)
	}
}
