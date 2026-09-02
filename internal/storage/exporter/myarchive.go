package exporter

import (
	"io"

	"github.com/juruen/rmapi/archive"
	"github.com/juruen/rmapi/log"
)

// rmapi's logging stuff
func init() {
	log.InitLog()
}

// V6Page holds the raw bytes of a .rm page that the bundled
// v3/v5 parser cannot handle (reMarkable firmware 3.x writes v6).
type V6Page struct {
	PageID string
	Data   []byte
}

// MyArchive but having the payload reader
type MyArchive struct {
	archive.Zip
	PayloadReader io.ReadSeekCloser
	// V6Pages contains raw .rm v6 pages in document order
	V6Pages []V6Page
}

// HasV6Pages reports whether the document contains v6 pages
func (f *MyArchive) HasV6Pages() bool {
	return len(f.V6Pages) > 0
}

func (f *MyArchive) Close() {
	if f.PayloadReader != nil {
		f.PayloadReader.Close()
	}
}
