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
	// EnvRMCBin points to an rmc executable (optional). If unset, PATH is used.
	EnvRMCBin = "RMFAKECLOUD_RMC_BIN"
	// EnvRMCSrc points to the "src" directory of the rmc repo checkout (optional).
	// Example: /home/user/Downloads/rmc-main/src
	EnvRMCSrc = "RMFAKECLOUD_RMC_SRC"
	// EnvRMSSceneSrc optionally points to rmscene's Python source directory.
	// Example: /path/to/rmscene/src
	EnvRMSSceneSrc = "RMFAKECLOUD_RMSCENE_SRC"
)

// RenderV6SVGWithRMC renders a v6 .rm page to SVG by invoking rmc.
// It prefers an rmc binary, and falls back to "python3 -m rmc.cli" if EnvRMCSrc is set.
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
			return fmt.Errorf("rmc binary not found (set %s or install rmc): %w", EnvRMCBin, err)
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
	rmcSrc := strings.TrimSpace(os.Getenv(EnvRMCSrc))
	if rmcSrc == "" {
		return fmt.Errorf("module mode disabled (set %s)", EnvRMCSrc)
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
