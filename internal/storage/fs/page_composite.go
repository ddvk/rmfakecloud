package fs

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"os"
	"path"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/rmdecode"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/exporter"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	log "github.com/sirupsen/logrus"
)

func docHasRm(doc *models.HashDoc) bool {
	if doc == nil {
		return false
	}
	for _, f := range doc.Files {
		if f != nil && strings.HasSuffix(strings.ToLower(f.EntryName), storage.RmFileExt) {
			return true
		}
	}
	return false
}

// exportInkLayerPNG renders .rm ink on a white page for multiply compositing.
// Missing/failed pages become blank white (multiply leaves the background unchanged).
func exportInkLayerPNG(doc *models.HashDoc, ls *LocalBlobStorage, docid string, pageNum int) ([]byte, error) {
	return exportInkLayerPNGSized(doc, ls, docid, pageNum, 0, 0)
}

func exportInkLayerPNGSized(doc *models.HashDoc, ls *LocalBlobStorage, docid string, pageNum int, pdfWPt, pdfHPt float64) ([]byte, error) {
	rmData, err := readPageRmBlob(doc, ls, docid, pageNum)
	if err != nil || len(rmData) == 0 {
		return rmdecode.RenderBlankNotebookPNG()
	}
	images := readPageImages(doc, ls, pageIDFromRmBlob(doc, ls, docid, pageNum))
	var b []byte
	if pdfWPt >= 1 && pdfHPt >= 1 {
		b, err = rmdecode.EncodeRmPageToPNGForPDFWithImages(rmData, images, pdfWPt, pdfHPt)
	} else {
		b, err = rmdecode.EncodeRmPageToPNGWithImages(rmData, images)
	}
	if err != nil {
		log.Warn("ink layer png: ", err)
		return rmdecode.RenderBlankNotebookPNG()
	}
	return b, nil
}

func imageBytesToPNG(raw []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// exportPDFPageCompositePNG: PDF page background × .rm ink (multiply).
func exportPDFPageCompositePNG(doc *models.HashDoc, ls *LocalBlobStorage, docid string, pageNum int) ([]byte, error) {
	archive, err := models.ArchiveFromHashDoc(doc, ls)
	if err != nil {
		return nil, err
	}
	if archive.PayloadReader == nil {
		return nil, fmt.Errorf("no payload reader")
	}
	if _, err := archive.PayloadReader.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	pdfBytes, err := io.ReadAll(archive.PayloadReader)
	if err != nil {
		return nil, err
	}
	bg, err := exporter.RenderPDFBytesToPNG(pdfBytes, pageNum)
	if err != nil {
		return nil, err
	}
	if !docHasRm(doc) {
		return bg, nil
	}
	pdfW, pdfH, err := exporter.PDFPageSizePts(pdfBytes, pageNum)
	if err != nil {
		log.Warn("pdf page size: ", err)
		pdfW, pdfH = 0, 0
	}
	ink, err := exportInkLayerPNGSized(doc, ls, docid, pageNum, pdfW, pdfH)
	if err != nil || len(ink) == 0 {
		return bg, nil
	}
	comp, err := exporter.MultiplyBlendPNGs(bg, ink, nil)
	if err != nil {
		log.Warn("pdf composite blend failed, returning background: ", err)
		return bg, nil
	}
	return comp, nil
}

// exportEpubPageCompositePNG: EPUB cover/page image × .rm ink (multiply).
func exportEpubPageCompositePNG(fs *FileSystemStorage, uid string, doc *models.HashDoc, ls *LocalBlobStorage, docid string, pageNum int) ([]byte, error) {
	bgPNG, err := epubBackgroundPNG(fs, uid, docid, pageNum)
	if err != nil || len(bgPNG) == 0 {
		blank, e2 := rmdecode.RenderBlankNotebookPNG()
		if e2 != nil {
			return nil, err
		}
		bgPNG = blank
	}
	if !docHasRm(doc) {
		return bgPNG, nil
	}
	ink, err := exportInkLayerPNG(doc, ls, docid, pageNum)
	if err != nil || len(ink) == 0 {
		return bgPNG, nil
	}
	comp, err := exporter.MultiplyBlendPNGs(bgPNG, ink, nil)
	if err != nil {
		log.Warn("epub composite blend failed, returning background: ", err)
		return bgPNG, nil
	}
	return comp, nil
}

func epubBackgroundPNG(fs *FileSystemStorage, uid, docid string, pageNum int) ([]byte, error) {
	idx0 := pageNum - 1
	if idx0 < 0 {
		idx0 = 0
	}
	rc, _, err := fs.GetEpubPageThumb(uid, docid, idx0)
	if err != nil || rc == nil {
		return nil, fmt.Errorf("epub background: %w", err)
	}
	defer rc.Close()
	raw, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty epub background")
	}
	if png, err := imageBytesToPNG(raw); err == nil {
		return png, nil
	}
	// Already PNG or exotic — try as-is for blend decode.
	if _, err := png.Decode(bytes.NewReader(raw)); err == nil {
		return raw, nil
	}
	return imageBytesToPNG(raw)
}

