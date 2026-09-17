package fs

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/ddvk/rmfakecloud/internal/storage/models"
)

func TestValidRasterImage(t *testing.T) {
	if validRasterImage(nil) || validRasterImage([]byte("not-an-image")) {
		t.Fatal("expected invalid")
	}
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	if !validRasterImage(buf.Bytes()) {
		t.Fatal("expected valid png")
	}
	webp := make([]byte, 12)
	copy(webp[0:4], []byte("RIFF"))
	copy(webp[8:12], []byte("WEBP"))
	if !validRasterImage(webp) {
		t.Fatal("expected webp magic accepted")
	}
}

func TestFilterValidPageImages(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	in := map[string][]byte{
		"ok.png":    buf.Bytes(),
		"bad.bin":   []byte("nope"),
		"empty.png": nil,
	}
	out := filterValidPageImages(in)
	if len(out) != 1 || out["ok.png"] == nil {
		t.Fatalf("got %#v", out)
	}
}

func TestIsPageImageExt(t *testing.T) {
	if !isPageImageExt("a.png") || !isPageImageExt("b.JPEG") {
		t.Fatal("expected image extensions")
	}
	if isPageImageExt("x.rm") || isPageImageExt("y") {
		t.Fatal("unexpected image extension")
	}
}

func TestReadPageImagesNil(t *testing.T) {
	if readPageImages(nil, nil, "abc") != nil {
		t.Fatal("expected nil")
	}
	doc := &models.HashDoc{}
	if readPageImages(doc, nil, "") != nil {
		t.Fatal("expected nil for empty page")
	}
}
