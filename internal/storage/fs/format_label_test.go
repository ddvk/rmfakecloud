package fs

import (
	"testing"

	"github.com/ddvk/rmfakecloud/internal/storage/models"
)

func TestParsePDFVersion(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"%PDF-1.7\n", "1.7"},
		{"%PDF-1.7\n%\xe2\xe3\xcf\xd3", "1.7"},
		{"%PDF-1.4\r\n", "1.4"},
		{"%PDF-2.0\n", "2.0"},
		{"\xef\xbb\xbf%PDF-1.5\n", "1.5"},
		{"not a pdf", ""},
		{"%PDF-", ""},
	}
	for _, tc := range cases {
		if got := ParsePDFVersion([]byte(tc.in)); got != tc.want {
			t.Fatalf("%q: got %q want %q", tc.in, got, tc.want)
		}
	}
}

func TestSniffFormatLabelFallbacks(t *testing.T) {
	if got := sniffFormatLabel(nil, &models.HashDoc{}); got != "RM" {
		t.Fatalf("empty notebook: %q", got)
	}
	if got := sniffFormatLabel(nil, &models.HashDoc{PayloadType: "pdf"}); got != "PDF" {
		t.Fatalf("pdf: %q", got)
	}
	if got := sniffFormatLabel(nil, &models.HashDoc{PayloadType: "epub"}); got != "EPUB" {
		t.Fatalf("epub: %q", got)
	}
}
