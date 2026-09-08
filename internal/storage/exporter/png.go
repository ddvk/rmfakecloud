package exporter

import (
	"bytes"
	"fmt"
	"image/png"
	"io"

	pdf "github.com/unidoc/unipdf/v3/model"
	"github.com/unidoc/unipdf/v3/render"
)

// RenderPagePNG renders a single page of the archive (as exported to PDF) to PNG.
// pageNum is 1-based. Returns PNG bytes.
func RenderPagePNG(a *MyArchive, pageNum int) ([]byte, error) {
	var pdfBuf bytes.Buffer
	if err := RenderRmapi(a, &pdfBuf); err != nil {
		return nil, err
	}
	pdfReader, err := pdf.NewPdfReader(bytes.NewReader(pdfBuf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("open pdf: %w", err)
	}
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, err
	}
	if pageNum < 1 || pageNum > numPages {
		return nil, fmt.Errorf("page %d out of range (1-%d)", pageNum, numPages)
	}
	page, err := pdfReader.GetPage(pageNum)
	if err != nil {
		return nil, err
	}
	device := render.NewImageDevice()
	device.OutputWidth = 1404
	img, err := device.Render(page)
	if err != nil {
		return nil, fmt.Errorf("render page: %w", err)
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// RenderPagePNGReader is like RenderPagePNG but returns a ReadCloser.
func RenderPagePNGReader(a *MyArchive, pageNum int) (io.ReadCloser, error) {
	b, err := RenderPagePNG(a, pageNum)
	if err != nil {
		return nil, err
	}
	return NewSeekCloser(b), nil
}
