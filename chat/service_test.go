// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package chat

import (
	"log"
	"strings"
	"testing"

	"github.com/jumonmd/gengo/chat"
	"github.com/jumonmd/jumon/internal/testutil"
)

func TestChatService(t *testing.T) {
	// setup test server
	nc, _, _, cleanup, err := testutil.NewNATSServer()
	if err != nil {
		t.Fatalf("failed to start NATS server: %v", err)
	}
	defer cleanup()

	// setup service
	svc, err := NewService(nc)
	if err != nil {
		t.Fatalf("failed to create prompt service: %v", err)
	}
	defer svc.Stop()

	// setup test llm
	testllm := testutil.NewMockOpenAIServer()
	defer testllm.Close()

	log.Println("testllm.URL", testllm.URL)

	// request
	req := &chat.Request{
		Model:    "gpt-4o-mini",
		Messages: []chat.Message{chat.NewTextMessage(chat.MessageRoleHuman, "say hello")},
	}

	resp, err := Generate(t.Context(), nc, req, chat.WithBaseURL(testllm.URL))
	if err != nil {
		t.Fatalf("failed to generate: %v", err)
	}

	if !strings.Contains(resp.Messages[0].ContentString(), "ello") {
		t.Fatalf("expected %q, got %q", "hello", resp.Messages[0].ContentString())
	}
}
