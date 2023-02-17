package models

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/common"
	log "github.com/sirupsen/logrus"
)

// HashDoc a document in a hash tree
type HashDoc struct {
	Files []*HashEntry
	HashEntry

	//extra fields that are serialized
	MetadataFile
	PayloadType string
	PayloadSize int64
}

func NewHashDocWithMeta(documentID string, meta MetadataFile) *HashDoc {
	return &HashDoc{
		MetadataFile: meta,
		HashEntry: HashEntry{
			EntryName: documentID,
		},
	}

}
func NewHashDoc(name, documentID string, docType common.EntryType) *HashDoc {
	return &HashDoc{
		MetadataFile: MetadataFile{
			DocumentName:   name,
			CollectionType: docType,
		},
		HashEntry: HashEntry{
			EntryName: documentID,
		},
	}
}

// Rehash re-calculates the hash
func (d *HashDoc) Rehash() error {
	hash, err := HashEntries(d.Files)
	if err != nil {
		return err
	}
	log.Debug("rehash: ", d.DocumentName, " new doc hash: ", hash)
	d.Hash = hash
	return nil
}

func (d *HashDoc) MetadataReader() (hash string, reader io.Reader, err error) {
	jsn, err := json.Marshal(d.MetadataFile)
	if err != nil {
		return
	}
	sha := sha256.New()
	sha.Write(jsn)
	hash = hex.EncodeToString(sha.Sum(nil))
	log.Info("new hash: ", hash)
	reader = bytes.NewReader(jsn)
	found := false
	for _, f := range d.Files {
		if f.IsMetadata() {
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

// AddFile adds an entry
func (d *HashDoc) AddFile(e *HashEntry) error {
	d.Files = append(d.Files, e)
	return d.Rehash()
}

// Add  adds a doc to the tree
func (t *HashTree) Add(d *HashDoc) error {
	if len(d.Files) == 0 {
		return errors.New("no files")
	}
	t.Docs = append(t.Docs, d)
	return t.Rehash()
}

// IndexReader reader of the document index
func (d *HashDoc) IndexReader() (io.ReadCloser, error) {
	if len(d.Files) == 0 {
		return nil, errors.New("no files")
	}
	pipeReader, pipeWriter := io.Pipe()
	w := bufio.NewWriter(pipeWriter)
	go func() {
		defer pipeWriter.Close()
		w.WriteString(schemaVersion)
		w.WriteString("\n")
		for _, d := range d.Files {
			w.WriteString(d.Line())
			w.WriteString("\n")
		}
		w.Flush()
	}()

	return pipeReader, nil
}

func (d *HashDoc) readMetadata(fileEntry string, r RemoteStorage) error {
	log.Println("Reading metadata: " + d.EntryName)

	metadata := MetadataFile{}

	meta, err := r.GetReader(fileEntry)
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
		log.Printf("cannot read metadata %s %v", fileEntry, err)
	}
	log.Println("name from metadata: ", metadata.DocumentName)
	d.MetadataFile = metadata

	return nil
}

func (d *HashDoc) readContent(hash string, r RemoteStorage) error {
	log.Println("Reading content: " + d.EntryName)

	contentFile := ContentFile{}

	meta, err := r.GetReader(hash)
	if err != nil {
		return err
	}
	defer meta.Close()
	contentBytes, err := ioutil.ReadAll(meta)
	if err != nil {
		return err
	}
	err = json.Unmarshal(contentBytes, &contentFile)
	if err != nil {
		log.Printf("cannot read content %s %v", hash, err)
	}
	d.PayloadType = contentFile.FileType

	if len(contentFile.SizeInBytes) > 0 {
		d.Size, err = strconv.ParseInt(contentFile.SizeInBytes, 10, 64)
		if err != nil {
			log.Printf("invalid SizeInBytes: %s %s %v", contentFile.SizeInBytes, hash, err)
		}
	}

	return nil
}

// ReadMetadata tries to read the metadata blob if this entry is metadata
func (d *HashDoc) ReadMetadata(fileEntry *HashEntry, r RemoteStorage) error {
	if fileEntry.IsMetadata() {
		return d.readMetadata(fileEntry.Hash, r)
	}
	if fileEntry.IsContent() {
		return d.readContent(fileEntry.Hash, r)
	}
	return nil

}

// Line index line
func (d *HashDoc) Line() string {
	var sb strings.Builder
	if d.Hash == "" {
		log.Print("missing hash for: ", d.EntryName)
	}
	sb.WriteString(d.Hash)
	sb.WriteRune(delimiter)
	sb.WriteString(docType)
	sb.WriteRune(delimiter)
	sb.WriteString(d.EntryName)
	sb.WriteRune(delimiter)

	numFilesStr := strconv.Itoa(len(d.Files))
	sb.WriteString(numFilesStr)
	sb.WriteRune(delimiter)
	sb.WriteString("0")
	return sb.String()
}

// Mirror mirror on the wall
func (d *HashDoc) Mirror(e *HashEntry, r RemoteStorage) error {
	d.HashEntry = *e
	entryIndex, err := r.GetReader(e.Hash)
	if err != nil {
		return err
	}
	defer entryIndex.Close()
	entries, err := parseIndex(entryIndex)
	if err != nil {
		return err
	}

	head := make([]*HashEntry, 0)
	current := make(map[string]*HashEntry)
	new := make(map[string]*HashEntry)

	for _, e := range entries {
		new[e.EntryName] = e
	}

	//updated and existing
	for _, currentEntry := range d.Files {
		if newEntry, ok := new[currentEntry.EntryName]; ok {
			if newEntry.Hash != currentEntry.Hash {
				err = d.ReadMetadata(newEntry, r)
				if err != nil {
					return err
				}
				currentEntry.Hash = newEntry.Hash
			}
			head = append(head, currentEntry)
			current[currentEntry.EntryName] = currentEntry
		}
	}

	//add missing
	for k, newEntry := range new {
		if _, ok := current[k]; !ok {
			err = d.ReadMetadata(newEntry, r)
			if err != nil {
				return err
			}
			head = append(head, newEntry)
		}
	}
	sort.Slice(head, func(i, j int) bool { return head[i].EntryName < head[j].EntryName })
	d.Files = head
	return nil

}
