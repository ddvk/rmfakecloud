package models

import (
	"strconv"
	"strings"
)

func NewFileHashEntry(hash, documentId string) *HashEntry {
	return &HashEntry{
		Hash:       hash,
		DocumentID: documentId,
		Type:       FileType,
	}
}

type HashEntry struct {
	Hash       string
	Type       string
	DocumentID string
	Subfiles   int
	Size       int64
}

func (d *HashEntry) Line() string {
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
