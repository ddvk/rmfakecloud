package epub

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"io"
	"path"
	"strings"
)

// ErrNotEpub is returned when the archive does not appear to be a valid EPUB.
var ErrNotEpub = errors.New("not a valid EPUB")

// Manifest holds the reading order (spine) and base path for resolving relative URLs.
type Manifest struct {
	Spine    []string `json:"spine"`    // paths relative to zip root (e.g. OEBPS/ch1.xhtml)
	BasePath string   `json:"basePath"` // directory of the OPF, for relative resolution
}

type containerRoot struct {
	XMLName   xml.Name `xml:"container"`
	RootFiles struct {
		Rootfile []struct {
			FullPath  string `xml:"full-path,attr"`
			MediaType string `xml:"media-type,attr"`
		} `xml:"rootfile"`
	} `xml:"rootfiles"`
}

type opfPackage struct {
	XMLName  xml.Name `xml:"http://www.idpf.org/2007/opf package"`
	Manifest struct {
		Items []struct {
			ID        string `xml:"id,attr"`
			Href      string `xml:"href,attr"`
			MediaType string `xml:"media-type,attr"`
		} `xml:"item"`
	} `xml:"manifest"`
	Spine struct {
		Itemrefs []struct {
			IDRef string `xml:"idref,attr"`
		} `xml:"itemref"`
	} `xml:"spine"`
}

// ReadManifest parses the EPUB zip and returns the spine (ordered content paths) and OPF base path.
func ReadManifest(zr *zip.Reader) (*Manifest, error) {
	// 1. Find and parse META-INF/container.xml
	containerFile, err := openZipFile(zr, "META-INF/container.xml")
	if err != nil {
		return nil, err
	}
	defer containerFile.Close()
	containerBytes, err := io.ReadAll(containerFile)
	if err != nil {
		return nil, err
	}
	var c containerRoot
	if err := xml.Unmarshal(containerBytes, &c); err != nil {
		return nil, err
	}
	if len(c.RootFiles.Rootfile) == 0 {
		return nil, ErrNotEpub
	}
	rootPath := c.RootFiles.Rootfile[0].FullPath
	rootPath = strings.TrimPrefix(path.Clean("/"+rootPath), "/")

	// 2. Parse OPF
	opfFile, err := openZipFile(zr, rootPath)
	if err != nil {
		return nil, err
	}
	defer opfFile.Close()
	opfBytes, err := io.ReadAll(opfFile)
	if err != nil {
		return nil, err
	}
	var pkg opfPackage
	if err := xml.Unmarshal(opfBytes, &pkg); err != nil {
		return nil, err
	}
	opfDir := path.Dir(rootPath)
	if opfDir == "." {
		opfDir = ""
	}
	idToHref := make(map[string]string)
	for _, it := range pkg.Manifest.Items {
		idToHref[it.ID] = it.Href
	}
	var spine []string
	for _, ref := range pkg.Spine.Itemrefs {
		href, ok := idToHref[ref.IDRef]
		if !ok {
			continue
		}
		// Resolve href relative to OPF directory
		fullPath := path.Join(opfDir, href)
		fullPath = path.Clean(fullPath)
		if strings.HasPrefix(fullPath, "..") {
			continue
		}
		spine = append(spine, fullPath)
	}
	if len(spine) == 0 {
		return nil, ErrNotEpub
	}
	return &Manifest{
		Spine:    spine,
		BasePath: opfDir,
	}, nil
}

func openZipFile(zr *zip.Reader, name string) (io.ReadCloser, error) {
	name = path.Clean(name)
	if strings.HasPrefix(name, "..") {
		return nil, errors.New("invalid path")
	}
	for _, f := range zr.File {
		clean := path.Clean(f.Name)
		if clean == name || f.Name == name {
			return f.Open()
		}
	}
	return nil, errors.New("file not found")
}

// OpenZipFile returns a reader for a file inside the zip by path (relative to zip root).
// Path must not contain "..".
func OpenZipFile(zr *zip.Reader, filePath string) (io.ReadCloser, error) {
	filePath = path.Clean(filePath)
	if filePath == "." || filePath == ".." || strings.HasPrefix(filePath, "..") {
		return nil, errors.New("invalid path")
	}
	for _, f := range zr.File {
		clean := path.Clean(f.Name)
		if clean == filePath {
			return f.Open()
		}
	}
	return nil, errors.New("file not found")
}

// ContentType returns a suitable Content-Type for an EPUB resource by extension.
func ContentType(filePath string) string {
	ext := strings.ToLower(path.Ext(filePath))
	switch ext {
	case ".xhtml", ".html", ".htm":
		return "application/xhtml+xml"
	case ".css":
		return "text/css"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".otf":
		return "font/otf"
	default:
		return "application/octet-stream"
	}
}
