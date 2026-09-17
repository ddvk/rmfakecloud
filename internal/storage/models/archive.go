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
			err = json.Unmarshal(contentBytes, &a.Content)
			if err != nil {
				return nil, err
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

	pageIDs := a.Content.Pages
	if len(pageIDs) == 0 {
		// Newer tablets store page UUIDs under cPages; rmapi's Content.Pages stays empty.
		// Fall back to whatever .rm blobs we found so notebook PDF export can still run.
		for name := range pageMap {
			pageIDs = append(pageIDs, name)
		}
	}

	for _, p := range pageIDs {
		if hash, ok := pageMap[p]; ok {
			log.Debug("page ", hash)
			reader, err := rs.GetReader(hash)
			if err != nil {
				return nil, err
			}
			pageBin, err := io.ReadAll(reader)
			_ = reader.Close()
			if err != nil {
				return nil, err
			}
			rmpage := rm.New()
			err = rmpage.UnmarshalBinary(pageBin)
			if err != nil {
				// v6 (and other) pages are not understood by rmapi; keep the
				// page slot so PDF export can still emit the background.
				log.Warn("skipping unreadable .rm page ", p, ": ", err)
				a.Pages = append(a.Pages, archive.Page{Pagedata: "Blank"})
				continue
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
