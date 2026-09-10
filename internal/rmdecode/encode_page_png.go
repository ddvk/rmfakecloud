package rmdecode

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// EnvV6PNGScript is the full path to scripts/rmscene_v6_to_png.py (optional).
const EnvV6PNGScript = "RMFAKECLOUD_V6_PNG_SCRIPT"

// ScreenDPI is the reMarkable display density used for inch→pixel mapping.
const ScreenDPI = 226

// EncodeRmPageToPNG converts raw .rm page bytes to a device-canvas PNG (notebooks).
func EncodeRmPageToPNG(data []byte) ([]byte, error) {
	return encodeRmPageToPNG(data, nil, 0, 0)
}

// EncodeRmPageToPNGWithImages is EncodeRmPageToPNG with inserted page image blobs
// (basename → PNG/JPEG bytes), used for reMarkable v6 image items.
func EncodeRmPageToPNGWithImages(data []byte, images map[string][]byte) ([]byte, error) {
	return encodeRmPageToPNG(data, images, 0, 0)
}

// EncodeRmPageToPNGForPDF renders .rm ink in PDF-page coordinates scaled to the
// same width-fitted raster used for the PDF background (1404 × aspect height).
// widthPt/heightPt are the PDF page size in points (e.g. 612×792 for Letter).
func EncodeRmPageToPNGForPDF(data []byte, widthPt, heightPt float64) ([]byte, error) {
	if widthPt < 1 || heightPt < 1 {
		return nil, fmt.Errorf("invalid pdf page size %gx%g", widthPt, heightPt)
	}
	return encodeRmPageToPNG(data, nil, widthPt, heightPt)
}

// EncodeRmPageToPNGForPDFWithImages is EncodeRmPageToPNGForPDF with page images.
func EncodeRmPageToPNGForPDFWithImages(data []byte, images map[string][]byte, widthPt, heightPt float64) ([]byte, error) {
	if widthPt < 1 || heightPt < 1 {
		return nil, fmt.Errorf("invalid pdf page size %gx%g", widthPt, heightPt)
	}
	return encodeRmPageToPNG(data, images, widthPt, heightPt)
}

func encodeRmPageToPNG(data []byte, images map[string][]byte, pdfWPt, pdfHPt float64) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty .rm data")
	}
	ver, err := ParseVersion(data)
	if err != nil {
		return nil, err
	}
	var ink []byte
	switch ver {
	case 3:
		if b, err := RenderV3PNGWithLines2PNG(data); err == nil {
			ink = b
		} else if Lines2PNGAvailable() {
			return nil, fmt.Errorf("lines2png available but failed: %w", err)
		} else {
			page, err := DecodeLegacy(data)
			if err != nil {
				return nil, err
			}
			ink, err = RenderWritingsPNG(page)
			if err != nil {
				return nil, err
			}
		}
	case 5:
		page, err := DecodeLegacy(data)
		if err != nil {
			return nil, err
		}
		ink, err = RenderWritingsPNG(page)
		if err != nil {
			return nil, err
		}
	case 6:
		ink, err = encodeV6RmToPNG(data, images, pdfWPt, pdfHPt)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported .rm version %d", ver)
	}

	meta := ReadInkDeviceMeta(ink)
	if meta.ViewW < 1 || meta.ViewH < 1 {
		meta = DefaultInkDeviceMeta()
	}
	out, err := WriteInkDeviceMeta(ink, meta)
	if err != nil {
		return ink, nil
	}
	return out, nil
}

func encodeV6RmToPNG(data []byte, images map[string][]byte, pdfWPt, pdfHPt float64) ([]byte, error) {
	// Prefer rmc SVG→PNG so inserted images (and full scene items) render.
	if png, err := RenderV6PNGWithRMC(data, images); err == nil {
		return png, nil
	} else {
		rmcErr := err
		if len(images) > 0 {
			return nil, fmt.Errorf("v6 png rmc (required for page images): %v", rmcErr)
		}
		if png, err := renderV6PNGWithScript(data, pdfWPt, pdfHPt); err == nil {
			return png, nil
		} else {
			return nil, fmt.Errorf("v6 png: rmc: %v; script: %v", rmcErr, err)
		}
	}
}

func renderV6PNGWithScript(data []byte, pdfWPt, pdfHPt float64) ([]byte, error) {
	script := os.Getenv(EnvV6PNGScript)
	if script == "" {
		if root := os.Getenv(EnvRepoRoot); root != "" {
			script = filepath.Join(root, "scripts", "rmscene_v6_to_png.py")
		} else if root := FindRepoRoot(); root != "" {
			script = filepath.Join(root, "scripts", "rmscene_v6_to_png.py")
		}
	}
	if script == "" {
		return nil, fmt.Errorf("no %s / %s for scripts/rmscene_v6_to_png.py", EnvV6PNGScript, EnvRepoRoot)
	}
	if st, err := os.Stat(script); err != nil || st.IsDir() {
		return nil, fmt.Errorf("v6 PNG script %q: %w", script, err)
	}

	args := []string{script}
	if pdfWPt >= 1 && pdfHPt >= 1 {
		args = append(args,
			"--pdf-pts",
			strconv.FormatFloat(pdfWPt, 'f', -1, 64),
			strconv.FormatFloat(pdfHPt, 'f', -1, 64),
		)
	}
	args = append(args, "-")
	cmd := exec.Command("python3", args...)
	cmd.Stdin = bytes.NewReader(data)
	if py := buildPythonPathForV6(""); py != "" {
		cmd.Env = append(os.Environ(), "PYTHONPATH="+py)
	}
	out, err := cmd.Output()
	if err != nil {
		if x, ok := err.(*exec.ExitError); ok && len(x.Stderr) > 0 {
			return nil, fmt.Errorf("v6 python: %w: %s", err, string(x.Stderr))
		}
		return nil, fmt.Errorf("v6 python: %w", err)
	}
	if len(out) < 32 {
		return nil, fmt.Errorf("v6 python: empty png")
	}
	return out, nil
}
