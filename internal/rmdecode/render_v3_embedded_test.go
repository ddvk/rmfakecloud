package rmdecode

import (
	"strings"
	"testing"

	"github.com/juruen/rmapi/encoding/rm"
)

func TestRenderV3SVGOverlayEmbedded_Nil(t *testing.T) {
	_, err := RenderV3SVGOverlayEmbedded(nil)
	if err == nil {
		t.Fatal("expected error for nil page")
	}
}

func TestRenderV3SVGOverlayEmbedded_EraserMask(t *testing.T) {
	page := &rm.Rm{
		Layers: []rm.Layer{
			{
				Lines: []rm.Line{
					{
						BrushType:  rm.TiltPencil,
						BrushSize:  20,
						BrushColor: rm.Black,
						Points: []rm.Point{
							{X: 10, Y: 10},
							{X: 20, Y: 20},
						},
					},
					{
						BrushType: rm.EraseArea,
						BrushSize: 20,
						Points: []rm.Point{
							{X: 11, Y: 11},
							{X: 12, Y: 12},
						},
					},
					{
						BrushType:  rm.TiltPencil,
						BrushSize:  20,
						BrushColor: rm.Black,
						Points: []rm.Point{
							{X: 30, Y: 30},
							{X: 40, Y: 40},
						},
					},
				},
			},
		},
	}
	s, err := RenderV3SVGOverlayEmbedded(page)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "<mask ") {
		t.Fatalf("expected eraser mask in svg: %s", s)
	}
	if !strings.Contains(s, "layer-0") {
		t.Fatalf("expected layer id in svg: %s", s)
	}
}
