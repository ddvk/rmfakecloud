package rmdecode

import (
	"bytes"
	"testing"

	"github.com/juruen/rmapi/encoding/rm"
)

func TestRenderWritingsPNG_Empty(t *testing.T) {
	page := &rm.Rm{Layers: []rm.Layer{}}
	b, err := RenderWritingsPNG(page)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 100 || !bytes.HasPrefix(b, []byte{0x89, 0x50, 0x4e, 0x47}) {
		t.Fatalf("expected PNG signature, got %d bytes", len(b))
	}
}
