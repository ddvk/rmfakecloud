//go:build !ci

package rmfakecloud

import "embed"

//go:embed build/*
var Assets embed.FS
