// +build !cairo

package exporter

import (
	"errors"
	"io"
)

const (
	DeviceWidth  = 1404
	DeviceHeight = 1872
)

type PdfGenerator struct {
	options PdfGeneratorOptions
}

type PdfGeneratorOptions struct {
	AddPageNumbers  bool
	AllPages        bool
	AnnotationsOnly bool
}

func (p *PdfGenerator) Generate(zip *MyArchive, output io.Writer, options PdfGeneratorOptions) error {
	return errors.New("PDF generation with annotations requires building with Cairo support. Build with: go build -tags cairo")
}
