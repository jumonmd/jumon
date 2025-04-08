// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package tracer

import (
	"context"
	"encoding/hex"
	"log/slog"

	"github.com/nats-io/nats.go"
)

type Attributes map[string]any

type ContextKey string

const (
	ContextKeyTraceParent ContextKey = "traceparent"
	ContextKeyNotifyTo    ContextKey = "notify-to"
)

type Headers interface {
	Get(name string) string
}

// NewContext creates a new context with the traceparent and notify-to headers.
func NewContext(h Headers) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, ContextKeyTraceParent, h.Get("traceparent"))
	ctx = context.WithValue(ctx, ContextKeyNotifyTo, h.Get("notify-to"))
	slog.Debug("new context", "traceparent", h.Get("traceparent"), "notify-to", h.Get("notify-to"))
	return ctx
}

// HeadersFromContext creates a new nats.Header from the context.
func HeadersFromContext(ctx context.Context) nats.Header {
	return nats.Header{
		"traceparent": {ContextValueTraceParent(ctx)},
		"notify-to":   {ContextValueNotifyTo(ctx)},
	}
}

// CreateNotifyContext creates a new context with a random notify-to value.
// It returns the context and the notify-to value.
func CreateNotifyContext() (context.Context, string) {
	notifyTo := hex.EncodeToString(generateRandomBytes(8))
	return context.WithValue(context.Background(), ContextKeyNotifyTo, notifyTo), notifyTo
}

func ContextValueNotifyTo(ctx context.Context) string {
	notifyTo, ok := ctx.Value(ContextKeyNotifyTo).(string)
	if !ok {
		return ""
	}
	return notifyTo
}

func ContextValueTraceParent(ctx context.Context) string {
	tp, ok := ctx.Value(ContextKeyTraceParent).(string)
	if !ok {
		return ""
	}
	return tp
}
