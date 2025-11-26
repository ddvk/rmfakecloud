package models

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const schemaVersionV3 = "3"
const schemaVersionV4 = "4"
const schemaVersion = schemaVersionV4
const docType = "80000000"
const fileType = "0"
const delimiter = ':'

func HashEntries(entries []*HashEntry) (string, error) {
	sort.Slice(entries, func(i, j int) bool { return entries[i].EntryName < entries[j].EntryName })
	hasher := sha256.New()
	for _, d := range entries {
		//TODO: back and forth converting
		bh, err := hex.DecodeString(d.Hash)
		if err != nil {
			return "", err
		}
		hasher.Write(bh)
	}
	hash := hasher.Sum(nil)
	hashStr := hex.EncodeToString(hash)
	return hashStr, nil
}

func Hash(r io.Reader) (string, int64, error) {
	hasher := sha256.New()
	w, err := io.Copy(hasher, r)
	if err != nil {
		return "", w, err
	}
	h := hasher.Sum(nil)
	hstr := hex.EncodeToString(h)
	return hstr, w, err
}
func FileHashAndSize(file string) ([]byte, int64, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	hasher := sha256.New()
	io.Copy(hasher, f)
	h := hasher.Sum(nil)
	size, err := f.Seek(0, io.SeekEnd)
	return h, size, err
}

// LoadTree loads a cached tree to avoid parsing all the blobs
func LoadTree(cacheFile string) (*HashTree, error) {
	tree := HashTree{}
	if _, err := os.Stat(cacheFile); err == nil {
		b, err := os.ReadFile(cacheFile)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(b, &tree)
		if err != nil {
			log.Warn("cached tree corrupt, returning empty tree")
			return &HashTree{}, nil
		}
		log.Info("cached tree loaded: ", cacheFile)
	}

	return &tree, nil
}

// Save saves
func (t *HashTree) Save(cacheFile string) error {
	log.Println("Writing cache: ", cacheFile)
	b, err := json.MarshalIndent(t, "", "")
	if err != nil {
		return err
	}
	err = os.WriteFile(cacheFile, b, 0644)
	return err
}

func parseEntry(line string) (*HashEntry, error) {
	entry := HashEntry{}
	rdr := NewFieldReader(line)
	numFields := len(rdr.fields)
	if numFields != 5 {
		return nil, fmt.Errorf("parseEntry: wrong number of fields %d", numFields)
	}
	var err error
	entry.Hash, err = rdr.Next()
	if err != nil {
		return nil, err
	}
	entry.Type, err = rdr.Next()
	if err != nil {
		return nil, err
	}
	entry.EntryName, err = rdr.Next()
	if err != nil {
		return nil, err
	}
	tmp, err := rdr.Next()
	if err != nil {
		return nil, err
	}
	entry.Subfiles, err = strconv.Atoi(tmp)
	if err != nil {
		return nil, fmt.Errorf("cannot read subfiles %s %v", line, err)
	}
	tmp, err = rdr.Next()
	if err != nil {
		return nil, err
	}
	entry.Size, err = strconv.ParseInt(tmp, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("cannot read size %s %v", line, err)
	}
	return &entry, nil
}

func parseSchemaV4SummaryLine(line string) (entriesCount int, totalSize int64, err error) {
	rdr := NewFieldReader(line)
	if len(rdr.fields) != 4 {
		return 0, 0, fmt.Errorf("invalid v4 summary line, expected 4 fields, got %d", len(rdr.fields))
	}
	if _, err = rdr.Next(); err != nil {
		return 0, 0, fmt.Errorf("cannot read field 1: %w", err)
	}
	if _, err = rdr.Next(); err != nil {
		return 0, 0, fmt.Errorf("cannot read field 2: %w", err)
	}
	entriesCountStr, err := rdr.Next()
	if err != nil {
		return 0, 0, fmt.Errorf("cannot read entry count field: %w", err)
	}
	totalSizeStr, err := rdr.Next()
	if err != nil {
		return 0, 0, fmt.Errorf("cannot read total size field: %w", err)
	}

	entriesCount, err = strconv.Atoi(entriesCountStr)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot parse entry count: %w", err)
	}
	totalSize, err = strconv.ParseInt(totalSizeStr, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot parse total size: %w", err)
	}

	return entriesCount, totalSize, nil
}

func parseIndex(r io.Reader) ([]*HashEntry, error) {
	var entries []*HashEntry
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	schema := scanner.Text()

	expectedCount := 0
	count := 0

	switch schema {
	case schemaVersionV4:
		if !scanner.Scan() {
			return nil, fmt.Errorf("expecting v4 summary line after schema version")
		}
		line := scanner.Text()
		var err error
		expectedCount, _, err = parseSchemaV4SummaryLine(line)
		if err != nil {
			return nil, fmt.Errorf("cannot parse v4 summary line: %w", err)
		}
		fallthrough
	case schemaVersionV3:
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				log.Warn("empty line in index file, ignored")
				continue
			}
			count++
			entry, err := parseEntry(line)
			if err != nil {
				return nil, fmt.Errorf("cant parse line '%s', %w", line, err)
			}
			entries = append(entries, entry)
		}
	default:
		return nil, fmt.Errorf("parseInde unknown schema: %s", schema)
	}

	if schema == schemaVersionV4 && count != expectedCount {
		return nil, fmt.Errorf("v4 index entry count mismatch: expected %d, got %d", expectedCount, count)
	}

	return entries, nil
}

