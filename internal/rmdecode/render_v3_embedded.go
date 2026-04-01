package rmdecode

import (
	"bytes"
	"fmt"

	"github.com/juruen/rmapi/encoding/rm"
)

type v3RenderGroup struct {
	strokes  []rm.Line
	erasures []rm.Line
}

// RenderV3SVGOverlayEmbedded renders a v3 page to transparent SVG.
// It mirrors lines-are-beautiful layering semantics by applying erasures
// as nested masks that only affect prior strokes in the same layer.
func RenderV3SVGOverlayEmbedded(page *rm.Rm) (string, error) {
	if page == nil {
		return "", fmt.Errorf("page is nil")
	}
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">`,
		PageWidthPt, PageHeightPt, PageWidthPt, PageHeightPt,
	))
	maskID := 0
	for layerID, layer := range page.Layers {
		maskID = renderV3Layer(layerID, layer, maskID, &out)
	}
	out.WriteString(`</svg>`)
	return out.String(), nil
}

func renderV3Layer(layerID int, layer rm.Layer, maskSeed int, out *bytes.Buffer) int {
	groups := make([]v3RenderGroup, 0)
	cur := v3RenderGroup{}
	for _, ln := range layer.Lines {
		if isEraserBrush(ln.BrushType) {
			if len(cur.strokes) > 0 {
				cur.erasures = append(cur.erasures, ln)
			}
			continue
		}
		if len(cur.erasures) > 0 {
			groups = append(groups, cur)
			cur = v3RenderGroup{}
		}
		cur.strokes = append(cur.strokes, ln)
	}
	groups = append(groups, cur)

	layerLabel := fmt.Sprintf("layer-%d", layerID)
	for i := len(groups) - 1; i >= 0; i-- {
		g := groups[i]
		if len(g.erasures) > 0 {
			maskLabel := fmt.Sprintf("%s-mask-%d", layerLabel, maskSeed)
			maskSeed++
			out.WriteString(`<mask id="` + maskLabel + `">`)
			out.WriteString(`<rect width="100%" height="100%" fill="white"/>`)
			for _, e := range g.erasures {
				renderV3ErasePath(e, out)
			}
			out.WriteString(`</mask>`)
			out.WriteString(`<g mask="url(#` + maskLabel + `)"`)
		} else {
			out.WriteString(`<g`)
		}
		if i == len(groups)-1 {
			out.WriteString(` id="` + layerLabel + `"`)
		}
		out.WriteString(`>`)
	}
	for _, g := range groups {
		for _, s := range g.strokes {
			renderV3StrokePath(s, out)
		}
		out.WriteString(`</g>`)
	}
	return maskSeed
}

func renderV3ErasePath(line rm.Line, out *bytes.Buffer) {
	if len(line.Points) < 2 {
		return
	}
	out.WriteString(`<path fill="black" stroke="none" d="`)
	renderPathData(line, out)
	out.WriteString(`"/>`)
}

func renderV3StrokePath(line rm.Line, out *bytes.Buffer) {
	if len(line.Points) < 2 {
		return
	}
	out.WriteString(`<path fill="none" stroke="`)
	switch line.BrushColor {
	case rm.Grey:
		out.WriteString(`grey`)
	case rm.White:
		out.WriteString(`white`)
	default:
		out.WriteString(`#000`)
	}
	w := legacyStrokeWidth(line)
	out.WriteString(fmt.Sprintf(`" stroke-width="%g" d="`, w))
	renderPathData(line, out)
	out.WriteString(`"`)
	if isHighlighterBrush(line.BrushType) {
		out.WriteString(` opacity="0.25"`)
	}
	out.WriteString(` stroke-linejoin="round" stroke-linecap="round"/>`)
}

func renderPathData(line rm.Line, out *bytes.Buffer) {
	for i, p := range line.Points {
		if i == 0 {
			out.WriteString(fmt.Sprintf("M%g,%g", p.X, p.Y))
		} else {
			out.WriteString(fmt.Sprintf("L%g,%g", p.X, p.Y))
		}
	}
}

func legacyStrokeWidth(line rm.Line) float64 {
	base := float64(line.BrushSize)
	if base > 10 {
		base = base / 10
	}
	if base <= 0 {
		base = 1
	}
	if isHighlighterBrush(line.BrushType) || isEraserBrush(line.BrushType) {
		return 20*base + 0.01
	}
	w := 18*base - 32
	if w <= 0 {
		return base
	}
	return w
}

func isEraserBrush(bt rm.BrushType) bool {
	return bt == rm.Eraser || bt == rm.EraseArea
}

func isHighlighterBrush(bt rm.BrushType) bool {
	return bt == rm.Highlighter || bt == rm.HighlighterV5 || int(bt) == 5 || int(bt) == 18
}
