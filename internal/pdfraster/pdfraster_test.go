package pdfraster

import (
	"os"
	"testing"
)

func TestRenderPagePdftoppm(t *testing.T) {
	pdfPath := "../../../test/test.pdf"
	b, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Skip("test.pdf missing:", err)
	}
	t.Setenv(EnvPDFRenderer, "pdftoppm")
	out, err := RenderPage(b, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) < 100 || out[0] != 0x89 {
		t.Fatalf("expected PNG output, got %d bytes", len(out))
	}
}
