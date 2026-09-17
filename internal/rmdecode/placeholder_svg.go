package rmdecode

import (
	"fmt"
	"strings"
)

// RenderNotebookPlaceholderSVG is a ruled-paper page used when a .rm page
// cannot be decoded (missing file, unsupported v6, etc.).
func RenderNotebookPlaceholderSVG() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">`,
		PageWidthPt, PageHeightPt, PageWidthPt, PageHeightPt,
	))
	b.WriteString(`<rect width="100%" height="100%" fill="#f3efe4"/>`)
	marginX := float64(PageWidthPt) * 0.12
	b.WriteString(fmt.Sprintf(
		`<path fill="none" stroke="#d9a39a" stroke-width="3" d="M %g 0 V %d"/>`,
		marginX, PageHeightPt,
	))
	b.WriteString(`<g stroke="#d9d2c4" stroke-width="2">`)
	for y := 80; y < PageHeightPt; y += 48 {
		b.WriteString(fmt.Sprintf(`<path d="M 0 %d H %d"/>`, y, PageWidthPt))
	}
	b.WriteString(`</g></svg>`)
	return b.String()
}
