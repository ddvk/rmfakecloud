package fs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	rmtool "github.com/akeil/rmtool"
	rmrender "github.com/akeil/rmtool/pkg/render"

	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
)

type rmtoolMeta struct {
	id      string
	version uint
	name    string
	typ     rmtool.NotebookType
	parent  string
	pinned  bool
	mod     time.Time
}

func (m *rmtoolMeta) ID() string                 { return m.id }
func (m *rmtoolMeta) Version() uint             { return m.version }
func (m *rmtoolMeta) Name() string              { return m.name }
func (m *rmtoolMeta) SetName(n string)          { m.name = n }
func (m *rmtoolMeta) Type() rmtool.NotebookType { return m.typ }
func (m *rmtoolMeta) Pinned() bool              { return m.pinned }
func (m *rmtoolMeta) SetPinned(p bool)          { m.pinned = p }
func (m *rmtoolMeta) LastModified() time.Time   { return m.mod }
func (m *rmtoolMeta) Parent() string            { return m.parent }
func (m *rmtoolMeta) Validate() error           { return nil }

// rmtoolRepo is an in-memory repository for a single document.
// It implements rmtool.Repository so we can use rmtool's PDF renderer.
type rmtoolRepo struct {
	files   map[string][]byte // relative path -> bytes (e.g. "<docid>.content", "<docid>.pdf", "<pageid>.rm")
	pageIDs []string
}

func (r *rmtoolRepo) List() ([]rmtool.Meta, error)          { return nil, fmt.Errorf("not implemented") }
func (r *rmtoolRepo) Update(meta rmtool.Meta) error         { return fmt.Errorf("not implemented") }
func (r *rmtoolRepo) Upload(d *rmtool.Document) error       { return fmt.Errorf("not implemented") }
func (r *rmtoolRepo) Reader(id string, version uint, p ...string) (io.ReadCloser, error) {
	key := path.Join(p...)
	b, ok := r.files[key]
	if !ok {
		return nil, fmt.Errorf("not found: %s", key)
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}
func (r *rmtoolRepo) PagePrefix(pageID string, pageIndex int) string {
	// rmtool Document.Drawing asks for PagePrefix(docID, idx) and then appends ".rm".
	// We map index -> actual page UUID from the .content file.
	if pageIndex >= 0 && pageIndex < len(r.pageIDs) {
		return r.pageIDs[pageIndex]
	}
	return pageID
}

type contentPages struct {
	FileType string   `json:"fileType"`
	Pages    []string `json:"pages"`
}

func readAll(rc io.ReadCloser) ([]byte, error) {
	defer rc.Close()
	return io.ReadAll(rc)
}

type hashReader interface {
	GetReader(hash string) (io.ReadCloser, error)
}

// renderPDFRmtool renders a multi-page PDF on the fly using rmtool's renderer.
// For PDFs it overlays drawings as a separate PDF layer; for notebooks it renders drawings as bitmaps.
func renderPDFRmtool(doc *models.HashDoc, ls hashReader) (io.ReadCloser, error) {
	if doc == nil {
		return nil, fmt.Errorf("doc is nil")
	}

	docid := doc.EntryName
	files := map[string][]byte{}

	// Read and store .content (needed to list page IDs and file type).
	var contentBytes []byte
	for _, f := range doc.Files {
		if strings.EqualFold(path.Ext(f.EntryName), storage.ContentFileExt) {
			rc, err := ls.GetReader(f.Hash)
			if err != nil {
				return nil, err
			}
			b, err := readAll(rc)
			if err != nil {
				return nil, err
			}
			contentBytes = b
			files[docid+storage.ContentFileExt] = b
			break
		}
	}
	if contentBytes == nil {
		return nil, fmt.Errorf("missing content file")
	}

	cp := contentPages{}
	if err := json.Unmarshal(contentBytes, &cp); err != nil {
		return nil, fmt.Errorf("parse content: %w", err)
	}

	// Store all .rm page files. In rmfakecloud storage they are stored as "<pageid>.rm".
	for _, f := range doc.Files {
		if strings.EqualFold(path.Ext(f.EntryName), storage.RmFileExt) {
			rc, err := ls.GetReader(f.Hash)
			if err != nil {
				return nil, err
			}
			b, err := readAll(rc)
			if err != nil {
				return nil, err
			}
			files[path.Base(f.EntryName)] = b
		}
	}

	// Store payload PDF if present.
	for _, f := range doc.Files {
		if strings.EqualFold(path.Ext(f.EntryName), storage.PdfFileExt) {
			rc, err := ls.GetReader(f.Hash)
			if err != nil {
				return nil, err
			}
			b, err := readAll(rc)
			if err != nil {
				return nil, err
			}
			files[docid+storage.PdfFileExt] = b
			break
		}
	}

	repo := &rmtoolRepo{files: files, pageIDs: cp.Pages}
	meta := &rmtoolMeta{
		id:      docid,
		version: uint(doc.MetadataFile.Version),
		name:    doc.MetadataFile.DocumentName,
		typ:     rmtool.DocumentType,
		mod:     time.Now(),
	}
	d, err := rmtool.ReadDocument(repo, meta)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err := rmrender.Pdf(d, &out); err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(out.Bytes())), nil
}

