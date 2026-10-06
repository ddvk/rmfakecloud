package fs

import (
	"os"
	"path"
	"testing"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/config"
)

func TestRenameBlobDocument(t *testing.T) {
	testuser := "test"
	dir := t.TempDir()
	err := os.MkdirAll(path.Join(dir, userDir, testuser), 0700)
	if err != nil {
		t.Fatal(err)
	}
	fs := NewStorage(&config.Config{DataDir: dir})

	folder, err := fs.CreateBlobFolder(testuser, "old", "parentid")
	if err != nil {
		t.Fatal(err)
	}

	err = fs.RenameBlobDocument(testuser, folder.ID, "new")
	if err != nil {
		t.Fatal(err)
	}

	tree, err := fs.GetCachedTree(testuser)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := tree.FindDoc(folder.ID)
	if err != nil {
		t.Fatal(err)
	}
	if doc.DocumentName != "new" {
		t.Errorf("name not updated: %q", doc.DocumentName)
	}
	if doc.Version != 2 {
		t.Errorf("version not bumped: %d", doc.Version)
	}
	if doc.Parent != "parentid" {
		t.Errorf("parent changed: %q", doc.Parent)
	}
	if doc.CollectionType != common.CollectionType {
		t.Errorf("type changed: %q", doc.CollectionType)
	}

	err = fs.RenameBlobDocument(testuser, "missing", "x")
	if err == nil {
		t.Error("expected an error for a missing document")
	}
}
