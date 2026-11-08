package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

const serverSettingsFile = "server_settings.json"

// ServerSettings holds admin-editable values persisted under DATADIR.
// Environment variables remain the bootstrap defaults; GUI values override them.
type ServerSettings struct {
	// RmcSrc is the rmc Python package source root (…/rmc-main/src).
	RmcSrc string `json:"rmcSrc"`
}

var serverSettingsMu sync.Mutex

func (cfg *Config) serverSettingsPath() string {
	return filepath.Join(cfg.DataDir, serverSettingsFile)
}

// LoadServerSettings overlays DATADIR/server_settings.json onto cfg (after env).
// When the file exists, GUI-managed fields from the file win (including empty clears).
func (cfg *Config) LoadServerSettings() {
	if cfg == nil || cfg.DataDir == "" {
		return
	}
	path := cfg.serverSettingsPath()
	b, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Warnf("read %s: %v", path, err)
		}
		return
	}
	var s ServerSettings
	if err := json.Unmarshal(b, &s); err != nil {
		log.Warnf("parse %s: %v", path, err)
		return
	}
	cfg.RmcSrc = strings.TrimSpace(s.RmcSrc)
}

// ApplyRuntimeEnv pushes GUI/config values into process env so packages that
// read os.Getenv (e.g. rmdecode) pick them up without an import cycle.
func (cfg *Config) ApplyRuntimeEnv() {
	if cfg == nil {
		return
	}
	if v := strings.TrimSpace(cfg.RmcSrc); v != "" {
		_ = os.Setenv(EnvRMCSrc, v)
	} else {
		_ = os.Unsetenv(EnvRMCSrc)
	}
}

// GetServerSettings returns the current admin-editable settings snapshot.
func (cfg *Config) GetServerSettings() ServerSettings {
	serverSettingsMu.Lock()
	defer serverSettingsMu.Unlock()
	return ServerSettings{
		RmcSrc: strings.TrimSpace(cfg.RmcSrc),
	}
}

// UpdateServerSettings validates, persists, and applies admin-editable settings.
func (cfg *Config) UpdateServerSettings(s ServerSettings) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	rmcSrc := strings.TrimSpace(s.RmcSrc)
	if rmcSrc != "" {
		st, err := os.Stat(rmcSrc)
		if err != nil {
			return fmt.Errorf("rmcSrc %q: %w", rmcSrc, err)
		}
		if !st.IsDir() {
			return fmt.Errorf("rmcSrc %q is not a directory", rmcSrc)
		}
	}

	serverSettingsMu.Lock()
	defer serverSettingsMu.Unlock()

	out := ServerSettings{RmcSrc: rmcSrc}
	if err := os.MkdirAll(cfg.DataDir, 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	path := cfg.serverSettingsPath()
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	cfg.RmcSrc = rmcSrc
	cfg.ApplyRuntimeEnv()
	log.Infof("server settings updated (%s=%q)", EnvRMCSrc, rmcSrc)
	return nil
}
