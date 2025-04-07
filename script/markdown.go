// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package script

import (
	"bytes"
	"regexp"
	"strings"

	markdown "github.com/teekennedy/goldmark-markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var checkPrefixes = regexp.MustCompile(`(?i)(check:|verify:|確認[:：])`)

// parseSymbols parses markdown and extracts code spans as symbols.
// e.g. "This is a `function`" -> ["function"].
func parseSymbols(doc ast.Node, r text.Reader) ([]Symbols, error) {
	symbols := []Symbols{}
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if span, ok := n.(*ast.CodeSpan); ok {
			var sb strings.Builder
			for c := span.FirstChild(); c != nil; c = c.NextSibling() {
				if text, ok := c.(*ast.Text); ok {
					sb.Write(text.Segment.Value(r.Source()))
				}
			}

			if s := strings.TrimSpace(sb.String()); s != "" {
				symbols = append(symbols, Symbols{
					Type: "function",
					Name: s,
				})
			}
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, err
	}

	return symbols, nil
}

// parseList parses the given markdown script and returns a root [listItem].
// It also returns the text that appears before the list begins as preface.
func parseList(doc ast.Node, r text.Reader) (root *Step, preface string, err error) {
	root = &Step{Level: 0}
	currentItem := root
	currentLevel := 0
	var prefaceBuilder strings.Builder
	foundFirstList := false

	err = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch node := n.(type) {
			case *ast.List:
				foundFirstList = true
				currentLevel++
			case *ast.ListItem:
				currentItem, err = parseListItem(node, r, currentLevel, root, currentItem)
				if err != nil {
					return ast.WalkStop, err
				}
			case *ast.Heading,
				*ast.Paragraph,
				*ast.CodeBlock,
				*ast.FencedCodeBlock,
				*ast.Blockquote:
				if foundFirstList {
					return ast.WalkContinue, nil
				}

				text := n.Lines().Value(r.Source())
				prefaceBuilder.Write(text)
				prefaceBuilder.WriteString("\n\n")
			}
		} else {
			if _, ok := n.(*ast.List); ok {
				currentLevel--
			}
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, "", err
	}

	return root, strings.TrimSpace(prefaceBuilder.String()), nil
}

// getListItemText returns the text of the list item with markdown formatting preserved.
func getListItemText(parent ast.Node, r text.Reader) (string, error) {
	if parent == nil || !parent.HasChildren() {
		return "", nil
	}

	renderer := goldmark.New(goldmark.WithRenderer(markdown.NewRenderer()))

	var buf bytes.Buffer

	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() == ast.KindList {
			continue
		}

		if err := renderer.Renderer().Render(&buf, r.Source(), child); err != nil {
			return "", err
		}
	}

	result := strings.TrimSpace(buf.String())
	return result, nil
}

// getListItemMarker returns the marker (-, *, +) of the list item.
func getListItemMarker(parent ast.Node) string {
	var sb strings.Builder
	_ = ast.Walk(parent, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if n.Parent().Kind() == ast.KindList {
				if list, ok := n.Parent().(*ast.List); ok {
					sb.WriteString(string(list.Marker))
					return ast.WalkStop, nil
				}
			}
		}
		return ast.WalkContinue, nil
	})
	return sb.String()
}

// parseListItem parses the given list item and returns the updated current item.
func parseListItem(node ast.Node, r text.Reader, currentLevel int, root, currentItem *Step) (*Step, error) {
	text, err := getListItemText(node, r)
	if err != nil {
		return nil, err
	}

	item := &Step{
		Level:    currentLevel,
		Content:  getListItemMarker(node) + " " + text,
		Children: []*Step{},
	}

	if item.Level > currentItem.Level {
		// This is a child of the current item
		currentItem.Children = append(currentItem.Children, item)
	} else {
		// Go up the tree until we find the right parent
		parent := findItemParentAtLevel(root, item.Level-1)
		parent.Children = append(parent.Children, item)
	}

	return item, nil
}

// findItemParentAtLevel finds the parent of the item at the given level.
func findItemParentAtLevel(root *Step, targetLevel int) *Step {
	if targetLevel <= 0 || root.Level == targetLevel {
		return root
	}

	for i := len(root.Children) - 1; i >= 0; i-- {
		child := root.Children[i]
		if child.Level == targetLevel {
			return child
		}

		if len(child.Children) > 0 {
			if result := findItemParentAtLevel(child, targetLevel); result != nil && result.Level == targetLevel {
				return result
			}
		}
	}

	return root
}

// parseChecks parses check items from markdown document.
func parseChecks(md string) string {
	checks := []string{}
	lines := strings.Split(md, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if checkPrefixes.MatchString(trimmed) {
			// extract the check item text after the prefix
			idx := checkPrefixes.FindStringIndex(trimmed)
			if idx == nil {
				continue
			}
			checkText := strings.TrimSpace(trimmed[idx[1]:])
			if checkText != "" {
				checks = append(checks, checkText)
			}
		}
	}

	return strings.Join(checks, "\n")
}

// removeChecks removes check items from markdown document.
func removeChecks(md string) string {
	if md == "" {
		return md
	}

	lines := strings.Split(md, "\n")
	filteredLines := make([]string, 0, len(lines))

	for _, line := range lines {
		if checkPrefixes.MatchString(line) {
			continue
		}
		filteredLines = append(filteredLines, line)
	}

	return strings.Join(filteredLines, "\n")
}
