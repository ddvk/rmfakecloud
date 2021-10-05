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

// MyArchive but having the payload reader
type MyArchive struct {
	archive.Zip
	PayloadReader io.ReadSeekCloser
}

func (f *MyArchive) Close() {
	if f.PayloadReader != nil {
		f.PayloadReader.Close()
	}
}
