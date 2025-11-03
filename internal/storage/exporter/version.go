package exporter

import (
	"bytes"
	"fmt"
	"io"
)

const (
	// HeaderV3 is the header for version 3 .rm files
	HeaderV3 = "reMarkable .lines file, version=3"
	// HeaderV5 is the header for version 5 .rm files
	HeaderV5 = "reMarkable .lines file, version=5"
	// HeaderV6 is the header for version 6 .rm files (43 bytes with padding)
	HeaderV6 = "reMarkable .lines file, version=6"
	// HeaderSizeV6 is the exact size of v6 header
	HeaderSizeV6 = 43
)

// RmVersion represents the version of a .rm file
type RmVersion int

const (
	// VersionUnknown indicates the version could not be determined
	VersionUnknown RmVersion = 0
	// VersionV3 indicates version 3 format
	VersionV3 RmVersion = 3
	// VersionV5 indicates version 5 format
	VersionV5 RmVersion = 5
	// VersionV6 indicates version 6 format
	VersionV6 RmVersion = 6
)

// String returns the string representation of the version
func (v RmVersion) String() string {
	switch v {
	case VersionV3:
		return "v3"
	case VersionV5:
		return "v5"
	case VersionV6:
		return "v6"
	default:
		return "unknown"
	}
}

// DetectRmVersion reads the header from .rm file to determine version
// The reader should be positioned at the start of the file
func DetectRmVersion(reader io.Reader) (RmVersion, error) {
	// Read enough bytes to detect any version
	// v6 header is 43 bytes, v3/v5 are shorter
	headerBuf := make([]byte, HeaderSizeV6)
	n, err := io.ReadAtLeast(reader, headerBuf, len(HeaderV3))
	if err != nil && err != io.ErrUnexpectedEOF {
		return VersionUnknown, fmt.Errorf("failed to read header: %w", err)
	}

	header := headerBuf[:n]

	// Check v6 first (most recent, longest header)
	if bytes.HasPrefix(header, []byte(HeaderV6)) {
		return VersionV6, nil
	}

	// Check v5
	if bytes.Contains(header, []byte("version=5")) {
		return VersionV5, nil
	}

	// Check v3
	if bytes.Contains(header, []byte("version=3")) {
		return VersionV3, nil
	}

	return VersionUnknown, fmt.Errorf("unknown .rm file format")
}

// DetectRmVersionFromBytes detects version from byte slice
func DetectRmVersionFromBytes(data []byte) (RmVersion, error) {
	return DetectRmVersion(bytes.NewReader(data))
}

// DetectArchiveVersion detects the .rm file version from an archive
// by examining the first page data
func DetectArchiveVersion(arch *MyArchive) (RmVersion, error) {
	if len(arch.Pages) == 0 {
		return VersionUnknown, fmt.Errorf("no pages in archive")
	}

	// Try to marshal first page and detect from header
	if arch.Pages[0].Data != nil {
		data, err := arch.Pages[0].Data.MarshalBinary()
		if err != nil {
			return VersionUnknown, fmt.Errorf("failed to marshal page data: %w", err)
		}
		return DetectRmVersion(bytes.NewReader(data))
	}

	return VersionUnknown, fmt.Errorf("no page data available")
}
