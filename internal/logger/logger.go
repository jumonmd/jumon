// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/nats-io/nats.go"
)

// NatsHandler wraps a slog.Handler to add context values to log records.
type NatsHandler struct {
	handler slog.Handler
	nc      *nats.Conn
}

func New(nc *nats.Conn, debug bool, isClient bool) *NatsHandler {
	loglevel := slog.LevelInfo
	if debug {
		loglevel = slog.LevelDebug
	}
	slogh := tint.NewHandler(os.Stderr, &tint.Options{
		Level:      loglevel,
		TimeFormat: time.TimeOnly,
		AddSource:  loglevel == slog.LevelDebug,
	})
	if isClient {
		// for client, no logger output by default
		w := io.Discard
		if debug {
			w = os.Stderr
		}
		slogh = tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.TimeOnly,
			AddSource:  true,
		})
	}

	return NewHandler(slogh, nc)
}

func NewHandler(h slog.Handler, nc *nats.Conn) *NatsHandler {
	handler := &NatsHandler{handler: h, nc: nc}
	return handler
}

func (h *NatsHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs := getAttrsFromContext(ctx); attrs != nil {
		for k, v := range attrs {
			r.AddAttrs(slog.Any(k, v))
		}
	}

	return h.handler.Handle(ctx, r)
}

func (h *NatsHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &NatsHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *NatsHandler) WithGroup(name string) slog.Handler {
	return &NatsHandler{handler: h.handler.WithGroup(name)}
}

func (h *NatsHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// getAttrs retrieves the attribute map from context.
func getAttrsFromContext(ctx context.Context) map[string]any {
	if ctx == nil {
		return nil
	}

	attrs := map[string]any{}
	if traceparent, ok := ctx.Value("traceparent").(string); ok {
		attrs["traceparent"] = traceparent
	}

	return attrs
}
