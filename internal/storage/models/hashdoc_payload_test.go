package models

import "testing"

func TestEffectivePayloadType(t *testing.T) {
	cases := []struct {
		name    string
		payload string
		files   []string
		want    string
	}{
		{name: "pdf content", payload: "pdf", want: "pdf"},
		{name: "epub dotted", payload: ".epub", want: "epub"},
		{name: "notebook", payload: "notebook", want: "notebook"},
		{name: "visible name with pdf file", payload: "Quarterly Report", files: []string{"abc.metadata", "abc.content", "abc.pdf"}, want: "pdf"},
		{name: "visible name with epub file", payload: "A Novel", files: []string{"abc.epub"}, want: "epub"},
		{name: "visible name notebook", payload: "Quick Notes", files: []string{"abc.metadata", "abc.content", "p1.rm"}, want: "notebook"},
		{name: "empty", payload: "", want: "notebook"},
	}
	for _, tc := range cases {
		d := &HashDoc{PayloadType: tc.payload}
		for _, n := range tc.files {
			d.Files = append(d.Files, &HashEntry{EntryName: n})
		}
		if got := d.EffectivePayloadType(); got != tc.want {
			t.Fatalf("%s: got %q want %q", tc.name, got, tc.want)
		}
	}
}
