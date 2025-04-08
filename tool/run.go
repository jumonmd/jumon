// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jumonmd/jumon/internal/tracer"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Run runs a tool using NATS service.
func Run(ctx context.Context, nc *nats.Conn, tl Tool) (json.RawMessage, error) {
	slog.Info("run tool", "status", "start", "tool", tl.Name)

	data, err := json.Marshal(tl)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tool: %w", err)
	}
	slog.Debug("run tool", "tool", tl.Name, "inputsize", len(tl.InputURL))

	resp, err := nc.RequestMsgWithContext(ctx, &nats.Msg{
		Subject: "tool.run",
		Data:    data,
		Header:  tracer.HeadersFromContext(ctx),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to request tool: %w", err)
	}
	if errorCode := resp.Header.Get("Nats-Service-Error-Code"); errorCode != "" {
		errorMessage := resp.Header.Get("Nats-Service-Error")
		return nil, fmt.Errorf("tool error: %s: %s", errorCode, errorMessage)
	}

	slog.Info("run tool", "status", "end", "tool", tl.Name, "headers", resp.Header)
	slog.Debug("run tool", "output", string(resp.Data)[:min(len(string(resp.Data)), 50)])

	return json.RawMessage(resp.Data), nil
}

// run executes the tool with the given input and returns the output.
//  1. Validates the tool.
//  2. Loads resources.
//  3. Executes the tool plugin.
//  4. Cleans up resources.
func run(ctx context.Context, tl *Tool, nc *nats.Conn, obs jetstream.ObjectStore) ([]byte, error) {
	if err := tl.Validate(); err != nil {
		return nil, fmt.Errorf("validate tool: %w", err)
	}

	if err := tl.LoadResources(ctx, obs); err != nil {
		return nil, fmt.Errorf("load resources: %w", err)
	}

	defer func() {
		if err := tl.CloseResources(); err != nil {
			slog.Error("close resources", "error", err)
		}
	}()

	var output []byte
	var err error
	switch tl.Type {
	case "wasm":
		output, err = runWasmPlugin(ctx, tl)
		if err != nil {
			return nil, err
		}
	case "nats":
		output, err = runNatsPlugin(ctx, nc, tl)
		if err != nil {
			return nil, err
		}
	case "script":
		output, err = runScriptPlugin(ctx, nc, tl)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrUnknownToolType.Wrap(fmt.Errorf("unknown type: %s", tl.Type))
	}

	return output, nil
}
