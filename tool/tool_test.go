// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package tool

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestNewResource(t *testing.T) {
	reader := io.NopCloser(strings.NewReader("test data"))
	r := NewResource("test", "http://example.com", reader)

	want := &Resource{
		Name:   "test",
		URL:    "http://example.com",
		reader: reader,
	}

	if diff := cmp.Diff(want, r, cmpopts.IgnoreUnexported(Resource{})); diff != "" {
		t.Errorf("NewResource() mismatch (-want +got):\n%s", diff)
	}

	data, err := io.ReadAll(r.reader)
	if err != nil {
		t.Errorf("Data() error = %v", err)
	}
	if string(data) != "test data" {
		t.Errorf("Data() = %v, want %v", string(data), "test data")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		tool    Tool
		wantErr bool
	}{
		{
			name: "valid tool",
			tool: Tool{
				Name: "test-tool",
				Type: "script",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			tool: Tool{
				Type: "script",
			},
			wantErr: true,
		},
		{
			name: "missing type",
			tool: Tool{
				Name: "test-tool",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tool.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceFetch(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		expected := "test resource data"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(expected))
		}))
		defer server.Close()

		r := &Resource{URL: server.URL}
		err := r.load(t.Context(), nil)
		if err != nil {
			t.Errorf("Fetch() error = %v", err)
		}

		data, err := io.ReadAll(r.reader)
		if err != nil {
			t.Errorf("Data() error = %v", err)
		}
		if string(data) != expected {
			t.Errorf("Data() = %v, want %v", string(data), expected)
		}
	})

	t.Run("failed fetch - bad status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		r := &Resource{
			URL: server.URL,
		}

		err := r.load(t.Context(), nil)
		if err == nil {
			t.Errorf("Fetch() error = nil, want error")
		}
	})

	t.Run("failed fetch - invalid URL", func(t *testing.T) {
		r := &Resource{
			URL: "invalid://url",
		}

		err := r.load(t.Context(), nil)
		if err == nil {
			t.Errorf("Fetch() error = nil, want error")
		}
	})
}
