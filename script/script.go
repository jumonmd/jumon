// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package script

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jumonmd/gengo/chat"
	"github.com/jumonmd/gengo/jsonschema"
	"github.com/jumonmd/jumon/internal/dataurl"
	"github.com/jumonmd/jumon/tool"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
)

const (
	defaultTimeoutSeconds = 300
)

// Script is a definition for multi-step AI prompt.
type Script struct {
	// Name is the name of the script (also referred to as a symbol)
	Name        string `json:"name"`
	Description string `json:"description"`
	// Config is the configuration such as timeout or limit.
	Config Config `json:"config,omitempty"`
	// Model is the model name.
	Model string `json:"model"`
	// ModelConfig is the model configuration to the model
	ModelConfig  *chat.ModelConfig `json:"model_config,omitempty"`
	InputSchema  jsonschema.Schema `json:"input_schema,omitempty"`
	OutputSchema jsonschema.Schema `json:"output_schema,omitempty"`
	Tools        []tool.Tool       `json:"tools,omitempty"`
	// Content is the markdown content of the script.
	Content string `json:"content"`
	// InputURL is the input as a data URL.
	InputURL string `json:"input_url"`
}

type Config struct {
	TimeoutSeconds int `json:"timeout_seconds"`
}

// Step defines an executable step in a script. Steps can be nested to create hierarchical structures.
type Step struct {
	// Level normally starts from 1. (root is 0)
	Level int
	// Type for the extension of the step.
	Type     string
	Content  string
	Children []*Step
}

// Symbol represents a definition in the script that can be referenced by name as a variable or tool.
type Symbols struct {
	Type string
	Name string
}

// Steps parses the script and returns the steps.
// It also returns the preface which is the text before the first step.
func (s *Script) Steps() ([]*Step, string, error) {
	parser := goldmark.New().Parser()
	doc := parser.Parse(text.NewReader([]byte(s.Content)))
	root, preface, err := parseList(doc, text.NewReader([]byte(s.Content)))
	if err != nil {
		return nil, "", fmt.Errorf("parse list: %w", err)
	}

	return root.Children, preface, nil
}

// Markdown returns the text of this list item and all its children recursively.
// Each line is indented according to its level, and a prefix (like "- ") is added based on the marker.
func (s *Step) Markdown() string {
	var builder strings.Builder

	if s.Level > 0 {
		for range s.Level - 1 {
			builder.WriteString("  ")
		}
		builder.WriteString(s.Content)
		builder.WriteString("\n")
	}

	for _, child := range s.Children {
		builder.WriteString(child.Markdown())
	}

	return builder.String()
}

// SetInput sets the input as a data URL.
// mime is automatically detected.
func (s *Script) SetInput(input []byte) {
	mime := http.DetectContentType(input)
	s.InputURL = dataurl.Encode(mime, input)
}

// Symbols parses the script and returns the symbols.
func (s *Script) Symbols() ([]Symbols, error) {
	parser := goldmark.New().Parser()
	doc := parser.Parse(text.NewReader([]byte(s.Content)))

	symbols, err := parseSymbols(doc, text.NewReader([]byte(s.Content)))
	if err != nil {
		return nil, fmt.Errorf("parse symbols: %w", err)
	}

	return symbols, nil
}

func (s *Script) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("name is required")
	}
	if s.Model == "" {
		return fmt.Errorf("model is required")
	}
	return nil
}

// AsTool converts the script to a [tool.Tool].
func (s *Script) AsTool() (tool.Tool, error) {
	scrdata, err := json.Marshal(s)
	if err != nil {
		return tool.Tool{}, fmt.Errorf("marshal script: %w", err)
	}

	return tool.Tool{
		Name:         s.Name,
		Description:  s.Description,
		Type:         "script",
		InputSchema:  s.InputSchema,
		OutputSchema: s.OutputSchema,
		Arguments: tool.Arguments{
			"script": string(scrdata),
		},
	}, nil
}

func (s *Script) String() string {
	return fmt.Sprintf("name: %s\ndescription: %s\nmodel: %s\nmodel_config: %v\ninput_schema: %v\noutput_schema: %v\ntools: %v\ncontent: %s", s.Name, s.Description, s.Model, s.ModelConfig, s.InputSchema, s.OutputSchema, s.Tools, s.Content)
}
