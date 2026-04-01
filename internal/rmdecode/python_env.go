package rmdecode

import (
	"os"
	"path/filepath"
	"strings"
)

func normalizePythonImportRoot(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	// If user passes ".../src/rmscene", use ".../src" as PYTHONPATH root.
	if filepath.Base(p) == "rmscene" {
		return filepath.Dir(p)
	}
	return p
}

// buildPythonPathForV6 builds a PYTHONPATH that can import rmscene from:
// - RMFAKECLOUD_RMSCENE_SRC
// - RMFAKECLOUD_ROOT/third_party/rmscene/src
// plus an optional extra path (e.g. rmc source).
func buildPythonPathForV6(extra string) string {
	parts := make([]string, 0, 4)
	if e := normalizePythonImportRoot(extra); e != "" {
		parts = append(parts, e)
	}
	if s := normalizePythonImportRoot(os.Getenv(EnvRMSSceneSrc)); s != "" {
		parts = append(parts, s)
	}
	if root := strings.TrimSpace(os.Getenv(EnvRepoRoot)); root != "" {
		parts = append(parts, filepath.Join(root, "third_party", "rmscene", "src"))
	}
	if existing := strings.TrimSpace(os.Getenv("PYTHONPATH")); existing != "" {
		parts = append(parts, existing)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, string(os.PathListSeparator))
}
