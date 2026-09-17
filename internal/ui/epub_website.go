package ui

import (
	"bytes"
	"html"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/common"
	"github.com/ddvk/rmfakecloud/internal/storage/epub"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func epubAssetURL(docID, zipPath string) string {
	zipPath = strings.TrimPrefix(path.Clean("/"+strings.ReplaceAll(zipPath, "\\", "/")), "/")
	if zipPath == "" || zipPath == "." {
		return ""
	}
	segs := strings.Split(zipPath, "/")
	for i, s := range segs {
		segs[i] = url.PathEscape(s)
	}
	return "/ui/api/documents/" + url.PathEscape(docID) + "/epub/" + strings.Join(segs, "/")
}

func epubSpineLabel(zipPath string, index int) string {
	base := path.Base(zipPath)
	base = strings.TrimSuffix(base, path.Ext(base))
	base = strings.ReplaceAll(base, "_", " ")
	base = strings.TrimSpace(base)
	if base == "" {
		return "Chapter " + strconv.Itoa(index+1)
	}
	return base
}

func (app *ReactAppWrapper) getEpubPath(c *gin.Context) {
	uid := userID(c)
	docid := common.ParamS(docIDParam, c)
	pathParam := c.Param("path")
	pathParam = strings.TrimPrefix(path.Clean("/"+pathParam), "/")
	if pathParam == ".." || strings.HasPrefix(pathParam, "../") {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	type epubBackend interface {
		GetEpubManifest(uid, docid string) (*epub.Manifest, error)
		GetEpubFile(uid, docid, filePath string) (io.ReadCloser, string, error)
		GetEpubCoverThumb(uid, docid string) (io.ReadCloser, string, error)
		GetEpubPageThumb(uid, docid string, pageIndex0 int) (io.ReadCloser, string, error)
	}
	backend := app.getBackend(c)
	eb, ok := backend.(epubBackend)
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	switch pathParam {
	case "cover-thumb":
		reader, contentType, err := eb.GetEpubCoverThumb(uid, docid)
		if err != nil {
			log.Debug("epub cover-thumb: ", err)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		defer reader.Close()
		c.Header("Cache-Control", "private, max-age=3600")
		c.Header("X-Content-Type-Options", "nosniff")
		c.DataFromReader(http.StatusOK, -1, contentType, reader, nil)
		return
	case "thumb":
		idx := 0
		if tree, err := backend.GetDocumentTree(uid); err == nil && tree != nil {
			if d := findDocument(tree.Entries, docid); d != nil {
				idx = d.CurrentPage
			} else if d := findDocument(tree.Trash, docid); d != nil {
				idx = d.CurrentPage
			}
		}
		reader, contentType, err := eb.GetEpubPageThumb(uid, docid, idx)
		if err != nil {
			log.Debug("epub thumb: ", err)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		defer reader.Close()
		c.Header("Cache-Control", "private, max-age=3600")
		c.Header("X-Content-Type-Options", "nosniff")
		c.DataFromReader(http.StatusOK, -1, contentType, reader, nil)
		return
	case "manifest":
		manifest, err := eb.GetEpubManifest(uid, docid)
		if err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, manifest)
		return
	case "", ".":
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	reader, contentType, err := eb.GetEpubFile(uid, docid, pathParam)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	defer reader.Close()

	if strings.Contains(contentType, "html") {
		body, err := io.ReadAll(reader)
		if err != nil {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		dir := path.Dir("/ui/api/documents/" + docid + "/epub/" + pathParam)
		if !strings.HasSuffix(dir, "/") {
			dir += "/"
		}
		root := "/ui/api/documents/" + docid + "/epub"
		body = prepareEpubHTML(body, dir, root)
		c.Header("Content-Type", contentType)
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Cache-Control", "private, max-age=300")
		c.Data(http.StatusOK, contentType, body)
		return
	}

	c.Header("Content-Type", contentType)
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "private, max-age=300")
	c.DataFromReader(http.StatusOK, -1, contentType, reader, nil)
}

func prepareEpubHTML(htmlBytes []byte, baseURL, epubRoot string) []byte {
	htmlBytes = injectHTMLBase(htmlBytes, baseURL)
	return rewriteRootRelative(htmlBytes, epubRoot)
}

func injectHTMLBase(htmlBytes []byte, baseURL string) []byte {
	lower := bytes.ToLower(htmlBytes)
	if bytes.Contains(lower, []byte("<base")) {
		return htmlBytes
	}
	tag := []byte(`<base href="` + html.EscapeString(baseURL) + `">`)
	head := bytes.Index(lower, []byte("<head"))
	if head >= 0 {
		gt := bytes.IndexByte(htmlBytes[head:], '>')
		if gt < 0 {
			return htmlBytes
		}
		at := head + gt + 1
		out := make([]byte, 0, len(htmlBytes)+len(tag))
		out = append(out, htmlBytes[:at]...)
		out = append(out, tag...)
		out = append(out, htmlBytes[at:]...)
		return out
	}
	htmlTag := bytes.Index(lower, []byte("<html"))
	if htmlTag >= 0 {
		gt := bytes.IndexByte(htmlBytes[htmlTag:], '>')
		if gt < 0 {
			return htmlBytes
		}
		at := htmlTag + gt + 1
		headWrap := append([]byte("<head>"), tag...)
		headWrap = append(headWrap, []byte("</head>")...)
		out := make([]byte, 0, len(htmlBytes)+len(headWrap))
		out = append(out, htmlBytes[:at]...)
		out = append(out, headWrap...)
		out = append(out, htmlBytes[at:]...)
		return out
	}
	return htmlBytes
}

func rewriteRootRelative(htmlBytes []byte, epubRoot string) []byte {
	epubRoot = strings.TrimSuffix(epubRoot, "/")
	if epubRoot == "" {
		return htmlBytes
	}
	lower := bytes.ToLower(htmlBytes)
	keys := [][]byte{
		[]byte(`href="`), []byte(`href='`),
		[]byte(`src="`), []byte(`src='`),
		[]byte(`poster="`), []byte(`poster='`),
	}
	var out []byte
	i := 0
	for i < len(htmlBytes) {
		hit := -1
		for _, k := range keys {
			end := i + len(k)
			if end <= len(lower) && bytes.Equal(lower[i:end], k) {
				hit = end
				break
			}
		}
		if hit < 0 {
			out = append(out, htmlBytes[i])
			i++
			continue
		}
		out = append(out, htmlBytes[i:hit]...)
		i = hit
		if i < len(htmlBytes) && htmlBytes[i] == '/' && (i+1 >= len(htmlBytes) || htmlBytes[i+1] != '/') {
			out = append(out, epubRoot...)
			out = append(out, '/')
			i++
		}
	}
	return out
}
