// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package module

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
)

// importModuleTools resolves and imports tools defined as modules, adding them to the current module's tools.
func importModuleTools(ctx context.Context, nc *nats.Conn, mod *Module) error {
	slog.Debug("importing module tools", "module", mod.Name, "tools", mod.Tools)
	for _, tl := range mod.Tools {
		if tl.Module == "" { // skip if the tool is not defined as a module
			continue
		}
		slog.Debug("importing module", "module", tl.Module)
		importmod, err := getModule(ctx, nc, tl.Module)
		if err != nil {
			return fmt.Errorf("get module: %w", err)
		}

		for _, t := range importmod.Tools {
			if t.Name != tl.Name {
				continue
			}
			mod.Tools = append(mod.Tools, t)
		}
	}
	return nil
}

// importScriptSymbolTools adds scripts that are defined as symbols in other scripts as tools.
// e.g.
// > ### ScriptA
// > Call `ScriptB` and show the result.
// > ### ScriptB
// > 1. Solve the quadratic equation
// - > ScriptA.Tools will be added ScriptB tool.
func importScriptSymbolTools(mod *Module, defaultModel string) error {
	for _, scr := range mod.Scripts {
		for _, tgt := range mod.Scripts {
			if scr.Name == tgt.Name {
				continue
			}
			symbols, err := scr.Symbols()
			if err != nil {
				return fmt.Errorf("get symbols: %w", err)
			}
			for _, s := range symbols {
				if tgt.Name == s.Name {
					if tgt.Model == "" {
						tgt.Model = defaultModel
					}
					tool, err := tgt.AsTool()
					if err != nil {
						return fmt.Errorf("convert tool: %w", err)
					}
					scr.Tools = append(scr.Tools, tool)
				}
			}
		}
	}
	return nil
}
