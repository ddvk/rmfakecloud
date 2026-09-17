package rmdecode

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"math"
)

// ZoomPNGAboutCenter scales image content about its center by factor (<1 zooms out),
// filling new margins with white. Factor <=0 or ~1 is a no-op.
func ZoomPNGAboutCenter(pngBytes []byte, factor float64) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, err
	}
	src := toNRGBAImg(img)
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	return zoomPNGAbout(src, float64(w)/2, float64(h)/2, factor)
}

// ZoomPNGAboutPoint scales about (cx, cy) in image pixel coordinates.
func ZoomPNGAboutPoint(pngBytes []byte, cx, cy, factor float64) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, err
	}
	return zoomPNGAbout(toNRGBAImg(img), cx, cy, factor)
}

func zoomPNGAbout(src *image.NRGBA, cx, cy, factor float64) ([]byte, error) {
	if factor <= 0 || math.Abs(factor-1) < 1e-6 {
		var buf bytes.Buffer
		if err := png.Encode(&buf, src); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	if w < 1 || h < 1 {
		var buf bytes.Buffer
		if err := png.Encode(&buf, src); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(dst.Pix); i += 4 {
		dst.Pix[i+0] = 255
		dst.Pix[i+1] = 255
		dst.Pix[i+2] = 255
		dst.Pix[i+3] = 255
	}
	inv := 1 / factor
	for y := 0; y < h; y++ {
		sy := cy + (float64(y)-cy)*inv
		siy := int(math.Floor(sy))
		if siy < 0 || siy >= h {
			continue
		}
		for x := 0; x < w; x++ {
			sx := cx + (float64(x)-cx)*inv
			six := int(math.Floor(sx))
			if six < 0 || six >= w {
				continue
			}
			si := src.PixOffset(six, siy)
			di := dst.PixOffset(x, y)
			copy(dst.Pix[di:di+4], src.Pix[si:si+4])
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SquishPNG scales content about the image center. Positive shrink compresses;
// negative stretch (edges clip on the fixed canvas).
func SquishPNG(pngBytes []byte, shrinkXPx, shrinkYPx int) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, err
	}
	src := toNRGBAImg(img)
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	return squishPNGAbout(src, float64(w)/2, float64(h)/2, shrinkXPx, shrinkYPx)
}

// SquishPNGAboutPoint scales about (cx, cy).
func SquishPNGAboutPoint(pngBytes []byte, cx, cy float64, shrinkXPx, shrinkYPx int) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, err
	}
	return squishPNGAbout(toNRGBAImg(img), cx, cy, shrinkXPx, shrinkYPx)
}

func squishPNGAbout(src *image.NRGBA, cx, cy float64, shrinkXPx, shrinkYPx int) ([]byte, error) {
	if shrinkXPx == 0 && shrinkYPx == 0 {
		var buf bytes.Buffer
		if err := png.Encode(&buf, src); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	sx := 1.0
	sy := 1.0
	if shrinkXPx > 0 {
		if w <= shrinkXPx {
			var buf bytes.Buffer
			if err := png.Encode(&buf, src); err != nil {
				return nil, err
			}
			return buf.Bytes(), nil
		}
		sx = float64(w-shrinkXPx) / float64(w)
	} else if shrinkXPx < 0 {
		sx = float64(w-shrinkXPx) / float64(w)
	}
	if shrinkYPx > 0 {
		if h <= shrinkYPx {
			var buf bytes.Buffer
			if err := png.Encode(&buf, src); err != nil {
				return nil, err
			}
			return buf.Bytes(), nil
		}
		sy = float64(h-shrinkYPx) / float64(h)
	} else if shrinkYPx < 0 {
		sy = float64(h-shrinkYPx) / float64(h)
	}
	return scalePNGXYAbout(src, cx, cy, sx, sy)
}

func scalePNGXYAbout(src *image.NRGBA, cx, cy, sx, sy float64) ([]byte, error) {
	if sx <= 0 || sy <= 0 {
		return nil, nil
	}
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(dst.Pix); i += 4 {
		dst.Pix[i+0] = 255
		dst.Pix[i+1] = 255
		dst.Pix[i+2] = 255
		dst.Pix[i+3] = 255
	}
	invX := 1 / sx
	invY := 1 / sy
	for y := 0; y < h; y++ {
		sySrc := cy + (float64(y)-cy)*invY
		siy := int(math.Floor(sySrc))
		if siy < 0 || siy >= h {
			continue
		}
		for x := 0; x < w; x++ {
			sxSrc := cx + (float64(x)-cx)*invX
			six := int(math.Floor(sxSrc))
			if six < 0 || six >= w {
				continue
			}
			si := src.PixOffset(six, siy)
			di := dst.PixOffset(x, y)
			copy(dst.Pix[di:di+4], src.Pix[si:si+4])
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// NudgePNG translates image content by (dx, dy) pixels (dy negative = up).
// Expands the canvas so no ink is cropped; vacated areas are white.
// Returns the new PNG and how much the top-left of the previous image shifted
// into the new canvas (always >= 0).
func NudgePNG(pngBytes []byte, dx, dy int) ([]byte, error) {
	out, _, _, err := NudgePNGExt(pngBytes, dx, dy)
	return out, err
}

// NudgePNGExt is NudgePNG plus the draw offset of the old image in the new canvas.
func NudgePNGExt(pngBytes []byte, dx, dy int) ([]byte, int, int, error) {
	if dx == 0 && dy == 0 {
		return pngBytes, 0, 0, nil
	}
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, 0, 0, err
	}
	src := toNRGBAImg(img)
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	minX := 0
	if dx < 0 {
		minX = dx
	}
	minY := 0
	if dy < 0 {
		minY = dy
	}
	maxX := w
	if dx+w > maxX {
		maxX = dx + w
	}
	maxY := h
	if dy+h > maxY {
		maxY = dy + h
	}
	nw, nh := maxX-minX, maxY-minY
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	dst := image.NewNRGBA(image.Rect(0, 0, nw, nh))
	for i := 0; i < len(dst.Pix); i += 4 {
		dst.Pix[i+0] = 255
		dst.Pix[i+1] = 255
		dst.Pix[i+2] = 255
		dst.Pix[i+3] = 255
	}
	ox, oy := dx-minX, dy-minY
	draw.Draw(dst, image.Rect(ox, oy, ox+w, oy+h), src, image.Point{}, draw.Src)
	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, 0, 0, err
	}
	return buf.Bytes(), ox, oy, nil
}

func toNRGBAImg(src image.Image) *image.NRGBA {
	if n, ok := src.(*image.NRGBA); ok {
		return n
	}
	b := src.Bounds()
	dst := image.NewNRGBA(b)
	draw.Draw(dst, b, src, b.Min, draw.Src)
	return dst
}
