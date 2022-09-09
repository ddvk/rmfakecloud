package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/ddvk/rmfakecloud/internal/app"
	"github.com/ddvk/rmfakecloud/internal/cli"
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var version string

func configureLogging() io.Closer {
	logfileName := os.Getenv(config.EnvLogFile)
	loglevel := os.Getenv(config.EnvLogLevel)
	logformat := os.Getenv(config.EnvLogFormat)

	logger := logrus.StandardLogger()
	var formatter logrus.Formatter
	switch logformat {
	case "json":
		formatter = &logrus.JSONFormatter{}
	default:
		formatter = &logrus.TextFormatter{}
	}
	logger.SetFormatter(formatter)

	var logFile io.Closer
	if logfileName != "" {
		logFile, err := os.OpenFile(logfileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot open log file '%s' for writing\n", logfileName)
		} else {
			hook := lfshook.NewHook(logFile, &logrus.TextFormatter{DisableColors: true})
			logger.Hooks.Add(hook)
		}
	}

	if lvl, err := logrus.ParseLevel(loglevel); err == nil {
		fmt.Println("Log level:", lvl)
		logger.SetLevel(lvl)
	}
	gin.DefaultWriter = logger.Writer()
	return logFile
}

func main() {
	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Println("Version: ", version)
		fmt.Println(cli.Usage())
		fmt.Println(config.EnvVars())
	}

	flag.Parse()

	logging := configureLogging()
	if logging != nil {
		defer logging.Close()
	}

	cfg := config.FromEnv()

	//cli
	cmd := cli.New(cfg)
	if cmd.Handle(os.Args) {
		return
	}

	cfg.Verify()

	logrus.Info("Version: ", version)

	a := app.NewApp(cfg)
	go a.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logrus.Println("Stopping the service...")
	a.Stop()
	logrus.Println("Stopped")
}
