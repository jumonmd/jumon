// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package tool

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	extism "github.com/extism/go-sdk"
)

type wasmRunner struct {
	plugin   *extism.Plugin
	funcname string
}

func newWASMRunner(ctx context.Context, arguments Arguments, resources []*Resource) (*wasmRunner, error) {
	config := extism.PluginConfig{
		EnableWasi: true,
	}

	if len(resources) == 0 {
		return nil, fmt.Errorf("a wasm resource is required")
	}
	if len(arguments) == 0 {
		return nil, fmt.Errorf("no arguments")
	}
	funcname, ok := arguments["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name is not a string")
	}

	wasmres := resources[0]
	data, err := io.ReadAll(wasmres.reader)
	if err != nil {
		return nil, fmt.Errorf("read wasm resource: %w", err)
	}
	defer wasmres.Close()
	slog.Debug("read wasm file", "size", len(data))

	wasmManifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmData{
				Data: data,
				Hash: wasmres.Hash,
			},
		},
	}

	plugin, err := extism.NewPlugin(ctx, wasmManifest, config, []extism.HostFunction{})
	if err != nil {
		return nil, fmt.Errorf("create wasm plugin: %w", err)
	}

	return &wasmRunner{plugin: plugin, funcname: funcname}, nil
}

// Run runs the wasm function using extism.
func (r *wasmRunner) Run(ctx context.Context, input []byte) (output []byte, err error) {
	exit, out, err := r.plugin.CallWithContext(ctx, r.funcname, input)
	if err != nil {
		return nil, fmt.Errorf("call wasm (exit=%d, funcname=%s, input=%s): %w", exit, r.funcname, string(input), err)
	}

	return out, nil
}
