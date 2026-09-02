package models

import (
	"encoding/json"
	"io"
	"path"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/exporter"
	"github.com/juruen/rmapi/archive"
	"github.com/juruen/rmapi/encoding/rm"
	log "github.com/sirupsen/logrus"
)

// cPagesPage mirrors one entry of the "cPages.pages" array found in
// .content files written by reMarkable firmware 3.x (Paper Pro/Pure).
type cPagesPage struct {
	ID string `json:"id"`
}

type cPages struct {
	Pages []cPagesPage `json:"pages"`
}

// rawContent embeds the rmapi Content and adds the newer cPages field
type rawContent struct {
	archive.Content
	CPages cPages `json:"cPages"`
}

// ArchiveFromHashDoc reads an archive
func ArchiveFromHashDoc(doc *HashDoc, rs RemoteStorage) (*exporter.MyArchive, error) {
	uuid := doc.EntryName
	a := exporter.MyArchive{
		Zip: archive.Zip{
			UUID: uuid,
		},
	}

	pageMap := make(map[string]string)
	for _, f := range doc.Files {
		filext := path.Ext(f.EntryName)
		name := strings.TrimSuffix(path.Base(f.EntryName), filext)
		switch filext {
		case storage.ContentFileExt:
			blob, err := rs.GetReader(f.Hash)
			if err != nil {
				return nil, err
			}
			defer blob.Close()
			contentBytes, err := io.ReadAll(blob)
			if err != nil {
				return nil, err
			}
			var rc rawContent
			err = json.Unmarshal(contentBytes, &rc)
			if err != nil {
				return nil, err
			}
			a.Content = rc.Content
			// firmware 3.x stores the page list in cPages instead of
			// the flat pages field; fall back to it when absent
			if len(a.Content.Pages) == 0 && len(rc.CPages.Pages) > 0 {
				a.Content.Pages = make([]string, 0, len(rc.CPages.Pages))
				for _, p := range rc.CPages.Pages {
					a.Content.Pages = append(a.Content.Pages, p.ID)
				}
			}
		case storage.EpubFileExt:
			fallthrough
		case storage.PdfFileExt:
			blob, err := rs.GetReader(f.Hash)
			if err != nil {
				return nil, err
			}
			// defer blob.Close()
			// contentBytes, err := ioutil.ReadAll(blob)
			// if err != nil {
			// 	return nil, err
			// }
			// a.Payload = contentBytes
			//HACK:
			a.PayloadReader = blob.(io.ReadSeekCloser)

		case ".json":
			//metadata
		case storage.RmFileExt:
			log.Debug("adding page ", name)
			pageMap[name] = f.Hash
		}
	}

	for _, p := range a.Content.Pages {
		if hash, ok := pageMap[p]; ok {
			log.Debug("page ", hash)
			reader, err := rs.GetReader(hash)
			if err != nil {
				return nil, err
			}
			pageBin, err := io.ReadAll(reader)
			if err != nil {
				return nil, err
			}
			rmpage := rm.New()
			err = rmpage.UnmarshalBinary(pageBin)
			if err != nil {
				// firmware 3.x (Paper Pro/Pure) writes .rm v6 pages that
				// the bundled parser cannot read; keep the raw bytes so a
				// v6-capable renderer can process them later instead of
				// aborting the whole export
				if isV6Page(pageBin) {
					log.Warnln("keeping unsupported v6 rm page", p)
					a.V6Pages = append(a.V6Pages, exporter.V6Page{
						PageID: p,
						Data:   pageBin,
					})
					continue
				}
				return nil, err
			}

			page := archive.Page{
				Data:     rmpage,
				Pagedata: "Blank",
			}
			a.Pages = append(a.Pages, page)
		}
	}

	return &a, nil
}

// v6Header is the magic string .rm v6 files start with
const v6Header = "reMarkable .lines file, version=6"

func isV6Page(data []byte) bool {
	return len(data) >= len(v6Header) && string(data[:len(v6Header)]) == v6Header
}
