// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package dataurl

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		mimeType string
		want     string
	}{
		{
			name:     "JSON data",
			data:     []byte(`{"key":"value"}`),
			mimeType: "application/json",
			want:     "data:application/json;base64,eyJrZXkiOiJ2YWx1ZSJ9",
		},
		{
			name:     "empty data",
			data:     []byte{},
			mimeType: "text/plain",
			want:     "data:text/plain;base64,",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Encode(tt.mimeType, tt.data)
			if got != tt.want {
				t.Errorf("encodeDATAURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name      string
		dataURL   string
		wantData  []byte
		wantMime  string
		wantError bool
	}{
		{
			name:     "valid JSON data URL",
			dataURL:  "data:application/json;base64,eyJrZXkiOiJ2YWx1ZSJ9",
			wantData: []byte(`{"key":"value"}`),
			wantMime: "application/json;base64",
		},
		{
			name:      "invalid format - no comma",
			dataURL:   "data:application/json;base64",
			wantError: true,
		},
		{
			name:      "invalid base64",
			dataURL:   "data:application/json;base64,!@#$",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotData, gotMime, err := Decode(tt.dataURL)
			if (err != nil) != tt.wantError {
				t.Errorf("decodeDATAURL() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.wantError {
				return
			}
			if !cmp.Equal(gotData, tt.wantData) {
				t.Errorf("decodeDATAURL() gotData = %v, want %v", gotData, tt.wantData)
			}
			if gotMime != tt.wantMime {
				t.Errorf("decodeDATAURL() gotMime = %v, want %v", gotMime, tt.wantMime)
			}
		})
	}
}
