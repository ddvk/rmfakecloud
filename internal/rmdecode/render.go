package rmdecode

import (
	"bytes"
	"fmt"

	"github.com/jung-kurt/gofpdf"
	"github.com/juruen/rmapi/encoding/rm"
)

// Nominal reMarkable 2 page size in points (matches exporter).
const (
	PageWidthPt  = 1404
	PageHeightPt = 1872
)

// RenderWritingsSVG returns a standalone SVG document with a white background and all
// non-eraser strokes from a v3/v5 decoded page.
func RenderWritingsSVG(page *rm.Rm) (string, error) {
	if page == nil {
		return "", fmt.Errorf("page is nil")
	}
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">`,
		PageWidthPt, PageHeightPt, PageWidthPt, PageHeightPt,
	))
	buf.WriteString(fmt.Sprintf(`<rect width="100%%" height="100%%" fill="white"/>`))

	for _, layer := range page.Layers {
		for _, line := range layer.Lines {
			if len(line.Points) < 2 {
				continue
			}
			if line.BrushType == rm.Eraser || line.BrushType == rm.EraseArea {
				continue
			}
			w := float64(line.BrushSize) / 10.0
			if w <= 0 {
				w = 1
			}
			buf.WriteString(`<path fill="none" stroke="#000" stroke-linecap="round" stroke-linejoin="round"`)
			buf.WriteString(fmt.Sprintf(` stroke-width="%g" d="`, w))
			for i, pt := range line.Points {
				if i == 0 {
					buf.WriteString(fmt.Sprintf("M %g %g ", pt.X, pt.Y))
				} else {
					buf.WriteString(fmt.Sprintf("L %g %g ", pt.X, pt.Y))
				}
			}
			buf.WriteString(`"/>`)
		}
	}
	buf.WriteString(`</svg>`)
	return buf.String(), nil
}

// RenderWritingsPDF returns a single-page PDF with a white background and all
// non-eraser strokes (v3/v5 only).
func RenderWritingsPDF(page *rm.Rm) ([]byte, error) {
	if page == nil {
		return nil, fmt.Errorf("page is nil")
	}
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "pt",
		Size:           gofpdf.SizeType{Wd: float64(PageWidthPt), Ht: float64(PageHeightPt)},
	})
	pdf.AddPage()
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	// White page
	pdf.SetFillColor(255, 255, 255)
	pdf.Rect(0, 0, float64(PageWidthPt), float64(PageHeightPt), "F")

	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineCapStyle("round")
	pdf.SetLineJoinStyle("round")

	for _, layer := range page.Layers {
		for _, line := range layer.Lines {
			if len(line.Points) < 2 {
				continue
			}
			if line.BrushType == rm.Eraser || line.BrushType == rm.EraseArea {
				continue
			}
			w := float64(line.BrushSize) / 10.0
			if w <= 0 {
				w = 1
			}
			pdf.SetLineWidth(w)
			for i := 1; i < len(line.Points); i++ {
				a, b := line.Points[i-1], line.Points[i]
				pdf.Line(float64(a.X), float64(a.Y), float64(b.X), float64(b.Y))
			}
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
