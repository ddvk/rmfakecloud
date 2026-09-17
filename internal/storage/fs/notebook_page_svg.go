package fs

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/rmdecode"
	"github.com/ddvk/rmfakecloud/internal/storage/exporter"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	log "github.com/sirupsen/logrus"
)

func notebookPlaceholderSVGReader() io.ReadCloser {
	return exporter.NewSeekCloser([]byte(rmdecode.RenderNotebookPlaceholderSVG()))
}

func exportNotebookPageSVG(doc *models.HashDoc, ls *LocalBlobStorage, docid string, pageNum int) (svg []byte, cacheable bool, err error) {
	rmData, err := readPageRmBlob(doc, ls, docid, pageNum)
	if err != nil {
		return nil, false, err
	}
	if len(rmData) == 0 {
		return []byte(rmdecode.RenderNotebookPlaceholderSVG()), false, nil
	}
	images := readPageImages(doc, ls, pageIDFromRmBlob(doc, ls, docid, pageNum))
	b, err := rmdecode.EncodeRmPageToSVGWithImages(rmData, images)
	if err != nil {
		log.Warn("notebook page svg: ", err)
		return []byte(rmdecode.RenderNotebookPlaceholderSVG()), false, nil
	}
	return b, true, nil
}

func notebookPageCount(doc *models.HashDoc, ls *LocalBlobStorage, docid string) int {
	ids := pageIDsFromDoc(doc, ls, docid)
	if n := len(ids); n > 0 {
		return n
	}
	if doc != nil && doc.PageCount > 0 {
		return doc.PageCount
	}
	return 1
}

// NotebookPageCount is the number of pages in a notebook (at least 1).
func (fs *FileSystemStorage) NotebookPageCount(uid, docid string) int {
	tree, err := fs.GetCachedTree(uid)
	if err != nil {
		return 1
	}
	doc, err := tree.FindDoc(docid)
	if err != nil {
		return 1
	}
	return notebookPageCount(doc, fs.BlobStorage(uid), docid)
}

// ExportPageSVG renders one notebook page as SVG (1-based).
func (fs *FileSystemStorage) ExportPageSVG(uid, docid string, pageNum int) (io.ReadCloser, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	tree, err := fs.GetCachedTree(uid)
	if err != nil {
		return notebookPlaceholderSVGReader(), nil
	}
	doc, err := tree.FindDoc(docid)
	if err != nil {
		return notebookPlaceholderSVGReader(), nil
	}
	ls := fs.BlobStorage(uid)
	pages := notebookPageCount(doc, ls, docid)
	if pageNum > pages {
		pageNum = pages
	}
	docHash := doc.Hash
	cacheDir := fs.getPathFromUser(uid, CacheDir)
	_ = os.MkdirAll(cacheDir, 0700)
	safeDoc := common.Sanitize(docid)
	cachePath := path.Join(cacheDir, "renders", safeDoc, fmt.Sprintf("page-%d-svg-v2-%s.svg", pageNum, docHash))
	if docHash != "" {
		if r, err := os.Open(cachePath); err == nil {
			return r, nil
		}
	}
	b, cacheable, err := exportNotebookPageSVG(doc, ls, docid, pageNum)
	if err != nil || len(b) == 0 {
		log.Warn("notebook page svg: ", err)
		return notebookPlaceholderSVGReader(), nil
	}
	if cacheable && docHash != "" {
		_ = os.MkdirAll(path.Dir(cachePath), 0700)
		_ = os.WriteFile(cachePath, b, 0600)
	}
	return exporter.NewSeekCloser(b), nil
}
