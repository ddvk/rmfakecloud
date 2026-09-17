package rmdecode

import (
	"bytes"
	"image/png"
	"testing"
)

func TestRenderNotebookPlaceholderPNG(t *testing.T) {
	b, err := RenderNotebookPlaceholderPNG()
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
}
