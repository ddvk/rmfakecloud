package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ddvk/rmfakecloud/internal/app"
	"github.com/ddvk/rmfakecloud/internal/cli"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/storage/fs"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var version string

func main() {
	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Println("Version: ", version)
		fmt.Printf(`
Commands:
	setuser		create users / reset passwords
	listusers	lists available users
`)
		fmt.Println(config.EnvVars())
	}

	flag.Parse()
	fmt.Fprintln(os.Stderr, "run with -h for all available env variables")

	cfg := config.FromEnv()
	//cli
	cmd := cli.New(cfg)
	if cmd.Handle(os.Args) {
		return
	}

	logger := logrus.StandardLogger()
	logger.SetFormatter(&logrus.TextFormatter{})

	if lvl, err := log.ParseLevel(os.Getenv(config.EnvLogLevel)); err == nil {
		fmt.Println("Log level:", lvl)
		logger.SetLevel(lvl)
	}

	log.Println("Version: ", version)
	// configs
	log.Println("Documents will be saved in:", cfg.DataDir)
	log.Println("Url the device should use:", cfg.StorageURL)
	log.Println("Listening on port:", cfg.Port)

	fsStorage := fs.NewStorage(cfg)
	usrs, err := fsStorage.GetUsers()

	if err != nil {
		log.Warn(err)
	}

	if len(usrs) == 0 {
		log.Warn("No users found, the first login will create a user")
		cfg.CreateFirstUser = true

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
