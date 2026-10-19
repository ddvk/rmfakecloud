package ui

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/gin-gonic/gin"
)

type renameBackend struct {
	backend
	renamedID   string
	renamedName string
	err         error
	synced      bool
}

func (b *renameBackend) RenameDocument(uid, docID, name string) error {
	b.renamedID = docID
	b.renamedName = name
	return b.err
}

func (b *renameBackend) Sync(uid string) {
	b.synced = true
}

func doRename(b *renameBackend, docID, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	app := &ReactAppWrapper{
		backends: map[common.SyncVersion]backend{common.Sync15: b},
	}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(userIDContextKey, "test")
		c.Set(backendVersionKey, common.Sync15)
	})
	r.PATCH("/documents/:docid", app.renameDocument)

	req := httptest.NewRequest(http.MethodPatch, "/documents/"+docID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRenameDocument(t *testing.T) {
	b := &renameBackend{}
	w := doRename(b, "abc", `{"name":"  New name  "}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if b.renamedID != "abc" || b.renamedName != "New name" {
		t.Errorf("unexpected rename: %q %q", b.renamedID, b.renamedName)
	}
	if !b.synced {
		t.Error("expected a sync notification")
	}
}

func TestRenameDocumentEmptyName(t *testing.T) {
	for _, body := range []string{`{"name":""}`, `{"name":"   "}`, `{}`} {
		b := &renameBackend{}
		w := doRename(b, "abc", body)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", body, w.Code)
		}
		if b.renamedID != "" || b.synced {
			t.Errorf("%s: backend must not be called", body)
		}
	}
}

func TestRenameDocumentBackendError(t *testing.T) {
	b := &renameBackend{err: errors.New("not found")}
	w := doRename(b, "abc", `{"name":"x"}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	if b.synced {
		t.Error("must not notify on failure")
	}
}
