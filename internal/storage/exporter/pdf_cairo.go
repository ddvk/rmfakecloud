//go:build cairo

package exporter

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"unsafe"

	"github.com/ddvk/rmfakecloud/internal/encoding/rm"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/sirupsen/logrus"
	"github.com/ungerik/go-cairo"
)

/*
#cgo pkg-config: cairo
#include <stdlib.h>
#include <cairo.h>
#include <cairo-pdf.h>
*/
import "C"

const (
	DeviceWidth  = 1404
	DeviceHeight = 1872
)

// rmPageSize is the default page size for blank templates (in PDF points: 1/72 inch)
var rmPageSize = struct{ Width, Height float64 }{445, 594}

type PdfGenerator struct {
	options       PdfGeneratorOptions
	backgroundPDF []byte
	template      bool
}

type PdfGeneratorOptions struct {
	AddPageNumbers  bool
	AllPages        bool
	AnnotationsOnly bool //export the annotations without the background/pdf
}

func normalized(p1 rm.Point, scale float64) (float64, float64) {
	return float64(p1.X) * scale, float64(p1.Y) * scale
}

// setPDFPageSize sets the size for the current page in a PDF surface
func setPDFPageSize(surface *cairo.Surface, width, height float64) {
	surfacePtr, _ := surface.Native()
	C.cairo_pdf_surface_set_size((*C.cairo_surface_t)(unsafe.Pointer(surfacePtr)), C.double(width), C.double(height))
}

func (p *PdfGenerator) Generate(zip *MyArchive, output io.Writer, options PdfGeneratorOptions) error {
	p.options = options

	if len(zip.Pages) == 0 {
		if zip.PayloadReader != nil {
			_, err := io.Copy(output, zip.PayloadReader)
			return err
		}
		return fmt.Errorf("the document has no pages")
	}

	if err := p.initBackgroundPages(zip.PayloadReader); err != nil {
		return err
	}

	// If we have a background PDF and not annotations-only mode, we need a two-step process
	if p.backgroundPDF != nil && !p.options.AnnotationsOnly {
		return p.generateWithBackground(zip, output)
	}

	// Otherwise, simple case: just annotations or blank pages
	return p.generateAnnotationsOnly(zip, output)
}

func (p *PdfGenerator) generateAnnotationsOnly(zip *MyArchive, output io.Writer) error {
	// Create a temporary file for PDF output (Cairo requires a file path)
	tmpFile, err := os.CreateTemp("", "rmfakecloud-annotations-*.pdf")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Determine first page dimensions
	var firstWidth, firstHeight float64
	if p.template {
		firstWidth, firstHeight = rmPageSize.Width, rmPageSize.Height
	} else {
		// TODO: Get dimensions from background PDF
		firstWidth, firstHeight = rmPageSize.Width, rmPageSize.Height
	}

	// Create PDF surface
	pdfSurface := cairo.NewPDFSurface(tmpPath, firstWidth, firstHeight, cairo.PDF_VERSION_1_5)
	defer pdfSurface.Finish()

	pageCount := 0
	for _, pageAnnotations := range zip.Pages {
		hasContent := pageAnnotations.Data != nil

		// Skip pages without content unless AllPages is set
		if !p.options.AllPages && !hasContent {
			continue
		}

		pageCount++

		// Set page size (for pages after the first)
		if pageCount > 1 {
			var pageWidth, pageHeight float64
			if p.template {
				pageWidth, pageHeight = rmPageSize.Width, rmPageSize.Height
			} else {
				// TODO: Get dimensions from background PDF page
				pageWidth, pageHeight = rmPageSize.Width, rmPageSize.Height
			}
			setPDFPageSize(pdfSurface, pageWidth, pageHeight)
		}

		// Calculate scale
		pageWidth := firstWidth
		pageHeight := firstHeight
		ratio := pageHeight / pageWidth

		var scale float64
		if ratio < 1.33 {
			scale = pageWidth / DeviceWidth
		} else {
			scale = pageHeight / DeviceHeight
		}

		// Draw annotations if present
		if hasContent {
			if err := p.drawAnnotations(pdfSurface, pageAnnotations.Data, scale, pageHeight); err != nil {
				return err
			}
		}

		// Add page numbers if requested
		if p.options.AddPageNumbers {
			p.drawPageNumber(pdfSurface, pageCount, pageWidth, pageHeight)
		}

		// Show page (prepare for next page)
		if pageCount < len(zip.Pages) || p.options.AllPages {
			pdfSurface.ShowPage()
		}
	}

	pdfSurface.Finish()

	// Copy temp file to output
	tmpFileRead, err := os.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to open temp file: %w", err)
	}
	defer tmpFileRead.Close()

	_, err = io.Copy(output, tmpFileRead)
	return err
}

