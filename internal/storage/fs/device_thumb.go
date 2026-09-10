package fs

import (
	"path"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/storage/models"
)

func isDeviceThumbExt(name string) bool {
	switch path.Ext(name) {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	}
	return false
}

func matchDeviceThumb(entryName, pageID string) bool {
	pageID = strings.ToLower(strings.TrimSpace(pageID))
	if pageID == "" {
		return false
	}
	n := strings.ToLower(strings.ReplaceAll(entryName, "\\", "/"))
	n = strings.TrimPrefix(n, "./")
	base := path.Base(n)
	ext := path.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	if stem == pageID && isDeviceThumbExt(base) {
		return true
	}
	if stem == pageID+".thumb" && isDeviceThumbExt(base) {
		return true
	}
	if strings.Contains(n, "thumbnail") && stem == pageID {
		return true
	}
	return false
}

func findDeviceThumbEntry(doc *models.HashDoc, pageID string) *models.HashEntry {
	if doc == nil {
		return nil
	}
	for _, f := range doc.Files {
		if f == nil {
			continue
		}
		if matchDeviceThumb(f.EntryName, pageID) {
			return f
		}
	}
	return nil
}
