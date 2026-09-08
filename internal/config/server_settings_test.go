package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestServerSettingsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{DataDir: dir, RmcSrc: "from-env"}
	src := filepath.Join(dir, "rmc-src")
	if err := os.MkdirAll(src, 0700); err != nil {
		t.Fatal(err)
	}
	if err := cfg.UpdateServerSettings(ServerSettings{RmcSrc: src}); err != nil {
		t.Fatal(err)
	}
	if cfg.RmcSrc != src {
		t.Fatalf("cfg.RmcSrc=%q", cfg.RmcSrc)
	}
	if os.Getenv(EnvRMCSrc) != src {
		t.Fatalf("env=%q", os.Getenv(EnvRMCSrc))
	}

	cfg2 := &Config{DataDir: dir, RmcSrc: "from-env"}
	cfg2.LoadServerSettings()
	if cfg2.RmcSrc != src {
		t.Fatalf("loaded RmcSrc=%q want %q", cfg2.RmcSrc, src)
	}

	if err := cfg.UpdateServerSettings(ServerSettings{RmcSrc: ""}); err != nil {
		t.Fatal(err)
	}
	if cfg.RmcSrc != "" {
		t.Fatalf("cleared RmcSrc=%q", cfg.RmcSrc)
	}
	if os.Getenv(EnvRMCSrc) != "" {
		t.Fatalf("env still set: %q", os.Getenv(EnvRMCSrc))
	}
}

func TestUpdateServerSettingsRejectsMissingDir(t *testing.T) {
	cfg := &Config{DataDir: t.TempDir()}
	err := cfg.UpdateServerSettings(ServerSettings{RmcSrc: filepath.Join(t.TempDir(), "nope")})
	if err == nil {
		t.Fatal("expected error")
	}
}
