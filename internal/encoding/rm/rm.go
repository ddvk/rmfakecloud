// Package rm provides primitives for encoding and decoding
// the .rm format which is a proprietary format created by
// Remarkable to store the data of a drawing made with the device.
//
// Axel Huebl has made a great job of understanding this binary format and
// has written an excellent blog post that helped a lot for writting this package.
// https://plasma.ninja/blog/devices/remarkable/binary/format/2017/12/26/reMarkable-lines-file-format.html
// As well, he has its own implementation of this decoder in C++ at this repository.
// https://github.com/ax3l/lines-are-beautiful
//
// To mention that the format has since evolve to a new version labeled as v3 in the
// header. This implementation is targeting this new version.
//
// As Ben Johnson says, "In the Go standard library, we use the term encoding
// and marshaling for two separate but related ideas. An encoder in Go is an object
// that applies structure to a stream of bytes while marshaling refers
// to applying structure to bounded, in-memory bytes."
// https://medium.com/go-walkthrough/go-walkthrough-encoding-package-bc5e912232d
//
// We will follow this convention and refer to marshaling for this encoder/decoder
// because we want to transform a .rm binary into a bounded in-memory representation
// of a .rm file.
//
// To try to be as idiomatic as possible, this package implements the two following interfaces
// of the default encoding package (https://golang.org/pkg/encoding/).
//   - BinaryMarshaler
//   - BinaryUnmarshaler
//
// The scope of this package is defined as just the encoding/decoding of the .rm format.
// It will only deal with bytes and not files (one must take care of unzipping the archive
// taken from the device, extracting and providing the content of .rm file as bytes).
//
// This package won't be used for retrieving metadata or attached PDF, ePub files.
package rm

import (
	"fmt"
	"strings"
)

// Version defines the version number of a remarkable note.
type Version int

const (
	V3 Version = iota
	V5
)

// Header starting a .rm binary file. This can help recognizing a .rm file.
const (
	HeaderV3  = "reMarkable .lines file, version=3          "
	HeaderV5  = "reMarkable .lines file, version=5          "
	HeaderLen = 43
)

// Width and Height of the device in pixels.
const (
	Width  int = 1404
	Height int = 1872
)

// BrushColor defines the 3 colors of the brush.
type BrushColor uint32

// Mapping of the three colors.
const (
	Black BrushColor = 0
	Grey  BrushColor = 1
	White BrushColor = 2
)

// BrushType respresents the type of brush.
//
// The different types of brush are explained here:
// https://blog.remarkable.com/how-to-find-your-perfect-writing-instrument-for-notetaking-on-remarkable-f53c8faeab77
type BrushType uint32

// Mappings for brush types.
const (
	BallPoint   BrushType = 2
	Marker      BrushType = 3
	Fineliner   BrushType = 4
	SharpPencil BrushType = 7
	TiltPencil  BrushType = 1
	Brush       BrushType = 0
	Highlighter BrushType = 5
	Eraser      BrushType = 6
	EraseArea   BrushType = 8

	// v5 brings new brush type IDs
	BallPointV5   BrushType = 15
	MarkerV5      BrushType = 16
	FinelinerV5   BrushType = 17
	SharpPencilV5 BrushType = 13
	TiltPencilV5  BrushType = 14
	BrushV5       BrushType = 12
	HighlighterV5 BrushType = 18
)

// BrushSize represents the base brush sizes.
type BrushSize float32

// 3 different brush sizes are noticed.
const (
	Small  BrushSize = 1.875
	Medium BrushSize = 2.0
	Large  BrushSize = 2.125
)

// A Rm represents an entire .rm file
// and is composed of layers.
type Rm struct {
	Version Version
	Layers  []Layer
}

// A Layer contains lines.
type Layer struct {
	Lines []Line
}

// A Line is composed of points.
type Line struct {
	BrushType  BrushType
	BrushColor BrushColor
	Padding    uint32
	Unknown    float32
	BrushSize  BrushSize
	Points     []Point
}

// A Point has coordinates.
type Point struct {
	X         float32
	Y         float32
	Speed     float32
	Direction float32
	Width     float32
	Pressure  float32
}

// New helps creating an empty Rm page.
// By mashaling an empty Rm page and exporting it
// to the device, we should generate an empty page
// as if it were created using the device itself.
// TODO
func New() *Rm {
	return &Rm{}
}

// String implements the fmt.Stringer interface
// The aim is to create a textual representation of a page as in the following image.
// https://plasma.ninja/blog/assets/reMarkable/2017_12_21_reMarkableAll.png
// TODO
func (rm Rm) String() string {
	var o strings.Builder

	fmt.Fprintf(&o, "no of layers: %d\n", len(rm.Layers))
	for i, layer := range rm.Layers {
		fmt.Fprintf(&o, "layer %d\n", i)
		fmt.Fprintf(&o, "  nb of lines: %d\n", len(layer.Lines))
		for j, line := range layer.Lines {
			fmt.Fprintf(&o, "  line %d\n", j)
			fmt.Fprintf(&o, "    brush type: %d\n", line.BrushType)
			fmt.Fprintf(&o, "    brush color: %d\n", line.BrushColor)
			fmt.Fprintf(&o, "    padding: %d\n", line.Padding)
			fmt.Fprintf(&o, "    brush size: %f\n", line.BrushSize)
			fmt.Fprintf(&o, "    nb of points: %d\n", len(line.Points))
			for k, point := range line.Points {
				fmt.Fprintf(&o, "    point %d\n", k)
				fmt.Fprintf(&o, "      coords: %f, %f\n", point.X, point.Y)
				fmt.Fprintf(&o, "      speed: %f\n", point.Speed)
				fmt.Fprintf(&o, "      direction: %f\n", point.Direction)
				fmt.Fprintf(&o, "      width: %f\n", point.Width)
				fmt.Fprintf(&o, "      pressure: %f\n", point.Pressure)
			}
		}
	}
	return o.String()
}
