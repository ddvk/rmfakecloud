package exporter

import (
	"bytes"
	"fmt"

	"github.com/juruen/rmapi/encoding/rm"
)

const (
	rmWidth  = 1404
	rmHeight = 1872
)

func svgHeader() string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">`, rmWidth, rmHeight, rmWidth, rmHeight)
}

// RenderPageAnnotationsSVG renders the handwriting (.rm) content for a page as an SVG overlay.
// The output is intended to be layered on top of a background PNG of the original PDF page.
func RenderPageAnnotationsSVG(a *MyArchive, pageNum int) (string, error) {
	if a == nil {
		return "", fmt.Errorf("archive is nil")
	}
	if pageNum < 1 || pageNum > len(a.Pages) {
		return "", fmt.Errorf("page %d out of range (1-%d)", pageNum, len(a.Pages))
	}
	p := a.Pages[pageNum-1]
	if p.Data == nil {
		return svgHeader() + `</svg>`, nil
	}

	var buf bytes.Buffer
	buf.WriteString(svgHeader())
	// no background rect: we want a transparent overlay

	for _, layer := range p.Data.Layers {
		for _, line := range layer.Lines {
			if len(line.Points) < 2 {
				continue
			}
			if line.BrushType == rm.Eraser || line.BrushType == rm.EraseArea {
				continue
			}
			// Force black-only: stroke always black, no fill.
			w := float64(line.BrushSize) / 10.0
			if w <= 0 {
				w = 1
			}

			// Build a simple polyline path.
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

