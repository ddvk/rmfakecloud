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
	err := os.MkdirAll(path.Join(dir, userDir, testuser, SyncFolder), 0700)
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

func TestMoveBlobDocument(t *testing.T) {
	testuser := "test"
	dir := t.TempDir()
	err := os.MkdirAll(path.Join(dir, userDir, testuser, SyncFolder), 0700)
	if err != nil {
		t.Fatal(err)
	}
	fs := NewStorage(&config.Config{DataDir: dir})

	a, err := fs.CreateBlobFolder(testuser, "a", "")
	if err != nil {
		t.Fatal(err)
	}
	b, err := fs.CreateBlobFolder(testuser, "b", "")
	if err != nil {
		t.Fatal(err)
	}

	err = fs.MoveBlobDocument(testuser, b.ID, a.ID)
	if err != nil {
		t.Fatal(err)
	}

	tree, err := fs.GetCachedTree(testuser)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := tree.FindDoc(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Parent != a.ID {
		t.Errorf("parent not updated: %q", doc.Parent)
	}
	if doc.Version != 2 {
		t.Errorf("version not bumped: %d", doc.Version)
	}
	if doc.DocumentName != "b" {
		t.Errorf("name changed: %q", doc.DocumentName)
	}
	if doc.CollectionType != common.CollectionType {
		t.Errorf("type changed: %q", doc.CollectionType)
	}

	// a cannot go into its child b, into itself, or into a missing folder
	for _, parent := range []string{b.ID, a.ID, "missing", "trash"} {
		err = fs.MoveBlobDocument(testuser, a.ID, parent)
		if err == nil {
			t.Errorf("expected an error moving into %q", parent)
		}
	}
	err = fs.MoveBlobDocument(testuser, "missing", "")
	if err == nil {
		t.Error("expected an error for a missing document")
	}

	tree, err = fs.GetCachedTree(testuser)
	if err != nil {
		t.Fatal(err)
	}
	doc, err = tree.FindDoc(a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Parent != "" || doc.Version != 1 {
		t.Errorf("rejected move changed the folder: %q %d", doc.Parent, doc.Version)
	}

	// back to the root
	err = fs.MoveBlobDocument(testuser, b.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	tree, err = fs.GetCachedTree(testuser)
	if err != nil {
		t.Fatal(err)
	}
	doc, err = tree.FindDoc(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Parent != "" || doc.Version != 3 {
		t.Errorf("move to root failed: %q %d", doc.Parent, doc.Version)
	}
}
