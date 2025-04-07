// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package module

import (
	"testing"
)

func TestGetVCSPath(t *testing.T) {
	tests := []struct {
		name     string
		module   string
		wantRepo string
		wantPath string
		wantErr  bool
	}{
		{
			name:     "github with path",
			module:   "github.com/golang/go/doc",
			wantRepo: "https://github.com/golang/go",
			wantPath: "doc",
			wantErr:  false,
		},
		{
			name:     "github root",
			module:   "github.com/golang/go",
			wantRepo: "https://github.com/golang/go",
			wantPath: "",
			wantErr:  false,
		},
		{
			name:     "gitlab",
			module:   "gitlab.com/randaalex/gitlab-ce/doc",
			wantRepo: "https://gitlab.com/randaalex/gitlab-ce",
			wantPath: "doc",
			wantErr:  false,
		},
		{
			name:     "invalid module",
			module:   "invalid/module/path",
			wantRepo: "",
			wantPath: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, path, err := getVCSPath(tt.module)
			if (err != nil) != tt.wantErr {
				t.Fatalf("getVCSPath() error = %v, wantErr %v", err, tt.wantErr)
			}
			if repo != tt.wantRepo {
				t.Errorf("getVCSPath() repo = %v, want %v", repo, tt.wantRepo)
			}
			if path != tt.wantPath {
				t.Errorf("getVCSPath() path = %v, want %v", path, tt.wantPath)
			}
		})
	}
}
