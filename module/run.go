// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package module

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jumonmd/jumon/internal/config"
	"github.com/jumonmd/jumon/script"
	"github.com/nats-io/nats.go"
)

// Run executes a module with the given module URL using NATS service.
func Run(ctx context.Context, nc *nats.Conn, modurl string, input []byte) (json.RawMessage, error) {
	modname, scriptname := extractModScriptName(modurl)
	slog.Debug("run module", "modurl", modurl)

	mod, err := Get(ctx, nc, modname)
	if err != nil {
		return nil, ErrModuleNotFound.Wrap(fmt.Errorf("%w: %s", err, modname))
	}

	slog.Debug("run module", "modname", modname, "script", scriptname)
	scr := mod.GetScript(scriptname)
	if scr == nil {
		return nil, ErrScriptNotFound.Wrap(fmt.Errorf("script not found: %s", scriptname))
	}

	defaultModel, err := config.Get(ctx, nc, config.DefaultModel)
	if err != nil {
		return nil, fmt.Errorf("get default model: %w", err)
	}
	if scr.Model == "" {
		slog.Debug("using default model", "model", defaultModel)
		scr.Model = defaultModel
	}
	// currently, input is expected to be JSON
	scr.SetInput(input)

	scr.Tools = append(mod.Tools, scr.Tools...)
	return script.Run(ctx, nc, scr)
}

// Get returns a module with the given module name with resolved tools and scripts.
func Get(ctx context.Context, nc *nats.Conn, modname string) (*Module, error) {
	mod, err := getModule(ctx, nc, modname)
	if err != nil {
		return nil, fmt.Errorf("get module: %w", err)
	}

	err = mod.Validate()
	if err != nil {
		return nil, fmt.Errorf("validate module: %w", err)
	}

	err = importModuleTools(ctx, nc, mod)
	if err != nil {
		return nil, fmt.Errorf("prepare import tools: %w", err)
	}

	defaultModel, err := config.Get(ctx, nc, config.DefaultModel)
	if err != nil {
		return nil, fmt.Errorf("get default model: %w", err)
	}

	err = importScriptSymbolTools(mod, defaultModel)
	if err != nil {
		return nil, fmt.Errorf("prepare script tools: %w", err)
	}

	return mod, nil
}

// extractModScriptName extracts the module name and script name from the module URL.
// e.g. "jumonmd/jumon/example/hello#sayname" -> ("jumonmd/jumon/example/hello", "sayname").
func extractModScriptName(modurl string) (modname string, scriptname string) {
	parts := strings.SplitN(modurl, "#", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return modurl, ""
}
