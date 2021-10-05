package exporter

import (
	"io"

	"github.com/juruen/rmapi/archive"
	"github.com/juruen/rmapi/log"
)

func init() {
	log.InitLog()
}

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
