// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package tracer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nats-io/nats.go"
)

const traceSubject = "trace"

type SpanTracer struct {
	nc *nats.Conn
	// current span
	span *Span
	// nextTransparent is the next transparent passthrough to the next header.
	nextTransparent string
	// notifyTo is the subject to notify the span end event.
	// if empty, the span end event is not notified.
	notifyTo string
}

type Notification struct {
	TraceID  string `json:"trace_id"`
	SpanID   string `json:"span_id"`
	ParentID string `json:"parent_id"`
	// On is the event type.
	//  e.g. "request", "response", "error".
	On      string    `json:"on"`
	Date    time.Time `json:"date"`
	Name    string    `json:"name"`
	Content string    `json:"content"`
}

// Start starts a new span tracer. name is the event name of the span.
// It is also returns the context with the traceparent value.
func Start(ctx context.Context, nc *nats.Conn, name string) (context.Context, *SpanTracer) {
	traceparent := ContextValueTraceParent(ctx)
	notifyTo := ContextValueNotifyTo(ctx)
	tp, err := extractTraceparent(traceparent)
	if err != nil {
		// if traceparent is not provided, create a new one
		tp = newTraceParent(true)
		tp.ParentID = ""
	}
	// set current span
	span := NewSpan(tp.TraceID, tp.ParentID, name, SpanKindInternal)

	tp.ParentID = span.SpanID
	ctx = context.WithValue(ctx, ContextKeyTraceParent, tp.String())
	return ctx, &SpanTracer{nc: nc, span: span, nextTransparent: tp.String(), notifyTo: notifyTo}
}

func (t *SpanTracer) SetRequest(data any) {
	s := convertToString(data)
	t.span.SetAttribute("request", s)
	err := t.Notify(data)
	if err != nil {
		slog.Warn("tracer", "event", "span_end", "error", err)
	}
}

func (t *SpanTracer) SetResponse(data any) {
	if t.span.Status == StatusUnset {
		t.span.Status = StatusOK
	}
	s := convertToString(data)
	t.span.SetAttribute("response", s)
	err := t.Notify(data)
	if err != nil {
		slog.Warn("tracer", "event", "span_end", "error", err)
	}
}

func (t *SpanTracer) SetError(err error) {
	slog.Error("error", "message", err.Error())
	t.span.Status = StatusError
	t.span.SetAttribute("error", err.Error())
	t.span.StatusCode = parseStatuCode(err)
	err = t.Notify(err)
	if err != nil {
		slog.Warn("tracer", "event", "span_end", "error", err)
	}
}

func (t *SpanTracer) End() {
	t.span.End()
	subject := fmt.Sprintf("%s.%s.%s", traceSubject, t.span.TraceID, t.span.SpanID)
	err := t.nc.Publish(subject, t.span.JSON())
	if err != nil {
		slog.Warn("tracer", "event", "span_end", "error", err)
	}
}

func (t *SpanTracer) Headers() map[string][]string {
	return map[string][]string{
		string(ContextKeyTraceParent): {t.nextTransparent},
		string(ContextKeyNotifyTo):    {t.notifyTo},
	}
}

func (t *SpanTracer) NextContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, ContextKeyTraceParent, t.nextTransparent)
	return context.WithValue(ctx, ContextKeyNotifyTo, t.notifyTo)
}

func convertToString(data any) string {
	switch v := data.(type) {
	case string:
		return v
	case []byte:
		if utf8.Valid(v) {
			return string(v)
		}
		js, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(js)
	case json.RawMessage:
		return string(v)
	default:
		js, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(js)
	}
}

func (t *SpanTracer) Notify(data any) error {
	if t.notifyTo == "" {
		return nil
	}

	content := convertToString(data)
	n := Notification{
		TraceID:  t.span.TraceID,
		SpanID:   t.span.SpanID,
		ParentID: t.span.ParentID,
		Date:     time.Now(),
		Name:     t.span.Name,
		Content:  content,
		On:       "request",
	}
	if t.span.Status == StatusOK {
		n.On = "response"
	}
	if t.span.Status == StatusError {
		n.On = "error"
	}
	nd, err := json.Marshal(n)
	if err != nil {
		return err
	}
	slog.Debug("tracer", "notify", string(nd))
	return t.nc.Publish("notification."+t.notifyTo, nd)
}

func parseStatuCode(e error) int {
	parts := strings.Split(e.Error(), ":")
	if len(parts) > 1 {
		p := strings.TrimSpace(parts[0])
		if code, err := strconv.Atoi(p); err == nil {
			return code
		}
	}
	return 0
}
