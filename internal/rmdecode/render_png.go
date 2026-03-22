package rmdecode

import (
	"bytes"
	"fmt"
	"image/png"

	"github.com/fogleman/gg"
	"github.com/juruen/rmapi/encoding/rm"
)

// RenderWritingsPNG renders all non-eraser strokes to a white RGBA page and encodes PNG.
func RenderWritingsPNG(page *rm.Rm) ([]byte, error) {
	if page == nil {
		return nil, fmt.Errorf("page is nil")
	}
	dc := gg.NewContext(PageWidthPt, PageHeightPt)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)

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
			dc.SetLineWidth(w)
			dc.SetLineCap(gg.LineCapRound)
			dc.SetLineJoin(gg.LineJoinRound)
			p0 := line.Points[0]
			dc.MoveTo(float64(p0.X), float64(p0.Y))
			for i := 1; i < len(line.Points); i++ {
				p := line.Points[i]
				dc.LineTo(float64(p.X), float64(p.Y))
			}
			dc.Stroke()
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
