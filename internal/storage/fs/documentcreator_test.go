package fs

import (
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
)

func TestCreateDocument(t *testing.T) {
	testuser := "test"
	dir := path.Join(os.TempDir(), "rmfake")
	userdir := path.Join(dir, userDir, testuser)

	err := os.MkdirAll(userdir, 0700)

	if err != nil {
		t.Error(err)
	}
	cfg := &config.Config{
		DataDir: dir,
	}
	content := io.NopCloser(strings.NewReader("dummy"))
	fs := NewStorage(cfg)

	d, err := fs.CreateDocument(testuser, "blah.pdf", "", content)
	if err != nil {
		t.Error(err)
	}

	_, err = os.Stat(path.Join(userdir, d.ID+models.MetadataFileExt))
	if err != nil {
		t.Error(err)
	}

	_, err = os.Stat(path.Join(userdir, d.ID+models.ZipFileExt))
	if err != nil {
		t.Error(err)
	}

}
