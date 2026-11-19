package ui

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/epub"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	containerPath = "META-INF/container.xml"
)

type containerRootfile struct {
	FullPath string `xml:"full-path,attr"`
}

type containerRootfiles struct {
	Rootfile containerRootfile `xml:"rootfile"`
}

type containerXML struct {
	Rootfiles containerRootfiles `xml:"rootfiles"`
}

type opfItem struct {
	ID   string `xml:"id,attr"`
	Href string `xml:"href,attr"`
}

type opfManifest struct {
	Items []opfItem `xml:"item"`
}

type opfItemref struct {
	IDRef string `xml:"idref,attr"`
}

type opfSpine struct {
	ItemRefs []opfItemref `xml:"itemref"`
}

type opfPackage struct {
	Manifest opfManifest `xml:"manifest"`
	Spine    opfSpine    `xml:"spine"`
}

var epubContentTypes = map[string]string{
	".html": "text/html", ".xhtml": "text/html", ".htm": "text/html",
	".css": "text/css",
	".svg": "image/svg+xml",
	".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png", ".gif": "image/gif", ".webp": "image/webp",
	".woff": "font/woff", ".woff2": "font/woff2", ".ttf": "font/ttf", ".otf": "font/otf",
	".mp3": "audio/mpeg", ".mp4": "video/mp4",
	".ncx": "application/x-dtbncx+xml",
	".opf": "application/oebps-package+xml",
}

func getEpubFirstSpinePath(zr *zip.Reader) (string, error) {
	containerFile, err := openZipPath(zr, containerPath)
	if err != nil {
		return "", err
	}
	defer containerFile.Close()
	containerData, err := io.ReadAll(containerFile)
	if err != nil {
		return "", err
	}
	var c containerXML
	if err := xml.Unmarshal(containerData, &c); err != nil {
		return "", err
	}
	opfPath := strings.TrimSpace(c.Rootfiles.Rootfile.FullPath)
	if opfPath == "" {
		return "", nil
	}
	opfFile, err := openZipPath(zr, opfPath)
	if err != nil {
		return "", err
	}
	defer opfFile.Close()
	opfData, err := io.ReadAll(opfFile)
	if err != nil {
		return "", err
	}
	var pkg opfPackage
	if err := xml.Unmarshal(opfData, &pkg); err != nil {
		return "", err
	}
	manifestByID := make(map[string]string)
	for _, it := range pkg.Manifest.Items {
		manifestByID[it.ID] = it.Href
	}
	if len(pkg.Spine.ItemRefs) == 0 {
		return "", nil
	}
	firstID := pkg.Spine.ItemRefs[0].IDRef
	href, ok := manifestByID[firstID]
	if !ok || href == "" {
		return "", nil
	}
	opfDir := filepath.Dir(opfPath)
	if opfDir == "." {
		return strings.TrimLeft(href, "/"), nil
	}
	// href is relative to OPF directory; normalize to zip path
	joined := filepath.Join(opfDir, href)
	return filepath.ToSlash(joined), nil
}

func openZipPath(zr *zip.Reader, name string) (io.ReadCloser, error) {
	name = strings.TrimPrefix(filepath.ToSlash(path.Clean(name)), "/")
	if strings.HasPrefix(name, "..") {
		return nil, nil
	}
	for _, f := range zr.File {
		entry := strings.TrimPrefix(filepath.ToSlash(filepath.Clean(f.Name)), "/")
		if entry == name {
			return f.Open()
		}
	}
	return nil, nil
}

func getEpubContentType(name string) string {
	ext := strings.ToLower(path.Ext(name))
	if ct, ok := epubContentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

// getDocumentEpub serves the EPUB as a website: unpacked files by path, redirect / to first spine item.
func (app *ReactAppWrapper) getDocumentEpub(c *gin.Context) {
	uid := userID(c)
	docid := common.ParamS(docIDParam, c)
	pathParam := c.Param("path")
	pathParam = strings.TrimPrefix(pathParam, "/")
	pathParam = filepath.ToSlash(path.Clean(pathParam))
	if strings.HasPrefix(pathParam, "..") {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	backend := app.getBackend(c)

	// Path "manifest" returns EPUB spine/manifest as JSON (mymod API).
	if pathParam == "manifest" {
		type epubManifestBackend interface {
			GetEpubManifest(uid, docid string) (*epub.Manifest, error)
		}
		if eb, ok := backend.(epubManifestBackend); ok {
			manifest, err := eb.GetEpubManifest(uid, docid)
			if err != nil {
				log.Error(err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			c.JSON(http.StatusOK, manifest)
			return
		}
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	rc, err := backend.Export(uid, docid, "epub", storage.ExportWithAnnotations)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer rc.Close()

	body, err := io.ReadAll(rc)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if pathParam == "" || pathParam == "." {
		first, err := getEpubFirstSpinePath(zr)
		if err != nil || first == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Redirect(http.StatusFound, "./"+first)
		return
	}

	var found *zip.File
	for _, f := range zr.File {
		entry := filepath.ToSlash(filepath.Clean(f.Name))
		entry = strings.TrimPrefix(entry, "/")
		if entry == pathParam {
			found = f
			break
		}
	}
	if found == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	r, err := found.Open()
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer r.Close()
	contentType := getEpubContentType(found.Name)
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "private, max-age=300")
	_, _ = io.Copy(c.Writer, r)
}
