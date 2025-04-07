// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package tracer

import (
	"encoding/json"
	"testing"
)

func TestConvertToString(t *testing.T) {
	tests := []struct {
		name string
		data any
		want string
	}{
		{
			name: "string input",
			data: "hello world",
			want: "hello world",
		},
		{
			name: "valid utf8 bytes",
			data: []byte("hello world"),
			want: "hello world",
		},
		{
			name: "valid utf8 bytes json",
			data: []byte(`{"key":"value"}`),
			want: `{"key":"value"}`,
		},
		{
			name: "json.RawMessage",
			data: json.RawMessage(`{"key":"value"}`),
			want: `{"key":"value"}`,
		},
		{
			name: "struct",
			data: struct {
				Key string `json:"key"`
			}{
				Key: "value",
			},
			want: `{"key":"value"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToString(tt.data); got != tt.want {
				t.Errorf("convertToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
