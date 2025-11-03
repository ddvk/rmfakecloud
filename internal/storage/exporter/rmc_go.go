package exporter

import (
	"bytes"
	"fmt"
	"io"

	rmc "github.com/joagonca/rmc-go"
)

// ExportV6ToPdfNative converts v6 .rm file to PDF using rmc-go library (in-process)
// This uses the Cairo renderer for native PDF generation
func ExportV6ToPdfNative(rmData []byte, output io.Writer) error {
	opts := &rmc.Options{
		UseLegacy: false, // Always use Cairo renderer (not Inkscape)
	}

	// Convert from bytes to PDF bytes
	pdfData, err := rmc.ConvertFromBytes(rmData, rmc.FormatPDF, opts)
	if err != nil {
		return fmt.Errorf("failed to convert v6 rm to PDF: %w", err)
	}

	// Write to output
	_, err = io.Copy(output, bytes.NewReader(pdfData))
	if err != nil {
		return fmt.Errorf("failed to write PDF output: %w", err)
	}

	return nil
}

// ExportV6ToSvgNative converts v6 .rm file to SVG using rmc-go library
func ExportV6ToSvgNative(rmData []byte, output io.Writer) error {
	opts := &rmc.Options{}

	svgData, err := rmc.ConvertFromBytes(rmData, rmc.FormatSVG, opts)
	if err != nil {
		return fmt.Errorf("failed to convert v6 rm to SVG: %w", err)
	}

	_, err = io.Copy(output, bytes.NewReader(svgData))
	if err != nil {
		return fmt.Errorf("failed to write SVG output: %w", err)
	}

	return nil
}

// ExportV6MultiPageToPdfNative converts multiple v6 .rm pages to a single PDF
func ExportV6MultiPageToPdfNative(pages [][]byte, output io.Writer) error {
	if len(pages) == 0 {
		return fmt.Errorf("no pages provided")
	}

	opts := &rmc.Options{
		UseLegacy: false, // Use Cairo renderer
	}

	// Use rmc-go's multipage function
	pdfData, err := rmc.ConvertMultipleFromBytes(pages, opts)
	if err != nil {
		return fmt.Errorf("failed to convert multiple v6 pages to PDF: %w", err)
	}

	_, err = io.Copy(output, bytes.NewReader(pdfData))
	return err
}
