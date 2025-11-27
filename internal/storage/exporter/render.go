package exporter

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"

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

// RenderRmapi renders with Cairo-based PDF generator
func RenderRmapi(a *MyArchive, output io.Writer) error {
	pdfgen := PdfGenerator{}
	options := PdfGeneratorOptions{
		AllPages: true,
	}
	return pdfgen.Generate(a, output, options)
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
