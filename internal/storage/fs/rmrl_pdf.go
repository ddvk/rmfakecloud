package fs

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	log "github.com/sirupsen/logrus"
)

const rmrlRunTimeout = 4 * time.Minute

// renderPDFRmrlFile runs the rmrl CLI on a reMarkable document zip (cloud / rmapi-style bundle).
// pythonPath is the interpreter (e.g. /usr/bin/python3); the rmrl package must be installed for that interpreter.
func renderPDFRmrlFile(pythonPath, zipPath string) ([]byte, error) {
	pythonPath = strings.TrimSpace(pythonPath)
	if pythonPath == "" {
		return nil, fmt.Errorf("empty python path")
	}
	ctx, cancel := context.WithTimeout(context.Background(), rmrlRunTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, pythonPath, "-m", "rmrl", zipPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return nil, fmt.Errorf("%w: %s", err, msg)
		}
		return nil, err
	}
	if len(out) < 8 || !bytes.HasPrefix(out, []byte("%PDF")) {
		return nil, fmt.Errorf("rmrl did not produce PDF output")
	}
	return out, nil
}

// buildRmrlDocumentZip builds an in-memory zip in the layout rmrl's ZipSource expects:
//   {docId}.content, optional {docId}.pdf, {docId}.pagedata, and {docId}/{pageId}.rm
func buildRmrlDocumentZip(doc *models.HashDoc, ls hashReader) ([]byte, error) {
	if doc == nil {
		return nil, fmt.Errorf("doc is nil")
	}
	docid := doc.EntryName
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
			break
		}
	}
	if contentBytes == nil {
		return nil, fmt.Errorf("missing content file")
	}

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	w, err := zw.Create(docid + storage.ContentFileExt)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(contentBytes); err != nil {
		_ = zw.Close()
		return nil, err
	}

	for _, f := range doc.Files {
		ext := strings.ToLower(path.Ext(f.EntryName))
		switch ext {
		case storage.PageFileExt:
			rc, err := ls.GetReader(f.Hash)
			if err != nil {
				_ = zw.Close()
				return nil, err
			}
			b, err := readAll(rc)
			if err != nil {
				_ = zw.Close()
				return nil, err
			}
			name := docid + storage.PageFileExt
			fw, err := zw.Create(name)
			if err != nil {
				_ = zw.Close()
				return nil, err
			}
			if _, err := fw.Write(b); err != nil {
				_ = zw.Close()
				return nil, err
			}
		case storage.PdfFileExt:
			rc, err := ls.GetReader(f.Hash)
			if err != nil {
				_ = zw.Close()
				return nil, err
			}
			b, err := readAll(rc)
			if err != nil {
				_ = zw.Close()
				return nil, err
			}
			fw, err := zw.Create(docid + storage.PdfFileExt)
			if err != nil {
				_ = zw.Close()
				return nil, err
			}
			if _, err := fw.Write(b); err != nil {
				_ = zw.Close()
				return nil, err
			}
		case storage.RmFileExt:
			base := path.Base(f.EntryName)
			rc, err := ls.GetReader(f.Hash)
			if err != nil {
				_ = zw.Close()
				return nil, err
			}
			b, err := readAll(rc)
			if err != nil {
				_ = zw.Close()
				return nil, err
			}
			inner := docid + "/" + base
			fw, err := zw.Create(inner)
			if err != nil {
				_ = zw.Close()
				return nil, err
			}
			if _, err := fw.Write(b); err != nil {
				_ = zw.Close()
				return nil, err
			}
		}
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// renderPDFRmrlFromHashDoc writes a temporary zip and runs rmrl (for sync 1.5 blob documents).
func renderPDFRmrlFromHashDoc(pythonPath string, doc *models.HashDoc, ls hashReader) (io.ReadCloser, error) {
	zb, err := buildRmrlDocumentZip(doc, ls)
	if err != nil {
		return nil, err
	}
	tmp, err := os.CreateTemp("", "rmfakecloud-rmrl-*.zip")
	if err != nil {
		return nil, err
	}
	tmpPath := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpPath) }

	if _, err := tmp.Write(zb); err != nil {
		_ = tmp.Close()
		cleanup()
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return nil, err
	}

	out, err := renderPDFRmrlFile(pythonPath, tmpPath)
	cleanup()
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(out)), nil
}

// rmZipLooksRmrlCompatible returns true if the zip has a top-level *.content and
// at least one {docId}/*.rm path (rmrl's expected layout; see rmrl/sources.py).
func rmZipLooksRmrlCompatible(zr *zip.Reader) bool {
	var contentName string
	rmCount := 0
	for _, f := range zr.File {
		n := f.Name
		if strings.Contains(n, "..") {
			continue
		}
		if strings.HasSuffix(strings.ToLower(n), ".content") && !strings.Contains(n, "/") {
			contentName = path.Base(n)
		}
		low := strings.ToLower(n)
		if strings.HasSuffix(low, ".rm") {
			rmCount++
		}
	}
	if contentName == "" || rmCount == 0 {
		return false
	}
	docStem := strings.TrimSuffix(contentName, ".content")
	// Flat page files: PAGE.rm next to content (rmrl tries {ID}/pid.rm first, then {ID}/pagenum.rm)
	hasUnderDoc := false
	for _, f := range zr.File {
		low := strings.ToLower(f.Name)
		if !strings.HasSuffix(low, ".rm") {
			continue
		}
		dir := path.Dir(f.Name)
		baseDir := path.Base(dir)
		if baseDir == docStem {
			hasUnderDoc = true
			break
		}
	}
	// rmrl resolves strokes from {docId}/{pageId}.rm (see rmrl/document.py).
	return hasUnderDoc
}

// tryExportPDFViaRmrl runs rmrl on an on-disk zip if the bundle looks compatible.
func tryExportPDFViaRmrl(pythonPath, zipPath string) (io.ReadCloser, bool) {
	pythonPath = strings.TrimSpace(pythonPath)
	if pythonPath == "" {
		return nil, false
	}
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		log.Debug("rmrl: open zip: ", err)
		return nil, false
	}
	defer r.Close()
	if !rmZipLooksRmrlCompatible(&r.Reader) {
		return nil, false
	}
	out, err := renderPDFRmrlFile(pythonPath, zipPath)
	if err != nil {
		log.Warn("rmrl export failed (sync 1.0 zip): ", err)
		return nil, false
	}
	return io.NopCloser(bytes.NewReader(out)), true
}

