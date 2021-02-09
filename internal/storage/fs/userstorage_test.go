package fs

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ddvk/rmfakecloud/internal/config"
)

var cfg *config.Config
var s *Storage

const (
	tmpPath  = "/tmp/rmfake"
	testUser = "testuid"
)

func init() {
	cfg = &config.Config{
		DataDir: tmpPath,
	}

	s = &Storage{
		Cfg: cfg,
	}

}

func TestGetUserPath(t *testing.T) {
}

func TestUploadFile(t *testing.T) {
	r := ioutil.NopCloser(bytes.NewReader([]byte("dummy content")))

	path := s.getPathFromUser(testUser, "")
	if path != "/tmp/rmfake/users/testuid" {
		t.Log("wrong path")
		t.Fail()
	}
	err := os.MkdirAll(path, 0700)
	if err != nil {
		t.Error(err)
	}
	err = s.StoreDocument(testUser, r, "documentId")
	if err != nil {
		t.Error(err)
	}
	rr, err := s.GetDocument(testUser, "documentId")
	defer rr.Close()

	if err != nil {
		t.Error(err)
	}
}
