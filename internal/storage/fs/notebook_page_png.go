package fs

import (
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/rmdecode"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/exporter"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	log "github.com/sirupsen/logrus"
)

func loadContentFile(doc *models.HashDoc, ls *LocalBlobStorage, docid string) (models.ContentFile, error) {
	var cf models.ContentFile
	want := strings.ToLower(docid + storage.ContentFileExt)
	for _, f := range doc.Files {
		if f == nil {
			continue
		}
		if strings.ToLower(f.EntryName) != want && !strings.HasSuffix(strings.ToLower(f.EntryName), storage.ContentFileExt) {
			continue
		}
		rc, err := ls.GetReader(f.Hash)
		if err != nil {
			return cf, err
		}
		err = json.NewDecoder(rc).Decode(&cf)
		_ = rc.Close()
		return cf, err
	}
	return cf, fmt.Errorf("missing .content for document")
}

func pageIDsFromDoc(doc *models.HashDoc, ls *LocalBlobStorage, docid string) []string {
	cf, err := loadContentFile(doc, ls, docid)
	if err != nil {
		return nil
	}
	return cf.PageIDs()
}

func readPageRmBlob(doc *models.HashDoc, ls *LocalBlobStorage, docid string, pageNum int) (data []byte, err error) {
	ids := pageIDsFromDoc(doc, ls, docid)
	if pageNum < 1 {
		return nil, fmt.Errorf("page %d out of range", pageNum)
	}
	var pageID string
	if pageNum <= len(ids) {
		pageID = ids[pageNum-1]
	} else {
		var rms []string
		for _, f := range doc.Files {
			if f != nil && strings.HasSuffix(strings.ToLower(f.EntryName), storage.RmFileExt) {
				rms = append(rms, f.EntryName)
			}
		}
		if pageNum > len(rms) {
			if len(rms) == 0 {
				return nil, nil
			}
			pageNum = len(rms)
		}
		pageID = strings.TrimSuffix(path.Base(rms[pageNum-1]), storage.RmFileExt)
	}
	if pageID == "" {
		return nil, nil
	}
	want := strings.ToLower(pageID + storage.RmFileExt)
	for _, f := range doc.Files {
		if f == nil {
			continue
		}
		name := strings.ToLower(f.EntryName)
		if name != want && !strings.HasSuffix(name, "/"+want) {
			continue
		}
		rc, e := ls.GetReader(f.Hash)
		if e != nil {
			return nil, e
		}
		data, e = io.ReadAll(rc)
		_ = rc.Close()
		return data, e
	}
	return nil, nil
}

func readDeviceThumbBytes(doc *models.HashDoc, ls *LocalBlobStorage, pageID string) []byte {
	ent := findDeviceThumbEntry(doc, pageID)
	if ent == nil {
		return nil
	}
	rc, err := ls.GetReader(ent.Hash)
	if err != nil {
		return nil
	}
	b, err := io.ReadAll(rc)
	_ = rc.Close()
	if err != nil || len(b) < 32 {
		return nil
	}
	return b
}

func pageIDForNum(doc *models.HashDoc, ls *LocalBlobStorage, docid string, pageNum int) string {
	ids := pageIDsFromDoc(doc, ls, docid)
	if pageNum >= 1 && pageNum <= len(ids) {
		return ids[pageNum-1]
	}
	return ""
}

// pageIDFromRmBlob resolves the page UUID used for sibling image paths.
func pageIDFromRmBlob(doc *models.HashDoc, ls *LocalBlobStorage, docid string, pageNum int) string {
	if id := pageIDForNum(doc, ls, docid, pageNum); id != "" {
		return id
	}
	var rms []string
	for _, f := range doc.Files {
		if f != nil && strings.HasSuffix(strings.ToLower(f.EntryName), storage.RmFileExt) {
			rms = append(rms, f.EntryName)
		}
	}
	if pageNum < 1 || pageNum > len(rms) {
		return ""
	}
	return strings.TrimSuffix(path.Base(rms[pageNum-1]), storage.RmFileExt)
}

func isPageImageExt(name string) bool {
	switch strings.ToLower(path.Ext(name)) {
	case ".png", ".jpg", ".jpeg", ".webp", ".gif":
		return true
	}
	return false
}

// readPageImages loads inserted image blobs stored beside a page as
// {docid}/{pageID}/{filename}.png (basename → bytes for rmc staging).
func readPageImages(doc *models.HashDoc, ls *LocalBlobStorage, pageID string) map[string][]byte {
	pageID = strings.ToLower(strings.TrimSpace(pageID))
	if doc == nil || ls == nil || pageID == "" {
		return nil
	}
	prefix := pageID + "/"
	out := make(map[string][]byte)
	for _, f := range doc.Files {
		if f == nil {
			continue
		}
		name := strings.ToLower(strings.ReplaceAll(f.EntryName, "\\", "/"))
		idx := strings.Index(name, prefix)
		if idx < 0 {
			continue
		}
		// Only accept .../{pageID}/{file} — not deeper nests or the page .rm.
		rel := name[idx+len(prefix):]
		if rel == "" || strings.Contains(rel, "/") || !isPageImageExt(rel) {
			continue
		}
		rc, err := ls.GetReader(f.Hash)
		if err != nil {
			continue
		}
		b, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil || len(b) == 0 {
			continue
		}
		out[path.Base(rel)] = b
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func exportNotebookPagePNGWithRmdecode(doc *models.HashDoc, ls *LocalBlobStorage, docid string, pageNum int) (png []byte, cacheable bool, err error) {
	if pageID := pageIDForNum(doc, ls, docid, pageNum); pageID != "" {
		if b := readDeviceThumbBytes(doc, ls, pageID); len(b) > 0 {
			return b, true, nil
		}
	}
	rmData, err := readPageRmBlob(doc, ls, docid, pageNum)
	if err != nil {
		return nil, false, err
	}
	if len(rmData) == 0 {
		b, e := rmdecode.RenderNotebookPlaceholderPNG()
		return b, false, e
	}
	pageID := pageIDFromRmBlob(doc, ls, docid, pageNum)
	images := readPageImages(doc, ls, pageID)
	b, err := rmdecode.EncodeRmPageToPNGWithImages(rmData, images)
	if err != nil {
		log.Warn("notebook page png: ", err)
		ph, e := rmdecode.RenderNotebookPlaceholderPNG()
		if e != nil {
			return nil, false, err
		}
		return ph, false, nil
	}
	return b, true, nil
}

func notebookPlaceholderReader() (io.ReadCloser, error) {
	b, err := rmdecode.RenderNotebookPlaceholderPNG()
	if err != nil {
		return nil, err
	}
	return exporter.NewSeekCloser(b), nil
}
