// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package module

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jumonmd/gengo/jsonschema"
	"github.com/jumonmd/jumon/script"
	"github.com/jumonmd/jumon/tool"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func TestDetectSection(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		want     string
	}{
		{
			name:     "scripts section",
			markdown: "## Scripts",
			want:     "Scripts",
		},
		{
			name:     "tools section",
			markdown: "## Tools",
			want:     "Tools",
		},
		{
			name:     "events section",
			markdown: "## Events",
			want:     "Events",
		},
		{
			name:     "other section",
			markdown: "## Other",
			want:     "",
		},
		{
			name:     "level 3 heading",
			markdown: "### Scripts",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := goldmark.New().Parser()
			reader := text.NewReader([]byte(tt.markdown))
			doc := parser.Parse(reader)

			var node ast.Node
			// Get the first node (should be the heading)
			err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if entering && n.Kind() == ast.KindHeading {
					node = n
					return ast.WalkStop, nil
				}
				return ast.WalkContinue, nil
			})
			if err != nil {
				t.Fatalf("ast.Walk() error = %v", err)
			}
			got := detectSection(node, reader)
			if got != tt.want {
				t.Errorf("detectSection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		testFile string
		want     *Module
	}{
		{
			name:     "basic",
			testFile: "basic.md",
			want: &Module{
				Name: "basic",
				Scripts: []*script.Script{
					{Name: "ModuleA"},
					{Name: "ModuleB"},
				},
			},
		},
		{
			name:     "scripts",
			testFile: "scripts.md",
			want: &Module{
				Name: "",
				Scripts: []*script.Script{
					{Name: "ScriptA", Content: "- do something"},
					{Name: "ScriptB", Content: "- do something else\nfirst line\nsecond line"},
				},
			},
		},
		{
			name:     "tools",
			testFile: "tools.md",
			want: &Module{
				Tools: []tool.Tool{
					{Name: "ToolA", Description: "Tool A description"},
					{Name: "ToolB", Description: "Tool B description"},
				},
			},
		},
		{
			name:     "import",
			testFile: "import.md",
			want: &Module{
				Name: "import",
				Tools: []tool.Tool{
					{Name: "get_weather", Module: "anothermodule"},
				},
			},
		},
		{
			name:     "schemas",
			testFile: "schema.md",
			want: &Module{
				Name: "schema",
				Schemas: map[string]jsonschema.Schema{
					"main.input": {
						"type": "object",
						"properties": map[string]any{
							"query": map[string]any{
								"type": "string",
							},
						},
						"required": []any{"query"},
					},
					"main.output": {
						"type": "object",
						"properties": map[string]any{
							"result": map[string]any{
								"type": "string",
							},
						},
						"required": []any{"result"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read test file
			data, err := os.ReadFile(filepath.Join("testdata", tt.testFile))
			if err != nil {
				t.Fatalf("failed to read test file: %v", err)
			}

			got, err := ParseMarkdown(data)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			opts := cmp.Options{
				cmp.FilterPath(func(p cmp.Path) bool {
					return strings.HasSuffix(p.String(), "Metadata")
				}, cmp.Ignore()),
			}

			if diff := cmp.Diff(tt.want, got, opts); diff != "" {
				t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetCodeBlock(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantLang    string
		wantContent string
	}{
		{
			name:        "simple code block",
			input:       "```go\npackage main\n\nfunc main() {}\n```",
			wantLang:    "go",
			wantContent: "package main\n\nfunc main() {}",
		},
		{
			name:        "code block without language",
			input:       "```\nsome code\n```",
			wantLang:    "",
			wantContent: "some code",
		},
		{
			name:        "empty code block",
			input:       "```python\n```",
			wantLang:    "python",
			wantContent: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLang, gotContent, err := getCodeBlock([]byte(tt.input))
			if err != nil {
				t.Fatalf("getCodeBlock() error = %v", err)
			}
			if diff := cmp.Diff(tt.wantLang, gotLang); diff != "" {
				t.Errorf("language mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantContent, gotContent); diff != "" {
				t.Errorf("content mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetNodeHeadingContent(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		want     string
	}{
		{
			name: "basic text",
			markdown: `### Heading
This is content
More content`,
			want: "This is content\nMore content",
		},
		{
			name: "until next heading",
			markdown: `### Heading
Content 1
Content 2

### Next Heading
Should not include this`,
			want: "Content 1\nContent 2",
		},
		{
			name: "with code block",
			markdown: `### Heading
Some text
` + "```json" + `
{
  "key": "value"
}
` + "```" + `
More text`,
			want: `Some text
` + "```json" + `
{
  "key": "value"
}
` + "```" + `
More text`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := goldmark.New().Parser()
			reader := text.NewReader([]byte(tt.markdown))
			doc := parser.Parse(reader)

			var heading ast.Node
			err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if entering {
					if h, ok := n.(*ast.Heading); ok {
						heading = h
						return ast.WalkStop, nil
					}
				}
				return ast.WalkContinue, nil
			})
			if err != nil {
				t.Fatalf("ast.Walk() error = %v", err)
			}

			got := getNodeHeadingContent(heading, reader)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getNodeHeadingContent() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
