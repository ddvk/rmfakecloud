package exporter

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	rm2pdf "github.com/poundifdef/go-remarkable2pdf"
)

// RenderPoundifdef caligraphy pen is nice
func RenderPoundifdef(input, output string) (io.ReadCloser, error) {
	reader, err := zip.OpenReader(input)
	if err != nil {
		return nil, fmt.Errorf("can't open file %w", err)
	}
	defer reader.Close()

	writer, err := os.Create(output)
	if err != nil {
		return nil, fmt.Errorf("can't create outputfile %w", err)
	}

	err = rm2pdf.RenderRmNotebookFromZip(&reader.Reader, writer)
	if err != nil {
		writer.Close()
		return nil, fmt.Errorf("can't render file %w", err)
	}

	_, err = writer.Seek(0, 0)
	if err != nil {
		writer.Close()
		return nil, fmt.Errorf("can't rewind file %w", err)
	}

	return writer, nil
}

// RenderRmapi renders with rmapi
func RenderRmapi(a *MyArchive, output io.Writer) error {
	pdfgen := PdfGenerator{}
	options := PdfGeneratorOptions{
		AllPages: true,
	}
	return pdfgen.Generate(a, output, options)
}

// v6RendererCmd and v6RendererPy point to the external v6 renderer
// (python script + venv). They can be overridden via environment
// variables RMC_RENDER_CMD / RMC_RENDER_PY.
var (
	v6RendererCmd = envOrDefault("RMC_RENDER_CMD", "/usr/local/lib/rmfakecloud-render/render.sh")
	v6RendererPy  = envOrDefault("RMC_RENDER_PY", "/usr/local/lib/rmfakecloud-render/render_v6.py")
)

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// RenderV6Fallback renders documents containing .rm v6 pages by
// delegating to an external python renderer (rmc + pypdf). Pages that
// the bundled parser understands (v3/v5) are rendered first with the
// usual pipeline when a background pdf exists; v6 pages are then
// appended in document order.
func RenderV6Fallback(a *MyArchive, output io.Writer) error {
	if len(a.V6Pages) == 0 {
		return errors.New("no v6 pages")
	}

	tmpDir, err := os.MkdirTemp("", "rmv6")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	outPath := filepath.Join(tmpDir, "out.pdf")
	args := []string{v6RendererPy, "--output", outPath}
	for _, p := range a.V6Pages {
		rmPath := filepath.Join(tmpDir, p.PageID+".rm")
		if err := os.WriteFile(rmPath, p.Data, 0600); err != nil {
			return err
		}
		args = append(args, "--page", p.PageID+":"+rmPath)
	}

	cmd := exec.Command(v6RendererCmd, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("v6 renderer failed: %w, stderr: %s", err, stderr.String())
	}

	pdfFile, err := os.Open(outPath)
	if err != nil {
		return err
	}
	defer pdfFile.Close()

	_, err = io.Copy(output, pdfFile)
	return err
}

type SeekCloser struct {
	*bytes.Reader
}

// Close closes
func (*SeekCloser) Close() error {
	return nil
}

func NewSeekCloser(b []byte) io.ReadSeekCloser {

	r := &SeekCloser{
		Reader: bytes.NewReader(b),
	}
	return r
}
