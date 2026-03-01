package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/storage/fs"
	"github.com/ddvk/rmfakecloud/internal/storage/models"
)

func main() {
	var userId string
	var rootHash string

	flag.StringVar(&userId, "user", userId, "The user in which file need to be recovered")
	flag.StringVar(&rootHash, "root-hash", rootHash, "The historical root index in which looking for old documents")
	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Println(config.EnvVars())
	}

	flag.Parse()

	cfg := config.FromEnv()
	filesystem := fs.NewStorage(cfg)

	lbs := storage.NewBlobStorer(filesystem, filesystem)

	h := models.RootHistory{Hash: rootHash}
	oldtree, err := h.GetHashTree(lbs.RemoteStorage(userId))
	if err != nil {
		log.Fatalf("%s: %s", h.Hash, err.Error())
	}

	hash, gen, _ := lbs.RemoteStorage(userId).GetRootIndex()
	h = models.RootHistory{Hash: hash, Generation: gen}
	curtree, err := h.GetHashTree(lbs.RemoteStorage(userId))

	for _, doc := range oldtree.Docs {
		concerned := false
		for _, filename := range flag.Args() {
			if filename == doc.MetadataFile.DocumentName {
				concerned = true
				break
			}
		}

		if concerned {
			err = storage.UpdateTree(curtree, lbs, userId, func(t *models.HashTree) error {
				return t.Add(doc)
			})
			if err != nil {
				log.Fatal("Unable to updateTree:", err)
			}

			fmt.Println("File ", doc.DocumentName, " reverted.")
		}
	}
}
