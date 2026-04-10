package models

import (
	"fmt"
	"io"
	"strings"
)

func CreateContent(fileType string) string {
	fileType = strings.TrimPrefix(fileType, ".")
	str :=
		`
{
	"dummyDocument": false,
	"extraMetadata": {
		"LastPen": "Finelinerv2",
		"LastTool": "Finelinerv2",
		"ThicknessScale": "",
		"LastFinelinerv2Size": "1"
	},
	"fileType": "%s",
	"fontName": "",
	"lastOpenedPage": 0,
	"lineHeight": -1,
	"margins": 180,
	"orientation": "portrait",
	"pageCount": 0,
	"pages": [],
	"textScale": 1,
	"transform": {
		"m11": 1,
		"m12": 0,
		"m13": 0,
		"m21": 0,
		"m22": 1,
		"m23": 0,
		"m31": 0,
		"m32": 0,
		"m33": 1
	}
}
`
	return fmt.Sprintf(str, fileType)
}

func ExtractID(_ io.Reader) (string, error) {
	return "", nil
}
