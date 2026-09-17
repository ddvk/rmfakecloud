package models

import (
	"encoding/json"
	"testing"
)

func TestResolvedLastOpenedIndexFromCPagesUUID(t *testing.T) {
	c := ContentFile{
		LastOpenedPage: 0,
		CPages: CPages{
			LastOpened: &CPageValue{Value: json.RawMessage(`"page-b"`)},
			Pages: []CPage{
				{ID: "page-a"},
				{ID: "page-b"},
				{ID: "page-c"},
			},
		},
	}
	if got := c.ResolvedLastOpenedIndex(); got != 1 {
		t.Fatalf("got %d want 1", got)
	}
	if ids := c.PageIDs(); len(ids) != 3 || ids[1] != "page-b" {
		t.Fatalf("page ids %v", ids)
	}
}

func TestMergeLastOpenedPrefersCPages(t *testing.T) {
	c := ContentFile{
		LastOpenedPage: 0,
		CPages: CPages{
			LastOpened: &CPageValue{Value: json.RawMessage(`"p2"`)},
			Pages:      []CPage{{ID: "p1"}, {ID: "p2"}},
		},
	}
	if got := c.MergeLastOpenedIndex(9); got != 1 {
		t.Fatalf("got %d want 1", got)
	}
}

func TestMergeLastOpenedUsesMetadataWhenNoCPages(t *testing.T) {
	c := ContentFile{LastOpenedPage: 0, Pages: []interface{}{"a", "b", "c"}}
	if got := c.MergeLastOpenedIndex(2); got != 2 {
		t.Fatalf("got %d want 2", got)
	}
}

func TestThumbPage1(t *testing.T) {
	if got := ThumbPage1(0, 3); got != 1 {
		t.Fatalf("first page: %d", got)
	}
	if got := ThumbPage1(4, 3); got != 3 {
		t.Fatalf("clamp: %d", got)
	}
	if got := ThumbPage1(-1, 0); got != 1 {
		t.Fatalf("empty: %d", got)
	}
}
