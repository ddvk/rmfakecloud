package epub

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"io"
	"path"
	"regexp"
	"sort"
	"strings"
)

var (
	imgSrcRE     = regexp.MustCompile(`(?i)<img[^>]+src\s*=\s*["']([^"']+)["']`)
	imageHrefRE  = regexp.MustCompile(`(?i)<image[^>]+href\s*=\s*["']([^"']+)["']`)
	coverPropRE  = regexp.MustCompile(`(?is)<item\b[^>]*\bproperties\s*=\s*["'][^"']*cover-image[^"']*["'][^>]*>`)
	itemHrefRE   = regexp.MustCompile(`(?i)\bhref\s*=\s*["']([^"']+)["']`)
	metaCoverRE  = regexp.MustCompile(`(?is)<meta\b[^>]*\bname\s*=\s*["']cover["'][^>]*>`)
	itemIDHrefRE = regexp.MustCompile(`(?is)<item\b[^>]*\bid\s*=\s*["']([^"']+)["'][^>]*>`)
	attrContent  = regexp.MustCompile(`(?i)\bcontent\s*=\s*["']([^"']+)["']`)
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

func isThumbExt(p string) bool {
	switch strings.ToLower(path.Ext(p)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return true
	}
	return false
}

// FindThumbImagePath finds any image suitable for a library thumbnail.
// Order: XHTML cover pages, cover.* files, OPF cover-image, largest raster in the zip.
func FindThumbImagePath(zr *zip.Reader) (string, error) {
	if p, err := FindCoverImagePath(zr); err == nil {
		return p, nil
	}
	if p := findNamedCoverFile(zr); p != "" {
		return p, nil
	}
	if p := findOPFCoverImage(zr); p != "" {
		return p, nil
	}
	if p := findLargestRaster(zr); p != "" {
		return p, nil
	}
	return "", errors.New("no cover image found")
}

func findNamedCoverFile(zr *zip.Reader) string {
	var found []string
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		base := strings.ToLower(path.Base(f.Name))
		switch base {
		case "cover.jpg", "cover.jpeg", "cover.png", "cover.gif", "cover.webp", "cover.svg",
			"cover-image.jpg", "cover-image.jpeg", "cover-image.png":
			found = append(found, zipEntryName(f.Name))
		}
	}
	if len(found) == 0 {
		return ""
	}
	sort.Strings(found)
	return found[0]
}

func findOPFCoverImage(zr *zip.Reader) string {
	opfPath, opfBytes, err := readOPF(zr)
	if err != nil {
		return ""
	}
	opfDir := path.Dir(opfPath)
	if opfDir == "." {
		opfDir = ""
	}
	if m := coverPropRE.Find(opfBytes); m != nil {
		if hm := itemHrefRE.FindSubmatch(m); len(hm) >= 2 {
			p := path.Clean(path.Join(opfDir, string(hm[1])))
			if isThumbExt(p) && zipHas(zr, p) {
				return p
			}
		}
	}
	if m := metaCoverRE.Find(opfBytes); m != nil {
		if cm := attrContent.FindSubmatch(m); len(cm) >= 2 {
			id := string(cm[1])
			for _, item := range itemIDHrefRE.FindAllSubmatch(opfBytes, -1) {
				if len(item) < 2 || string(item[1]) != id {
					continue
				}
				if hm := itemHrefRE.FindSubmatch(item[0]); len(hm) >= 2 {
					p := path.Clean(path.Join(opfDir, string(hm[1])))
					if isThumbExt(p) && zipHas(zr, p) {
						return p
					}
				}
			}
		}
	}
	return ""
}

func findLargestRaster(zr *zip.Reader) string {
	var best string
	var bestSize int64
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := zipEntryName(f.Name)
		if !isThumbExt(name) {
			continue
		}
		if f.UncompressedSize64 < 32 {
			continue
		}
		if int64(f.UncompressedSize64) > bestSize {
			bestSize = int64(f.UncompressedSize64)
			best = name
		}
	}
	return best
}

func zipHas(zr *zip.Reader, name string) bool {
	_, err := OpenZipFile(zr, name)
	return err == nil
}

func readOPF(zr *zip.Reader) (string, []byte, error) {
	containerFile, err := openZipFile(zr, "META-INF/container.xml")
	if err != nil {
		return "", nil, err
	}
	b, err := io.ReadAll(containerFile)
	_ = containerFile.Close()
	if err != nil {
		return "", nil, err
	}
	var c containerRoot
	if err := xml.Unmarshal(b, &c); err != nil {
		return "", nil, err
	}
	if len(c.RootFiles.Rootfile) == 0 {
		return "", nil, ErrNotEpub
	}
	rootPath := strings.TrimPrefix(path.Clean("/"+c.RootFiles.Rootfile[0].FullPath), "/")
	opfFile, err := openZipFile(zr, rootPath)
	if err != nil {
		return "", nil, err
	}
	opfBytes, err := io.ReadAll(opfFile)
	_ = opfFile.Close()
	return rootPath, opfBytes, err
}

// FindFirstImageInFile returns the first raster/SVG image referenced by an HTML/XHTML zip entry.
func FindFirstImageInFile(zr *zip.Reader, htmlPath string) (string, error) {
	rc, err := OpenZipFile(zr, htmlPath)
	if err != nil {
		return "", err
	}
	b, err := io.ReadAll(rc)
	_ = rc.Close()
	if err != nil {
		return "", err
	}
	var src string
	if m := imgSrcRE.FindSubmatch(b); len(m) >= 2 {
		src = strings.TrimSpace(string(m[1]))
	} else if m := imageHrefRE.FindSubmatch(b); len(m) >= 2 {
		src = strings.TrimSpace(string(m[1]))
	}
	if src == "" {
		return "", errors.New("no image in file")
	}
	imgPath := resolveImgHref(htmlPath, src)
	if imgPath == "" {
		return "", errors.New("no image in file")
	}
	if _, err := OpenZipFile(zr, imgPath); err != nil {
		return "", err
	}
	ext := strings.ToLower(path.Ext(imgPath))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return imgPath, nil
	}
	return "", errors.New("no image in file")
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
