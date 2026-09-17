package fs

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/rmdecode"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
)

func peekBlob(ls *LocalBlobStorage, hash string, n int) []byte {
	if ls == nil || hash == "" || n < 1 {
		return nil
	}
	rc, err := ls.GetReader(hash)
	if err != nil {
		return nil
	}
	defer rc.Close()
	buf := make([]byte, n)
	got, _ := io.ReadFull(rc, buf)
	if got < 1 {
		return nil
	}
	return buf[:got]
}

func firstFileByExt(doc *models.HashDoc, ext string) *models.HashEntry {
	ext = strings.ToLower(ext)
	if doc == nil {
		return nil
	}
	for _, f := range doc.Files {
		if f == nil {
			continue
		}
		if strings.HasSuffix(strings.ToLower(f.EntryName), ext) {
			return f
		}
	}
	return nil
}

func firstRmFile(doc *models.HashDoc) *models.HashEntry {
	return firstFileByExt(doc, storage.RmFileExt)
}

// ParsePDFVersion reads the %PDF-x.y header from the start of a PDF.
func ParsePDFVersion(b []byte) string {
	if len(b) < 8 {
		return ""
	}
	i := bytes.Index(b, []byte("%PDF-"))
	if i < 0 || i > 16 {
		return ""
	}
	rest := b[i+5:]
	n := 0
	for n < len(rest) {
		c := rest[n]
		if c == '.' || (c >= '0' && c <= '9') {
			n++
			continue
		}
		break
	}
	if n == 0 {
		return ""
	}
	return string(rest[:n])
}

func sniffRmVersion(ls *LocalBlobStorage, doc *models.HashDoc) int {
	ent := firstRmFile(doc)
	if ent == nil {
		return 0
	}
	hdr := peekBlob(ls, ent.Hash, rmdecode.HeaderLen)
	v, err := rmdecode.ParseVersion(hdr)
	if err != nil {
		return 0
	}
	return v
}

func sniffFormatLabel(ls *LocalBlobStorage, doc *models.HashDoc) string {
	if doc == nil {
		return ""
	}
	if doc.CollectionType == common.CollectionType {
		return ""
	}
	kind := doc.EffectivePayloadType()
	rmVer := sniffRmVersion(ls, doc)
	switch kind {
	case "pdf":
		label := "PDF"
		if ent := firstFileByExt(doc, storage.PdfFileExt); ent != nil {
			if ver := ParsePDFVersion(peekBlob(ls, ent.Hash, 64)); ver != "" {
				label = "PDF " + ver
			}
		}
		if rmVer > 0 {
			return fmt.Sprintf("%s · v%d", label, rmVer)
		}
		return label
	case "epub":
		label := "EPUB"
		if rmVer > 0 {
			return fmt.Sprintf("%s · v%d", label, rmVer)
		}
		return label
	default:
		if rmVer > 0 {
			return fmt.Sprintf("RM v%d", rmVer)
		}
		return "RM"
	}
}

func (fs *FileSystemStorage) ensureFormatLabels(tree *models.HashTree, ls *LocalBlobStorage) bool {
	if tree == nil {
		return false
	}
	changed := false
	for _, d := range tree.Docs {
		if d == nil || d.Deleted {
			continue
		}
		if d.CollectionType == common.CollectionType {
			continue
		}
		if d.FormatLabel != "" {
			continue
		}
		if lab := sniffFormatLabel(ls, d); lab != "" {
			d.FormatLabel = lab
			changed = true
		}
	}
	return changed
}

// ensureTemplateMetadata re-reads .metadata for TemplateType entries when Source is
// missing from the cached tree (older caches predate the Source field). Needed so
// com.remarkable.methods + TemplateType classifies as Methods.
func (fs *FileSystemStorage) ensureTemplateMetadata(tree *models.HashTree, ls *LocalBlobStorage) bool {
	if tree == nil || ls == nil {
		return false
	}
	changed := false
	for _, d := range tree.Docs {
		if d == nil || d.Deleted {
			continue
		}
		if d.CollectionType != common.TemplateType {
			continue
		}
		if d.Source != "" {
			continue
		}
		var metaEnt *models.HashEntry
		for _, f := range d.Files {
			if f != nil && f.IsMetadata() {
				metaEnt = f
				break
			}
		}
		if metaEnt == nil {
			continue
		}
		if err := d.ReadMetadata(metaEnt, ls); err != nil {
			continue
		}
		changed = true
	}
	return changed
}
