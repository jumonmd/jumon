// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package tracer

import (
	"encoding/json"
	"time"
)

// SpanKind represents the role of a Span.
type SpanKind int

const (
	// SpanKindUnspecified is an unspecified SpanKind and is not a valid
	// SpanKind. SpanKindUnspecified should be replaced with SpanKindInternal
	// if it is received.
	SpanKindUnspecified SpanKind = 0
	// SpanKindInternal is a SpanKind for a Span that represents an internal
	// operation within an application.
	SpanKindInternal SpanKind = 1
	// SpanKindServer is a SpanKind for a Span that represents the operation
	// of handling a request from a client.
	SpanKindServer SpanKind = 2
	// SpanKindClient is a SpanKind for a Span that represents the operation
	// of client making a request to a server.
	SpanKindClient SpanKind = 3
	// SpanKindProducer is a SpanKind for a Span that represents the operation
	// of a producer sending a message to a message broker.
	SpanKindProducer SpanKind = 4
	// SpanKindConsumer is a SpanKind for a Span that represents the operation
	// of a consumer receiving a message from a message broker.
	SpanKindConsumer SpanKind = 5
)

// ValidateSpanKind returns a valid span kind value.
func ValidateSpanKind(spanKind SpanKind) SpanKind {
	switch spanKind {
	case SpanKindUnspecified,
		SpanKindInternal,
		SpanKindServer,
		SpanKindClient,
		SpanKindProducer,
		SpanKindConsumer:
		return spanKind
	default:
		return SpanKindInternal
	}
}

// String returns the specified name of the SpanKind in lower-case.
func (sk SpanKind) String() string {
	switch sk {
	case SpanKindInternal:
		return "internal"
	case SpanKindServer:
		return "server"
	case SpanKindClient:
		return "client"
	case SpanKindProducer:
		return "producer"
	case SpanKindConsumer:
		return "consumer"
	case SpanKindUnspecified:
		return "unspecified"
	default:
		return "unspecified"
	}
}

type Status int32

const (
	StatusUnset Status = 0
	StatusOK    Status = 1
	StatusError Status = 2
)

// Span is based on OpenTelemetry Span.
type Span struct {
	TraceID    string            `json:"trace_id"`
	SpanID     string            `json:"span_id"`
	ParentID   string            `json:"parent_id"`
	Name       string            `json:"name"`
	Kind       SpanKind          `json:"kind"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
	Status     Status            `json:"status"`
	StatusCode int               `json:"status_code"`
	Attributes map[string]string `json:"attributes"`
}

func NewSpan(traceID, parentID, name string, kind SpanKind) *Span {
	return &Span{
		TraceID:    traceID,
		ParentID:   parentID,
		SpanID:     newSpanID(),
		Name:       name,
		Kind:       kind,
		StartTime:  time.Now(),
		Status:     StatusUnset,
		Attributes: make(map[string]string),
	}
}

func (s *Span) JSON() []byte {
	data, err := json.Marshal(s)
	if err != nil {
		return nil
	}
	return data
}

// End completes the span.
func (s *Span) End() {
	s.EndTime = time.Now()
}

func (s *Span) IsEnd() bool {
	return !s.EndTime.IsZero()
}

func (s *Span) SetStatusOK() {
	s.Status = StatusOK
}

func (s *Span) SetStatusError() {
	s.Status = StatusError
}

func (s *Span) SetAttribute(key, value string) {
	s.Attributes[key] = value
}
