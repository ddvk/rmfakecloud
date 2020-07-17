package main

import (
	"log"
	"os"

	"github.com/ddvk/rmfakecloud/internal/app"
	"github.com/ddvk/rmfakecloud/internal/config"
)

func main() {
	log.SetOutput(os.Stdout)

	config := config.ConfigFromEnv()
	a := app.CreateApp(config)
	a.Start()
}
