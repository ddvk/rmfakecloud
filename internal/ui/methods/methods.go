package methods

import (
	"fmt"
	"math"
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
	{id: "cornell", name: "Cornell Notes", svg: cornellSVG()},
	{id: "outline", name: "Outline", svg: outlineSVG()},
	{id: "mindmap", name: "Mind Map", svg: mindmapSVG()},
	{id: "flowchart", name: "Flowchart", svg: flowchartSVG()},
	{id: "checklist", name: "Checklist", svg: checklistSVG()},
}

func cornellSVG() string {
	// Left column ~25%, right margin, bottom summary area
	cueW := width / 4
	summaryH := height / 5
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">
<rect width="100%%" height="100%%" fill="white"/>
<line x1="%d" y1="0" x2="%d" y2="%d" stroke="#c0c0c0" stroke-width="2"/>
<line x1="0" y1="%d" x2="%d" y2="%d" stroke="#c0c0c0" stroke-width="2"/>
<text x="%d" y="40" font-family="sans-serif" font-size="28" fill="#888">Cues</text>
<text x="%d" y="40" font-family="sans-serif" font-size="28" fill="#888">Notes</text>
<text x="20" y="%d" font-family="sans-serif" font-size="28" fill="#888">Summary</text>
</svg>`,
		width, height, width, height,
		cueW, cueW, height,
		height-summaryH, width, height-summaryH,
		cueW/2-20, cueW+40,
		height-summaryH+32,
	)
}

func outlineSVG() string {
	indent := 80
	lineH := 52
	lines := ""
	for i := 0; i < 28; i++ {
		y := 60 + i*lineH
		level := i % 4
		x := 40 + level*indent
		lines += fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#e0e0e0" stroke-width="1"/>`, x, y+20, width-40, y+20)
	}
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d"><rect width="100%%" height="100%%" fill="white"/>%s</svg>`, width, height, width, height, lines)
}

func mindmapSVG() string {
	cx, cy := width/2, height/2
	r := 120
	nodes := 6
	lines := ""
	for i := 0; i < nodes; i++ {
		angle := float64(i) * (2 * math.Pi / float64(nodes))
		x := int(float64(cx) + float64(r)*math.Cos(angle))
		y := int(float64(cy) + float64(r)*math.Sin(angle))
		lines += fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#b0b0b0" stroke-width="2"/>`, cx, cy, x, y)
		lines += fmt.Sprintf(`<circle cx="%d" cy="%d" r="24" fill="none" stroke="#808080" stroke-width="2"/>`, x, y)
	}
	lines += fmt.Sprintf(`<circle cx="%d" cy="%d" r="40" fill="none" stroke="#606060" stroke-width="2"/>`, cx, cy)
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d"><rect width="100%%" height="100%%" fill="white"/>%s<text x="%d" y="%d" text-anchor="middle" font-family="sans-serif" font-size="24" fill="#666">Topic</text></svg>`, width, height, width, height, lines, cx, cy+8)
}

func flowchartSVG() string {
	// Simple flowchart: start -> process -> decision -> end
	bw, bh := 200, 56
	x1, y := width/2-bw/2, 200
	x2 := width/2 - 80
	x3 := width/2 - 40
	lines := fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="28" fill="none" stroke="#70a070" stroke-width="2"/>`, x1, y, bw, bh)
	lines += fmt.Sprintf(`<rect x="%d" y="%d" width="160" height="%d" fill="none" stroke="#7090b0" stroke-width="2"/>`, x2, 320, bh)
	lines += fmt.Sprintf(`<polygon points="%d,%d %d,%d %d,%d" fill="none" stroke="#b09070" stroke-width="2"/>`, x3, 420, x3-50, 500, x3+50, 500)
	lines += fmt.Sprintf(`<rect x="%d" y="580" width="%d" height="%d" rx="28" fill="none" stroke="#a07070" stroke-width="2"/>`, width/2-80, 160, bh)
	lines += `<line x1="702" y1="256" x2="702" y2="318" stroke="#888" stroke-width="1"/>`
	lines += `<line x1="662" y1="376" x2="662" y2="418" stroke="#888" stroke-width="1"/>`
	lines += `<line x1="702" y1="476" x2="702" y2="538" stroke="#888" stroke-width="1"/>`
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d"><rect width="100%%" height="100%%" fill="white"/>%s</svg>`, width, height, width, height, lines)
}

func checklistSVG() string {
	lines := ""
	step := 56
	for y := 80; y < height-80; y += step {
		lines += fmt.Sprintf(`<circle cx="40" cy="%d" r="14" fill="none" stroke="#a0a0a0" stroke-width="2"/>`, y)
		lines += fmt.Sprintf(`<line x1="60" y1="%d" x2="%d" y2="%d" stroke="#e0e0e0" stroke-width="1"/>`, y+14, width-40, y+14)
	}
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d"><rect width="100%%" height="100%%" fill="white"/>%s</svg>`, width, height, width, height, lines)
}

// BuiltinMethodsDirectory returns a Directory entry "Methods" with method documents as children.
func BuiltinMethodsDirectory() *viewmodel.Directory {
	children := make([]viewmodel.Entry, 0, len(builtins))
	for _, b := range builtins {
		children = append(children, &viewmodel.Document{
			ID:           b.id,
			Name:         b.name,
			DocumentType: "method",
			LastModified: time.Time{},
			Size:         0,
		})
	}
	return &viewmodel.Directory{
		ID:           "methods",
		Name:         "rm Methods",
		Entries:      children,
		LastModified: time.Time{},
		IsFolder:     true,
	}
}

// GetSVG returns the SVG content for a method ID, or empty string if not found.
func GetSVG(id string) string {
	for _, b := range builtins {
		if b.id == id {
			return b.svg
		}
	}
	return ""
}
