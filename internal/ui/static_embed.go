package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/common"
)

//go:embed static/*
var staticEmbed embed.FS

func (app *ReactAppWrapper) initStatic() {
	sub, err := fs.Sub(staticEmbed, "static")
	if err != nil {
		panic("static embed missing: " + err.Error())
	}
	app.fs = common.NewLastModifiedFS(http.FS(sub), time.Now())
	app.prefix = "/assets"
}

func (app *ReactAppWrapper) readStatic(name string) ([]byte, error) {
	name = strings.TrimPrefix(name, "/")
	return staticEmbed.ReadFile(path.Join("static", name))
}
