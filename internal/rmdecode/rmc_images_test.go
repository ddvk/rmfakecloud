package rmdecode

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestEncodeRmPageToPNGWithImages_Sample(t *testing.T) {
	if !RMCAvailable() {
		t.Skip("rmc not available")
	}
	sample := "/tmp/rmv6-image-sample"
	rmPath := filepath.Join(sample, "page.rm")
	imgPath := filepath.Join(sample, "a1839a1e-b7d3-4cc3-bdff-0a27e1341fd4.png")
	rm, err := os.ReadFile(rmPath)
	if err != nil {
		t.Skip("sample page.rm missing:", err)
	}
	img, err := os.ReadFile(imgPath)
	if err != nil {
		t.Skip("sample image missing:", err)
	}
	images := map[string][]byte{filepath.Base(imgPath): img}
	out, err := EncodeRmPageToPNGWithImages(rm, images)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) < 1000 {
		t.Fatalf("png too small: %d", len(out))
	}
	svg, err := EncodeRmPageToSVGWithImages(rm, images)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(svg, []byte("<image")) || !bytes.Contains(svg, []byte("base64,")) {
		t.Fatalf("svg missing embedded image")
	}
}
