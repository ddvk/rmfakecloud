// +build ignore
//go:generate go run assets_generate.go
package main

import (
	"github.com/ddvk/rmfakecloud/internal/webassets"
	"github.com/shurcooL/vfsgen"
	"log"
)

func main() {
	err := vfsgen.Generate(webassets.Assets, vfsgen.Options{
		PackageName:  "webassets",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatal(err)
	}

}