// RootIndex reads the root index
func (t *HashTree) RootIndex() (io.ReadCloser, error) {
	pipeReader, pipeWriter := io.Pipe()
	w := bufio.NewWriter(pipeWriter)
	go func() {
		defer pipeWriter.Close()
		version := t.SchemaVersion
		if version == "" {
			version = schemaVersion
		}
		w.WriteString(version)
		w.WriteString("\n")

		if version == schemaVersionV4 {
			totalSize := int64(0)
			for _, d := range t.Docs {
				totalSize += d.Size
			}
			w.WriteString(fileType)
			w.WriteRune(delimiter)
			w.WriteString(".")
			w.WriteRune(delimiter)
			w.WriteString(strconv.Itoa(len(t.Docs)))
			w.WriteRune(delimiter)
			w.WriteString(strconv.FormatInt(totalSize, 10))
			w.WriteString("\n")
		}

		for _, d := range t.Docs {
			w.WriteString(d.Line())
			w.WriteString("\n")
		}
		w.Flush()
	}()

	return pipeReader, nil
}

// HashTree a tree of hashes
type HashTree struct {
	Hash          string
	Generation    int64
	Docs          []*HashDoc
	SchemaVersion string
}

// FindDoc finds a document by its name
func (t *HashTree) FindDoc(documentID string) (*HashDoc, error) {
	//O(n)
	for _, d := range t.Docs {
		if d.EntryName == documentID {
			return d, nil
		}
	}
	return nil, fmt.Errorf("treedoc '%s' not found", documentID)
}

// Remove removes
func (t *HashTree) Remove(documentID string) error {
	docIndex := -1
	for index, d := range t.Docs {
		if d.EntryName == documentID {
			docIndex = index
			break
		}
	}
	if docIndex > -1 {
		log.Infof("Removing %s", documentID)
		length := len(t.Docs) - 1
		t.Docs[docIndex] = t.Docs[length]
		t.Docs = t.Docs[:length]

		t.Rehash()
		return nil
	}
	return fmt.Errorf("%s not found", documentID)
}

// Rehash recalcualte the root hash from all docs
func (t *HashTree) Rehash() error {
	entries := []*HashEntry{}
	for _, e := range t.Docs {
		entries = append(entries, &e.HashEntry)
	}
	hash, err := HashEntries(entries)
	if err != nil {
		return err
	}
	log.Debug("New root hash: ", hash)
	t.Hash = hash
	return nil
}

// Mirror makes the tree look like the storage
func (t *HashTree) Mirror(r RemoteStorage) (changed bool, err error) {
	rootHash, gen, err := r.GetRootIndex()
	if err != nil {
		return
	}
	log.Debug("mirror: got root ", rootHash, gen)
	if rootHash == "" {
		log.Warn("empty root hash, empty cloud?")
		t.Docs = nil
		t.Generation = gen
		return
	}

	if rootHash == t.Hash {
		if gen != t.Generation {
			t.Generation = gen
			return true, nil
		}
		return
	}
	log.Debug("remote root hash is different")

	rdr, err := r.GetReader(rootHash)
	if err != nil {
		return
	}
	defer rdr.Close()

	entries, err := parseIndex(rdr)
	if err != nil {
		return
	}

	head := make([]*HashDoc, 0)
	current := make(map[string]*HashDoc)
	new := make(map[string]*HashEntry)
	for _, e := range entries {
		new[e.EntryName] = e
	}
	//current documents
	for _, doc := range t.Docs {
		if entry, ok := new[doc.HashEntry.EntryName]; ok {
			//hash different update
			if entry.Hash != doc.Hash {
				log.Debug("doc updated: ", doc.EntryName)
				doc.Mirror(entry, r)
			}
			if doc.Deleted {
				continue
			}
			head = append(head, doc)
			current[doc.EntryName] = doc
		}
	}

	//find new entries
	for k, newEntry := range new {
		if _, ok := current[k]; !ok {
			doc := &HashDoc{}
			log.Info("doc new: ", k)
			doc.Mirror(newEntry, r)

			if doc.Deleted {
				continue
			}
			head = append(head, doc)
		}
	}
	sort.Slice(head, func(i, j int) bool { return head[i].EntryName < head[j].EntryName })
	t.Docs = head
	t.Generation = gen
	t.Hash = rootHash
	return true, nil
}

// BuildTree from remote storage
func BuildTree(provider RemoteStorage) (*HashTree, error) {
	rootHash, gen, err := provider.GetRootIndex()

	if err != nil {
		return nil, err
	}

	rootIndex, err := provider.GetReader(rootHash)
	if err != nil {
		return nil, err
	}
	defer rootIndex.Close()

	return buildTreeFromFile(provider, rootIndex, rootHash, gen)
}

func buildTreeFromFile(provider RemoteStorage, rootFile io.Reader, rootHash string, gen int64) (*HashTree, error) {
	entries, _ := parseIndex(rootFile)

	tree := HashTree{}

	tree.Hash = rootHash
	tree.Generation = gen

	for _, e := range entries {
		r, _ := provider.GetReader(e.Hash)
		defer r.Close()

		doc := &HashDoc{}
		doc.HashEntry = *e

		items, _ := parseIndex(r)
		doc.Files = items
		for _, i := range items {
			doc.ReadMetadata(i, provider)
		}

		if doc.Deleted {
			continue
		}
		tree.Docs = append(tree.Docs, doc)

	}

	return &tree, nil
}
