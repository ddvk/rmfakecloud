package epub

import (
	"archive/zip"
	"errors"
	"io"
	"path"
	"regexp"
	"sort"
	"strings"
)

var (
	imgSrcRE = regexp.MustCompile(`(?i)<img[^>]+src\s*=\s*["']([^"']+)["']`)
	// SVG <image href="..."> (some EPUBs)
	imageHrefRE = regexp.MustCompile(`(?i)<image[^>]+href\s*=\s*["']([^"']+)["']`)
)

// FindCoverImagePath returns a zip-relative path to an image file suitable for a thumbnail.
// It scans for these XHTML/HTML files (case-insensitive basename, any directory):
// cover.xhtml, cover.html, cover.htm, then any *0000.xhtml (e.g. part0000.xhtml) — in that priority order.
// The first <img src="..."> (or <image href="...">) pointing to a raster or SVG inside the zip wins.
func FindCoverImagePath(zr *zip.Reader) (string, error) {
	type cand struct {
		path string
		pri  int
	}
	var cands []cand
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		base := strings.ToLower(path.Base(f.Name))
		var pri int
		switch base {
		case "cover.xhtml":
			pri = 1
		case "cover.html":
			pri = 2
		case "cover.htm":
			pri = 3
		default:
			if strings.HasSuffix(base, "0000.xhtml") {
				pri = 4
			} else {
				continue
			}
		}
		cands = append(cands, cand{f.Name, pri})
	}
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].pri != cands[j].pri {
			return cands[i].pri < cands[j].pri
		}
		return cands[i].path < cands[j].path
	})

	for _, c := range cands {
		rc, err := OpenZipFile(zr, c.path)
		if err != nil {
			continue
		}
		b, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			continue
		}
		var src string
		if m := imgSrcRE.FindSubmatch(b); len(m) >= 2 {
			src = strings.TrimSpace(string(m[1]))
		} else if m := imageHrefRE.FindSubmatch(b); len(m) >= 2 {
			src = strings.TrimSpace(string(m[1]))
		}
		if src == "" {
			continue
		}
		imgPath := resolveImgHref(c.path, src)
		if imgPath == "" {
			continue
		}
		if _, err := OpenZipFile(zr, imgPath); err != nil {
			continue
		}
		ext := strings.ToLower(path.Ext(imgPath))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
			return imgPath, nil
		}
	}
	return "", errors.New("no cover image found")
}

func resolveImgHref(htmlPath, src string) string {
	src = strings.TrimSpace(src)
	if i := strings.IndexByte(src, '?'); i >= 0 {
		src = src[:i]
	}
	if src == "" {
		return ""
	}
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		return ""
	}
	if strings.HasPrefix(src, "mailto:") {
		return ""
	}
	htmlDir := path.Dir(htmlPath)
	if strings.HasPrefix(src, "/") {
		src = strings.TrimPrefix(path.Clean(src), "/")
		return src
	}
	out := path.Join(htmlDir, src)
	out = path.Clean(out)
	if strings.HasPrefix(out, "..") {
		return ""
	}
	return out
}
