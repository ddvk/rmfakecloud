package fs

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/google/uuid"
)

func createZipContent(ext string) string {
	ext = strings.TrimPrefix(ext, ".")
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
	return fmt.Sprintf(str, ext)
}

func extractID(r io.Reader) (string, error) {
	return "", nil
}

// CreateDocument creates a new document
func (fs *Storage) CreateDocument(uid, filename string, stream io.ReadCloser) (doc *storage.Document, err error) {
	ext := path.Ext(filename)
	switch ext {
	case ".pdf":
	case ".epub":
	default:
		return nil, errors.New("unsupported extension: " + ext)
	}

	var docid string

	var isZip = false
	if ext == zipExtension {
		docid, err = extractID(stream)
		isZip = true
	} else {
		docid = uuid.New().String()
	}
	//create zip from pdf
	zipfile := fs.getPathFromUser(uid, docid+zipExtension)
	file, err := os.Create(zipfile)
	if err != nil {
		return
	}
	defer file.Close()

	if !isZip {
		w := zip.NewWriter(file)
		defer w.Close()

		documentPath := docid + ext
		var entry io.Writer
		entry, err = w.Create(documentPath)
		if err != nil {
			return
		}

		_, err = io.Copy(entry, stream)
		if err != nil {
			return
		}

		entry, err = w.Create(docid + ".pagedata")
		if err != nil {
			return
		}
		entry.Write([]byte{})

		entry, err = w.Create(docid + ".content")
		if err != nil {
			return
		}

		content := createZipContent(ext)
		entry.Write([]byte(content))
	} else {
		_, err = io.Copy(file, stream)
		if err != nil {
			return
		}
	}

	//create metadata
	name := strings.TrimSuffix(filename, ext)
	doc1 := createMedatadata(name, docid)

	jsn, err := json.Marshal(doc1)
	if err != nil {
		return
	}

	doc = &storage.Document{
		ID:     docid,
		Type:   doc1.Type,
		Parent: "",
		Name:   name,
	}
	//save metadata
	metafilePath := fs.getPathFromUser(uid, docid+metadataExtension)
	err = ioutil.WriteFile(metafilePath, jsn, 0600)
	return
}
