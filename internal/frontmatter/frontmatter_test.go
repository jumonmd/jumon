// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package frontmatter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type testData struct {
	Title       string
	Description string
	Tags        []string
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     testData
	}{
		{
			name:     "yaml frontmatter",
			filePath: "testdata/yaml.md",
			want: testData{
				Title:       "Test Title",
				Description: "Test Description",
				Tags:        []string{"test", "frontmatter"},
			},
		},
		{
			name:     "toml frontmatter",
			filePath: "testdata/toml.md",
			want: testData{
				Title:       "Test Title",
				Description: "Test Description",
				Tags:        []string{"test", "frontmatter"},
			},
		},
		{
			name:     "no frontmatter",
			filePath: "testdata/plain.md",
			want:     testData{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(".", tt.filePath))
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", tt.filePath, err)
			}

			var got testData
			content, err := Unmarshal(data, &got)
			if err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Unmarshal() mismatch (-want +got):\n%s", diff)
			}

			if len(content) == 0 {
				t.Errorf("Unmarshal() returned empty content")
			}
		})
	}
}
