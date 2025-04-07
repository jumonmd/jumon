// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/zchee/go-xdgbasedir"
)

const defaultRunTimeout = "5m"

// Config is the configuration for the client.
type Config struct {
	// ServerURL is the URL of the NATS server.
	ServerURL string `toml:"server_url"`
	// RunTimeout is the timeout for the run command. e.g. "5m".
	RunTimeout string `toml:"run_timeout"`
	// DisableTelemetry is the flag to disable telemetry.
	DisableTelemetry bool `toml:"disable_telemetry"`
}

// LoadConfig loads the configuration from the config file.
// If the file does not exist, a default configuration is returned.
// Config file is located in XDG Config Directory with subdirectory "jumon".
// e.g. ~/.config/jumon/client.toml.
func LoadConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return NewDefaultConfig(), nil
	}

	cfg := &Config{}

	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return cfg, nil
}

// NewDefaultConfig returns a default configuration for the client.
func NewDefaultConfig() *Config {
	return &Config{
		ServerURL:  "nats://localhost:4333",
		RunTimeout: defaultRunTimeout,
	}
}

// RunTimeoutDuration returns the run timeout as a duration.
func (c *Config) RunTimeoutDuration() time.Duration {
	d, err := time.ParseDuration(c.RunTimeout)
	if err != nil {
		d, _ = time.ParseDuration(defaultRunTimeout)
	}
	return d
}

// defaultConfigPath returns the path to the config file based on XDG Config Directory.
// e.g. ~/.config/jumon/client.toml.
func DefaultConfigPath() string {
	configDir := xdgbasedir.ConfigHome()
	return filepath.Join(configDir, "jumon", "client.toml")
}
