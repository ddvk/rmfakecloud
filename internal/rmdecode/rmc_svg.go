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
	// EnvRMCBin points to an rmc executable (optional). If unset, PATH then vendored module is used.
	EnvRMCBin = "RMFAKECLOUD_RMC_BIN"
	// EnvRMCSrc optionally overrides the vendored third_party/rmc/src directory.
	EnvRMCSrc = "RMFAKECLOUD_RMC_SRC"
	// EnvRMSSceneSrc optionally overrides vendored third_party/rmscene/src.
	EnvRMSSceneSrc = "RMFAKECLOUD_RMSCENE_SRC"
)

// RenderV6SVGWithRMC renders a v6 .rm page to SVG by invoking rmc.
// Order: rmc binary (PATH/env) → python3 -m rmc.cli with vendored (or overridden) sources.
func RenderV6SVGWithRMC(data []byte) (string, error) {
	ver, err := ParseVersion(data)
	if err != nil {
		return "", err
	}
	if ver != 6 {
		return "", fmt.Errorf("rmc rendering expects v6 .rm, got v%d", ver)
	}

	tmpDir, err := os.MkdirTemp("", "rmc-svg-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	inPath := filepath.Join(tmpDir, "page.rm")
	outPath := filepath.Join(tmpDir, "page.svg")
	if err := os.WriteFile(inPath, data, 0600); err != nil {
		return "", err
	}

	var errs []string
	if err := runRMCBinary(inPath, outPath); err == nil {
		return readSVG(outPath)
	} else {
		errs = append(errs, err.Error())
	}
	if err := runRMCModule(inPath, outPath); err == nil {
		return readSVG(outPath)
	} else {
		errs = append(errs, err.Error())
	}

	return "", fmt.Errorf("rmc svg failed: %s", strings.Join(errs, " | "))
}

func runRMCBinary(inPath, outPath string) error {
	rmcBin := strings.TrimSpace(os.Getenv(EnvRMCBin))
	if rmcBin == "" {
		var err error
		rmcBin, err = exec.LookPath("rmc")
		if err != nil {
			return fmt.Errorf("rmc binary not found: %w", err)
		}
	}
	cmd := exec.Command(rmcBin, "-t", "svg", "-o", outPath, inPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return fmt.Errorf("rmc binary: %w: %s", err, msg)
		}
		return fmt.Errorf("rmc binary: %w", err)
	}
	return nil
}

func runRMCModule(inPath, outPath string) error {
	rmcSrc := EffectiveRMCSrc()
	if rmcSrc == "" {
		return fmt.Errorf("no rmc source (vendored third_party/rmc missing; set %s)", EnvRMCSrc)
	}
	python, err := exec.LookPath("python3")
	if err != nil {
		return fmt.Errorf("python3 not found: %w", err)
	}
	cmd := exec.Command(python, "-m", "rmc.cli", "-t", "svg", "-o", outPath, inPath)
	if py := buildPythonPathForV6(rmcSrc); py != "" {
		cmd.Env = append(os.Environ(), "PYTHONPATH="+py)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return fmt.Errorf("python -m rmc.cli: %w: %s", err, msg)
		}
		return fmt.Errorf("python -m rmc.cli: %w", err)
	}
	return nil
}

func readSVG(outPath string) (string, error) {
	b, err := os.ReadFile(outPath)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// RMCAvailable reports whether rmc can be invoked (binary, override, or vendored source).
func RMCAvailable() bool {
	if strings.TrimSpace(os.Getenv(EnvRMCBin)) != "" {
		return true
	}
	if _, err := exec.LookPath("rmc"); err == nil {
		return true
	}
	return EffectiveRMCSrc() != ""
}
