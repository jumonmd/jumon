// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package module

import (
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/jumonmd/gengo/chat"
	"github.com/jumonmd/jumon/internal/testutil"
	"github.com/nats-io/nats.go/jetstream"
)

func TestModuleService(t *testing.T) {
	// Set slog to debug level with debug handler
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	// setup test server
	nc, js, _, cleanup, err := testutil.NewNATSServer()
	if err != nil {
		t.Fatalf("failed to start NATS server: %v", err)
	}
	defer cleanup()

	// setup kv
	kv, err := js.CreateKeyValue(t.Context(), jetstream.KeyValueConfig{
		Bucket: "module",
	})
	if err != nil {
		t.Fatalf("failed to create kv: %v", err)
	}

	_, err = js.CreateKeyValue(t.Context(), jetstream.KeyValueConfig{
		Bucket: "config",
	})
	if err != nil {
		t.Fatalf("failed to create kv: %v", err)
	}

	// setup service
	svc, err := NewService(nc)
	if err != nil {
		t.Fatalf("failed to create module service: %v", err)
	}
	defer svc.Stop()

	testresp := chat.Response{
		Model:        "gpt-4o-mini",
		FinishReason: "stop",
		Messages:     []chat.Message{chat.NewTextMessage(chat.MessageRoleAI, "hello")},
	}
	respdata, err := json.Marshal(testresp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}
	// setup test chat service
	chtsvc, err := testutil.NewMicroServer(nc, "chat.generate", respdata)
	if err != nil {
		t.Fatalf("failed to create mock script service: %v", err)
	}
	defer chtsvc.Stop()

	// put module to kv
	modmd := `
---
module: test/module
---

# test module
## Scripts
### main
1. say hello
`
	_, err = kv.Put(t.Context(), "test/module", []byte(modmd))
	if err != nil {
		t.Fatalf("failed to put module: %v", err)
	}

	// run module
	result, err := Run(t.Context(), nc, "test/module", nil)
	if err != nil {
		t.Fatalf("failed to run module: %v", err)
	}

	if string(result) != `"hello"` {
		t.Fatalf("expected hello, got %v", string(result))
	}
}
