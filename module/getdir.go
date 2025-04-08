// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package module

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/nats-io/nats.go/jetstream"
)

// GetByDir loads a module from a local directory and stores it.
func GetByDir(ctx context.Context, kv jetstream.KeyValue, dir string) (*Module, error) {
	slog.Info("jumon get by dir", "dir", dir)
	moddata, err := os.ReadFile(filepath.Join(dir, "JUMON.md"))
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	mod, err := ParseMarkdown(moddata)
	if err != nil {
		return nil, fmt.Errorf("parse module: %w", err)
	}

	if err := mod.Validate(); err != nil {
		return nil, fmt.Errorf("validate module: %w", err)
	}

	_, err = kv.Put(ctx, mod.Name, moddata)
	if err != nil {
		return nil, fmt.Errorf("put module failed: %w", err)
	}

	slog.Info("jumon get by dir", "module", mod.Name)
	slog.Debug("module", "module", string(moddata))
	return mod, nil
}
