package models

import (
	"strconv"
	"strings"
)

// NewHashEntry blah
func NewHashEntry(hash, documentID string, size int64) *HashEntry {
	return &HashEntry{
		Hash:      hash,
		EntryName: documentID,
		Type:      fileType,
		Size:      size,
	}
}

// HashEntry an entry in a doc (.content, .meta the payload etc) with a hash
type HashEntry struct {
	Hash      string
	Type      string
	EntryName string
	Subfiles  int
	Size      int64
}

// IsMetadata if this entry points to a metadata blob
func (h *HashEntry) IsMetadata() bool {
	return strings.HasSuffix(h.EntryName, MetadataFileExt)
}
func (h *HashEntry) IsContent() bool {
	return strings.HasSuffix(h.EntryName, ContentFileExt)
}

// Line a line in the index file
func (h *HashEntry) Line() string {
	var sb strings.Builder
	sb.WriteString(h.Hash)
	sb.WriteRune(delimiter)
	sb.WriteString(fileType)
	sb.WriteRune(delimiter)
	sb.WriteString(h.EntryName)
	sb.WriteRune(delimiter)
	sb.WriteString("0")
	sb.WriteRune(delimiter)
	sb.WriteString(strconv.FormatInt(h.Size, 10))
	return sb.String()
}
