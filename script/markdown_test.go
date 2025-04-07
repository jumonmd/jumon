// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package script

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
)

func TestParseSymbols(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "No Symbol",
			content:  "this is atest",
			expected: []string{},
		},
		{
			name:     "Single Symbol",
			content:  "this is a `test` symbol",
			expected: []string{"test"},
		},
		{
			name:     "Empty Symbol",
			content:  "this is a ` ` symbol",
			expected: []string{},
		},
		{
			name:     "Multiple Symbols",
			content:  "this is a `test` symbol\n and `another` symbol",
			expected: []string{"test", "another"},
		},
		{
			name:     "Unparsable Symbol",
			content:  "this is a ```test symbol`",
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := goldmark.New().Parser()
			doc := parser.Parse(text.NewReader([]byte(tc.content)))
			symbols, err := parseSymbols(doc, text.NewReader([]byte(tc.content)))
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if len(symbols) != len(tc.expected) {
				t.Errorf("expected %d symbol(s), got %d", len(tc.expected), len(symbols))
			}
			for i, symbol := range symbols {
				if symbol.Name != tc.expected[i] {
					t.Errorf("expected symbol to be '%s', got '%s'", tc.expected[i], symbol.Name)
				}
			}
		})
	}
}

func TestStepMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple list",
			input:    "- Item 1\n- Item 2\n- Item 3",
			expected: "- Item 1\n- Item 2\n- Item 3\n",
		},
		{
			name:     "nested list",
			input:    "- Parent 1\n  - Child 1.1\n  - Child 1.2\n- Parent 2\n  - Child 2.1",
			expected: "- Parent 1\n  - Child 1.1\n  - Child 1.2\n- Parent 2\n  - Child 2.1\n",
		},
		{
			name:     "deep nested list",
			input:    "- Level 1\n  - Level 2\n    - Level 3\n      - Level 4",
			expected: "- Level 1\n  - Level 2\n    - Level 3\n      - Level 4\n",
		},
		{
			name:     "with emphasis",
			input:    "- Parent 1\n  - **Child 1.1**\n",
			expected: "- Parent 1\n  - **Child 1.1**\n",
		},
		{
			name:     "with mixed formatting",
			input:    "- Item with *emphasis*\n- Item with **strong emphasis**\n- Item with `code`\n- Item with [link](https://example.com)",
			expected: "- Item with *emphasis*\n- Item with **strong emphasis**\n- Item with `code`\n- Item with [link](https://example.com)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := goldmark.New().Parser()
			list, _, err := parseList(parser.Parse(text.NewReader([]byte(tt.input))), text.NewReader([]byte(tt.input)))
			if err != nil {
				t.Fatalf("ParseMarkdown() error = %v", err)
			}

			// Create a virtual root to get complete text
			root := &Step{Level: 0, Children: list.Children}
			result := root.Markdown()

			if result != tt.expected {
				t.Errorf("GetCompleteText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseListWithPreface(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantPreface   string
		wantListItems []string
	}{
		{
			name:        "no preface",
			input:       "- Item 1\n- Item 2",
			wantPreface: "",
			wantListItems: []string{
				"- Item 1",
				"- Item 2",
			},
		},
		{
			name:        "with simple preface",
			input:       "This is a preface.\n- Item 1\n- Item 2",
			wantPreface: "This is a preface.",
			wantListItems: []string{
				"- Item 1",
				"- Item 2",
			},
		},
		{
			name:        "with multi-paragraph preface",
			input:       "First paragraph.\n\nSecond paragraph.\n- Item 1\n- Item 2",
			wantPreface: "First paragraph.\n\nSecond paragraph.",
			wantListItems: []string{
				"- Item 1",
				"- Item 2",
			},
		},
		{
			name:        "with formatted preface",
			input:       "This is a **bold** preface with `code`.\n- Item 1\n- Item 2",
			wantPreface: "This is a **bold** preface with `code`.",
			wantListItems: []string{
				"- Item 1",
				"- Item 2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := goldmark.New().Parser()
			doc := parser.Parse(text.NewReader([]byte(tt.input)))
			root, preface, err := parseList(doc, text.NewReader([]byte(tt.input)))
			if err != nil {
				t.Fatalf("parseList() error = %v", err)
			}

			if diff := cmp.Diff(tt.wantPreface, preface); diff != "" {
				t.Errorf("preface mismatch (-want +got):\n%s", diff)
			}

			var gotItems []string
			for _, child := range root.Children {
				gotItems = append(gotItems, child.Content)
			}

			if diff := cmp.Diff(tt.wantListItems, gotItems); diff != "" {
				t.Errorf("list items mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseChecks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
		{
			name:     "no check items",
			input:    "This is a text with `code span`",
			expected: []string{},
		},
		{
			name:     "single check item",
			input:    "Check: Is this a check item?\n# Heading",
			expected: []string{"Is this a check item?"},
		},
		{
			name:     "multiple check items",
			input:    "Check: First check\nText here\nVerify: Second check",
			expected: []string{"First check", "Second check"},
		},
		{
			name:     "check items with prefix",
			input:    "- 確認: これは確認事項です？\n# Heading",
			expected: []string{"これは確認事項です？"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Parse checks
			checksResult := parseChecks(tc.input)
			checks := []string{}
			if checksResult != "" {
				checks = strings.Split(checksResult, "\n")
			}

			if !cmp.Equal(checks, tc.expected) {
				t.Errorf("Checks mismatch:\nwant: %v\ngot:  %v\ndiff: %s",
					tc.expected, checks, cmp.Diff(tc.expected, checks))
			}
		})
	}
}

func TestRemoveChecks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "no check items",
			input:    "This is a text with `code span`",
			expected: "This is a text with `code span`",
		},
		{
			name:     "with check items",
			input:    "Check: Is this a check item?\n# Heading\n`test`",
			expected: "# Heading\n`test`",
		},
		{
			name:     "multiple check items",
			input:    "Check: First check\nVerify: Second check\n# Content\nNormal text",
			expected: "# Content\nNormal text",
		},
		{
			name:     "japanese check items",
			input:    "確認: これは確認事項です？\n# Heading\n`テスト`",
			expected: "# Heading\n`テスト`",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := removeChecks(tc.input)

			if result != tc.expected {
				t.Errorf("Result mismatch:\nwant: %v\ngot:  %v\ndiff: %s",
					tc.expected, result, cmp.Diff(tc.expected, result))
			}
		})
	}
}
