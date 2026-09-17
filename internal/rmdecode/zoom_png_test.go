package rmdecode

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestInkDeviceMetaRoundTrip(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	for i := range img.Pix {
		img.Pix[i] = 255
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	meta := InkDeviceMeta{OriginX: 12, OriginY: 34, ViewW: 1404, ViewH: 1872}
	out, err := WriteInkDeviceMeta(buf.Bytes(), meta)
	if err != nil {
		t.Fatal(err)
	}
	got := ReadInkDeviceMeta(out)
	if got != meta {
		t.Fatalf("got %+v want %+v", got, meta)
	}
}

func TestNudgePNGNoCropUp(t *testing.T) {
	const w, h, n = 40, 60, 15
	src := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	src.SetNRGBA(3, 0, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	src.SetNRGBA(5, h-1, color.NRGBA{R: 10, G: 20, B: 30, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatal(err)
	}
	out, err := NudgePNG(buf.Bytes(), 0, -n)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	b := img.Bounds()
	if b.Dx() != w || b.Dy() != h+n {
		t.Fatalf("size %dx%d, want %dx%d", b.Dx(), b.Dy(), w, h+n)
	}
	got := toNRGBAImg(img)
	r, g, bl, _ := got.At(3, 0).RGBA()
	if r>>8 != 0 || g>>8 != 0 || bl>>8 != 0 {
		t.Fatalf("top ink must be preserved, got %d,%d,%d", r>>8, g>>8, bl>>8)
	}
	r, g, bl, _ = got.At(5, h-1).RGBA()
	if r>>8 != 10 || g>>8 != 20 || bl>>8 != 30 {
		t.Fatalf("bottom ink misplaced, got %d,%d,%d", r>>8, g>>8, bl>>8)
	}
}
