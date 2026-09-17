package rmdecode

import (
	"bytes"
	"image/png"

	"github.com/fogleman/gg"
)

const (
	thumbWidth  = 180
	thumbHeight = 240
)

// RenderNotebookPlaceholderPNG is a ruled-paper thumbnail used when a page
// cannot be decoded (missing .rm, unsupported v6, etc.).
func RenderNotebookPlaceholderPNG() ([]byte, error) {
	dc := gg.NewContext(thumbWidth, thumbHeight)
	dc.SetRGB(0.953, 0.937, 0.894)
	dc.Clear()
	dc.SetRGB(0.851, 0.639, 0.604)
	dc.SetLineWidth(1.2)
	dc.DrawLine(thumbWidth*0.12, 0, thumbWidth*0.12, thumbHeight)
	dc.Stroke()
	dc.SetRGB(0.851, 0.824, 0.769)
	dc.SetLineWidth(1)
	for y := 18.0; y < thumbHeight; y += 18 {
		dc.DrawLine(0, y, thumbWidth, y)
		dc.Stroke()
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
