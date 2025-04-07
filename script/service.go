// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package script

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jumonmd/jumon/internal/errors"
	"github.com/jumonmd/jumon/internal/tracer"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

var (
	ErrValidateScript = errors.New(400300, "validate script failed")
	ErrRunScript      = errors.New(500300, "run script failed")
)

// NewService creates a new script service.
// subject: script.run
func NewService(nc *nats.Conn) (micro.Service, error) {
	svc, err := micro.AddService(nc, micro.Config{
		Name:        "jumon_script",
		Version:     "0.1.0",
		Description: `jumon script service`,
		QueueGroup:  "script",
	})
	if err != nil {
		slog.Error("script service", "status", "create service failed", "error", err)
		return nil, fmt.Errorf("create script service: %w", err)
	}

	scriptGroup := svc.AddGroup("script")
	scriptGroup.AddEndpoint("run", micro.HandlerFunc(func(r micro.Request) {
		go runHandler(nc, r)
	}))

	slog.Info("script service", "status", "started")
	return svc, nil
}

func runHandler(nc *nats.Conn, r micro.Request) {
	slog.Info("script.run", "status", "started")

	ctx, span := tracer.Start(tracer.NewContext(r.Headers()), nc, "script.run")
	defer span.End()

	scr := &Script{}
	err := json.Unmarshal(r.Data(), scr)
	if err != nil {
		span.SetError(ErrValidateScript.Wrap(err))
		r.Error(ErrValidateScript.ServiceError(err))
		return
	}

	err = scr.Validate()
	if err != nil {
		span.SetError(ErrValidateScript.Wrap(err))
		r.Error(ErrValidateScript.ServiceError(err))
		return
	}

	span.SetRequest(scr)
	timeoutSeconds := scr.Config.TimeoutSeconds
	if timeoutSeconds == 0 {
		timeoutSeconds = defaultTimeoutSeconds
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	resp, err := Run(ctx, nc, scr)
	if err != nil {
		span.SetError(ErrRunScript.Wrap(err))
		r.Error(ErrRunScript.ServiceError(err))
		return
	}
	span.SetResponse(resp)

	r.RespondJSON(resp, micro.WithHeaders(span.Headers()))
	slog.Info("script.run", "status", "finished")
}
