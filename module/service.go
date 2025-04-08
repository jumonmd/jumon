// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package module

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jumonmd/jumon/internal/errors"
	"github.com/jumonmd/jumon/internal/tracer"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

var (
	// ErrValidateModule is returned when module validation fails.
	ErrValidateModule = errors.New(400400, "validate module failed")
	// ErrModuleNotFound is returned when a requested module doesn't exist.
	ErrModuleNotFound = errors.New(404400, "module not found")
	// ErrScriptNotFound is returned when a script within a module doesn't exist.
	ErrScriptNotFound = errors.New(404401, "script not found")
	// ErrRunModule is returned when module execution fails.
	ErrRunModule = errors.New(500400, "run module failed")
)

// NewService creates a NATS microservice that handles module operations.
// It listens on the "module.>" subject pattern.
func NewService(nc *nats.Conn) (micro.Service, error) {
	svc, err := micro.AddService(nc, micro.Config{
		Name:        "jumon_module",
		Version:     "0.1.0",
		Description: `jumon module service`,
		Endpoint: &micro.EndpointConfig{
			Subject: "module.>",
			Handler: micro.HandlerFunc(func(r micro.Request) {
				if strings.HasPrefix(r.Subject(), "module.run") {
					go runHandler(nc, r)
				}
				if strings.HasPrefix(r.Subject(), "module.put") {
					go putHandler(nc, r)
				}
			}),
		},
	})
	if err != nil {
		slog.Error("module service", "status", "create service failed", "error", err)
		return nil, fmt.Errorf("create module service: %w", err)
	}

	slog.Info("module service", "status", "started")
	return svc, nil
}

// runHandler run given module with input.
// it returns the output of the module.
func runHandler(nc *nats.Conn, r micro.Request) {
	modurl := strings.TrimPrefix(r.Subject(), "module.run.")
	if modurl == "" {
		r.Error(ErrModuleNotFound.ServiceError(fmt.Errorf("module url is empty")))
		return
	}
	slog.Info("module.run", "status", "started", "modurl", modurl)

	ctx, span := tracer.Start(tracer.NewContext(r.Headers()), nc, "module.run")
	defer span.End()

	resp, err := Run(ctx, nc, modurl, r.Data())
	if err != nil {
		span.SetError(ErrRunModule.Wrap(err))
		r.Error(ErrRunModule.ServiceError(err))
		return
	}

	r.Respond(resp, micro.WithHeaders(r.Headers()))
	slog.Info("module.run", "status", "finished", "modurl", modurl)
}

func putHandler(nc *nats.Conn, r micro.Request) {
	slog.Info("module.put", "status", "started")
	modurl := strings.TrimPrefix(r.Subject(), "module.put.")
	if modurl == "" {
		r.Error(ErrModuleNotFound.ServiceError(fmt.Errorf("module url is empty")))
		return
	}
	ctx, cancel := context.WithTimeout(tracer.NewContext(r.Headers()), 5*time.Second)
	defer cancel()

	modkv, err := keyvalue(ctx, nc)
	if err != nil {
		r.Error(ErrModuleNotFound.ServiceError(fmt.Errorf("get keyvalue: %w", err)))
		return
	}

	mod, err := ParseMarkdown(r.Data())
	if err != nil {
		r.Error(ErrModuleNotFound.ServiceError(fmt.Errorf("parse module: %w", err)))
		return
	}

	slog.Info("module.put", "status", "parsed", "mod", mod.Name, "scripts", len(mod.Scripts))

	// data, err := json.Marshal(mod)
	// if err != nil {
	// 	r.Error(ErrModuleNotFound.ServiceError(fmt.Errorf("marshal module: %w", err)))
	// 	return
	// }
	_, err = modkv.Put(ctx, mod.Name, r.Data())
	if err != nil {
		r.Error(ErrModuleNotFound.ServiceError(fmt.Errorf("put module: %w", err)))
		return
	}
	r.Respond(nil, micro.WithHeaders(r.Headers()))
	slog.Info("module.put", "status", "finished", "modurl", modurl)
}
