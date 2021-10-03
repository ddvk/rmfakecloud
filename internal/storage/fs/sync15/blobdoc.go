package sync15

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/ddvk/rmfakecloud/internal/storage"
)

type BlobDoc struct {
	Files []*Entry
	Entry
	model.MetadataFile
}

func NewBlobDoc(name, documentId, colType string) *BlobDoc {
	return &BlobDoc{
		MetadataFile: model.MetadataFile{
			DocName:        name,
			CollectionType: colType,
		},
		Entry: Entry{
			DocumentID: documentId,
		},
	}

}

func (d *BlobDoc) Rehash() error {

	hash, err := HashEntries(d.Files)
	if err != nil {
		return err
	}
	log.Println("New doc hash: ", hash)
	d.Hash = hash
	return nil
}

func (d *BlobDoc) MetadataHashAndReader() (hash string, reader io.Reader, err error) {
	jsn, err := json.Marshal(d.MetadataFile)
	if err != nil {
		return
	}
	sha := sha256.New()
	sha.Write(jsn)
	hash = hex.EncodeToString(sha.Sum(nil))
	log.Println("new hash", hash)
	reader = bytes.NewReader(jsn)
	found := false
	for _, f := range d.Files {
		if strings.HasSuffix(f.DocumentID, storage.MetadataFileExt) {
			f.Hash = hash
			found = true
			break
		}
	}
	if !found {
		err = errors.New("metadata not found")
	}

	return
}

func (d *BlobDoc) AddFile(e *Entry) error {
	d.Files = append(d.Files, e)
	return d.Rehash()
}

func (t *HashTree) Add(d *BlobDoc) error {
	if len(d.Files) == 0 {
		return errors.New("no files")
	}
	t.Docs = append(t.Docs, d)
	return t.Rehash()
}

func (t *BlobDoc) IndexReader() (io.ReadCloser, error) {
	if len(t.Files) == 0 {
		return nil, errors.New("no files")
	}
	pipeReader, pipeWriter := io.Pipe()
	w := bufio.NewWriter(pipeWriter)
	go func() {
		defer pipeWriter.Close()
		w.WriteString(SchemaVersion)
		w.WriteString("\n")
		for _, d := range t.Files {
			w.WriteString(d.Line())
			w.WriteString("\n")
		}
		w.Flush()
	}()

	return pipeReader, nil
}

// Extract the documentname from metadata blob
func (doc *BlobDoc) ReadMetadata(fileEntry *Entry, r RemoteStorage) error {
	if strings.HasSuffix(fileEntry.DocumentID, ".metadata") {
		log.Println("Reading metadata: " + doc.DocumentID)

		metadata := model.MetadataFile{}

		meta, err := r.GetReader(fileEntry.Hash)
		if err != nil {
			return err
		}
		defer meta.Close()
		content, err := ioutil.ReadAll(meta)
		if err != nil {
			return err
		}
		err = json.Unmarshal(content, &metadata)
		if err != nil {
			log.Printf("cannot read metadata %s %v", fileEntry.DocumentID, err)
		}
		log.Println("name from metadata: ", metadata.DocName)
		doc.MetadataFile = metadata
	}

	return nil
}

func (d *BlobDoc) Line() string {
	var sb strings.Builder
	if d.Hash == "" {
		log.Print("missing hash for: ", d.DocumentID)
	}
	sb.WriteString(d.Hash)
	sb.WriteRune(Delimiter)
	sb.WriteString(DocType)
	sb.WriteRune(Delimiter)
	sb.WriteString(d.DocumentID)
	sb.WriteRune(Delimiter)

	numFilesStr := strconv.Itoa(len(d.Files))
	sb.WriteString(numFilesStr)
	sb.WriteRune(Delimiter)
	sb.WriteString("0")
	return sb.String()
}

func (doc *BlobDoc) Mirror(e *Entry, r RemoteStorage) error {
	doc.Entry = *e
	entryIndex, err := r.GetReader(e.Hash)
	if err != nil {
		return err
	}
	defer entryIndex.Close()
	entries, err := parseIndex(entryIndex)
	if err != nil {
		return err
	}

	head := make([]*Entry, 0)
	current := make(map[string]*Entry)
	new := make(map[string]*Entry)

	for _, e := range entries {
		new[e.DocumentID] = e
	}

	//updated and existing
	for _, currentEntry := range doc.Files {
		if newEntry, ok := new[currentEntry.DocumentID]; ok {
			if newEntry.Hash != currentEntry.Hash {
				err = doc.ReadMetadata(newEntry, r)
				if err != nil {
					return err
				}
				currentEntry.Hash = newEntry.Hash
			}
			head = append(head, currentEntry)
			current[currentEntry.DocumentID] = currentEntry
		}
	}

	//add missing
	for k, newEntry := range new {
		if _, ok := current[k]; !ok {
			err = doc.ReadMetadata(newEntry, r)
			if err != nil {
				return err
			}
			head = append(head, newEntry)
		}
	}
	sort.Slice(head, func(i, j int) bool { return head[i].DocumentID < head[j].DocumentID })
	doc.Files = head
	return nil

}
