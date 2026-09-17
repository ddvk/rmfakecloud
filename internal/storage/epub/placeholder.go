package epub

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
)

const placeholderSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 180 240" width="180" height="240">
  <rect width="180" height="240" fill="#f6f0e2"/>
  <rect x="0" y="0" width="14" height="240" fill="#2c2824"/>
  <text x="104" y="124" text-anchor="middle" font-size="16" font-family="system-ui,sans-serif" fill="#3f3b36">EPUB</text>
</svg>
`

// PlaceholderThumbPNG is a book-shaped raster used when an EPUB has no cover.
func PlaceholderThumbPNG() ([]byte, error) {
	const w, h = 180, 240
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	beige := color.RGBA{R: 0xf6, G: 0xf0, B: 0xe2, A: 0xff}
	spine := color.RGBA{R: 0x2c, G: 0x28, B: 0x24, A: 0xff}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if x < 14 {
				img.SetRGBA(x, y, spine)
			} else {
				img.SetRGBA(x, y, beige)
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// PlaceholderThumb is a book-shaped image used when an EPUB has no extractable cover.
func PlaceholderThumb() (io.ReadCloser, string) {
	b, err := PlaceholderThumbPNG()
	if err != nil || len(b) == 0 {
		return io.NopCloser(bytes.NewReader([]byte(placeholderSVG))), "image/svg+xml"
	}
	return io.NopCloser(bytes.NewReader(b)), "image/png"
}
