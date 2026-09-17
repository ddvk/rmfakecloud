package rmdecode

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/pdfraster"
)

// rmc SVG uses points (72/226 × screen units) with X origin at page center.
const (
	rmcScreenDPI   = 226.0
	rmcScale       = 72.0 / rmcScreenDPI
	rmcPageWidthPt = float64(PageWidthPt) * rmcScale
	rmcPageHeightPt = float64(PageHeightPt) * rmcScale
	rmcXMinPt      = -float64(PageWidthPt) / 2 * rmcScale
)

var (
	reSVGOpen = regexp.MustCompile(`(?is)<svg\b[^>]*>`)
)

// RenderV6PNGWithRMC renders a v6 .rm page to a fixed 1404×1872 device PNG.
// SVG is produced via rmc (with optional inserted page images), then rasterized.
// rmc's default SVG page is the stroke bounding box; we force the device
// viewport (centered X, Y from 0) before rasterizing so PDF composites align.
func RenderV6PNGWithRMC(data []byte, images map[string][]byte) ([]byte, error) {
	ver, err := ParseVersion(data)
	if err != nil {
		return nil, err
	}
	if ver != 6 {
		return nil, fmt.Errorf("rmc rendering expects v6 .rm, got v%d", ver)
	}

	svg, err := RenderV6SVGWithRMC(data, images)
	if err != nil {
		return nil, err
	}
	svg = ForceDeviceViewportSVG(svg)

	tmpDir, err := os.MkdirTemp("", "rmc-png-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	svgPath := filepath.Join(tmpDir, "page.svg")
	pngPath := filepath.Join(tmpDir, "page.png")
	if err := os.WriteFile(svgPath, []byte(svg), 0600); err != nil {
		return nil, err
	}

	if err := rasterSVGToPNG(svgPath, pngPath, PageWidthPt, PageHeightPt); err != nil {
		// Fallback: SVG → PDF (inkscape) → pdfraster at device width.
		pdfPath := filepath.Join(tmpDir, "page.pdf")
		if err2 := svgToPDFInkscape(svgPath, pdfPath); err2 != nil {
			return nil, fmt.Errorf("raster svg: %v; svg→pdf: %v", err, err2)
		}
		pdfBytes, err2 := os.ReadFile(pdfPath)
		if err2 != nil {
			return nil, err2
		}
		return pdfraster.RenderPage(pdfBytes, 1)
	}
	return os.ReadFile(pngPath)
}

// ForceDeviceViewportSVG rewrites the root <svg> to the reMarkable device
// viewport used for PDF/notebook compositing (1404×1872 screen units).
func ForceDeviceViewportSVG(svg string) string {
	viewBox := fmt.Sprintf("%.6f 0 %.6f %.6f", rmcXMinPt, rmcPageWidthPt, rmcPageHeightPt)
	repl := fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="%.6f" height="%.6f" viewBox="%s">`,
		rmcPageWidthPt, rmcPageHeightPt, viewBox,
	)
	if reSVGOpen.MatchString(svg) {
		return reSVGOpen.ReplaceAllString(svg, repl)
	}
	return repl + "\n" + svg + "\n</svg>\n"
}

func rasterSVGToPNG(svgPath, pngPath string, width, height int) error {
	var errs []string
	if bin, err := exec.LookPath("rsvg-convert"); err == nil {
		cmd := exec.Command(bin, "-w", fmt.Sprintf("%d", width), "-h", fmt.Sprintf("%d", height), "-o", pngPath, svgPath)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err == nil {
			return nil
		} else {
			errs = append(errs, fmt.Sprintf("rsvg-convert: %v %s", err, strings.TrimSpace(stderr.String())))
		}
	}
	if bin, err := exec.LookPath("inkscape"); err == nil {
		cmd := exec.Command(bin, svgPath,
			"--export-type=png",
			"--export-filename="+pngPath,
			fmt.Sprintf("--export-width=%d", width),
			fmt.Sprintf("--export-height=%d", height),
		)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err == nil {
			return nil
		} else {
			errs = append(errs, fmt.Sprintf("inkscape: %v %s", err, strings.TrimSpace(stderr.String())))
		}
	}
	if len(errs) == 0 {
		return fmt.Errorf("no rsvg-convert or inkscape")
	}
	return fmt.Errorf("%s", strings.Join(errs, " | "))
}

func svgToPDFInkscape(svgPath, pdfPath string) error {
	bin, err := exec.LookPath("inkscape")
	if err != nil {
		return err
	}
	cmd := exec.Command(bin, svgPath, "--export-type=pdf", "--export-filename="+pdfPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return fmt.Errorf("%w: %s", err, msg)
		}
		return err
	}
	return nil
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
	rmcSrc := EffectiveRMCSrc()
	if rmcSrc == "" {
		return fmt.Errorf("no rmc source (vendored third_party/rmc missing; set %s)", EnvRMCSrc)
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
