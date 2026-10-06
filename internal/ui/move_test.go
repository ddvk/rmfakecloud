package ui

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
	"github.com/gin-gonic/gin"
)

type moveBackend struct {
	backend
	movedID     string
	movedParent string
	moved       bool
	err         error
	synced      bool
}

// folders: a > b > c, folder d, document doc (in a)
func (b *moveBackend) GetDocumentTree(uid string) (*viewmodel.DocumentTree, error) {
	folder := common.CollectionType
	docs := []*viewmodel.InternalDoc{
		{ID: "a", Name: "a", Type: folder},
		{ID: "b", Name: "b", Type: folder, Parent: "a"},
		{ID: "c", Name: "c", Type: folder, Parent: "b"},
		{ID: "d", Name: "d", Type: folder},
		{ID: "doc", Name: "doc", Type: common.DocumentType, Parent: "a"},
	}
	return viewmodel.DocTreeFromRawMetadata(docs), nil
}

func (b *moveBackend) MoveDocument(uid, docID, parent string) error {
	b.moved = true
	b.movedID = docID
	b.movedParent = parent
	return b.err
}

func (b *moveBackend) Sync(uid string) {
	b.synced = true
}

func doMove(b *moveBackend, docID, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	app := &ReactAppWrapper{
		backends: map[common.SyncVersion]backend{common.Sync15: b},
	}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(userIDContextKey, "test")
		c.Set(backendVersionKey, common.Sync15)
	})
	r.PATCH("/documents/:docid/parent", app.moveDocument)

	req := httptest.NewRequest(http.MethodPatch, "/documents/"+docID+"/parent", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestMoveDocument(t *testing.T) {
	tests := []struct {
		name   string
		docID  string
		body   string
		status int
	}{
		{"into folder", "doc", `{"parent":"d"}`, http.StatusOK},
		{"to root", "doc", `{"parent":""}`, http.StatusOK},
		{"folder to folder", "b", `{"parent":"d"}`, http.StatusOK},
		{"folder to root", "c", `{"parent":""}`, http.StatusOK},
		{"missing target", "doc", `{"parent":"nope"}`, http.StatusBadRequest},
		{"target is a document", "d", `{"parent":"doc"}`, http.StatusBadRequest},
		{"into self", "a", `{"parent":"a"}`, http.StatusBadRequest},
		{"into child", "a", `{"parent":"b"}`, http.StatusBadRequest},
		{"into grandchild", "a", `{"parent":"c"}`, http.StatusBadRequest},
		{"into trash", "doc", `{"parent":"trash"}`, http.StatusBadRequest},
		{"no parent field", "doc", `{}`, http.StatusBadRequest},
		{"bad json", "doc", `nope`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		b := &moveBackend{}
		w := doMove(b, tt.docID, tt.body)
		if w.Code != tt.status {
			t.Errorf("%s: expected %d, got %d", tt.name, tt.status, w.Code)
		}
		ok := tt.status == http.StatusOK
		if b.moved != ok || b.synced != ok {
			t.Errorf("%s: moved=%v synced=%v, expected %v", tt.name, b.moved, b.synced, ok)
		}
	}
}

func TestMoveDocumentArguments(t *testing.T) {
	b := &moveBackend{}
	w := doMove(b, "doc", `{"parent":"d"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if b.movedID != "doc" || b.movedParent != "d" {
		t.Errorf("unexpected move: %q %q", b.movedID, b.movedParent)
	}
}

func TestMoveDocumentBackendError(t *testing.T) {
	b := &moveBackend{err: errors.New("not found")}
	w := doMove(b, "doc", `{"parent":"d"}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	if b.synced {
		t.Error("must not notify on failure")
	}
}
