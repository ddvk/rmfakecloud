package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ddvk/rmfakecloud/internal/app"
	"github.com/ddvk/rmfakecloud/internal/cli"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var version string

func main() {
	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Println("Version: ", version)
		fmt.Printf(`
Commands:
	setuser		create users / reset passwords
	listusers	list available users
`)
		fmt.Println(config.EnvVars())
	}

	flag.Parse()

	cfg := config.FromEnv()

	//cli
	cmd := cli.New(cfg)
	if cmd.Handle(os.Args) {
		return
	}

	fmt.Fprintln(os.Stderr, "run with -h for all available env variables")
	cfg.Verify()

	logger := logrus.StandardLogger()
	logger.SetFormatter(&logrus.TextFormatter{})

	if cfg.LogFile != "" {
		var file, err = os.OpenFile(cfg.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot open log file '%s' for writing\n", cfg.LogFile)
		} else {
			defer file.Close()
			hook := lfshook.NewHook(file, &logrus.TextFormatter{DisableColors: true})
			logger.Hooks.Add(hook)
		}
	}

	if lvl, err := logrus.ParseLevel(os.Getenv(config.EnvLogLevel)); err == nil {
		fmt.Println("Log level:", lvl)
		logger.SetLevel(lvl)
	}

	logrus.Info("Version: ", version)
	// configs
	logrus.Info("STORAGE_URL, The device should use this URL: ", cfg.StorageURL)
	logrus.Info("Documents will be saved in:", cfg.DataDir)
	logrus.Info("Listening on port:", cfg.Port)

	gin.DefaultWriter = logger.Writer()

	// invalidate user tokens on restart
	cfg.TokenVersion = int(time.Now().Unix())
	a := app.NewApp(cfg)
	go a.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logrus.Println("Stopping the service...")
	a.Stop()
	logrus.Println("Stopped")
}
