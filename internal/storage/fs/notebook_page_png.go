package fs

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/rmdecode"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
)

type contentPagesOnly struct {
	Pages []string `json:"pages"`
}

// readPageRmBlob loads the raw .rm bytes for a 1-based page index. If the page exists in
// .content but has no .rm file yet, returns (nil, "", nil). Other errors: missing content, bad index.
func readPageRmBlob(doc *models.HashDoc, ls *LocalBlobStorage, docid string, pageNum int) (data []byte, rmHash string, err error) {
	var cp contentPagesOnly
	found := false
	for _, f := range doc.Files {
		if strings.EqualFold(f.EntryName, docid+storage.ContentFileExt) {
			found = true
			rc, e := ls.GetReader(f.Hash)
			if e != nil {
				return nil, "", e
			}
			e = json.NewDecoder(rc).Decode(&cp)
			_ = rc.Close()
			if e != nil {
				return nil, "", e
			}
			break
		}
	}
	if !found {
		return nil, "", fmt.Errorf("missing .content for document")
	}
	if pageNum < 1 || pageNum > len(cp.Pages) {
		return nil, "", fmt.Errorf("page %d out of range (1-%d)", pageNum, len(cp.Pages))
	}
	pageID := cp.Pages[pageNum-1]
	want := pageID + storage.RmFileExt
	for _, f := range doc.Files {
		if strings.EqualFold(f.EntryName, want) {
			rmHash = f.Hash
			rc, e := ls.GetReader(f.Hash)
			if e != nil {
				return nil, "", e
			}
			data, e = io.ReadAll(rc)
			_ = rc.Close()
			if e != nil {
				return nil, "", e
			}
			return data, rmHash, nil
		}
	}
	return nil, "", nil
}

// exportNotebookPagePNGWithRmdecode renders strokes to PNG via rmdecode (v3/v5 Go; v6 optional python).
func exportNotebookPagePNGWithRmdecode(
	doc *models.HashDoc,
	ls *LocalBlobStorage,
	docid string,
	pageNum int,
) ([]byte, error) {
	rmData, _, err := readPageRmBlob(doc, ls, docid, pageNum)
	if err != nil {
		return nil, err
	}
	if len(rmData) == 0 {
		b, err := rmdecode.RenderBlankNotebookPNG()
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	b, err := rmdecode.EncodeRmPageToPNG(rmData)
	if err != nil {
		return nil, err
	}
	return b, nil
}
