package exporter

import (
	"io"

	"github.com/ddvk/rmfakecloud/internal/archive"
)

// MyArchive but having the payload reader
type MyArchive struct {
	archive.Zip
	PayloadReader io.ReadSeekCloser
	V6PageData    map[int][]byte // Raw v6 .rm data, indexed by page number
}

func (f *MyArchive) Close() {
	if f.PayloadReader != nil {
		f.PayloadReader.Close()
	}
}
