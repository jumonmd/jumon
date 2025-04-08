// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package tool

import (
	"errors"
	"testing"

	"github.com/jumonmd/jumon/internal/testutil"
)

func TestToolService(t *testing.T) {
	// setup test server
	nc, _, obs, cleanup, err := testutil.NewNATSServer()
	if err != nil {
		t.Fatalf("failed to start NATS server: %v", err)
	}
	defer cleanup()

	// setup tool service
	svc, err := NewService(nc, obs)
	if err != nil {
		t.Fatalf("failed to create prompt service: %v", err)
	}
	defer svc.Stop()

	// setup test service
	tsvc, err := testutil.NewMicroServer(nc, "test.>", []byte("hello"))
	if err != nil {
		t.Fatalf("failed to create test service: %v", err)
	}
	defer tsvc.Stop()

	tests := []struct {
		name     string
		tool     Tool
		expected []byte
		err      error
	}{
		{
			name: "nats test",
			tool: Tool{
				Name:      "nats-test",
				Type:      "nats",
				Arguments: Arguments{"subject": "test.hello"},
			},
			expected: []byte("hello"),
			err:      nil,
		},
		{
			name: "wasm error",
			tool: Tool{
				Name: "wasm-test",
				Type: "wasm",
			},
			expected: nil,
			err:      ErrWasmValidate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tool.SetInput([]byte("hello"))
			resp, err := Run(t.Context(), nc, tt.tool)

			if tt.err != nil {
				if err == nil || errors.Is(err, tt.err) {
					t.Errorf("expected error %v, got %v", tt.err, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(resp) != string(tt.expected) {
				t.Errorf("expected %q, got %q", tt.expected, resp)
			}
		})
	}
}
