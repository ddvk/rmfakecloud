// +build ignore

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

// embeds the ui/build into WebAssets
func main() {
	fmt.Println("generating assets")
	fs := http.Dir("../../ui/build")
	err := vfsgen.Generate(fs, vfsgen.Options{
		PackageName:  "webassets",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatal(err)
	}

}
