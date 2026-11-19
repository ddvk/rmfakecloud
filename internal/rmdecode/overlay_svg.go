package rmdecode

import (
	"fmt"
	"regexp"
	"strings"
)

// opaquePageBackground matches full-page white/light rects that some exporters
// embed (breaks transparent overlay on PDF/EPUB backgrounds).
var opaquePageBackground = regexp.MustCompile(
	`(?i)<rect\b[^>]*\b(?:width\s*=\s*["'](?:100%|1404)["'][^>]*height\s*=\s*["'](?:100%|1872)["']|height\s*=\s*["'](?:100%|1872)["'][^>]*width\s*=\s*["'](?:100%|1404)["'])[^>]*\bfill\s*=\s*["'](?:#?fff(?:fff)?|white|rgb\(\s*255\s*,\s*255\s*,\s*255\s*\))["'][^>]*/>`,
)

// EnsureTransparentOverlaySVG strips opaque full-page backgrounds so the SVG
// can sit on PDF/EPUB (or blank paper) as a transparent ink layer.
func EnsureTransparentOverlaySVG(svg string) string {
	s := opaquePageBackground.ReplaceAllString(svg, "")
	// Also drop a leading white rect with no size attrs that fills via CSS.
	s = regexp.MustCompile(`(?i)<rect\b[^>]*\bfill\s*=\s*["'](?:#?fff(?:fff)?|white)["'][^>]*\bwidth\s*=\s*["']100%["'][^>]*/>`).ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

// RenderRmOverlaySVG renders raw .rm bytes to a transparent SVG overlay.
// This is the app’s primary ink path for PDF/EPUB overlays and standalone notebooks.
//
//	v6 → rmc (RMFAKECLOUD_RMC_SRC / rmc binary)
//	v3 → lines2svg when available, else embedded eraser-aware SVG
//	v5 → Go polyline overlay
func RenderRmOverlaySVG(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty .rm data")
	}
	ver, err := ParseVersion(data)
	if err != nil {
		return "", err
	}
	switch ver {
	case 6:
		s, err := RenderV6SVGWithRMC(data)
		if err != nil {
			return "", err
		}
		return EnsureTransparentOverlaySVG(s), nil
	case 3:
		if s, err := RenderV3SVGWithLines2SVG(data); err == nil {
			return EnsureTransparentOverlaySVG(s), nil
		} else if Lines2SVGAvailable() {
			return "", fmt.Errorf("lines2svg available but failed: %w", err)
		}
		page, err := DecodeLegacy(data)
		if err != nil {
			return "", err
		}
		s, err := RenderV3SVGOverlayEmbedded(page)
		if err != nil {
			return "", err
		}
		return EnsureTransparentOverlaySVG(s), nil
	case 5:
		page, err := DecodeLegacy(data)
		if err != nil {
			return "", err
		}
		s, err := RenderV3SVGOverlayEmbedded(page)
		if err != nil {
			return "", err
		}
		return EnsureTransparentOverlaySVG(s), nil
	default:
		return "", fmt.Errorf("unsupported .rm version %d", ver)
	}
}
