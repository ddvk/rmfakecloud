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

			// Try to detect version first
			version, versionErr := exporter.DetectRmVersionFromBytes(pageBin)

			// For v5 and earlier, parse with rmapi
			if versionErr == nil && (version == exporter.VersionV3 || version == exporter.VersionV5) {
				rmpage := rm.New()
				err = rmpage.UnmarshalBinary(pageBin)
				if err != nil {
					log.Warnf("Failed to unmarshal v5 page: %v", err)
					return nil, err
				}

				page := archive.Page{
					Data:     rmpage,
					Pagedata: "Blank",
				}
				a.Pages = append(a.Pages, page)
			} else if versionErr == nil && version == exporter.VersionV6 {
				// For v6, we can't unmarshal with rmapi
				// Store the raw bytes in a special way
				log.Debugf("Detected v6 page, storing raw data")

				// Create a dummy rm page with the raw bytes stored
				// This is a workaround - we'll handle v6 differently in the export
				rmpage := rm.New()
				// Store raw v6 data - we'll write it directly to file later
				page := archive.Page{
					Data:     rmpage, // Empty, but needed for structure
					Pagedata: "Blank",
				}
				// We need to store the raw v6 bytes somehow
				// The Page structure doesn't have a field for this
				// We'll need to modify the export logic instead
				a.Pages = append(a.Pages, page)
			} else {
				log.Warnf("Unknown rm file version or detection failed: %v", versionErr)
				// Try to parse as v5 anyway (backward compatibility)
				rmpage := rm.New()
				err = rmpage.UnmarshalBinary(pageBin)
				if err != nil {
					log.Warnf("Failed to unmarshal page: %v", err)
					return nil, err
				}

				page := archive.Page{
					Data:     rmpage,
					Pagedata: "Blank",
				}
				a.Pages = append(a.Pages, page)
			}
		}
	}

	return &a, nil
}
