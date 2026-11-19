package rmdecode

import (
	"strings"
	"testing"
)

func TestParseVersion_V6SampleHeader(t *testing.T) {
	// v6 header is exactly 43 bytes: prefix + "6" + spaces
	hdr := []byte("reMarkable .lines file, version=6" + strings.Repeat(" ", 10))
	if len(hdr) != HeaderLen {
		t.Fatalf("test header length %d, want %d", len(hdr), HeaderLen)
	}
	v, err := ParseVersion(hdr)
	if err != nil {
		t.Fatal(err)
	}
	if v != 6 {
		t.Fatalf("version %d, want 6", v)
	}
}

func TestParseVersion_V5(t *testing.T) {
	hdr := []byte("reMarkable .lines file, version=5          ")
	if len(hdr) != HeaderLen {
		t.Fatal("bad test header length")
	}
	v, err := ParseVersion(hdr)
	if err != nil {
		t.Fatal(err)
	}
	if v != 5 {
		t.Fatalf("version %d, want 5", v)
	}
}
