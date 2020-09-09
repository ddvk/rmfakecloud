package webassets

import "net/http"

var Assets http.FileSystem = http.Dir("ui/build/")
