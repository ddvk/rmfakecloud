package exporter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/juruen/rmapi/archive"
	log "github.com/sirupsen/logrus"
)

// RmcConfig holds configuration for RMC tool execution
type RmcConfig struct {
	RmcPath      string        // Path to rmc binary
	TempDir      string        // Temporary directory for processing
	Timeout      time.Duration // Command timeout
	InkscapePath string        // Path to inkscape (optional, for custom location)
}

// DefaultRmcConfig returns default configuration
func DefaultRmcConfig() RmcConfig {
	return RmcConfig{
		RmcPath: "rmc",                // Assume in PATH
		TempDir: os.TempDir(),         // Use system temp
		Timeout: 60 * time.Second,     // 60 second timeout
		InkscapePath: "",              // Auto-detect
	}
}

// ExportV6ToPdf converts v6 .rm file to PDF using rmc tool
func ExportV6ToPdf(rmFilePath, outputPath string, cfg RmcConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Validate input file exists
	if _, err := os.Stat(rmFilePath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", rmFilePath)
	}

	// Check if rmc exists
	rmcPath := cfg.RmcPath
	if rmcPath == "" {
		rmcPath = "rmc"
	}

	// Build command: rmc input.rm -o output.pdf
	cmd := exec.CommandContext(ctx, rmcPath, rmFilePath, "-o", outputPath)

	// Set environment - add inkscape to PATH if specified
	if cfg.InkscapePath != "" {
		inkscapeDir := filepath.Dir(cfg.InkscapePath)
		currentPath := os.Getenv("PATH")
		newPath := fmt.Sprintf("%s:%s", inkscapeDir, currentPath)
		cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s", newPath))
	} else {
		cmd.Env = os.Environ()
	}

	log.Debugf("Executing rmc command: %s %s -o %s", rmcPath, rmFilePath, outputPath)

	// Capture output for logging
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("rmc failed: %v, output: %s", err, string(output))

		// Check for timeout
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("rmc conversion timeout after %v", cfg.Timeout)
		}

		return fmt.Errorf("rmc conversion failed: %w (output: %s)", err, string(output))
	}

	log.Debugf("rmc output: %s", string(output))

	// Verify output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("rmc did not create output file: %s", outputPath)
	}

	return nil
}

// ExportV6ArchiveToPdf handles conversion of v6 archive to PDF
// This function extracts .rm files from the archive and converts them
func ExportV6ArchiveToPdf(arch *MyArchive, outputPath string, cfg RmcConfig) error {
	// For v6 files in archive format, we need to extract the raw .rm data
	// The archive contains Pages with Data that needs to be written to temp files

	if len(arch.Pages) == 0 {
		return fmt.Errorf("archive contains no pages")
	}

	// Create temp directory for extraction
	tempDir := filepath.Join(cfg.TempDir, fmt.Sprintf("rmfakecloud-v6-%d", time.Now().UnixNano()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up

	log.Debugf("Extracting v6 archive to temp dir: %s", tempDir)

	// For single page documents
	if len(arch.Pages) == 1 {
		rmFile := filepath.Join(tempDir, "page.rm")
		if err := writePageToFile(arch.Pages[0], rmFile); err != nil {
			return err
		}
		return ExportV6ToPdf(rmFile, outputPath, cfg)
	}

	// For multi-page documents, we need to convert each page and merge
	// This is complex - for now, we'll convert the first page only
	// TODO: Implement multi-page PDF merging
	log.Warnf("Multi-page v6 document detected (%d pages), converting first page only", len(arch.Pages))

	rmFile := filepath.Join(tempDir, "page_0.rm")
	if err := writePageToFile(arch.Pages[0], rmFile); err != nil {
		return err
	}

	return ExportV6ToPdf(rmFile, outputPath, cfg)
}

// writePageToFile writes a page's data to a .rm file
func writePageToFile(page archive.Page, filepath string) error {
	if page.Data == nil {
		return fmt.Errorf("page has no data")
	}

	// Marshal the page data to binary
	data, err := page.Data.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal page data: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write rm file: %w", err)
	}

	return nil
}

// CheckRmcAvailable checks if rmc command is available
func CheckRmcAvailable(rmcPath string) error {
	if rmcPath == "" {
		rmcPath = "rmc"
	}

	cmd := exec.Command(rmcPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rmc not found or not executable: %w (try: pip install rmc)", err)
	}

	log.Debugf("rmc version: %s", string(output))
	return nil
}

// ExportV6ToSvg converts v6 .rm file to SVG using rmc tool
// This is an alternative to PDF that doesn't require Inkscape
func ExportV6ToSvg(rmFilePath, outputPath string, cfg RmcConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	rmcPath := cfg.RmcPath
	if rmcPath == "" {
		rmcPath = "rmc"
	}

	// Build command: rmc input.rm -t svg -o output.svg
	cmd := exec.CommandContext(ctx, rmcPath, rmFilePath, "-t", "svg", "-o", outputPath)
	cmd.Env = os.Environ()

	log.Debugf("Executing rmc SVG command: %s %s -t svg -o %s", rmcPath, rmFilePath, outputPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("rmc SVG conversion failed: %v, output: %s", err, string(output))

		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("rmc SVG conversion timeout after %v", cfg.Timeout)
		}

		return fmt.Errorf("rmc SVG conversion failed: %w (output: %s)", err, string(output))
	}

	log.Debugf("rmc SVG output: %s", string(output))

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("rmc did not create SVG output file: %s", outputPath)
	}

	return nil
}