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
	if !strings.Contains(s, "<svg") || !strings.Contains(s, "white") {
		t.Fatalf("unexpected svg: %s", s[:min(200, len(s))])
	}
}

func TestRenderWritingsPDF_Empty(t *testing.T) {
	page := &rm.Rm{Layers: []rm.Layer{}}
	b, err := RenderWritingsPDF(page)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 100 || !strings.HasPrefix(string(b[:5]), "%PDF") {
		t.Fatalf("expected PDF header, got %d bytes", len(b))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
