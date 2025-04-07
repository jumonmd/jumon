// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package server

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/zchee/go-xdgbasedir"
)

const defaultConfig = `
listen: 127.0.0.1:4222
jetstream {
	store_dir = "%s"
}
pid_file = "%s"
`

// createConfig creates a default config file if it doesn't exist.
func createConfig() error {
	cfgpath := configPath()
	cfgdir := filepath.Dir(cfgpath)
	if _, err := os.Stat(cfgdir); os.IsNotExist(err) {
		if err := os.MkdirAll(cfgdir, 0o755); err != nil {
			return fmt.Errorf("create config dir: %w", err)
		}
	}

	if _, err := os.Stat(cfgpath); err == nil {
		return nil
	}

	slog.Info("creating default config", "path", cfgpath)

	cfg := fmt.Sprintf(defaultConfig, storeDir(), pidPath())
	err := os.WriteFile(cfgpath, []byte(cfg), 0o600)
	if err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// configPath returns the path to the config file based on XDG Config Directory.
// e.g. ~/.config/jumon/server.conf.
func configPath() string {
	configDir := xdgbasedir.ConfigHome()
	path := filepath.Join(configDir, "jumon", "server.conf")
	return path
}

// storeDir returns the directory for the Jumon server.
// XDG Data Directory with subdirectory "jumon" is used.
// e.g. ~/.local/share/jumon.
func storeDir() string {
	dataDir := xdgbasedir.DataHome()
	dir := filepath.Join(dataDir, "jumon")

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatal("failed to create store dir: %w", err)
		}
	}
	return dir
}

// pidPath returns the path to the pid file.
// e.g. ~/.local/share/jumon/jumon.pid.
func pidPath() string {
	storeDir := storeDir()
	return filepath.Join(storeDir, "jumon.pid")
}
