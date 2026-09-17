package epub

import (
	"bytes"
	"image/png"
	"testing"
)

func TestPlaceholderThumbPNG(t *testing.T) {
	b, err := PlaceholderThumbPNG()
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != 180 || img.Bounds().Dy() != 240 {
		t.Fatalf("size %v", img.Bounds())
	}
	rc, ct := PlaceholderThumb()
	defer rc.Close()
	if ct != "image/png" {
		t.Fatalf("content type %q", ct)
	}
}
