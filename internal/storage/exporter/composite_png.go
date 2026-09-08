package exporter

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"
)

// DeviceThumbSize matches on-device thumbnail storage (~384×512 portrait).
var (
	DeviceThumbPortrait  = image.Pt(384, 512)
	DeviceThumbLandscape = image.Pt(512, 384)
)

// MultiplyBlendPNGs composites a white-backed ink PNG over a background PNG using
// multiply blending (same idea as the web DocumentPageThumb and the device’s
// highlighter darken / ink-over-paper composite).
//
// Both inputs are PNG bytes. The result is scaled to fit target (nil = ink size).
func MultiplyBlendPNGs(bgPNG, inkPNG []byte, target *image.Point) ([]byte, error) {
	bgImg, err := png.Decode(bytes.NewReader(bgPNG))
	if err != nil {
		return nil, fmt.Errorf("decode background: %w", err)
	}
	inkImg, err := png.Decode(bytes.NewReader(inkPNG))
	if err != nil {
		return nil, fmt.Errorf("decode ink: %w", err)
	}

	bg := toNRGBA(bgImg)
	ink := toNRGBA(inkImg)

	// Scale ink to background bounds (PDF page aspect vs notebook 1404×1872).
	if ink.Bounds().Dx() != bg.Bounds().Dx() || ink.Bounds().Dy() != bg.Bounds().Dy() {
		ink = scaleNRGBA(ink, bg.Bounds().Dx(), bg.Bounds().Dy())
	}

	out := image.NewNRGBA(bg.Bounds())
	w, h := bg.Bounds().Dx(), bg.Bounds().Dy()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			bi := bg.PixOffset(x, y)
			ii := ink.PixOffset(x, y)
			out.Pix[bi+0] = mul8(bg.Pix[bi+0], ink.Pix[ii+0])
			out.Pix[bi+1] = mul8(bg.Pix[bi+1], ink.Pix[ii+1])
			out.Pix[bi+2] = mul8(bg.Pix[bi+2], ink.Pix[ii+2])
			out.Pix[bi+3] = 255
		}
	}

	if target != nil && target.X > 0 && target.Y > 0 {
		out = scaleNRGBA(out, target.X, target.Y)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, out); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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
