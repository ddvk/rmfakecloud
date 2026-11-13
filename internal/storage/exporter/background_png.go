package exporter

import (
	"bytes"
	"fmt"
	"image/png"
	"io"

	"github.com/ddvk/rmfakecloud/internal/pdfraster"
	pdf "github.com/unidoc/unipdf/v3/model"
	"github.com/unidoc/unipdf/v3/render"
)

func init() {
	pdfraster.UnipdfFallback = renderPayloadPDFBytesUnipdf
}

// RenderPDFBytesToPNG rasters one PDF page (1-based) to PNG via pdfraster.
func RenderPDFBytesToPNG(pdfBytes []byte, pageNum int) ([]byte, error) {
	return pdfraster.RenderPage(pdfBytes, pageNum)
}

// RenderPayloadPagePNG renders a single page of the *payload* PDF to PNG.
// This is intended for PDF documents where we want the background without handwriting.
// pageNum is 1-based. Returns PNG bytes.
//
// Prefer out-of-process PDF engines (pdftoppm / mutool), matching the device’s
// xochitl_pdf_renderer architecture; fall back to unipdf when those are missing.
func RenderPayloadPagePNG(a *MyArchive, pageNum int) ([]byte, error) {
	if a == nil || a.PayloadReader == nil {
		return nil, fmt.Errorf("no payload reader")
	}
	if _, err := a.PayloadReader.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	b, err := io.ReadAll(a.PayloadReader)
	if err != nil {
		return nil, err
	}
	return pdfraster.RenderPage(b, pageNum)
}

// RenderPayloadPagePNGReader is like RenderPayloadPagePNG but returns a ReadCloser.
func RenderPayloadPagePNGReader(a *MyArchive, pageNum int) (io.ReadCloser, error) {
	b, err := RenderPayloadPagePNG(a, pageNum)
	if err != nil {
		return nil, err
	}
	return NewSeekCloser(b), nil
}

func renderPayloadPDFBytesUnipdf(pdfBytes []byte, pageNum int) ([]byte, error) {
	pdfReader, err := pdf.NewPdfReader(bytes.NewReader(pdfBytes))
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
	device.OutputWidth = pdfraster.DevicePortraitWidth
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
