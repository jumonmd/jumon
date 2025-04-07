// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"testing"
	"time"

	"github.com/jumonmd/jumon/internal/testutil"
	"github.com/jumonmd/jumon/internal/tracer"
	"github.com/jumonmd/jumon/module"
)

func TestPrepareContext(t *testing.T) {
	// Test with valid timeout value
	t.Run("valid timeout", func(t *testing.T) {
		cfg := &Config{
			RunTimeout: "5s",
		}

		ctx, cancel := notifyContext(cfg)
		defer cancel()

		// Verify context has a deadline
		deadline, hasDeadline := ctx.Deadline()
		if !hasDeadline {
			t.Error("Context should have a deadline")
		}

		// Calculate duration from now
		now := time.Now()
		duration := deadline.Sub(now)

		// Log the timeout duration
		t.Logf("Timeout duration: %v", duration)

		expected, _ := time.ParseDuration("5s")
		if duration < expected-time.Second || duration > expected {
			t.Errorf("Expected timeout around %v, got %v", expected, duration)
		}
	})

	// Test with invalid timeout value
	t.Run("invalid timeout", func(t *testing.T) {
		cfg := &Config{
			RunTimeout: "invalid",
		}

		ctx, cancel := notifyContext(cfg)
		defer cancel()

		// Verify context has a deadline
		deadline, hasDeadline := ctx.Deadline()
		if !hasDeadline {
			t.Error("Context should have a deadline")
		}

		// Calculate duration from now
		now := time.Now()
		duration := deadline.Sub(now)

		// Log the timeout duration
		t.Logf("Timeout duration: %v", duration)

		// For invalid timeout, default value (5 minutes) should be used
		if duration < 4*time.Minute || duration > 5*time.Minute {
			t.Errorf("Expected timeout around 5 minutes, got %v", duration)
		}
	})
}

func TestExecuteModule(t *testing.T) {
	// Setup NATS server for testing
	nc, _, _, cleanup, err := testutil.NewNATSServer()
	if err != nil {
		t.Fatalf("Failed to setup NATS server: %v", err)
	}
	defer cleanup()

	// Setup a micro service for testing
	resp := []byte("Success response")
	svc, err := testutil.NewMicroServer(nc, "module.run.test-module", resp)
	if err != nil {
		t.Fatalf("Failed to setup micro server: %v", err)
	}
	defer svc.Stop()

	// Create test module
	mod := &module.Module{
		Name: "test-module",
	}

	// Create notification context
	ctx, _ := tracer.CreateNotifyContext()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Execute module
	err = executeModule(ctx, nc, mod, nil)
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}
}

func TestSetupNotification(t *testing.T) {
	// Setup NATS server for testing
	nc, _, _, cleanup, err := testutil.NewNATSServer()
	if err != nil {
		t.Fatalf("Failed to setup NATS server: %v", err)
	}
	defer cleanup()

	// Create notification context
	ctx, _ := tracer.CreateNotifyContext()

	// Setup notification
	err = subscribeNotification(ctx, nc)
	if err != nil {
		t.Errorf("Failed to setup notification: %v", err)
	}

	// It's difficult to verify the subscription is unsubscribed when the context is canceled,
	// so we only check that no error is returned
}
