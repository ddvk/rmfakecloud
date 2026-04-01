package rmdecode

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// EnvLines2SVGBin points to a lines-are-beautiful `lines2svg` executable.
	EnvLines2SVGBin = "RMFAKECLOUD_LINES2SVG_BIN"
)

// RenderV3SVGWithLines2SVG renders a v3 .rm page to SVG via lines-are-beautiful.
// The tool only understands the old v3 layout; newer versions should use other renderers.
func RenderV3SVGWithLines2SVG(data []byte) (string, error) {
	ver, err := ParseVersion(data)
	if err != nil {
		return "", err
	}
	if ver != 3 {
		return "", fmt.Errorf("lines2svg renderer supports v3 only, got v%d", ver)
	}

	bin := strings.TrimSpace(os.Getenv(EnvLines2SVGBin))
	if bin == "" {
		var lookErr error
		bin, lookErr = exec.LookPath("lines2svg")
		if lookErr != nil {
			return "", fmt.Errorf("lines2svg not found (set %s): %w", EnvLines2SVGBin, lookErr)
		}
	}

	tmpDir, err := os.MkdirTemp("", "lines2svg-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	// lines2svg expects a notebook directory with pages as 0.rm, 1.rm, ...
	inPath := filepath.Join(tmpDir, "0.rm")
	if err := os.WriteFile(inPath, data, 0600); err != nil {
		return "", err
	}

	cmd := exec.Command(bin, tmpDir)
	cmd.Dir = tmpDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return "", fmt.Errorf("lines2svg: %w: %s", err, msg)
		}
		return "", fmt.Errorf("lines2svg: %w", err)
	}

	outPath := filepath.Join(tmpDir, "test-0.svg")
	out, err := os.ReadFile(outPath)
	if err != nil {
		return "", fmt.Errorf("lines2svg output not found: %w", err)
	}
	return string(out), nil
}
