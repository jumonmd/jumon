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

	var out string
	if err := json.Unmarshal(got, &out); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}
	want := `{"buz":"bar"}`
	if out != want {
		t.Errorf("got %s, want %s", out, want)
	}
}
