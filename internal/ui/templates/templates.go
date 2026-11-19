package templates

import (
	"fmt"
	"time"

	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
)

// reMarkable 2 nominal dimensions (points at 1404x1872)
const width = 1404
const height = 1872

var builtins = []struct {
	id   string
	name string
	svg  string
}{
	{
		id:   "blank",
		name: "Blank",
		svg:  blankSVG(),
	},
	{
		id:   "lined",
		name: "Lined",
		svg:  linedSVG(),
	},
	{
		id:   "grid",
		name: "Grid",
		svg:  gridSVG(),
	},
	{
		id:   "dotted",
		name: "Dotted",
		svg:  dottedSVG(),
	},
}

func blankSVG() string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d"><rect width="100%%" height="100%%" fill="white"/></svg>`, width, height, width, height)
}

func linedSVG() string {
	lines := ""
	step := 48
	for y := step; y < height; y += step {
		lines += fmt.Sprintf(`<line x1="0" y1="%d" x2="%d" y2="%d" stroke="#000" stroke-width="1"/>`, y, width, y)
	}
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d"><rect width="100%%" height="100%%" fill="white"/>%s</svg>`, width, height, width, height, lines)
}

func gridSVG() string {
	lines := ""
	step := 36
	for x := step; x < width; x += step {
		lines += fmt.Sprintf(`<line x1="%d" y1="0" x2="%d" y2="%d" stroke="#000" stroke-width="1"/>`, x, x, height)
	}
	for y := step; y < height; y += step {
		lines += fmt.Sprintf(`<line x1="0" y1="%d" x2="%d" y2="%d" stroke="#000" stroke-width="1"/>`, y, width, y)
	}
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d"><rect width="100%%" height="100%%" fill="white"/>%s</svg>`, width, height, width, height, lines)
}

func dottedSVG() string {
	dots := ""
	step := 24
	for y := step; y < height; y += step {
		for x := step; x < width; x += step {
			dots += fmt.Sprintf(`<circle cx="%d" cy="%d" r="1" fill="#000"/>`, x, y)
		}
	}
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d"><rect width="100%%" height="100%%" fill="white"/>%s</svg>`, width, height, width, height, dots)
}

// BuiltinTemplatesDirectory returns a Directory entry "Templates" with template documents as children.
func BuiltinTemplatesDirectory() *viewmodel.Directory {
	children := make([]viewmodel.Entry, 0, len(builtins))
	for _, b := range builtins {
		children = append(children, &viewmodel.Document{
			ID:           b.id,
			Name:         b.name,
			DocumentType: "template",
			LastModified: time.Time{},
			Size:         0,
		})
	}
	return &viewmodel.Directory{
		ID:           "templates",
		Name:         "Templates",
		Entries:      children,
		LastModified: time.Time{},
		IsFolder:     true,
	}
}

// GetSVG returns the SVG content for a template ID, or empty string if not found.
func GetSVG(id string) string {
	for _, b := range builtins {
		if b.id == id {
			return b.svg
		}
	}
	return ""
}
