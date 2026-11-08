package rmdecode

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/pdfraster"
)

// RenderV6PNGWithRMC renders a v6 .rm page to PNG via rmc → PDF → out-of-process raster.
// Prefers production-grade stroke rendering (rmc) over the crude rmscene polyline script.
func RenderV6PNGWithRMC(data []byte) ([]byte, error) {
	ver, err := ParseVersion(data)
	if err != nil {
		return nil, err
	}
	if ver != 6 {
		return nil, fmt.Errorf("rmc rendering expects v6 .rm, got v%d", ver)
	}

	tmpDir, err := os.MkdirTemp("", "rmc-png-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	inPath := filepath.Join(tmpDir, "page.rm")
	pdfPath := filepath.Join(tmpDir, "page.pdf")
	if err := os.WriteFile(inPath, data, 0600); err != nil {
		return nil, err
	}

	var errs []string
	if err := runRMCBinaryFmt(inPath, pdfPath, "pdf"); err != nil {
		errs = append(errs, err.Error())
		if err2 := runRMCModuleFmt(inPath, pdfPath, "pdf"); err2 != nil {
			errs = append(errs, err2.Error())
			return nil, fmt.Errorf("rmc pdf failed: %s", strings.Join(errs, " | "))
		}
	}

	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, err
	}
	return pdfraster.RenderPage(pdfBytes, 1)
}

func runRMCBinaryFmt(inPath, outPath, format string) error {
	rmcBin := strings.TrimSpace(os.Getenv(EnvRMCBin))
	if rmcBin == "" {
		var err error
		rmcBin, err = exec.LookPath("rmc")
		if err != nil {
			return fmt.Errorf("rmc binary not found (set %s or install rmc): %w", EnvRMCBin, err)
		}
	}
	cmd := exec.Command(rmcBin, "-t", format, "-o", outPath, inPath)
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

func runRMCModuleFmt(inPath, outPath, format string) error {
	rmcSrc := strings.TrimSpace(os.Getenv(EnvRMCSrc))
	if rmcSrc == "" {
		return fmt.Errorf("module mode disabled (set %s)", EnvRMCSrc)
	}
	python, err := exec.LookPath("python3")
	if err != nil {
		return fmt.Errorf("python3 not found: %w", err)
	}
	cmd := exec.Command(python, "-m", "rmc.cli", "-t", format, "-o", outPath, inPath)
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
