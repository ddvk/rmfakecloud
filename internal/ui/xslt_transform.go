package ui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	xsltOnce   sync.Once
	xsltMethod string // "python" or ""
)

func detectXSLT() {
	xsltOnce.Do(func() {
		if _, err := exec.LookPath("python3"); err != nil {
			return
		}
		cmd := exec.Command("python3", "-c", "from lxml import etree")
		if err := cmd.Run(); err != nil {
			return
		}
		xsltMethod = "python"
	})
}

// transformPageXML applies shell.xsl to page XML.
// ok=false means caller should use the browser XSLT bootstrap.
func (app *ReactAppWrapper) transformPageXML(pageXML []byte) (html []byte, ok bool, err error) {
	detectXSLT()
	if xsltMethod != "python" {
		return nil, false, nil
	}
	xsl, err := app.readStatic("xslt/shell.xsl")
	if err != nil {
		return nil, false, err
	}

	tmpXML, err := os.CreateTemp("", "rmfc-page-*.xml")
	if err != nil {
		return nil, false, err
	}
	defer os.Remove(tmpXML.Name())
	if _, err := tmpXML.Write(pageXML); err != nil {
		_ = tmpXML.Close()
		return nil, false, err
	}
	_ = tmpXML.Close()

	tmpXSL, err := os.CreateTemp("", "rmfc-shell-*.xsl")
	if err != nil {
		return nil, false, err
	}
	defer os.Remove(tmpXSL.Name())
	if _, err := tmpXSL.Write(xsl); err != nil {
		_ = tmpXSL.Close()
		return nil, false, err
	}
	_ = tmpXSL.Close()

	script := fmt.Sprintf(`
from lxml import etree
import sys
xsl = etree.parse(%q)
transform = etree.XSLT(xsl)
page = etree.parse(%q)
result = transform(page)
html = etree.tostring(result, pretty_print=False, method="html", encoding="unicode")
if not html.lstrip().lower().startswith("<!doctype"):
    html = "<!DOCTYPE html>\n" + html
sys.stdout.write(html)
`, tmpXSL.Name(), tmpXML.Name())

	cmd := exec.Command("python3", "-c", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, false, fmt.Errorf("xslt python: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), true, nil
}

func bootstrapXSLTHTML(pageXML []byte) []byte {
	escaped := strings.ReplaceAll(string(pageXML), "</script>", "<\\/script>")
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"en\"><head>")
	b.WriteString("<meta charset=\"utf-8\"/>")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"/>")
	b.WriteString("<title>rmfakecloud</title>")
	b.WriteString("<link rel=\"stylesheet\" href=\"/assets/app.css\"/>")
	b.WriteString("</head><body>")
	b.WriteString("<p id=\"xslt-status\">Loading…</p>")
	b.WriteString("<script id=\"page-xml\" type=\"application/xml\">")
	b.WriteString(escaped)
	b.WriteString("</script>")
	b.WriteString("<script src=\"/assets/js/bootstrap-xslt.js\"></script>")
	b.WriteString("</body></html>")
	return []byte(b.String())
}
