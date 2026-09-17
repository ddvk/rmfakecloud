package models

import (
	"encoding/json"
	"strings"
)

// CPages is the software 3+ page tree in a .content file.
type CPages struct {
	LastOpened *CPageValue `json:"lastOpened"`
	Pages      []CPage     `json:"pages"`
}

// CPageValue is cPages.lastOpened.value (page UUID or index).
type CPageValue struct {
	Value json.RawMessage `json:"value"`
}

// CPage is one page in cPages.pages.
type CPage struct {
	ID string `json:"id"`
}

// PageIDs returns page UUIDs from cPages.pages or the legacy pages array.
func (c ContentFile) PageIDs() []string {
	if len(c.CPages.Pages) > 0 {
		out := make([]string, 0, len(c.CPages.Pages))
		for _, p := range c.CPages.Pages {
			if id := strings.TrimSpace(p.ID); id != "" {
				out = append(out, id)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	out := make([]string, 0, len(c.Pages))
	for _, p := range c.Pages {
		switch v := p.(type) {
		case string:
			if id := strings.TrimSpace(v); id != "" {
				out = append(out, id)
			}
		case map[string]interface{}:
			if id, ok := v["id"].(string); ok {
				if id = strings.TrimSpace(id); id != "" {
					out = append(out, id)
				}
			}
		}
	}
	return out
}

func (c ContentFile) lastOpenedFromCPages() (int, bool) {
	if c.CPages.LastOpened == nil {
		return 0, false
	}
	raw := strings.TrimSpace(string(c.CPages.LastOpened.Value))
	if raw == "" || raw == "null" {
		return 0, false
	}
	ids := c.PageIDs()
	if strings.HasPrefix(raw, "\"") {
		var id string
		if json.Unmarshal(c.CPages.LastOpened.Value, &id) != nil {
			return 0, false
		}
		id = strings.TrimSpace(id)
		if id == "" {
			return 0, false
		}
		for i, p := range ids {
			if p == id {
				return i, true
			}
		}
		return 0, false
	}
	var n int
	if json.Unmarshal(c.CPages.LastOpened.Value, &n) != nil || n < 0 {
		return 0, false
	}
	if len(ids) > 0 && n >= len(ids) {
		n = len(ids) - 1
	}
	return n, true
}

// ResolvedLastOpenedIndex is the 0-based page the tablet last had open:
// cPages.lastOpened, then lastOpenedPage, clamped to the page list.
func (c ContentFile) ResolvedLastOpenedIndex() int {
	if idx, ok := c.lastOpenedFromCPages(); ok {
		return idx
	}
	ids := c.PageIDs()
	idx := c.LastOpenedPage
	if idx < 0 {
		idx = 0
	}
	if len(ids) > 0 && idx >= len(ids) {
		idx = len(ids) - 1
	}
	return idx
}

// MergeLastOpenedIndex prefers the per-file .content pref over metadata.
func (c ContentFile) MergeLastOpenedIndex(metaIndex int) int {
	if idx, ok := c.lastOpenedFromCPages(); ok {
		return idx
	}
	idx := metaIndex
	if c.LastOpenedPage > idx {
		idx = c.LastOpenedPage
	}
	if idx < 0 {
		idx = 0
	}
	ids := c.PageIDs()
	if len(ids) > 0 && idx >= len(ids) {
		idx = len(ids) - 1
	}
	return idx
}

// ThumbPage1 is the 1-based page to render as a thumbnail.
func ThumbPage1(opened0, pageCount int) int {
	n := opened0 + 1
	if n < 1 {
		n = 1
	}
	if pageCount > 0 && n > pageCount {
		n = pageCount
	}
	return n
}
