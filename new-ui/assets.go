//go:build !ci

package rmfakecloud

import "embed"

//go:embed dist/*
var Assets embed.FS
