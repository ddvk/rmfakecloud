package rmdecode

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// EnvRepoRoot is the repository root containing scripts/ and third_party/ (optional).
const EnvRepoRoot = "RMFAKECLOUD_ROOT"

func normalizePythonImportRoot(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	// If user passes ".../src/rmscene" or ".../src/rmc", use ".../src" as PYTHONPATH root.
	base := filepath.Base(p)
	if base == "rmscene" || base == "rmc" {
		return filepath.Dir(p)
	}
	return p
}

var (
	repoRootOnce sync.Once
	repoRootVal  string
)

// FindRepoRoot walks up from cwd (and optionally RMFAKECLOUD_ROOT) looking for go.mod
// plus third_party/rmscene (vendored tooling).
func FindRepoRoot() string {
	repoRootOnce.Do(func() {
		if e := strings.TrimSpace(os.Getenv(EnvRepoRoot)); e != "" {
			if isRepoRoot(e) {
				repoRootVal = e
				return
			}
		}
		dir, err := os.Getwd()
		if err != nil {
			return
		}
		for i := 0; i < 32; i++ {
			if isRepoRoot(dir) {
				repoRootVal = dir
				return
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	})
	return repoRootVal
}

func isRepoRoot(dir string) bool {
	if st, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil || st.IsDir() {
		return false
	}
	// Prefer layout that includes vendored Python tooling.
	if st, err := os.Stat(filepath.Join(dir, "third_party", "rmscene", "src")); err == nil && st.IsDir() {
		return true
	}
	if st, err := os.Stat(filepath.Join(dir, "scripts", "rmscene_v6_to_png.py")); err == nil && !st.IsDir() {
		return true
	}
	return false
}

// VendoredRMCSrc returns third_party/rmc/src when present in the repo.
func VendoredRMCSrc() string {
	root := FindRepoRoot()
	if root == "" {
		return ""
	}
	p := filepath.Join(root, "third_party", "rmc", "src")
	if st, err := os.Stat(filepath.Join(p, "rmc")); err == nil && st.IsDir() {
		return p
	}
	return ""
}

// VendoredRMSceneSrc returns third_party/rmscene/src when present.
func VendoredRMSceneSrc() string {
	root := FindRepoRoot()
	if root == "" {
		return ""
	}
	p := filepath.Join(root, "third_party", "rmscene", "src")
	if st, err := os.Stat(filepath.Join(p, "rmscene")); err == nil && st.IsDir() {
		return p
	}
	return ""
}

// EffectiveRMCSrc returns RMFAKECLOUD_RMC_SRC if set, else the vendored third_party/rmc/src.
func EffectiveRMCSrc() string {
	if s := strings.TrimSpace(os.Getenv(EnvRMCSrc)); s != "" {
		return normalizePythonImportRoot(s)
	}
	return VendoredRMCSrc()
}

// buildPythonPathForV6 builds a PYTHONPATH that can import rmc + rmscene from:
// - optional extra path (usually rmc src)
// - RMFAKECLOUD_RMSCENE_SRC / vendored third_party/rmscene/src
// - RMFAKECLOUD_ROOT / discovered repo root
func buildPythonPathForV6(extra string) string {
	parts := make([]string, 0, 6)
	if e := normalizePythonImportRoot(extra); e != "" {
		parts = append(parts, e)
	}
	if s := EffectiveRMCSrc(); s != "" && s != normalizePythonImportRoot(extra) {
		parts = append(parts, s)
	}
	if s := normalizePythonImportRoot(os.Getenv(EnvRMSSceneSrc)); s != "" {
		parts = append(parts, s)
	}
	if s := VendoredRMSceneSrc(); s != "" {
		parts = append(parts, s)
	}
	if root := FindRepoRoot(); root != "" {
		parts = append(parts, filepath.Join(root, "third_party", "rmscene", "src"))
		parts = append(parts, filepath.Join(root, "third_party", "rmc", "src"))
	}
	if existing := strings.TrimSpace(os.Getenv("PYTHONPATH")); existing != "" {
		parts = append(parts, existing)
	}
	// Dedupe while preserving order.
	seen := map[string]struct{}{}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, string(os.PathListSeparator))
}
