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
	"github.com/gin-gonic/gin"
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

	logger := log.StandardLogger()
	logger.SetFormatter(&log.TextFormatter{})

	if lvl, err := log.ParseLevel(os.Getenv(config.EnvLogLevel)); err == nil {
		fmt.Println("Log level:", lvl)
		logger.SetLevel(lvl)
	}

	log.Println("Version: ", version)
	// configs
	log.Println("Documents will be saved in:", cfg.DataDir)
	log.Println("Url the device should use:", cfg.StorageURL)
	log.Println("Listening on port:", cfg.Port)

	gin.DefaultWriter = logger.Writer()

	a := app.NewApp(cfg)
	go a.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Stopping the service...")
	a.Stop()
	log.Println("Stopped")
}
