package fs

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/epub"
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
	for _, f := range doc.Files {
		if f == nil {
			continue
		}
		if strings.HasSuffix(strings.ToLower(f.EntryName), storage.EpubFileExt) {
			return ls.GetReader(f.Hash)
		}
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
		return nil, "", err
	}
	imgPath, err := epub.FindCoverImagePath(zr)
	if err != nil {
		return nil, "", err
	}
	return fs.GetEpubFile(uid, docid, imgPath)
}

// GetEpubPageThumb returns an image from the last-opened spine item, else the cover.
func (fs *FileSystemStorage) GetEpubPageThumb(uid, docid string, pageIndex0 int) (io.ReadCloser, string, error) {
	zr, _, err := fs.openEpubZip(uid, docid)
	if err != nil {
		return nil, "", err
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
			return fs.GetEpubFile(uid, docid, imgPath)
		}
	}
	imgPath, err := epub.FindCoverImagePath(zr)
	if err != nil {
		return nil, "", err
	}
	return fs.GetEpubFile(uid, docid, imgPath)
}
