package ui

import "testing"

func TestNormalizeDocType(t *testing.T) {
	cases := map[string]string{
		"pdf":                  "pdf",
		".PDF":                 "pdf",
		"application/pdf":      "pdf",
		"epub":                 "epub",
		"application/epub+zip": "epub",
		"notebook":             "notebook",
		"Quick Notes":          "notebook",
		"":                     "notebook",
	}
	for in, want := range cases {
		if got := normalizeDocType(in); got != want {
			t.Fatalf("%q: got %q want %q", in, got, want)
		}
	}
}
