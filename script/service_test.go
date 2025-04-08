// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package script

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/jumonmd/gengo/chat"
	"github.com/jumonmd/jumon/internal/testutil"
)

func TestScriptService(t *testing.T) {
	// setup test server
	nc, _, _, cleanup, err := testutil.NewNATSServer()
	if err != nil {
		t.Fatalf("failed to start NATS server: %v", err)
	}
	defer cleanup()

	// setup script service
	svc, err := NewService(nc)
	if err != nil {
		t.Fatalf("failed to create script service: %v", err)
	}
	defer svc.Stop()

	// setup test chat service
	testresp := chat.Response{
		Model:        "gpt-4o-mini",
		FinishReason: "stop",
		Messages:     []chat.Message{chat.NewTextMessage(chat.MessageRoleAI, "hello")},
	}
	respdata, err := json.Marshal(testresp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}
	chtsvc, err := testutil.NewMicroServer(nc, "chat.generate", respdata)
	if err != nil {
		t.Fatalf("failed to create test service: %v", err)
	}
	defer chtsvc.Stop()

	tests := []struct {
		name     string
		script   *Script
		expected []byte
	}{
		{
			name: "simple step",
			script: &Script{
				Name:        "test-script",
				Description: "Test script for testing",
				Model:       "gpt-4o-mini",
				Content:     "1. Say hello",
			},
			expected: []byte(`"hello"`),
		},
		{
			name: "no steps",
			script: &Script{
				Name:        "multi-step-script",
				Description: "Test script with multiple steps",
				Model:       "gpt-4o-mini",
				Content:     "Say hello",
			},
			expected: []byte(`"hello"`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := Run(t.Context(), nc, tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !bytes.Equal(resp, tt.expected) {
				t.Errorf("expected %q, got %q", tt.expected, resp)
			}
		})
	}
}
