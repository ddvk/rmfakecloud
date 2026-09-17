package rmdecode

import (
	"strings"
	"testing"

	"github.com/juruen/rmapi/encoding/rm"
)

func TestRenderWritingsSVG_Empty(t *testing.T) {
	page := &rm.Rm{Layers: []rm.Layer{}}
	s, err := RenderWritingsSVG(page)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(s, "<svg") || !strings.Contains(s, `xmlns="http://www.w3.org/2000/svg"`) {
		t.Fatalf("not svg: %s", s[:min(80, len(s))])
	}
	if !strings.Contains(s, "viewBox=") {
		t.Fatal("missing viewBox")
	}
}

func TestRenderNotebookPlaceholderSVG(t *testing.T) {
	s := RenderNotebookPlaceholderSVG()
	if !strings.Contains(s, "<svg") || !strings.Contains(s, "viewBox=") {
		t.Fatalf("placeholder: %s", s[:min(120, len(s))])
	}
}

func TestEncodeRmPageToSVG_Empty(t *testing.T) {
	if _, err := EncodeRmPageToSVG(nil); err == nil {
		t.Fatal("expected error")
	}
}
