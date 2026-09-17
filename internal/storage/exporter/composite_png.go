package exporter

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"

	"github.com/ddvk/rmfakecloud/internal/rmdecode"
)

// DeviceThumbSize matches on-device thumbnail storage (~384×512 portrait).
var (
	DeviceThumbPortrait  = image.Pt(384, 512)
	DeviceThumbLandscape = image.Pt(512, 384)
)

// MultiplyBlendPNGs composites a white-backed ink PNG over a background PNG using
// multiply blending. The PDF/EPUB background is placed with rmapi-style width- or
// height-fit (top-left, natural aspect) into the ink page viewport. The returned
// image is cropped to that viewport — for PDF annotations this is the full PDF
// page raster (correct Letter aspect, no device letterbox squish).
//
// Both inputs are PNG bytes. target (if set) scales the final viewport.
func MultiplyBlendPNGs(bgPNG, inkPNG []byte, target *image.Point) ([]byte, error) {
	bgImg, err := png.Decode(bytes.NewReader(bgPNG))
	if err != nil {
		return nil, fmt.Errorf("decode background: %w", err)
	}
	inkImg, err := png.Decode(bytes.NewReader(inkPNG))
	if err != nil {
		return nil, fmt.Errorf("decode ink: %w", err)
	}

	ink := toNRGBA(inkImg)
	meta := rmdecode.ReadInkDeviceMeta(inkPNG)
	if meta.ViewW < 1 {
		meta.ViewW = rmdecode.PageWidthPt
	}
	if meta.ViewH < 1 {
		meta.ViewH = rmdecode.PageHeightPt
	}
	bg := placeBackgroundOnInkCanvas(toNRGBA(bgImg), ink.Bounds().Dx(), ink.Bounds().Dy(), meta)

	full := image.NewNRGBA(ink.Bounds())
	w, h := ink.Bounds().Dx(), ink.Bounds().Dy()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bi := bg.PixOffset(x, y)
			ii := ink.PixOffset(x, y)
			full.Pix[ii+0] = mul8(bg.Pix[bi+0], ink.Pix[ii+0])
			full.Pix[ii+1] = mul8(bg.Pix[bi+1], ink.Pix[ii+1])
			full.Pix[ii+2] = mul8(bg.Pix[bi+2], ink.Pix[ii+2])
			full.Pix[ii+3] = 255
		}
	}

	ox, oy := meta.OriginX, meta.OriginY
	vw, vh := meta.ViewW, meta.ViewH
	if ox < 0 {
		ox = 0
	}
	if oy < 0 {
		oy = 0
	}
	if ox+vw > w {
		vw = w - ox
	}
	if oy+vh > h {
		vh = h - oy
	}
	if vw < 1 {
		vw = 1
	}
	if vh < 1 {
		vh = 1
	}
	out := cropNRGBA(full, ox, oy, vw, vh)

	if target != nil && target.X > 0 && target.Y > 0 {
		out = scaleNRGBA(out, target.X, target.Y)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, out); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// placeBackgroundOnInkCanvas draws the PDF/EPUB background into a white ink
// canvas using rmapi fit (width- or height-fill, top-left, natural aspect)
// inside the device viewport at meta.Origin.
func placeBackgroundOnInkCanvas(bg *image.NRGBA, iw, ih int, meta rmdecode.InkDeviceMeta) *image.NRGBA {
	if iw < 1 {
		iw = 1
	}
	if ih < 1 {
		ih = 1
	}
	viewW, viewH := meta.ViewW, meta.ViewH
	if viewW < 1 {
		viewW = rmdecode.PageWidthPt
	}
	if viewH < 1 {
		viewH = rmdecode.PageHeightPt
	}
	if viewW > iw {
		viewW = iw
	}
	if viewH > ih {
		viewH = ih
	}
	ox, oy := meta.OriginX, meta.OriginY
	if ox < 0 {
		ox = 0
	}
	if oy < 0 {
		oy = 0
	}
	if ox+viewW > iw {
		ox = iw - viewW
		if ox < 0 {
			ox = 0
		}
	}
	if oy+viewH > ih {
		oy = ih - viewH
		if oy < 0 {
			oy = 0
		}
	}

	bw, bh := bg.Bounds().Dx(), bg.Bounds().Dy()
	canvas := image.NewNRGBA(image.Rect(0, 0, iw, ih))
	for i := 0; i < len(canvas.Pix); i += 4 {
		canvas.Pix[i+0] = 255
		canvas.Pix[i+1] = 255
		canvas.Pix[i+2] = 255
		canvas.Pix[i+3] = 255
	}
	if bw < 1 || bh < 1 {
		return canvas
	}
	if bw == viewW && bh == viewH && ox == 0 && oy == 0 && iw == viewW && ih == viewH {
		return bg
	}

	deviceAspect := float64(viewH) / float64(viewW)
	pageAspect := float64(bh) / float64(bw)

	var dw, dh int
	if pageAspect < deviceAspect {
		dw = viewW
		dh = int(math.Round(float64(bh) * float64(viewW) / float64(bw)))
	} else {
		dh = viewH
		dw = int(math.Round(float64(bw) * float64(viewH) / float64(bh)))
	}
	if dw < 1 {
		dw = 1
	}
	if dh < 1 {
		dh = 1
	}
	scaled := bg
	if dw != bw || dh != bh {
		scaled = scaleNRGBA(bg, dw, dh)
	}
	draw.Draw(canvas, image.Rect(ox, oy, ox+dw, oy+dh), scaled, image.Point{}, draw.Src)
	return canvas
}

