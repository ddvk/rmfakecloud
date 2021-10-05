package exporter

import (
	"errors"
	"fmt"
	"io"

	"github.com/juruen/rmapi/encoding/rm"
	"github.com/sirupsen/logrus"
	"github.com/unidoc/unipdf/v3/annotator"
	"github.com/unidoc/unipdf/v3/contentstream"
	"github.com/unidoc/unipdf/v3/contentstream/draw"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/creator"
	pdf "github.com/unidoc/unipdf/v3/model"
)

const (
	DeviceWidth  = 1404
	DeviceHeight = 1872
)

var rmPageSize = creator.PageSize{445, 594}

type PdfGenerator struct {
	options   PdfGeneratorOptions
	pdfReader *pdf.PdfReader
	template  bool
}

type PdfGeneratorOptions struct {
	AddPageNumbers  bool
	AllPages        bool
	AnnotationsOnly bool //export the annotations without the background/pdf
}

func normalized(p1 rm.Point, ratioX float64) (float64, float64) {
	return float64(p1.X) * ratioX, float64(p1.Y) * ratioX
}

func (p *PdfGenerator) Generate(zip *MyArchive, output io.Writer, options PdfGeneratorOptions) (err error) {

	p.options = options

	if len(zip.Pages) == 0 {
		if zip.PayloadReader != nil {
			_, err := io.Copy(output, zip.PayloadReader)
			return err
		} else {
			return errors.New("the document has no pages")
		}
	}

	if err = p.initBackgroundPages(zip.PayloadReader); err != nil {
		return err
	}

	c := creator.New()
	if p.template {
		// use the standard page size
		c.SetPageSize(rmPageSize)
	}

	if p.pdfReader != nil && p.options.AllPages {
		logrus.Info("generating all pages")
		outlines := p.pdfReader.GetOutlineTree()
		c.SetOutlineTree(outlines)
	}

	for i, pageAnnotations := range zip.Pages {
		hasContent := pageAnnotations.Data != nil

		// do not add a page when there are no annotations
		if !p.options.AllPages && !hasContent {
			continue
		}

		page, err := p.addBackgroundPage(c, i+1)
		if err != nil {
			return err
		}

		ratio := c.Height() / c.Width()

		var scale float64
		if ratio < 1.33 {
			scale = c.Width() / DeviceWidth
		} else {
			scale = c.Height() / DeviceHeight
		}
		if page == nil {
			logrus.Fatal("page is null")
		}

		if err != nil {
			return err
		}
		if !hasContent {
			continue
		}

		contentCreator := contentstream.NewContentCreator()
		contentCreator.Add_q()

		for _, layer := range pageAnnotations.Data.Layers {
			for _, line := range layer.Lines {
				if len(line.Points) < 1 {
					continue
				}
				if line.BrushType == rm.Eraser {
					continue
				}

				if line.BrushType == rm.HighlighterV5 {
					last := len(line.Points) - 1
					x1, y1 := normalized(line.Points[0], scale)
					x2, _ := normalized(line.Points[last], scale)
					// make horizontal lines only, use y1
					width := scale * 30
					y1 += width / 2

					lineDef := annotator.LineAnnotationDef{X1: x1 - 1, Y1: c.Height() - y1, X2: x2, Y2: c.Height() - y1}
					lineDef.LineColor = pdf.NewPdfColorDeviceRGB(1.0, 1.0, 0.0) //yellow
					lineDef.Opacity = 0.5
					lineDef.LineWidth = width
					ann, err := annotator.CreateLineAnnotation(lineDef)
					if err != nil {
						return err
					}
					page.AddAnnotation(ann)
				} else {
					path := draw.NewPath()
					for i := 0; i < len(line.Points); i++ {
						x1, y1 := normalized(line.Points[i], scale)
						path = path.AppendPoint(draw.NewPoint(x1, c.Height()-y1))
					}

					contentCreator.Add_w(float64(line.BrushSize / 10))

					switch line.BrushColor {
					case rm.Black:
						contentCreator.Add_rg(1.0, 1.0, 1.0)
					case rm.White:
						contentCreator.Add_rg(0.0, 0.0, 0.0)
					case rm.Grey:
						contentCreator.Add_rg(0.8, 0.8, 0.8)
					}

					//TODO: use bezier
					draw.DrawPathWithCreator(path, contentCreator)

					contentCreator.Add_S()
				}
			}
		}
		contentCreator.Add_Q()
		drawingOperations := contentCreator.Operations().String()
		pageContentStreams, err := page.GetAllContentStreams()
		if err != nil {
			return err
		}
		//hack: wrap the page content in a context to prevent transformation matrix misalignment
		wrapper := []string{"q", pageContentStreams, "Q", drawingOperations}
		page.SetContentStreams(wrapper, core.NewFlateEncoder())
	}

	return c.Write(output)
}

func (p *PdfGenerator) initBackgroundPages(r io.ReadSeeker) error {
	if r != nil {
		logrus.Info("new seker")
		pdfReader, err := pdf.NewPdfReader(r)
		if err != nil {
			return err
		}

		encrypted, err := pdfReader.IsEncrypted()
		if err != nil {
			return nil
		}
		if encrypted {
			valid, err := pdfReader.Decrypt([]byte(""))
			if err != nil {
				return err
			}
			if !valid {
				return fmt.Errorf("cannot decrypt")
			}

		}

		p.pdfReader = pdfReader
		p.template = false
		return nil
	}

	logrus.Info("template")
	p.template = true
	return nil
}

func (p *PdfGenerator) addBackgroundPage(c *creator.Creator, pageNum int) (*pdf.PdfPage, error) {
	var page *pdf.PdfPage

	if !p.template && !p.options.AnnotationsOnly {
		tmpPage, err := p.pdfReader.GetPage(pageNum)
		if err != nil {
			return nil, err
		}
		mbox, err := tmpPage.GetMediaBox()
		if err != nil {
			return nil, err
		}

		// TODO: adjust the page if cropped
		pageHeight := mbox.Ury - mbox.Lly
		pageWidth := mbox.Urx - mbox.Llx
		// use the pdf's page size
		c.SetPageSize(creator.PageSize{pageWidth, pageHeight})
		c.AddPage(tmpPage)
		page = tmpPage
	} else {
		page = c.NewPage()
	}

	if p.options.AddPageNumbers {
		c.DrawFooter(func(block *creator.Block, args creator.FooterFunctionArgs) {
			p := c.NewParagraph(fmt.Sprintf("%d", args.PageNum))
			p.SetFontSize(8)
			w := block.Width() - 20
			h := block.Height() - 10
			p.SetPos(w, h)
			block.Draw(p)
		})
	}
	return page, nil
}