// ExportPagePNG renders one document page as PNG (1-based).
// PDF/EPUB with .rm: payload/cover background × ink multiply.
// Notebooks: ink (or device thumb / placeholder).
func (fs *FileSystemStorage) ExportPagePNG(uid, docid string, pageNum int) (io.ReadCloser, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	tree, err := fs.GetCachedTree(uid)
	if err != nil {
		return notebookPlaceholderReader()
	}
	doc, err := tree.FindDoc(docid)
	if err != nil {
		return notebookPlaceholderReader()
	}
	docHash := doc.Hash
	ls := fs.BlobStorage(uid)
	ids := pageIDsFromDoc(doc, ls, docid)
	if len(ids) > 0 && pageNum > len(ids) {
		pageNum = len(ids)
	}
	kind := doc.EffectivePayloadType()
	cacheDir := fs.getPathFromUser(uid, CacheDir)
	_ = os.MkdirAll(cacheDir, 0700)
	safeDoc := common.Sanitize(docid)

	var cachePath string
	var gen func() ([]byte, error)
	switch kind {
	case "pdf":
		cachePath = path.Join(cacheDir, "renders", safeDoc, fmt.Sprintf("page-%d-composite-v32-%s.png", pageNum, docHash))
		gen = func() ([]byte, error) {
			return exportPDFPageCompositePNG(doc, ls, docid, pageNum)
		}
	case "epub":
		cachePath = path.Join(cacheDir, "renders", safeDoc, fmt.Sprintf("page-%d-epub-composite-v30-%s.png", pageNum, docHash))
		gen = func() ([]byte, error) {
			return exportEpubPageCompositePNG(fs, uid, doc, ls, docid, pageNum)
		}
	default:
		cachePath = path.Join(cacheDir, "renders", safeDoc, fmt.Sprintf("page-%d-thumb-v6-%s.bin", pageNum, docHash))
		gen = func() ([]byte, error) {
			b, _, e := exportNotebookPagePNGWithRmdecode(doc, ls, docid, pageNum)
			return b, e
		}
	}

	if docHash != "" {
		if r, err := os.Open(cachePath); err == nil {
			return r, nil
		}
	}
	b, err := gen()
	if err != nil || len(b) == 0 {
		log.Warn("ExportPagePNG: ", err)
		return notebookPlaceholderReader()
	}
	if docHash != "" {
		_ = os.MkdirAll(path.Dir(cachePath), 0700)
		_ = os.WriteFile(cachePath, b, 0600)
	}
	return exporter.NewSeekCloser(b), nil
}

// ExportPageThumbPNG returns a device-sized (~384×512) composite thumbnail.
func (fs *FileSystemStorage) ExportPageThumbPNG(uid, docid string, pageNum int) (io.ReadCloser, error) {
	rc, err := fs.ExportPagePNG(uid, docid, pageNum)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	full, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	thumb, err := exporter.ScalePNGToThumb(full)
	if err != nil {
		return exporter.NewSeekCloser(full), nil
	}
	return exporter.NewSeekCloser(thumb), nil
}
