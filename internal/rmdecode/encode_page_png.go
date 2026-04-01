package rmdecode

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// EnvV6PNGScript is the full path to scripts/rmscene_v6_to_png.py (optional).
const EnvV6PNGScript = "RMFAKECLOUD_V6_PNG_SCRIPT"

// EnvRepoRoot is the repository root containing scripts/ (optional; used to find v6 script).
const EnvRepoRoot = "RMFAKECLOUD_ROOT"

// EncodeRmPageToPNG converts raw .rm page bytes to PNG (1404×1872).
// v3/v5: rendered in Go (RenderWritingsPNG).
// v6: requires python3 and scripts/rmscene_v6_to_png.py (see EnvV6PNGScript / EnvRepoRoot).
func EncodeRmPageToPNG(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty .rm data")
	}
	ver, err := ParseVersion(data)
	if err != nil {
		return nil, err
	}
	switch ver {
	case 3:
		// Prefer lines-are-beautiful for stroke-faithful v3 rendering when installed.
		if png, err := RenderV3PNGWithLines2PNG(data); err == nil {
			return png, nil
		}
		page, err := DecodeLegacy(data)
		if err != nil {
			return nil, err
		}
		return RenderWritingsPNG(page)
	case 5:
		page, err := DecodeLegacy(data)
		if err != nil {
			return nil, err
		}
		return RenderWritingsPNG(page)
	case 6:
		return encodeV6RmToPNG(data)
	default:
		return nil, fmt.Errorf("unsupported .rm version %d", ver)
	}
}

func encodeV6RmToPNG(data []byte) ([]byte, error) {
	script := os.Getenv(EnvV6PNGScript)
	if script == "" {
		if root := os.Getenv(EnvRepoRoot); root != "" {
			script = filepath.Join(root, "scripts", "rmscene_v6_to_png.py")
		}
	}
	if script == "" {
		return nil, fmt.Errorf("v6 .rm requires %s or %s pointing to scripts/rmscene_v6_to_png.py", EnvV6PNGScript, EnvRepoRoot)
	}
	if st, err := os.Stat(script); err != nil || st.IsDir() {
		return nil, fmt.Errorf("v6 PNG script %q: %w", script, err)
	}
	cmd := exec.Command("python3", script, "-")
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
	return out, nil
}
