package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ddvk/rmfakecloud/internal/app"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/storage/fs"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var version string

func main() {
	//logrus := logrus.New()
	logger := logrus.StandardLogger()
	logger.SetFormatter(&logrus.TextFormatter{})

	if lvl, err := log.ParseLevel(os.Getenv("LOGLEVEL")); err == nil {
		fmt.Println("Log level:", lvl)
		logger.SetLevel(lvl)
	}
	cfg := config.FromEnv()

	log.Println("Version: ", version)
	// configs
	log.Println("Documents will be saved in:", cfg.DataDir)
	log.Println("Url the device should use:", cfg.StorageURL)
	log.Println("Port", cfg.Port)

	fsStorage := &fs.Storage{
		Cfg: *cfg,
	}
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	gin.DefaultWriter = logger.Writer()
	a := app.NewApp(cfg, fsStorage, fsStorage, fsStorage)
	go a.Start()
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Stopping the service...")
	a.Stop()
	log.Println("Stopped")
}
