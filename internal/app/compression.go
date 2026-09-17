package app

import (
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// HTTP uses gzip Content-Encoding (DEFLATE / zlib). Request bodies may also
// arrive as Content-Encoding: gzip, deflate, or zlib.

var gzipWriterPool = sync.Pool{
	New: func() any {
		w, err := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		if err != nil {
			panic(err)
		}
		return w
	},
}

var compressedExt = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".webp": true, ".ico": true, ".wasm": true, ".zip": true,
	".gz": true, ".br": true, ".pdf": true, ".mp4": true,
	".woff": true, ".woff2": true,
}

type gzipResponseWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	w.Header().Del("Content-Length")
	return w.writer.Write(data)
}

func (w *gzipResponseWriter) WriteString(s string) (int, error) {
	w.Header().Del("Content-Length")
	return w.writer.Write([]byte(s))
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	w.Header().Del("Content-Length")
	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Flush() {
	_ = w.writer.Flush()
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// compressionMiddleware compresses responses with gzip (zlib DEFLATE) when the
// client advertises Accept-Encoding: gzip, and decompresses gzip/deflate/zlib
// request bodies. Always continues the handler chain.
func compressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := decompressRequestBody(c); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid compressed request body"})
			return
		}

		if shouldGzipResponse(c.Request) {
			gz := gzipWriterPool.Get().(*gzip.Writer)
			defer gzipWriterPool.Put(gz)
			defer gz.Reset(io.Discard)

			gz.Reset(c.Writer)
			c.Header("Content-Encoding", "gzip")
			c.Writer.Header().Add("Vary", "Accept-Encoding")
			c.Writer = &gzipResponseWriter{ResponseWriter: c.Writer, writer: gz}
			defer func() { _ = gz.Close() }()
		}

		c.Next()
	}
}

func shouldGzipResponse(req *http.Request) bool {
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		return false
	}
	if strings.Contains(strings.ToLower(req.Header.Get("Connection")), "upgrade") {
		return false
	}
	if strings.Contains(req.Header.Get("Accept"), "text/event-stream") {
		return false
	}
	ext := strings.ToLower(filepath.Ext(req.URL.Path))
	return !compressedExt[ext]
}

func decompressRequestBody(c *gin.Context) error {
	if c.Request.Body == nil {
		return nil
	}
	enc := strings.ToLower(strings.TrimSpace(c.GetHeader("Content-Encoding")))
	if enc == "" || enc == "identity" {
		return nil
	}

	var (
		r   io.ReadCloser
		err error
	)
	switch enc {
	case "gzip":
		r, err = gzip.NewReader(c.Request.Body)
	case "deflate", "zlib":
		r, err = zlib.NewReader(c.Request.Body)
	default:
		// Leave unknown encodings (e.g. br) for handlers / proxies.
		return nil
	}
	if err != nil {
		return err
	}

	c.Request.Body = r
	c.Request.Header.Del("Content-Encoding")
	c.Request.Header.Del("Content-Length")
	return nil
}