// fitInkToBackground samples the ink device canvas onto a background-sized image.
// Kept for tests/thumbs that still want a page-cropped blend.
func fitInkToBackground(ink *image.NRGBA, bw, bh int) *image.NRGBA {
	if bw < 1 {
		bw = 1
	}
	if bh < 1 {
		bh = 1
	}
	iw, ih := ink.Bounds().Dx(), ink.Bounds().Dy()
	if iw == bw && ih == bh {
		return ink
	}

	deviceAspect := float64(ih) / float64(iw)
	pageAspect := float64(bh) / float64(bw)

	var dispW, dispH, ox, oy int
	if pageAspect < deviceAspect {
		dispW = iw
		dispH = int(math.Round(float64(bh) * float64(iw) / float64(bw)))
		if dispH > ih {
			dispH = ih
		}
		ox, oy = 0, 0
	} else {
		dispH = ih
		dispW = int(math.Round(float64(bw) * float64(ih) / float64(bh)))
		if dispW > iw {
			dispW = iw
		}
		ox, oy = 0, 0
	}
	if dispW < 1 {
		dispW = 1
	}
	if dispH < 1 {
		dispH = 1
	}

	cropped := cropNRGBA(ink, ox, oy, dispW, dispH)
	if dispW == bw && dispH == bh {
		return cropped
	}
	return scaleNRGBA(cropped, bw, bh)
}

func cropNRGBA(src *image.NRGBA, x, y, w, h int) *image.NRGBA {
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	if x+w > sw {
		w = sw - x
	}
	if y+h > sh {
		h = sh - y
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	for row := 0; row < h; row++ {
		si := src.PixOffset(x, y+row)
		di := dst.PixOffset(0, row)
		copy(dst.Pix[di:di+4*w], src.Pix[si:si+4*w])
	}
	return dst
}

// ScalePNGToThumb scales a PNG to device thumbnail size, preserving orientation.
func ScalePNGToThumb(pngBytes []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, err
	}
	b := img.Bounds()
	var target image.Point
	if b.Dx() >= b.Dy() {
		target = DeviceThumbLandscape
	} else {
		target = DeviceThumbPortrait
	}
	scaled := scaleNRGBA(toNRGBA(img), target.X, target.Y)
	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func mul8(a, b uint8) uint8 {
	return uint8((uint16(a) * uint16(b)) / 255)
}

func toNRGBA(src image.Image) *image.NRGBA {
	if n, ok := src.(*image.NRGBA); ok {
		return n
	}
	b := src.Bounds()
	dst := image.NewNRGBA(b)
	draw.Draw(dst, b, src, b.Min, draw.Src)
	return dst
}

func scaleNRGBA(src *image.NRGBA, tw, th int) *image.NRGBA {
	if tw < 1 {
		tw = 1
	}
	if th < 1 {
		th = 1
	}
	dst := image.NewNRGBA(image.Rect(0, 0, tw, th))
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	for y := 0; y < th; y++ {
		sy := int(math.Floor(float64(y) * float64(sh) / float64(th)))
		if sy >= sh {
			sy = sh - 1
		}
		for x := 0; x < tw; x++ {
			sx := int(math.Floor(float64(x) * float64(sw) / float64(tw)))
			if sx >= sw {
				sx = sw - 1
			}
			si := src.PixOffset(sx, sy)
			di := dst.PixOffset(x, y)
			copy(dst.Pix[di:di+4], src.Pix[si:si+4])
		}
	}
	return dst
}
