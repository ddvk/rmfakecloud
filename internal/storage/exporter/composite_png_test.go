package exporter

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestMultiplyBlendPNGs(t *testing.T) {
	bg := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	ink := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			bg.SetNRGBA(x, y, color.NRGBA{R: 200, G: 200, B: 200, A: 255})
			ink.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	ink.SetNRGBA(0, 0, color.NRGBA{R: 0, G: 0, B: 0, A: 255})

	var bgBuf, inkBuf bytes.Buffer
	if err := png.Encode(&bgBuf, bg); err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(&inkBuf, ink); err != nil {
		t.Fatal(err)
	}

	out, err := MultiplyBlendPNGs(bgBuf.Bytes(), inkBuf.Bytes(), nil)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	n := toNRGBA(img)
	// Black ink × gray bg → black
	r, g, b, _ := n.At(0, 0).RGBA()
	if r>>8 != 0 || g>>8 != 0 || b>>8 != 0 {
		t.Fatalf("expected black at ink pixel, got %d,%d,%d", r>>8, g>>8, b>>8)
	}
	// White ink × gray bg → gray unchanged
	r, g, b, _ = n.At(1, 1).RGBA()
	if r>>8 != 200 || g>>8 != 200 || b>>8 != 200 {
		t.Fatalf("expected gray where ink white, got %d,%d,%d", r>>8, g>>8, b>>8)
	}
}

func TestMultiplyBlendFullOverlayNoCrop(t *testing.T) {
	// Device ink 1404×1872; PDF raster same width, shorter height (US Letter).
	const iw, ih = 1404, 1872
	const bw, bh = 1404, 1817
	ink := image.NewNRGBA(image.Rect(0, 0, iw, ih))
	bg := image.NewNRGBA(image.Rect(0, 0, bw, bh))
	for y := 0; y < ih; y++ {
		for x := 0; x < iw; x++ {
			ink.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	for y := 0; y < bh; y++ {
		for x := 0; x < bw; x++ {
			bg.SetNRGBA(x, y, color.NRGBA{R: 180, G: 180, B: 180, A: 255})
		}
	}
	ink.SetNRGBA(99, 0, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	// Stroke below the PDF page must remain visible on the full canvas.
	ink.SetNRGBA(42, bh+10, color.NRGBA{R: 0, G: 0, B: 0, A: 255})

	var bgBuf, inkBuf bytes.Buffer
	if err := png.Encode(&bgBuf, bg); err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(&inkBuf, ink); err != nil {
		t.Fatal(err)
	}
	out, err := MultiplyBlendPNGs(bgBuf.Bytes(), inkBuf.Bytes(), nil)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	b := img.Bounds()
	if b.Dx() != iw || b.Dy() != ih {
		t.Fatalf("expected device viewport %dx%d, got %dx%d", iw, ih, b.Dx(), b.Dy())
	}
	n := toNRGBA(img)
	r, g, bl, _ := n.At(99, 0).RGBA()
	if r>>8 != 0 || g>>8 != 0 || bl>>8 != 0 {
		t.Fatalf("expected black at top-aligned edge, got %d,%d,%d", r>>8, g>>8, bl>>8)
	}
	r, g, bl, _ = n.At(42, bh+10).RGBA()
	if r>>8 != 0 || g>>8 != 0 || bl>>8 != 0 {
		t.Fatalf("expected below-page ink preserved (no crop), got %d,%d,%d", r>>8, g>>8, bl>>8)
	}
}

func TestMultiplyBlendExpandedInkKeepsTopStroke(t *testing.T) {
	// Expanded ink taller than device; without metadata origin, viewport is top-left.
	viewH := 1872
	pad := 50
	iw, ih := 1404, viewH+pad
	ink := image.NewNRGBA(image.Rect(0, 0, iw, ih))
	bg := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < ih; y++ {
		for x := 0; x < iw; x++ {
			ink.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			bg.SetNRGBA(x, y, color.NRGBA{R: 200, G: 200, B: 200, A: 255})
		}
	}
	ink.SetNRGBA(10, 0, color.NRGBA{R: 0, G: 0, B: 0, A: 255})

	var bgBuf, inkBuf bytes.Buffer
	if err := png.Encode(&bgBuf, bg); err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(&inkBuf, ink); err != nil {
		t.Fatal(err)
	}
	out, err := MultiplyBlendPNGs(bgBuf.Bytes(), inkBuf.Bytes(), nil)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != iw || img.Bounds().Dy() != viewH {
		t.Fatalf("expected device crop %dx%d, got %dx%d", iw, viewH, img.Bounds().Dx(), img.Bounds().Dy())
	}
	n := toNRGBA(img)
	r, g, b, _ := n.At(10, 0).RGBA()
	if r>>8 != 0 || g>>8 != 0 || b>>8 != 0 {
		t.Fatalf("top ink in device viewport must remain, got %d,%d,%d", r>>8, g>>8, b>>8)
	}
}

func TestScalePNGToThumb(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 1404, 1872))
	src.SetNRGBA(0, 0, color.NRGBA{R: 10, G: 20, B: 30, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatal(err)
	}
	out, err := ScalePNGToThumb(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	b := img.Bounds()
	if b.Dx() != DeviceThumbPortrait.X || b.Dy() != DeviceThumbPortrait.Y {
		t.Fatalf("thumb size %dx%d, want %dx%d", b.Dx(), b.Dy(), DeviceThumbPortrait.X, DeviceThumbPortrait.Y)
	}
}
