// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package tool

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jumonmd/jumon/internal/dataurl"
	"github.com/jumonmd/jumon/internal/tracer"
	"github.com/nats-io/nats.go"
)

// runWasmPlugin runs a WASM plugin and returns its output.
func runWasmPlugin(ctx context.Context, tl *Tool) ([]byte, error) {
	wasmRunner, err := newWASMRunner(ctx, tl.Arguments, tl.Resources)
	if err != nil {
		return nil, ErrWasmValidate.Wrap(fmt.Errorf("wasm runner create failed: %w", err))
	}

	input, _, err := dataurl.Decode(tl.InputURL)
	if err != nil {
		return nil, ErrWasmValidate.Wrap(fmt.Errorf("input decode failed: %w", err))
	}
	slog.Debug("run tool", "wasm input", string(input))

	resp, err := wasmRunner.Run(ctx, input)
	if err != nil {
		return nil, ErrRunWasm.Wrap(fmt.Errorf("wasm runner run failed: %w", err))
	}

	slog.Debug("run tool", "wasm output", string(resp))
	return resp, nil
}

// runNatsPlugin runs a NATS plugin and returns its output.
func runNatsPlugin(ctx context.Context, nc *nats.Conn, tl *Tool) ([]byte, error) {
	if tl.Arguments == nil {
		return nil, ErrNatsValidate.Wrap(fmt.Errorf("nats subject is not set"))
	}
	subject, ok := tl.Arguments["subject"].(string)
	if !ok {
		return nil, ErrNatsValidate.Wrap(fmt.Errorf("nats subject is not set"))
	}

	input, _, err := dataurl.Decode(tl.InputURL)
	if err != nil {
		return nil, ErrNatsValidate.Wrap(fmt.Errorf("input decode failed: %w", err))
	}

	slog.Debug("run tool", "nats subject", subject)

	resp, err := nc.RequestMsgWithContext(ctx, &nats.Msg{
		Subject: subject,
		Data:    input,
		Header:  tracer.HeadersFromContext(ctx),
	})
	if err != nil {
		return nil, ErrNatsValidate.Wrap(fmt.Errorf("nats request failed: %w", err))
	}
	if errorCode := resp.Header.Get("Nats-Service-Error-Code"); errorCode != "" {
		errorMessage := resp.Header.Get("Nats-Service-Error")
		return nil, fmt.Errorf("%s: %s", errorCode, errorMessage)
	}

	slog.Debug("run tool", "nats response", string(resp.Data))
	return resp.Data, nil
}

// runScriptPlugin runs a script plugin and returns its output.
func runScriptPlugin(ctx context.Context, nc *nats.Conn, tl *Tool) ([]byte, error) {
	if tl.Arguments == nil {
		return nil, ErrScriptValidate.Wrap(fmt.Errorf("script name is not set"))
	}

	script, ok := tl.Arguments["script"].(string)
	if !ok {
		return nil, ErrScriptValidate.Wrap(fmt.Errorf("script name is not set"))
	}

	resp, err := nc.RequestMsgWithContext(ctx, &nats.Msg{
		Subject: "script.run",
		Data:    []byte(script),
		Header:  tracer.HeadersFromContext(ctx),
	})
	if err != nil {
		return nil, ErrRunScript.Wrap(fmt.Errorf("script run failed: %w", err))
	}
	if errorCode := resp.Header.Get("Nats-Service-Error-Code"); errorCode != "" {
		errorMessage := resp.Header.Get("Nats-Service-Error")
		return nil, ErrRunScript.Wrap(fmt.Errorf("script error: %s", errorMessage))
	}

	slog.Debug("run tool", "script output", string(resp.Data))
	return resp.Data, nil
}
