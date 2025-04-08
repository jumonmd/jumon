// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package tracer

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

type Traceparent struct {
	Version   string
	TraceID   string
	ParentID  string
	TraceFlag string
}

func (tp *Traceparent) String() string {
	return fmt.Sprintf("%s-%s-%s-%s", tp.Version, tp.TraceID, tp.ParentID, tp.TraceFlag)
}

func newTraceParent(trace bool) *Traceparent {
	traceflag := "00"
	if trace {
		traceflag = "01"
	}

	return &Traceparent{
		Version:   "00",
		TraceID:   hex.EncodeToString(generateRandomBytes(16)),
		ParentID:  hex.EncodeToString(generateRandomBytes(8)),
		TraceFlag: traceflag,
	}
}

func extractTraceparent(traceparent string) (*Traceparent, error) {
	traceparent = strings.TrimSpace(traceparent)

	parts := strings.Split(traceparent, "-")
	if len(parts) != 4 || len(parts[0]) != 2 || len(parts[1]) != 32 || len(parts[2]) != 16 || len(parts[3]) != 2 {
		return nil, fmt.Errorf("invalid traceparent format: expected 4 parts, got %d", len(parts))
	}

	return &Traceparent{
		Version:   parts[0],
		TraceID:   parts[1],
		ParentID:  parts[2],
		TraceFlag: parts[3],
	}, nil
}

func newSpanID() string {
	return hex.EncodeToString(generateRandomBytes(8))
}

func generateRandomBytes(length int) []byte {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return bytes
}
