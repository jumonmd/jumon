// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package module

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jumonmd/jumon/internal/frontmatter"
	"github.com/jumonmd/jumon/script"
	"github.com/jumonmd/jumon/tool"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

const (
	SectionScripts = "Scripts"
	SectionTools   = "Tools"
	SectionEvents  = "Events"
)

func ParseMarkdown(markdown []byte) (*Module, error) {
	fm := &Module{}
	body, err := frontmatter.Unmarshal(markdown, fm)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal frontmatter: %w", err)
	}

	parser := goldmark.New().Parser()
	doc := parser.Parse(text.NewReader(body))

	mod, err := parseMarkdown(doc, text.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse scripts: %w", err)
	}

	mod.Name = fm.Name

	return mod, nil
}

// detectSection returns the section type if the node is a level 2 heading with a specific name.
func detectSection(node ast.Node, r text.Reader) string {
	if heading, ok := node.(*ast.Heading); ok && heading.Level == 2 {
		text := getNodeText(heading, r)
		trimmedText := strings.TrimSpace(text)

		switch trimmedText {
		case SectionScripts:
			return SectionScripts
		case SectionTools:
			return SectionTools
		case SectionEvents:
			return SectionEvents
		}
	}

	return ""
}

func handleLevel3Heading(mod *Module, node *ast.Heading, r text.Reader, currentSection string) error {
	if node.Level != 3 {
		return nil
	}

	name := strings.TrimSpace(getNodeText(node, r))

	switch currentSection {
	case SectionScripts:
		content := getNodeHeadingContent(node, r)
		newScript := &script.Script{
			Name:    name,
			Content: content,
		}
		mod.Scripts = append(mod.Scripts, newScript)

	case SectionTools:
		content := getNodeHeadingContent(node, r)
		var tl tool.Tool

		if m := getMap([]byte(content)); m["import"] != "" {
			importname := m["import"]
			slog.Debug("import", "module", importname)
			tl.Module = importname
		}

		lang, code, err := getCodeBlock([]byte(content))
		if err != nil {
			slog.Error("failed to get code block", "error", err)
			return err
		}
		slog.Debug("tool", "lang", lang, "code", code)
		if lang == "json" {
			err := json.Unmarshal([]byte(code), &tl)
			if err != nil {
				slog.Error("failed to unmarshal tool data", "error", err)
				return err
			}
			slog.Debug("parsed tool", "tool", tl)
		}

		tl.Name = name
		mod.Tools = append(mod.Tools, tl)
	}
	return nil
}

func parseMarkdown(doc ast.Node, r text.Reader) (*Module, error) {
	mod := &Module{}
	currentSection := ""

	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if node, ok := n.(*ast.Heading); ok {
			// Check for level 2 headings (sections)
			if node.Level == 2 {
				currentSection = detectSection(node, r)
				return ast.WalkContinue, nil
			}

			// Handle level 3 headings based on current section
			if node.Level == 3 {
				err := handleLevel3Heading(mod, node, r, currentSection)
				if err != nil {
					return ast.WalkStop, err
				}
			}
		}

		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, err
	}

	return mod, nil
}

func getNodeText(n ast.Node, r text.Reader) string {
	var sb strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if text, ok := c.(*ast.Text); ok {
			sb.Write(text.Segment.Value(r.Source()))
		}
	}
	return strings.TrimSpace(sb.String())
}

// getCodeBlock extracts the language and content from a fenced code block.
func getCodeBlock(md []byte) (lang, content string, err error) {
	r := text.NewReader(md)
	parser := goldmark.New().Parser()
	n := parser.Parse(r)

	err = ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if fence, ok := n.(*ast.FencedCodeBlock); ok {
				if fence.Info != nil {
					lang = string(fence.Info.Value(r.Source()))
				}
				var lines []string
				for i := 0; i < fence.Lines().Len(); i++ {
					line := fence.Lines().At(i)
					lineText := strings.TrimRight(string(line.Value(r.Source())), "\n\r")
					lines = append(lines, lineText)
				}
				content = strings.Join(lines, "\n")
			}
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return "", "", err
	}
	return lang, content, nil
}

// getMap converts markdown text to a key-value map.
func getMap(md []byte) map[string]string {
	m := map[string]string{}
	lines := strings.Split(string(md), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			m[key] = value
		}
	}

	return m
}

// getNodeHeadingContent extracts content between the current heading and the next heading.
// e.g.
// > ## Heading
// > Inner Content
// > ## Next Heading
// output -> Inner Content

func getNodeHeadingContent(n ast.Node, r text.Reader) string {
	heading, ok := n.(*ast.Heading)
	if !ok {
		return ""
	}
	if heading.Lines().Len() == 0 {
		return ""
	}

	headerLine := heading.Lines().At(0)
	headerEnd := headerLine.Stop

	var nextHeaderStart int64 = -1
	next := n.NextSibling()

	for next != nil {
		if h, isHeading := next.(*ast.Heading); isHeading {
			if h.Lines().Len() > 0 {
				nextHeaderStart = int64(h.Lines().At(0).Start)
				break
			}
		}
		next = next.NextSibling()
	}

	if nextHeaderStart == -1 {
		// end of document
		nextHeaderStart = int64(len(r.Source()))
	}

	content := r.Source()[headerEnd:nextHeaderStart]
	// trim next header hash and space
	trimmedContent := strings.TrimRight(string(content), "# ")
	trimmedContent = strings.TrimSpace(trimmedContent)

	return trimmedContent
}
