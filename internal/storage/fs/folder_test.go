package fs

import (
	"os"
	"path"
	"testing"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
)

func TestFileCreateFolder_CreateFolder(t *testing.T) {
	testuser := "test"
	dir := path.Join(os.TempDir(), "rmfake")
	userdir := path.Join(dir, userDir, testuser)

	err := os.MkdirAll(userdir, 0700)

	if err != nil {
		t.Error(err)
	}

	defer os.RemoveAll(userDir)

	cfg := &config.Config{
		DataDir: dir,
	}

	fs := NewStorage(cfg)

	doc, err := fs.CreateFolder(testuser, "TestDir", "")
	if err != nil {
		t.Error(err)
	}

	if doc == nil {
		t.Error("document should not be nil")
	}

	_, err = os.Stat(path.Join(userdir, doc.ID+models.MetadataFileExt))
	if err != nil {
		t.Error(err)
	}

	_, err = os.Stat(path.Join(userdir, doc.ID+models.ZipFileExt))
	if err != nil {
		t.Error(err)
	}

}
