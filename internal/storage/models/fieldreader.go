package models

import (
	"fmt"
	"strings"
)

// FieldReader iterates over delimted fields
type FieldReader struct {
	index  int
	fields []string
}

// HasNext are there more fields
func (fr *FieldReader) HasNext() bool {
	return fr.index < len(fr.fields)
}

// Next read the next field
func (fr *FieldReader) Next() (string, error) {
	if fr.index >= len(fr.fields) {
		return "", fmt.Errorf("out of bounds")
	}
	res := fr.fields[fr.index]
	fr.index++
	return res, nil
}

// NewFieldReader reader from string line
func NewFieldReader(line string) FieldReader {
	fld := strings.FieldsFunc(line, func(r rune) bool { return r == delimiter })

	fr := FieldReader{
		index:  0,
		fields: fld,
	}
	return fr
}
