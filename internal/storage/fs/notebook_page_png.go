package fs

import (
	"encoding/json"
	"fmt"
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
			return nil, fmt.Errorf("page %d out of range (1-%d)", pageNum, len(rms))
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

func exportNotebookPagePNGWithRmdecode(doc *models.HashDoc, ls *LocalBlobStorage, docid string, pageNum int) ([]byte, error) {
	rmData, err := readPageRmBlob(doc, ls, docid, pageNum)
	if err != nil {
		return nil, err
	}
	if len(rmData) == 0 {
		return rmdecode.RenderBlankNotebookPNG()
	}
	b, err := rmdecode.EncodeRmPageToPNG(rmData)
	if err != nil {
		log.Warn("notebook page png: ", err)
		return rmdecode.RenderBlankNotebookPNG()
	}
	return b, nil
}

// ExportPagePNG renders one document page as PNG (1-based). Notebooks use the
// Go .rm stroke renderer; the page index should be the file's last-opened page.
func (fs *FileSystemStorage) ExportPagePNG(uid, docid string, pageNum int) (io.ReadCloser, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	tree, err := fs.GetCachedTree(uid)
	if err != nil {
		return nil, err
	}
	doc, err := tree.FindDoc(docid)
	if err != nil {
		return nil, err
	}
	docHash := doc.Hash
	cacheDir := fs.getPathFromUser(uid, CacheDir)
	_ = os.MkdirAll(cacheDir, 0700)
	safeDoc := common.Sanitize(docid)
	cachePath := path.Join(cacheDir, "renders", safeDoc, fmt.Sprintf("page-%d-rmdecode-%s.png", pageNum, docHash))
	if docHash != "" {
		if r, err := os.Open(cachePath); err == nil {
			return r, nil
		}
	}

	ls := fs.BlobStorage(uid)
	b, err := exportNotebookPagePNGWithRmdecode(doc, ls, docid, pageNum)
	if err != nil {
		return nil, err
	}
	if docHash != "" {
		_ = os.MkdirAll(path.Dir(cachePath), 0700)
		_ = os.WriteFile(cachePath, b, 0600)
	}
	return exporter.NewSeekCloser(b), nil
}