func (p *PdfGenerator) generateWithBackground(zip *MyArchive, output io.Writer) error {
	// Step 1: Create annotations-only PDF
	tmpAnnotations, err := os.CreateTemp("", "rmfakecloud-annotations-*.pdf")
	if err != nil {
		return fmt.Errorf("failed to create temp annotations file: %w", err)
	}
	tmpAnnotationsPath := tmpAnnotations.Name()
	tmpAnnotations.Close()
	defer os.Remove(tmpAnnotationsPath)

	// Generate annotations PDF to temp file
	annotationsFile, err := os.Create(tmpAnnotationsPath)
	if err != nil {
		return fmt.Errorf("failed to create annotations file: %w", err)
	}
	if err := p.generateAnnotationsOnly(zip, annotationsFile); err != nil {
		annotationsFile.Close()
		return err
	}
	annotationsFile.Close()

	// Step 2: Write background PDF to temp file
	tmpBackground, err := os.CreateTemp("", "rmfakecloud-background-*.pdf")
	if err != nil {
		return fmt.Errorf("failed to create temp background file: %w", err)
	}
	tmpBackgroundPath := tmpBackground.Name()
	if _, err := tmpBackground.Write(p.backgroundPDF); err != nil {
		tmpBackground.Close()
		os.Remove(tmpBackgroundPath)
		return fmt.Errorf("failed to write background PDF: %w", err)
	}
	tmpBackground.Close()
	defer os.Remove(tmpBackgroundPath)

	// Step 3: Merge background and annotations using pdfcpu
	tmpOutput, err := os.CreateTemp("", "rmfakecloud-merged-*.pdf")
	if err != nil {
		return fmt.Errorf("failed to create temp output file: %w", err)
	}
	tmpOutputPath := tmpOutput.Name()
	tmpOutput.Close()
	defer os.Remove(tmpOutputPath)

	outFile, err := os.Create(tmpOutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Open both PDFs as ReadSeekers
	bgFile, err := os.Open(tmpBackgroundPath)
	if err != nil {
		return fmt.Errorf("failed to open background PDF: %w", err)
	}
	defer bgFile.Close()

	annFile, err := os.Open(tmpAnnotationsPath)
	if err != nil {
		return fmt.Errorf("failed to open annotations PDF: %w", err)
	}
	defer annFile.Close()

	// Merge: background first, then overlay annotations
	conf := model.NewDefaultConfiguration()
	rsc := []io.ReadSeeker{bgFile, annFile}
	if err := api.MergeRaw(rsc, outFile, false, conf); err != nil {
		return fmt.Errorf("failed to merge PDFs: %w", err)
	}
	outFile.Close()

	// Copy merged result to output
	mergedFile, err := os.Open(tmpOutputPath)
	if err != nil {
		return fmt.Errorf("failed to open merged file: %w", err)
	}
	defer mergedFile.Close()

	_, err = io.Copy(output, mergedFile)
	return err
}

func (p *PdfGenerator) drawAnnotations(surface *cairo.Surface, rmData *rm.Rm, scale, pageHeight float64) error {
	surface.Save()
	defer surface.Restore()

	for _, layer := range rmData.Layers {
		for _, line := range layer.Lines {
			if len(line.Points) < 1 {
				continue
			}
			if line.BrushType == rm.Eraser || line.BrushType == rm.EraseArea {
				continue
			}

			if line.BrushType == rm.HighlighterV5 {
				// Draw highlighter as semi-transparent rectangle
				p.drawHighlighter(surface, line, scale, pageHeight)
			} else {
				// Draw regular stroke
				p.drawStroke(surface, line, scale, pageHeight)
			}
		}
	}

	return nil
}

func (p *PdfGenerator) drawHighlighter(surface *cairo.Surface, line rm.Line, scale, pageHeight float64) {
	if len(line.Points) < 2 {
		return
	}

	last := len(line.Points) - 1
	x1, y1 := normalized(line.Points[0], scale)
	x2, _ := normalized(line.Points[last], scale)

	// Highlighter width
	width := scale * 30
	y1 += width / 2

	// Convert Y coordinate (Cairo origin is top-left, PDF is bottom-left)
	y := pageHeight - y1

	// Yellow color with 50% opacity
	surface.SetSourceRGBA(1.0, 1.0, 0.0, 0.5)
	surface.SetLineWidth(width)
	surface.SetLineCap(cairo.LINE_CAP_BUTT)

	surface.MoveTo(x1, y)
	surface.LineTo(x2, y)
	surface.Stroke()
}

func (p *PdfGenerator) drawStroke(surface *cairo.Surface, line rm.Line, scale, pageHeight float64) {
	if len(line.Points) < 1 {
		return
	}

	// Set stroke color
	var r, g, b float64
	switch line.BrushColor {
	case rm.Black:
		r, g, b = 0.0, 0.0, 0.0
	case rm.White:
		r, g, b = 1.0, 1.0, 1.0
	case rm.Grey:
		r, g, b = 0.5, 0.5, 0.5
	default:
		r, g, b = 0.0, 0.0, 0.0
	}
	surface.SetSourceRGB(r, g, b)

	// Set stroke width
	// Formula from original: line.BrushSize*6.0 - 10.8
	strokeWidth := float64(line.BrushSize)*6.0 - 10.8
	if strokeWidth < 0.5 {
		strokeWidth = 0.5
	}
	surface.SetLineWidth(strokeWidth)

	// Set line cap
	surface.SetLineCap(cairo.LINE_CAP_ROUND)
	surface.SetLineJoin(cairo.LINE_JOIN_ROUND)

	// Draw path
	for i, point := range line.Points {
		x, y := normalized(point, scale)
		// Convert Y coordinate
		y = pageHeight - y

		if i == 0 {
			surface.MoveTo(x, y)
		} else {
			surface.LineTo(x, y)
		}
	}

	surface.Stroke()
}

func (p *PdfGenerator) drawPageNumber(surface *cairo.Surface, pageNum int, pageWidth, pageHeight float64) {
	surface.Save()
	defer surface.Restore()

	surface.SelectFontFace("sans-serif", cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_NORMAL)
	surface.SetFontSize(8.0)
	surface.SetSourceRGB(0, 0, 0)

	text := fmt.Sprintf("%d", pageNum)
	surface.MoveTo(pageWidth-20, pageHeight-10)
	surface.ShowText(text)
}

func (p *PdfGenerator) initBackgroundPages(r io.ReadSeeker) error {
	if r != nil {
		// Read the PDF into memory
		pdfBytes, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("failed to read background PDF: %w", err)
		}

		// Check if PDF is encrypted and handle with pdfcpu
		rs := bytes.NewReader(pdfBytes)
		ctx, err := api.ReadContext(rs, model.NewDefaultConfiguration())
		if err != nil {
			return fmt.Errorf("failed to read PDF: %w", err)
		}

		// Check if encrypted by checking if Encrypt field exists
		if ctx.XRefTable.Encrypt != nil {
			logrus.Info("PDF is encrypted - pdfcpu will handle decryption")
			// pdfcpu's ReadContext already handles decryption with empty password
		}

		p.backgroundPDF = pdfBytes
		p.template = false
		return nil
	}

	logrus.Info("template")
	p.template = true
	return nil
}
