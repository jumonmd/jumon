// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0
package event

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestFormatJSONNoTemplate(t *testing.T) {
	input := []byte("\n```json\n{\"foo\":\"bar\"}\n```\n")
	got, err := FormatJSON(input, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := json.RawMessage(`{"foo":"bar"}`)
	if !bytes.Equal(got, want) {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestFormatJSONWithTemplate(t *testing.T) {
	input := []byte(`{"foo":"bar"}`)
	got, err := FormatJSON(input, `{"buz":"{{.foo}}"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := json.RawMessage(`{"buz":"bar"}`)
	if !bytes.Equal(got, want) {
		t.Errorf("got %s, want %s", got, want)
	}
}
