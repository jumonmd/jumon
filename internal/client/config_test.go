// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	expected := &Config{
		ServerURL:  "nats://localhost:4333",
		RunTimeout: "5m",
	}

	if !cmp.Equal(cfg, expected) {
		t.Errorf("Default config mismatch:\n%s", cmp.Diff(expected, cfg))
	}
}

func TestRunTimeoutDuration(t *testing.T) {
	tests := []struct {
		name     string
		timeout  string
		expected time.Duration
	}{
		{
			name:     "valid timeout",
			timeout:  "10s",
			expected: 10 * time.Second,
		},
		{
			name:     "invalid timeout",
			timeout:  "invalid",
			expected: 5 * time.Minute, // default value
		},
		{
			name:     "empty timeout",
			timeout:  "",
			expected: 5 * time.Minute, // default value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{RunTimeout: tt.timeout}
			duration := cfg.RunTimeoutDuration()

			if duration != tt.expected {
				t.Errorf("Expected duration %v, got %v", tt.expected, duration)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Use the test file
	path := filepath.Join("testdata", "client.toml")
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// Expected configuration matches the test file
	expected := &Config{
		ServerURL:        "nats://testserver:4222",
		RunTimeout:       "30s",
		DisableTelemetry: true,
	}

	// Verify the configuration matches expectations
	if !cmp.Equal(cfg, expected) {
		t.Errorf("Config mismatch:\n%s", cmp.Diff(expected, cfg))
	}
}
