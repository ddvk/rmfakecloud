package models

import (
	"fmt"
	"strings"
)

type FieldReader struct {
	index  int
	fields []string
}

func (fr *FieldReader) HasNext() bool {
	return fr.index < len(fr.fields)
}

func (fr *FieldReader) Next() (string, error) {
	if fr.index >= len(fr.fields) {
		return "", fmt.Errorf("out of bounds")
	}
	res := fr.fields[fr.index]
	fr.index++
	return res, nil
}

func NewFieldReader(line string) FieldReader {
	fld := strings.FieldsFunc(line, func(r rune) bool { return r == Delimiter })

	fr := FieldReader{
		index:  0,
		fields: fld,
	}
	return fr
}
