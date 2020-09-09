package main

import (
	"github.com/ddvk/rmfakecloud/internal/app"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/storage/fs"
	"log"
	"os"
)

var version string

func main() {
	log.SetOutput(os.Stdout)

	cfg := config.FromEnv()

	log.Println("Version: ", version)
	// configs
	log.Println("Documents will be saved in:", cfg.DataDir)
	log.Println("Url the device should use:", cfg.StorageURL)
	log.Println("Port", cfg.Port)

	fsStorage := &fs.Storage{
		Cfg: *cfg,
	}

	a := app.NewApp(cfg, fsStorage, fsStorage)
	a.Start()

	//todo: ctrl-c handler
}
