package ui

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/antchfx/xmlquery"
)

// themeToCSS ports theme-to-css.xsl in pure Go (server-side, no CGO).
func themeToCSS(themeXML []byte, formFactor string) (string, error) {
	doc, err := xmlquery.Parse(bytes.NewReader(themeXML))
	if err != nil {
		return "", err
	}
	theme := xmlquery.FindOne(doc, "/theme")
	if theme == nil {
		return "", fmt.Errorf("missing theme root")
	}

	colors := xmlquery.FindOne(theme, "colors")
	layoutPath := "layout"
	if formFactor == "mobile" {
		if xmlquery.FindOne(theme, "layout/mobile") != nil {
			layoutPath = "layout/mobile"
		}
	}
	chrome := xmlquery.FindOne(theme, layoutPath+"/chrome")
	if chrome == nil {
		chrome = xmlquery.FindOne(theme, "layout/chrome")
	}
	connect := xmlquery.FindOne(theme, layoutPath+"/connect")
	if connect == nil {
		connect = xmlquery.FindOne(theme, "layout/connect")
	}

	var b strings.Builder
	b.WriteString(":root {\n")
	writePaletteVars(&b, colors, chrome, connect, true)
	b.WriteString("}\n")

	if dark := xmlquery.FindOne(theme, "colors-dark"); dark != nil {
		b.WriteString("@media (prefers-color-scheme: dark) {\n:root {\n")
		writePaletteVars(&b, dark, chrome, connect, false)
		b.WriteString("}\n}\n")
	}
	return b.String(), nil
}

func writePaletteVars(b *strings.Builder, colors, chrome, connect *xmlquery.Node, useConnectColors bool) {
	bg1 := nodeAttr(colors, "background1", "#212529")
	bg2 := nodeAttr(colors, "background2", "#0f0f0f")
	fg1 := nodeAttr(colors, "foreground1", "#f8f7f6")
	fg2 := nodeAttr(colors, "foreground2", "#e8e3d9")
	fg3 := nodeAttr(colors, "foreground3", "#e4dbaf")
	action := nodeAttr(colors, "action", "#EE7B30")
	accept := nodeAttr(colors, "accept", "#00e676")
	reject := nodeAttr(colors, "reject", "#ff1744")
	scheme := "dark"
	if hexLuminance(bg1) > 0.5 {
		scheme = "light"
	}

	track, fill, text, pbg, pfg, pborder, code := bg2, action, fg1, bg1, fg2, fg2, fg1
	if useConnectColors {
		track = nodeAttr(connect, "track-color", track)
		fill = nodeAttr(connect, "fill-color", fill)
		text = nodeAttr(connect, "text-color", text)
		pbg = nodeAttr(connect, "prompt-bg", pbg)
		pfg = nodeAttr(connect, "prompt-fg", pfg)
		pborder = nodeAttr(connect, "prompt-border", pborder)
		code = nodeAttr(connect, "code-color", code)
	}

	fontFamily := nodeAttr(connect, "font-family", "system")
	fontSize := nodeAttr(connect, "font-size", "lg")
	ffCSS := `system-ui, -apple-system, "Segoe UI", Roboto, sans-serif`
	switch fontFamily {
	case "serif":
		ffCSS = `Georgia, "Times New Roman", Times, serif`
	case "mono":
		ffCSS = `ui-monospace, SFMono-Regular, Menlo, Consolas, monospace`
	case "rounded":
		ffCSS = `"Trebuchet MS", "Segoe UI", sans-serif`
	}
	fsCSS := "2.75rem"
	switch fontSize {
	case "sm":
		fsCSS = "1.5rem"
	case "md":
		fsCSS = "2rem"
	case "xl":
		fsCSS = "3.5rem"
	}

	fmt.Fprintf(b, "  color-scheme: %s;\n", scheme)
	fmt.Fprintf(b, "  --rm-color-scheme: %s;\n", scheme)
	fmt.Fprintf(b, "  --rm-bg-1: %s;\n", bg1)
	fmt.Fprintf(b, "  --rm-bg-2: %s;\n", bg2)
	fmt.Fprintf(b, "  --rm-fg-1: %s;\n", fg1)
	fmt.Fprintf(b, "  --rm-fg-2: %s;\n", fg2)
	fmt.Fprintf(b, "  --rm-fg-3: %s;\n", fg3)
	fmt.Fprintf(b, "  --rm-action: %s;\n", action)
	fmt.Fprintf(b, "  --rm-on-action: %s;\n", contrastOn(action))
	fmt.Fprintf(b, "  --rm-accept: %s;\n", accept)
	fmt.Fprintf(b, "  --rm-reject: %s;\n", reject)
	fmt.Fprintf(b, "  --rm-connect-gauge-track: %s;\n", track)
	fmt.Fprintf(b, "  --rm-connect-gauge-fill: %s;\n", fill)
	fmt.Fprintf(b, "  --rm-connect-gauge-text: %s;\n", text)
	fmt.Fprintf(b, "  --rm-connect-prompt-bg: %s;\n", pbg)
	fmt.Fprintf(b, "  --rm-connect-prompt-fg: %s;\n", pfg)
	fmt.Fprintf(b, "  --rm-connect-prompt-border: %s;\n", pborder)
	fmt.Fprintf(b, "  --rm-connect-code-color: %s;\n", code)
	fmt.Fprintf(b, "  --rm-connect-font-family: %s;\n", ffCSS)
	fmt.Fprintf(b, "  --rm-connect-font-size: %s;\n", fsCSS)
	fmt.Fprintf(b, "  --rm-chrome-folder: %s;\n", nodeAttr(chrome, "folder-color", action))
	fmt.Fprintf(b, "  --rm-chrome-tab: %s;\n", nodeAttr(chrome, "tab-color", action))
	fmt.Fprintf(b, "  --rm-chrome-outline: %s;\n", nodeAttr(chrome, "outline-color", fg2))
	fmt.Fprintf(b, "  --rm-chrome-drawer: %s;\n", nodeAttr(chrome, "drawer-color", bg2))
	fmt.Fprintf(b, "  --rm-chrome-label: %s;\n", nodeAttr(chrome, "label-color", fg1))
	fmt.Fprintf(b, "  --bs-body-bg: %s !important;\n", bg1)
	fmt.Fprintf(b, "  --bs-body-color: %s !important;\n", fg1)
	fmt.Fprintf(b, "  --bs-border-color: %s !important;\n", fg2)
	fmt.Fprintf(b, "  --bs-tertiary-bg: %s !important;\n", bg1)
	fmt.Fprintf(b, "  --bs-code-color: %s !important;\n", action)
}

