package ui

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestTemplateIconSVG(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><rect width="10" height="10"/></svg>`
	payload := `{"iconData":"` + base64.StdEncoding.EncodeToString([]byte(svg)) + `"}`
	got, err := templateIconSVG(strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "<svg") {
		t.Fatalf("unexpected svg: %q", got)
	}
}
