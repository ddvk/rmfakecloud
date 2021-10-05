package models

import (
	"strconv"
	"strings"
)

// NewFileHashEntry blah
func NewFileHashEntry(hash, documentID string) *HashEntry {
	return &HashEntry{
		Hash:      hash,
		EntryName: documentID,
		Type:      fileType,
	}
}

// HashEntry an entry with a hash
type HashEntry struct {
	Hash      string
	Type      string
	EntryName string
	Subfiles  int
	Size      int64
}

// Line a line in the index file
func (d *HashEntry) Line() string {
	var sb strings.Builder
	sb.WriteString(d.Hash)
	sb.WriteRune(delimiter)
	sb.WriteString(fileType)
	sb.WriteRune(delimiter)
	sb.WriteString(d.EntryName)
	sb.WriteRune(delimiter)
	sb.WriteString("0")
	sb.WriteRune(delimiter)
	sb.WriteString(strconv.FormatInt(d.Size, 10))
	return sb.String()
}
