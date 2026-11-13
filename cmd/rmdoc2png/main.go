// Command rmdoc2png extracts every .rm page from a .rmdoc zip and writes one PNG per page
// with zero-padded page numbers (paging).
//
// - v3/v5 .rm: rendered in Go (internal/rmdecode, github.com/fogleman/gg).
// - v6 .rm: rendered via scripts/rmscene_v6_to_png.py (Pillow + vendored rmscene).
//
// Run from the repository root (or set RMFAKECLOUD_ROOT).
//
//	go run ./cmd/rmdoc2png -o outdir document.rmdoc
package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ddvk/rmfakecloud/internal/rmdecode"
	"github.com/juruen/rmapi/encoding/rm"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: rmdoc2png [-o outdir] <file.rmdoc>")
		os.Exit(1)
	}
	outDir := ""
	args := os.Args[1:]
	for len(args) > 0 && args[0] == "-o" {
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: rmdoc2png [-o outdir] <file.rmdoc>")
			os.Exit(1)
		}
		outDir = args[1]
		args = args[2:]
	}
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: rmdoc2png [-o outdir] <file.rmdoc>")
		os.Exit(1)
	}
	rmdocPath := args[0]
	if outDir == "" {
		base := strings.TrimSuffix(filepath.Base(rmdocPath), filepath.Ext(rmdocPath))
		outDir = base + "_png"
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	repoRoot := findRepoRoot()
	if repoRoot == "" {
		if e := os.Getenv("RMFAKECLOUD_ROOT"); e != "" {
			repoRoot = e
		}
	}
	v6Script := filepath.Join(repoRoot, "scripts", "rmscene_v6_to_png.py")
	if _, err := os.Stat(v6Script); err != nil {
		fmt.Fprintf(os.Stderr, "rmdoc2png: need repo root with scripts/rmscene_v6_to_png.py (cwd or RMFAKECLOUD_ROOT): %v\n", err)
		os.Exit(1)
	}

	zr, err := zip.OpenReader(rmdocPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer zr.Close()

	var rmEntries []*zip.File
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if !strings.EqualFold(filepath.Ext(f.Name), ".rm") {
			continue
		}
		rmEntries = append(rmEntries, f)
	}
	if len(rmEntries) == 0 {
		fmt.Fprintln(os.Stderr, "no .rm files found in archive")
		os.Exit(1)
	}

	// Stable order: zip iteration order (usually alphabetical within archive).
	// Page numbers are 1-based for output filenames.
	for n, f := range rmEntries {
		pageNum := n + 1
		rc, err := f.Open()
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", f.Name, err)
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", f.Name, err)
			continue
		}

		stem := strings.TrimSuffix(f.Name, filepath.Ext(f.Name))
		stem = strings.TrimPrefix(stem, "/")
		safe := strings.ReplaceAll(stem, "/", "_")
		safe = strings.ReplaceAll(safe, "\\", "_")
		if safe == "" {
			safe = "page"
		}
		outPath := filepath.Join(outDir, fmt.Sprintf("%03d_%s.png", pageNum, safe))

		ver, err := rmdecode.ParseVersion(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", f.Name, err)
			continue
		}

		var pngData []byte
		switch ver {
		case 3, 5:
			var page rm.Rm
			if err := page.UnmarshalBinary(data); err != nil {
				fmt.Fprintf(os.Stderr, "skip %s: %v\n", f.Name, err)
				continue
			}
			pngData, err = rmdecode.RenderWritingsPNG(&page)
			if err != nil {
				fmt.Fprintf(os.Stderr, "skip %s: %v\n", f.Name, err)
				continue
			}
		case 6:
			cmd := exec.Command("python3", v6Script, "-")
			cmd.Stdin = bytes.NewReader(data)
			cmd.Dir = repoRoot
			out, err := cmd.Output()
			if err != nil {
				if x, ok := err.(*exec.ExitError); ok {
					fmt.Fprintf(os.Stderr, "skip %s: python: %s\n", f.Name, string(x.Stderr))
				} else {
					fmt.Fprintf(os.Stderr, "skip %s: %v\n", f.Name, err)
				}
				continue
			}
			pngData = out
		default:
			fmt.Fprintf(os.Stderr, "skip %s: unsupported lines version %d\n", f.Name, ver)
			continue
		}

		if err := os.WriteFile(outPath, pngData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", outPath, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "wrote %s (page %d, version %d)\n", outPath, pageNum, ver)
	}
}

func findRepoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for i := 0; i < 32; i++ {
		goMod := filepath.Join(dir, "go.mod")
		script := filepath.Join(dir, "scripts", "rmscene_v6_to_png.py")
		if st, err := os.Stat(goMod); err == nil && !st.IsDir() {
			if _, err := os.Stat(script); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
