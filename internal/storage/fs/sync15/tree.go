package sync15

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"

	"github.com/ddvk/rmfakecloud/internal/storage"
	log "github.com/sirupsen/logrus"
)

const SchemaVersion = "3"
const DocType = "80000000"
const FileType = "0"
const Delimiter = ':'

func HashEntries(entries []*Entry) (string, error) {
	sort.Slice(entries, func(i, j int) bool { return entries[i].DocumentID < entries[j].DocumentID })
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

func Hash(r io.Reader) (string, error) {
	hasher := sha256.New()
	_, err := io.Copy(hasher, r)
	if err != nil {
		return "", err
	}
	h := hasher.Sum(nil)
	hstr := hex.EncodeToString(h)
	return hstr, err
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
	size, err := f.Seek(0, os.SEEK_CUR)
	return h, size, err
}
func LoadTree(cacheFile string) (*HashTree, error) {
	tree := HashTree{}
	if _, err := os.Stat(cacheFile); err == nil {
		b, err := ioutil.ReadFile(cacheFile)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(b, &tree)
		if err != nil {
			log.Println("cache corrupt")
			return nil, err
		}
		log.Println("Cache loaded: ", cacheFile)
	}

	return &tree, nil
}

func (tree *HashTree) Save(cacheFile string) error {
	log.Println("Writing cache: ", cacheFile)
	b, err := json.MarshalIndent(tree, "", "")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(cacheFile, b, 0644)
	return err
}

func parseEntry(line string) (*Entry, error) {
	entry := Entry{}
	rdr := NewFieldReader(line)
	numFields := len(rdr.fields)
	if numFields != 5 {
		return nil, fmt.Errorf("wrong number of fields %d", numFields)

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
	entry.DocumentID, err = rdr.Next()
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

func parseIndex(f io.Reader) ([]*Entry, error) {
	var entries []*Entry
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	schema := scanner.Text()

	if schema != SchemaVersion {
		return nil, errors.New("wrong schema")
	}
	for scanner.Scan() {
		line := scanner.Text()
		entry, err := parseEntry(line)
		if err != nil {
			return nil, fmt.Errorf("cant parse line '%s', %w", line, err)
		}

		entries = append(entries, entry)
	}
	return entries, nil
}

func (t *HashTree) RootIndex() (io.ReadCloser, error) {
	pipeReader, pipeWriter := io.Pipe()
	w := bufio.NewWriter(pipeWriter)
	go func() {
		defer pipeWriter.Close()
		w.WriteString(SchemaVersion)
		w.WriteString("\n")
		for _, d := range t.Docs {
			w.WriteString(d.Line())
			w.WriteString("\n")
		}
		w.Flush()
	}()

	return pipeReader, nil
}

type HashTree struct {
	Hash       string
	Generation int64
	Docs       []*BlobDoc
}

func (t *HashTree) FindDoc(id string) (*BlobDoc, error) {
	//O(n)
	for _, d := range t.Docs {
		if d.DocumentID == id {
			return d, nil
		}
	}
	return nil, fmt.Errorf("doc %s not found", id)
}

func (t *HashTree) Remove(id string) error {
	docIndex := -1
	for index, d := range t.Docs {
		if d.DocumentID == id {
			docIndex = index
			break
		}
	}
	if docIndex > -1 {
		log.Info("Removing %s", id)
		length := len(t.Docs) - 1
		t.Docs[docIndex] = t.Docs[length]
		t.Docs = t.Docs[:length]

		t.Rehash()
		return nil
	}
	return fmt.Errorf("%s not found", id)
}

func (t *HashTree) Rehash() error {
	entries := []*Entry{}
	for _, e := range t.Docs {
		entries = append(entries, &e.Entry)
	}
	hash, err := HashEntries(entries)
	if err != nil {
		return err
	}
	log.Println("New root hash: ", hash)
	t.Hash = hash
	return nil
}

/// Mirror makes the tree look like the storage
func (t *HashTree) Mirror(r storage.RemoteStorage) (changed bool, err error) {
	rootHash, gen, err := r.GetRootIndex()
	if err != nil {
		return
	}
	if rootHash == "" && gen == 0 {
		log.Println("Empty cloud")
		t.Docs = nil
		t.Generation = 0
		return
	}

	if rootHash == t.Hash {
		if gen != t.Generation {
			t.Generation = gen
			return true, nil

		}
		return
	}
	log.Printf("remote root hash different")

	rdr, err := r.GetReader(rootHash)
	if err != nil {
		return
	}
	defer rdr.Close()

	entries, err := parseIndex(rdr)
	if err != nil {
		return
	}

	head := make([]*BlobDoc, 0)
	current := make(map[string]*BlobDoc)
	new := make(map[string]*Entry)
	for _, e := range entries {
		new[e.DocumentID] = e
	}
	//current documents
	for _, doc := range t.Docs {
		if entry, ok := new[doc.Entry.DocumentID]; ok {
			//hash different update
			if entry.Hash != doc.Hash {
				log.Println("doc updated: " + doc.DocumentID)
				doc.Mirror(entry, r)
			}
			head = append(head, doc)
			current[doc.DocumentID] = doc
		}

	}

	//find new entries
	for k, newEntry := range new {
		if _, ok := current[k]; !ok {
			doc := &BlobDoc{}
			log.Println("doc new: " + k)
			doc.Mirror(newEntry, r)
			head = append(head, doc)
		}
	}
	sort.Slice(head, func(i, j int) bool { return head[i].DocumentID < head[j].DocumentID })
	t.Docs = head
	t.Generation = gen
	t.Hash = rootHash
	return true, nil
}

func BuildTree(provider storage.RemoteStorage) (*HashTree, error) {
	tree := HashTree{}

	rootHash, gen, err := provider.GetRootIndex()

	if err != nil {
		return nil, err
	}
	tree.Hash = rootHash
	tree.Generation = gen

	rootIndex, err := provider.GetReader(rootHash)
	if err != nil {
		return nil, err
	}

	defer rootIndex.Close()
	entries, _ := parseIndex(rootIndex)

	for _, e := range entries {
		f, _ := provider.GetReader(e.Hash)
		defer f.Close()

		doc := &BlobDoc{}
		doc.Entry = *e
		tree.Docs = append(tree.Docs, doc)

		items, _ := parseIndex(f)
		doc.Files = items
		for _, i := range items {
			doc.ReadMetadata(i, provider)
		}
	}

	return &tree, nil

}