func nodeAttr(n *xmlquery.Node, name, fallback string) string {
	if n == nil {
		return fallback
	}
	v := strings.TrimSpace(n.SelectAttr(name))
	if v == "" {
		return fallback
	}
	return v
}

func hexLuminance(s string) float64 {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) < 6 {
		return 0
	}
	r, err1 := strconv.ParseUint(s[0:2], 16, 8)
	g, err2 := strconv.ParseUint(s[2:4], 16, 8)
	b, err3 := strconv.ParseUint(s[4:6], 16, 8)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0
	}
	return (0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)) / 255
}

func contrastOn(bg string) string {
	if hexLuminance(bg) > 0.55 {
		return "#111111"
	}
	return "#ffffff"
}

func applyColorOverridesXML(themeXML []byte, overrides map[string]string) []byte {
	if len(overrides) == 0 {
		return themeXML
	}
	s := string(themeXML)
	idx := strings.Index(s, "<colors")
	if idx < 0 {
		return themeXML
	}
	end := strings.Index(s[idx:], "/>")
	if end < 0 {
		end = strings.Index(s[idx:], ">")
	}
	if end < 0 {
		return themeXML
	}
	end += idx
	seg := s[idx:end]
	for k, v := range overrides {
		if strings.TrimSpace(v) == "" {
			continue
		}
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(k) + `="[^"]*"`)
		if re.MatchString(seg) {
			seg = re.ReplaceAllString(seg, k+`="`+v+`"`)
		} else {
			seg = strings.TrimRight(seg, " \t") + ` ` + k + `="` + v + `"`
		}
	}
	return []byte(s[:idx] + seg + s[end:])
}

func normalizeChromeStyle(style string) string {
	switch strings.TrimSpace(strings.ToLower(style)) {
	case "remarkable", "googledocs", "icloud", "os":
		return strings.ToLower(style)
	default:
		return "remarkable"
	}
}

func chromeStyleFromTheme(themeXML []byte, formFactor string) string {
	doc, err := xmlquery.Parse(bytes.NewReader(themeXML))
	if err != nil {
		return "remarkable"
	}
	path := "layout/chrome"
	if formFactor == "mobile" && xmlquery.FindOne(doc, "/theme/layout/mobile/chrome") != nil {
		path = "layout/mobile/chrome"
	}
	n := xmlquery.FindOne(doc, "/theme/"+path)
	if n == nil {
		return "remarkable"
	}
	return normalizeChromeStyle(n.SelectAttr("style"))
}

func connectAttrsFromTheme(themeXML []byte, formFactor string) (gaugeType, promptLoc, promptStyle, prompt string) {
	gaugeType, promptLoc, promptStyle = "circle", "above", "plain"
	prompt = "Enter this one-time code on your tablet under Account → Connect."
	doc, err := xmlquery.Parse(bytes.NewReader(themeXML))
	if err != nil {
		return
	}
	path := "layout/connect"
	if formFactor == "mobile" && xmlquery.FindOne(doc, "/theme/layout/mobile/connect") != nil {
		path = "layout/mobile/connect"
	}
	n := xmlquery.FindOne(doc, "/theme/"+path)
	if n == nil {
		return
	}
	if v := n.SelectAttr("gauge-type"); v != "" {
		gaugeType = v
	}
	if v := n.SelectAttr("prompt-location"); v != "" {
		promptLoc = v
	}
	if v := n.SelectAttr("prompt-style"); v != "" {
		promptStyle = v
	}
	return
}
