// Command rmdecode decodes a reMarkable .rm page file to text/markdown/SVG/PDF.
//
//	go run ./cmd/rmdecode page.rm
//	go run ./cmd/rmdecode -format svg -o out.svg page.rm
//	go run ./cmd/rmdecode -o out.pdf page.rm
//
// Version 3/5: decoded in Go (juruen/rmapi); SVG/PDF rendered in Go.
// Version 6: markdown/SVG/PDF via `rmc` (pip install rmc / pipx install rmc); optional text summary via scripts/rmdecode_v6_summary.py.
package main

import (
	"flag"
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
	outPath := flag.String("o", "", "output file (default: stdout for text); extension .svg/.pdf/.md selects format")
	formatFlag := flag.String("format", "", "output type: text (default), markdown, svg, or pdf (optional if inferred from -o)")
	v6Script := flag.String("v6-script", "", "path to rmdecode_v6_summary.py (default: ./scripts/rmdecode_v6_summary.py)")
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: rmdecode [-format text|markdown|svg|pdf] [-o path] <file.rm>")
		os.Exit(1)
	}
	path := args[0]
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ver, err := rmdecode.ParseVersion(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "header version: %d\n", ver)

	format := strings.ToLower(strings.TrimSpace(*formatFlag))
	if format == "" && *outPath != "" {
		switch strings.ToLower(filepath.Ext(*outPath)) {
		case ".svg":
			format = "svg"
		case ".pdf":
			format = "pdf"
		case ".md", ".markdown":
			format = "markdown"
		case ".txt":
			format = "text"
		default:
			format = "text"
		}
	}
	if format == "" {
		format = "text"
	}

	switch format {
	case "text", "markdown", "svg", "pdf":
	default:
		fmt.Fprintf(os.Stderr, "invalid -format %q (use text, markdown, svg, or pdf)\n", format)
		os.Exit(1)
	}

	if format == "svg" || format == "pdf" || format == "markdown" {
		if *outPath == "" {
			fmt.Fprintln(os.Stderr, "markdown/svg/pdf output require -o <file> (or use -o - for stdout)")
			os.Exit(1)
		}
	}

	switch ver {
	case 3, 5:
		page, err := rmdecode.DecodeLegacy(data)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := writeLegacy(format, page, *outPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case 6:
		if err := writeV6(format, path, *outPath, *v6Script); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unsupported lines format version %d\n", ver)
		os.Exit(1)
	}
}

func writeLegacy(format string, page *rm.Rm, outPath string) error {
	switch format {
	case "text":
		out := os.Stdout
		if outPath != "" {
			f, err := os.Create(outPath)
			if err != nil {
				return err
			}
			defer f.Close()
			out = f
		}
		_, err := io.WriteString(out, page.String())
		return err
	case "svg":
		s, err := rmdecode.RenderWritingsSVG(page)
		if err != nil {
			return err
		}
		return writeStringOut(outPath, s)
	case "pdf":
		b, err := rmdecode.RenderWritingsPDF(page)
		if err != nil {
			return err
		}
		return writeBytesOut(outPath, b)
	case "markdown":
		return fmt.Errorf("markdown is only supported for v6 via rmc")
	}
	return fmt.Errorf("unknown format %q", format)
}

func writeV6(format, inPath, outPath, v6Script string) error {
	switch format {
	case "text":
		script := v6Script
		if script == "" {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			script = filepath.Join(wd, "scripts", "rmdecode_v6_summary.py")
		}
		if _, err := os.Stat(script); err != nil {
			return fmt.Errorf("v6 text: script not found: %s (use -v6-script or run from repo root)", script)
		}
		cmd := exec.Command("python3", script, inPath)
		if py := buildPythonPathForV6(); py != "" {
			cmd.Env = append(os.Environ(), "PYTHONPATH="+py)
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	case "markdown", "svg", "pdf":
		rmc, err := exec.LookPath("rmc")
		if err != nil {
			return fmt.Errorf("v6 %s: install the `rmc` tool (pip install rmc or pipx install rmc): %w", format, err)
		}
		if outPath == "-" {
			return fmt.Errorf("v6 %s: use a file path with -o (rmc does not stream to stdout here)", format)
		}
		cmd := exec.Command(rmc, "-t", format, "-o", outPath, inPath)
		if py := buildPythonPathForV6(); py != "" {
			cmd.Env = append(os.Environ(), "PYTHONPATH="+py)
		}
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return fmt.Errorf("unknown format %q", format)
}

func buildPythonPathForV6() string {
	parts := make([]string, 0, 3)
	if p := strings.TrimSpace(os.Getenv("RMFAKECLOUD_RMSCENE_SRC")); p != "" {
		if filepath.Base(p) == "rmscene" {
			p = filepath.Dir(p)
		}
		parts = append(parts, p)
	}
	if root := strings.TrimSpace(os.Getenv("RMFAKECLOUD_ROOT")); root != "" {
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

func writeStringOut(outPath, s string) error {
	if outPath == "-" {
		_, err := io.WriteString(os.Stdout, s)
		return err
	}
	return os.WriteFile(outPath, []byte(s), 0644)
}

func writeBytesOut(outPath string, b []byte) error {
	if outPath == "-" {
		_, err := os.Stdout.Write(b)
		return err
	}
	return os.WriteFile(outPath, b, 0644)
}
