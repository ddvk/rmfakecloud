package fs

import (
	"testing"

	"github.com/ddvk/rmfakecloud/internal/storage/models"
)

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
