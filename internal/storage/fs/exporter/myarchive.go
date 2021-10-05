package exporter

import (
	"io"

	"github.com/juruen/rmapi/archive"
)

/// Archive but having the payload in reader
type MyArchive struct {
	archive.Zip
	PayloadReader io.ReadSeekCloser
}

func (f *MyArchive) Close() {
	if f.PayloadReader != nil {
		f.PayloadReader.Close()
	}

}
