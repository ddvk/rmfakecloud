// +build ignore

package main

import (
	"github.com/ddvk/rmfakecloud/internal/webassets"
	"github.com/shurcooL/vfsgen"
	"log"
)

// embeds the ui/build into WebAssets
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
