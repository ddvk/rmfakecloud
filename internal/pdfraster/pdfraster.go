package pdfraster

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// EnvPDFRenderer selects the PDF→PNG backend.
	// Values: "auto" (default), "pdftoppm", "mutool", "unipdf".
	// Out-of-process tools mirror xochitl_pdf_renderer (crash-isolated PDF engine).
	EnvPDFRenderer = "RMFAKECLOUD_PDF_RENDERER"

	// DevicePortraitWidth is the reMarkable portrait page width used for raster targets.
	DevicePortraitWidth = 1404
)

// RenderFunc optionally overrides the final unipdf fallback (injected by exporter to
// avoid an import cycle with unidoc).
var UnipdfFallback func(pdfBytes []byte, pageNum int) ([]byte, error)

// RenderPage rasters one PDF page (1-based) to PNG.
func RenderPage(pdfBytes []byte, pageNum int) ([]byte, error) {
	if len(pdfBytes) == 0 {
		return nil, fmt.Errorf("empty pdf")
	}
	if pageNum < 1 {
		return nil, fmt.Errorf("page %d out of range", pageNum)
	}

	prefer := strings.ToLower(strings.TrimSpace(os.Getenv(EnvPDFRenderer)))
	if prefer == "" {
		prefer = "auto"
	}

	var errs []string
	try := func(name string, fn func() ([]byte, error)) ([]byte, error) {
		b, err := fn()
		if err == nil {
			return b, nil
		}
		errs = append(errs, name+": "+err.Error())
		return nil, err
	}

	runAuto := func() ([]byte, error) {
		if b, err := try("pdftoppm", func() ([]byte, error) { return withPdftoppm(pdfBytes, pageNum) }); err == nil {
			return b, nil
		}
		if b, err := try("mutool", func() ([]byte, error) { return withMutool(pdfBytes, pageNum) }); err == nil {
			return b, nil
		}
		if UnipdfFallback != nil {
			if b, err := try("unipdf", func() ([]byte, error) { return UnipdfFallback(pdfBytes, pageNum) }); err == nil {
				return b, nil
			}
		}
		return nil, fmt.Errorf("pdf raster failed: %s", strings.Join(errs, " | "))
	}

	switch prefer {
	case "pdftoppm":
		if b, err := try("pdftoppm", func() ([]byte, error) { return withPdftoppm(pdfBytes, pageNum) }); err == nil {
			return b, nil
		}
	case "mutool":
		if b, err := try("mutool", func() ([]byte, error) { return withMutool(pdfBytes, pageNum) }); err == nil {
			return b, nil
		}
	case "unipdf":
		if UnipdfFallback == nil {
			return nil, fmt.Errorf("unipdf fallback not registered")
		}
		return UnipdfFallback(pdfBytes, pageNum)
	default:
		return runAuto()
	}

	// Preferred tool failed — still try the rest.
	return runAuto()
}

func withPdftoppm(pdfBytes []byte, pageNum int) ([]byte, error) {
	bin, err := exec.LookPath("pdftoppm")
	if err != nil {
		return nil, err
	}
	tmpDir, err := os.MkdirTemp("", "pdftoppm-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	inPath := filepath.Join(tmpDir, "in.pdf")
	if err := os.WriteFile(inPath, pdfBytes, 0600); err != nil {
		return nil, err
	}
	outPrefix := filepath.Join(tmpDir, "page")
	page := strconv.Itoa(pageNum)
	cmd := exec.Command(bin,
		"-png",
		"-f", page,
		"-l", page,
		"-singlefile",
		"-cropbox",
		"-scale-to-x", strconv.Itoa(DevicePortraitWidth),
		"-scale-to-y", "-1",
		inPath,
		outPrefix,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return nil, fmt.Errorf("%w: %s", err, msg)
		}
		return nil, err
	}
	return os.ReadFile(outPrefix + ".png")
}

func withMutool(pdfBytes []byte, pageNum int) ([]byte, error) {
	bin, err := exec.LookPath("mutool")
	if err != nil {
		return nil, err
	}
	tmpDir, err := os.MkdirTemp("", "mutool-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	inPath := filepath.Join(tmpDir, "in.pdf")
	outPath := filepath.Join(tmpDir, "page.png")
	if err := os.WriteFile(inPath, pdfBytes, 0600); err != nil {
		return nil, err
	}
	cmd := exec.Command(bin, "draw",
		"-o", outPath,
		"-F", "png",
		"-w", strconv.Itoa(DevicePortraitWidth),
		"-b", "CropBox",
		inPath,
		strconv.Itoa(pageNum),
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return nil, fmt.Errorf("%w: %s", err, msg)
		}
		return nil, err
	}
	return os.ReadFile(outPath)
}
