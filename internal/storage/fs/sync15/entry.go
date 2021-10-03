package sync15

import (
	"strconv"
	"strings"
)

func NewFileEntry(hash, documentId string) *Entry {
	return &Entry{
		Hash:       hash,
		DocumentID: documentId,
		Type:       FileType,
	}
}

type Entry struct {
	Hash       string
	Type       string
	DocumentID string
	Subfiles   int
	Size       int64
}

func (d *Entry) Line() string {
	var sb strings.Builder
	sb.WriteString(d.Hash)
	sb.WriteRune(Delimiter)
	sb.WriteString(FileType)
	sb.WriteRune(Delimiter)
	sb.WriteString(d.DocumentID)
	sb.WriteRune(Delimiter)
	sb.WriteString("0")
	sb.WriteRune(Delimiter)
	sb.WriteString(strconv.FormatInt(d.Size, 10))
	return sb.String()
}
