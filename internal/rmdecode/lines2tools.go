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
	// EnvLines2SVGBin points to a lines-are-beautiful lines2svg executable.
	EnvLines2SVGBin = "RMFAKECLOUD_LINES2SVG_BIN"
	// EnvLines2PNGBin points to a lines-are-beautiful lines2png executable.
	EnvLines2PNGBin = "RMFAKECLOUD_LINES2PNG_BIN"
)

func renderV3WithLinesTool(data []byte, envVar, fallbackBin, outName string) ([]byte, error) {
	ver, err := ParseVersion(data)
	if err != nil {
		return nil, err
	}
	if ver != 3 {
		return nil, fmt.Errorf("%s renderer supports v3 only, got v%d", fallbackBin, ver)
	}

	bin := strings.TrimSpace(os.Getenv(envVar))
	if bin == "" {
		bin, err = exec.LookPath(fallbackBin)
		if err != nil {
			return nil, fmt.Errorf("%s not found (set %s): %w", fallbackBin, envVar, err)
		}
	}

	tmpDir, err := os.MkdirTemp("", fallbackBin+"-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	// lines-are-beautiful expects notebook pages as 0.rm, 1.rm...
	inPath := filepath.Join(tmpDir, "0.rm")
	if err := os.WriteFile(inPath, data, 0600); err != nil {
		return nil, err
	}

	cmd := exec.Command(bin, tmpDir)
	cmd.Dir = tmpDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return nil, fmt.Errorf("%s: %w: %s", fallbackBin, err, msg)
		}
		return nil, fmt.Errorf("%s: %w", fallbackBin, err)
	}

	outPath := filepath.Join(tmpDir, outName)
	out, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("%s output not found: %w", fallbackBin, err)
	}
	return out, nil
}

func RenderV3SVGWithLines2SVG(data []byte) (string, error) {
	b, err := renderV3WithLinesTool(data, EnvLines2SVGBin, "lines2svg", "test-0.svg")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func RenderV3PNGWithLines2PNG(data []byte) ([]byte, error) {
	return renderV3WithLinesTool(data, EnvLines2PNGBin, "lines2png", "test.png")
}
