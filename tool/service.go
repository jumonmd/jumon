// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package tool

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jumonmd/jumon/internal/errors"
	"github.com/jumonmd/jumon/internal/tracer"
	"github.com/jumonmd/jumon/tool/std"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nats.go/micro"
)

var (
	ErrToolValidate    = errors.New(400200, "tool validation failed")
	ErrUnknownToolType = errors.New(400201, "unknown tool type")
	ErrWasmValidate    = errors.New(400202, "wasm validation failed")
	ErrNatsValidate    = errors.New(400203, "nats validation failed")
	ErrScriptValidate  = errors.New(400204, "script validation failed")

	ErrLoadResources = errors.New(500200, "load resources failed")
	ErrRunTool       = errors.New(500201, "tool execution failed")
	ErrRunWasm       = errors.New(500202, "wasm execution failed")
	ErrRunNats       = errors.New(500203, "nats execution failed")
	ErrRunScript     = errors.New(500204, "script execution failed")
)

// NewService creates a new tool service.
// subject: tool.>
func NewService(nc *nats.Conn, obs jetstream.ObjectStore) (micro.Service, error) {
	svc, err := micro.AddService(nc, micro.Config{
		Name:        "jumon_tool",
		Version:     "0.1.0",
		Description: "jumon tool service",
		QueueGroup:  "tool",
	})
	if err != nil {
		slog.Error("tool service", "status", "create service failed", "error", err)
		return nil, fmt.Errorf("create tool service: %w", err)
	}

	toolGroup := svc.AddGroup("tool")
	toolGroup.AddEndpoint("run", micro.HandlerFunc(func(r micro.Request) {
		go runHandler(nc, obs, r)
	}))

	std.New(toolGroup)

	slog.Info("tool service", "status", "started")
	return svc, nil
}

// runHandler run given tool with input.
// cache is for loading resources.
// it returns the output of the tool.
func runHandler(nc *nats.Conn, obs jetstream.ObjectStore, r micro.Request) {
	slog.Info("tool.run", "status", "started")

	ctx, span := tracer.Start(tracer.NewContext(r.Headers()), nc, "tool.run")
	defer span.End()

	tl := &Tool{}
	err := json.Unmarshal(r.Data(), tl)
	if err != nil {
		span.SetError(ErrToolValidate.Wrap(err))
		r.Error(ErrToolValidate.ServiceError(err))
		return
	}

	span.SetRequest(tl)

	output, err := run(ctx, tl, nc, obs)
	if err != nil {
		span.SetError(err)
		r.Error(ErrRunTool.ServiceError(err))
		return
	}

	span.SetResponse(output)
	slog.Info("tool.run", "status", "finished")
	r.Respond(output, micro.WithHeaders(span.Headers()))
}
