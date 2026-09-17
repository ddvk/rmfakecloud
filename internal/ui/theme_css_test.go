package ui

import (
	"strings"
	"testing"
)

func TestThemeToCSSPalettes(t *testing.T) {
	cases := []struct {
		id     string
		want   string
		scheme string
	}{
		{"light", "--rm-bg-1: #f7f4ef", "color-scheme: light"},
		{"dark", "--rm-bg-1: #212529", "color-scheme: dark"},
		{"hicontrast", "--rm-fg-1: #ffffff", "color-scheme: dark"},
	}
	s := newThemeStore(t.TempDir())
	for _, tc := range cases {
		xml, _, err := s.getXML(tc.id)
		if err != nil {
			t.Fatal(err)
		}
		css, err := themeToCSS(xml, "desktop")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(css, tc.want) {
			t.Fatalf("%s missing %q in %s", tc.id, tc.want, css)
		}
		if !strings.Contains(css, tc.scheme) {
			t.Fatalf("%s missing %q", tc.id, tc.scheme)
		}
		if !strings.Contains(css, "--rm-on-action:") {
			t.Fatalf("%s missing on-action", tc.id)
		}
	}
}

func TestThemeToCSSSystemFollowsPrefersColorScheme(t *testing.T) {
	s := newThemeStore(t.TempDir())
	xml, _, err := s.getXML("system")
	if err != nil {
		t.Fatal(err)
	}
	css, err := themeToCSS(xml, "desktop")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(css, "@media (prefers-color-scheme: dark)") {
		t.Fatalf("system theme should emit dark media query:\n%s", css)
	}
	if !strings.Contains(css, "--rm-bg-1: #f7f4ef") || !strings.Contains(css, "--rm-bg-1: #212529") {
		t.Fatalf("system theme should include light and dark backgrounds:\n%s", css)
	}
}

func TestHexLuminance(t *testing.T) {
	if hexLuminance("#ffffff") < 0.9 {
		t.Fatal("white should be light")
	}
	if hexLuminance("#000000") > 0.1 {
		t.Fatal("black should be dark")
	}
	if contrastOn("#ffff00") != "#111111" {
		t.Fatal("yellow needs dark text")
	}
}
