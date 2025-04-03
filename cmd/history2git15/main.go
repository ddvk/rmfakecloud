package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/fs"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
	"github.com/ddvk/rmfakecloud/internal/ui/viewmodel"
)

type BlobRemoteStorage struct {
	directory string
}

func (brs *BlobRemoteStorage) GetRootIndex() (hash string, generation int64, err error) {
	return brs.bs.root.GetRootIndex(brs.uid)
}

func (brs *BlobRemoteStorage) GetReader(hash string) (io.ReadCloser, error) {
	rc, _, _, err := brs.bs.LoadBlob(brs.uid, hash)
	return rc, err
}

func writeEntries(fd *strings.Builder, entries []viewmodel.Entry, currentLevel int) {
	for _, e := range entries {
		for i := currentLevel * 2; i > 0; i-- {
			fd.WriteString(" ")
		}
		if dir, ok := e.(*viewmodel.Directory); ok {
			fd.WriteString(dir.Name + "\n")
			writeEntries(fd, dir.Entries, currentLevel+1)
		} else if doc, ok := e.(*viewmodel.Document); ok {
			fd.WriteString(doc.Name + " @ " + doc.LastModified.Format(time.RFC3339) + "\n")
		}
	}

}

func main() {
	var tail int = 0
	flag.IntVar(&tail, "tail", tail, "Maximum number of history to take into account (0 = all)")

	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Println(config.EnvVars())
	}

	flag.Parse()

	cfg := config.FromEnv()

	for _, arg := range flag.Args() {
		if path.Base(arg) == ".root.history" {
			basedirectory := path.Dir(arg)

			historyDirectory := path.Join(basedirectory, "history")
			err := os.Mkdir(historyDirectory, 0755)
			if err != nil {
				log.Fatal("Unable to create history directory:", err)
			}

			cmd := exec.Command("git", "-C", historyDirectory, "init")
			err = cmd.Run()
			if err != nil {
				log.Fatal(err)
			}

			f1, _ := os.Create(path.Join(historyDirectory, "doctree"))
			f1.Close()
			f1, _ = os.Create(path.Join(historyDirectory, "tree"))
			f1.Close()

			cmd = exec.Command("git", "-C", historyDirectory, "add", "doctree", "tree")
			err = cmd.Run()
			if err != nil {
				log.Fatal(err)
			}

			cmd = exec.Command("git", "-C", historyDirectory, "commit", "-m", "Initial commit")
			err = cmd.Run()
			if err != nil {
				log.Fatal(err)
			}

			history, err := models.ReadRootHistory(arg)
			if err != nil {
				log.Fatalf("%s: %s", arg, err.Error())
			}

			userdirectory := path.Dir(basedirectory)
			cfg.DataDir = path.Dir(path.Dir(userdirectory))

			filesystem := fs.NewStorage(cfg)
			lbs := storage.NewBlobStorer(filesystem, filesystem)

			if tail != 0 && len(history) > tail {
				history = history[len(history)-tail:]
			}

			for _, h := range history {
				tree, err := h.GetHashTree(lbs.RemoteStorage())
				if err != nil {
					log.Fatalf("%s: %s: %s", arg, h.Hash, err.Error())
				}

				doctree := viewmodel.DocTreeFromHashTree(tree)

				tree.Save(path.Join(historyDirectory, "tree"))

				fd, err := os.Create(path.Join(historyDirectory, "doctree"))
				if err != nil {
					log.Fatalf("%s: %s: %s: %s", arg, h.Hash, h.Date.Format(time.RFC3339), err.Error())
				}

				var b strings.Builder

				b.WriteString("entries\n")
				writeEntries(&b, doctree.Entries, 1)
				b.WriteString("trash\n")
				writeEntries(&b, doctree.Trash, 1)

				strings.NewReader(b.String()).WriteTo(fd)
				fd.Close()

				cmd = exec.Command("git", "-C", historyDirectory, "commit", "--allow-empty", "-am", h.Hash, "--date", h.Date.Format(time.RFC3339))
				err = cmd.Run()
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
