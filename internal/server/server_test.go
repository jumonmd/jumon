// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package server

import (
	"testing"
)

func TestIsRunning(t *testing.T) {
	running, err := IsRunning()
	if err != nil && running {
		t.Errorf("unexpected result: got running=%v with err=%v", running, err)
	}
}

func TestSetupMetrics(t *testing.T) {
	// Test with telemetry disabled
	mc, err := setupMetrics(nil, true)
	if err != nil {
		t.Errorf("setupMetrics failed: %v", err)
	}
	if mc != nil {
		t.Errorf("expected nil metrics client when telemetry disabled, got %v", mc)
	}
}
