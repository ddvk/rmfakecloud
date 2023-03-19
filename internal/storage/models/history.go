package models

import (
	"bufio"
	"io"
	"os"
	"strings"
	"time"
)

type RootHistory struct {
	Generation int64
	Date       time.Time
	Hash       string
}

func ReadRootHistory(filename string) (history []*RootHistory, err error) {
	fd, err1 := os.Open(filename)
	if err1 != nil {
		return nil, err1
	}
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	var i int64 = 0
	for scanner.Scan() {
		line := scanner.Text()

		tokens := strings.Split(line, " ")
		if len(tokens) != 2 {
			continue
		}

		var item RootHistory
		item.Date, err = time.Parse(time.RFC3339, tokens[0])
		if err != nil {
			return
		}

		item.Hash = tokens[1]
		item.Generation = i

		history = append(history, &item)
		i += 1
	}

	return
}

func (h *RootHistory) OpenIndex(r RemoteStorage) (io.ReadCloser, error) {
	return r.GetReader(h.Hash)
}

func (h *RootHistory) GetHashTree(r RemoteStorage) (t *HashTree, err error) {
	t = &HashTree{
		Hash:       h.Hash,
		Generation: h.Generation,
	}

	rootFile, err := h.OpenIndex(r)
	if err != nil {
		return
	}

	return buildTreeFromFile(r, rootFile, h.Hash, h.Generation)
}
