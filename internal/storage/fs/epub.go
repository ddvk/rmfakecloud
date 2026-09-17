package fs

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/epub"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
)

func (fs *FileSystemStorage) openEpubZip(uid, docid string) (*zip.Reader, []byte, error) {
	rc, err := fs.GetEpub(uid, docid)
	if err != nil {
		return nil, nil, err
	}
	defer rc.Close()
	body, err := io.ReadAll(rc)
	if err != nil {
		return nil, nil, err
	}
	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, nil, err
	}
	return zr, body, nil
}

// GetEpub returns the raw .epub payload for a document.
func (fs *FileSystemStorage) GetEpub(uid, docid string) (io.ReadCloser, error) {
	tree, err := fs.GetCachedTree(uid)
	if err != nil {
		return nil, err
	}
	doc, err := tree.FindDoc(docid)
	if err != nil {
		return nil, err
	}
	ls := fs.BlobStorage(uid)
	var fallback *models.HashEntry
	for _, f := range doc.Files {
		if f == nil {
			continue
		}
		n := strings.ToLower(f.EntryName)
		if strings.HasSuffix(n, storage.EpubFileExt) {
			return ls.GetReader(f.Hash)
		}
		if strings.HasSuffix(n, storage.MetadataFileExt) || strings.HasSuffix(n, storage.ContentFileExt) ||
			strings.HasSuffix(n, storage.PageFileExt) || strings.HasSuffix(n, storage.RmFileExt) ||
			strings.HasSuffix(n, storage.PdfFileExt) || strings.HasSuffix(n, ".json") {
			continue
		}
		if fallback == nil || f.Size > fallback.Size {
			cp := f
			fallback = cp
		}
	}
	if strings.EqualFold(doc.EffectivePayloadType(), "epub") && fallback != nil {
		return ls.GetReader(fallback.Hash)
	}
	return nil, errors.New("epub not found")
}

// GetEpubManifest parses the EPUB spine.
func (fs *FileSystemStorage) GetEpubManifest(uid, docid string) (*epub.Manifest, error) {
	zr, _, err := fs.openEpubZip(uid, docid)
	if err != nil {
		return nil, err
	}
	return epub.ReadManifest(zr)
}

// GetEpubFile returns a file from inside the EPUB zip.
func (fs *FileSystemStorage) GetEpubFile(uid, docid, filePath string) (io.ReadCloser, string, error) {
	zr, _, err := fs.openEpubZip(uid, docid)
	if err != nil {
		return nil, "", err
	}
	f, err := epub.OpenZipFile(zr, filePath)
	if err != nil {
		return nil, "", err
	}
	data, err := io.ReadAll(f)
	_ = f.Close()
	if err != nil {
		return nil, "", err
	}
	return io.NopCloser(bytes.NewReader(data)), epub.ContentType(filePath), nil
}

// GetEpubCoverThumb returns a cover image from the EPUB.
func (fs *FileSystemStorage) GetEpubCoverThumb(uid, docid string) (io.ReadCloser, string, error) {
	zr, _, err := fs.openEpubZip(uid, docid)
	if err != nil {
		rc, ct := epub.PlaceholderThumb()
		return rc, ct, nil
	}
	return openThumbImage(zr)
}

// GetEpubPageThumb returns an image from the last-opened spine item, else any cover, else a placeholder.
func (fs *FileSystemStorage) GetEpubPageThumb(uid, docid string, pageIndex0 int) (io.ReadCloser, string, error) {
	zr, _, err := fs.openEpubZip(uid, docid)
	if err != nil {
		rc, ct := epub.PlaceholderThumb()
		return rc, ct, nil
	}
	man, err := epub.ReadManifest(zr)
	if err == nil && man != nil && len(man.Spine) > 0 {
		idx := pageIndex0
		if idx < 0 {
			idx = 0
		}
		if idx >= len(man.Spine) {
			idx = len(man.Spine) - 1
		}
		if imgPath, e := epub.FindFirstImageInFile(zr, man.Spine[idx]); e == nil {
			if rc, ct, e2 := openZipImage(zr, imgPath); e2 == nil {
				return rc, ct, nil
			}
		}
		for i, sp := range man.Spine {
			if i == idx {
				continue
			}
			if imgPath, e := epub.FindFirstImageInFile(zr, sp); e == nil {
				if rc, ct, e2 := openZipImage(zr, imgPath); e2 == nil {
					return rc, ct, nil
				}
			}
		}
	}
	return openThumbImage(zr)
}

func openThumbImage(zr *zip.Reader) (io.ReadCloser, string, error) {
	imgPath, err := epub.FindThumbImagePath(zr)
	if err != nil {
		rc, ct := epub.PlaceholderThumb()
		return rc, ct, nil
	}
	if rc, ct, err := openZipImage(zr, imgPath); err == nil {
		return rc, ct, nil
	}
	rc, ct := epub.PlaceholderThumb()
	return rc, ct, nil
}

func openZipImage(zr *zip.Reader, imgPath string) (io.ReadCloser, string, error) {
	f, err := epub.OpenZipFile(zr, imgPath)
	if err != nil {
		return nil, "", err
	}
	data, err := io.ReadAll(f)
	_ = f.Close()
	if err != nil {
		return nil, "", err
	}
	return io.NopCloser(bytes.NewReader(data)), epub.ContentType(imgPath), nil
}
