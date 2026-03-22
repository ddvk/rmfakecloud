// Package rmdecode decodes reMarkable page files (.rm / "lines" format).
//
// Version 3 and 5 use the linear format implemented in github.com/juruen/rmapi/encoding/rm.
// Version 6 uses a tagged block / scene tree format; use DecodeV6Summary with Python
// and the rmscene library, or see scripts/rmdecode_v6_summary.py.
//
// References:
//   - https://plasma.ninja/blog/devices/remarkable/binary/format/2017/12/26/reMarkable-lines-file-format.html
//   - https://github.com/chemag/maxio/blob/master/version6.md
//   - https://github.com/ricklupton/rmscene
package rmdecode

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/juruen/rmapi/encoding/rm"
)

// HeaderLen is the fixed size of the ASCII header prefix (v3/v5/v6).
const HeaderLen = 43

const headerPrefix = "reMarkable .lines file, version="

// ErrShortFile is returned when data is shorter than HeaderLen.
var ErrShortFile = errors.New("rm file too short for header")

// ParseVersion reads the version digit(s) from the first 43 bytes of a .rm file.
func ParseVersion(data []byte) (int, error) {
	if len(data) < HeaderLen {
		return 0, ErrShortFile
	}
	s := string(data[:HeaderLen])
	if !strings.HasPrefix(s, headerPrefix) {
		return 0, fmt.Errorf("not a reMarkable .lines file (expected header %q)", headerPrefix)
	}
	rest := strings.TrimPrefix(s, headerPrefix)
	i := 0
	for i < len(rest) && rest[i] >= '0' && rest[i] <= '9' {
		i++
	}
	if i == 0 {
		return 0, fmt.Errorf("no version number in header")
	}
	return strconv.Atoi(rest[:i])
}

// DecodeLegacy parses v3 or v5 .rm bytes into rmapi's Rm struct.
func DecodeLegacy(data []byte) (*rm.Rm, error) {
	v, err := ParseVersion(data)
	if err != nil {
		return nil, err
	}
	if v != 3 && v != 5 {
		return nil, fmt.Errorf("DecodeLegacy: expected version 3 or 5, got %d (use rmscene for v6)", v)
	}
	var page rm.Rm
	if err := page.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return &page, nil
}
