// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

// Tool package runs various tools with simple binary input/output.
package tool

import (
	"fmt"
	"net/http"

	"github.com/jumonmd/gengo/chat"
	"github.com/jumonmd/gengo/jsonschema"
	"github.com/jumonmd/jumon/internal/dataurl"
)

// Tool is a tool definition.
type Tool struct {
	// Type is the type of the tool. e.g. "script", "wasm", "nats"
	Type string `json:"type" yaml:"type" toml:"type" md:"type"`
	// Name is the name of the tool. It is also referred to as a symbol.
	Name string `json:"name" yaml:"name" toml:"name" md:"name"`
	// Module is the module of the tool. It is used for importing tools.
	Module       string            `json:"module" yaml:"module" toml:"module" md:"module"`
	Description  string            `json:"description" yaml:"description" toml:"description" md:"description"`
	InputSchema  jsonschema.Schema `json:"input_schema" yaml:"input_schema" toml:"input_schema" md:"input_schema,json"`
	OutputSchema jsonschema.Schema `json:"output_schema" yaml:"output_schema" toml:"output_schema" md:"output_schema,json"`
	// Arguments is the arguments of the tool.
	Arguments Arguments `json:"arguments" yaml:"arguments" toml:"arguments" md:"arguments"`
	// Resources is the plugin resources of the tool.
	Resources []*Resource `json:"resources" yaml:"resources" toml:"resources"`
	// InputURL is the input as a data URL.
	InputURL string `json:"input_url" yaml:"input_url" toml:"input_url" md:"input_url"`
}

type Arguments map[string]any

// ChatTool returns a chat tool struct.
func (t *Tool) ChatTool() chat.Tool {
	return chat.Tool{
		Name:        t.Name,
		Description: t.Description,
		InputSchema: t.InputSchema,
	}
}

func (t *Tool) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("name is required")
	}
	if t.Type == "" {
		return fmt.Errorf("type is required")
	}
	return nil
}

// SetInput sets the input of the tool.
// mime is automatically detected.
func (t *Tool) SetInput(input []byte) {
	mime := http.DetectContentType(input)
	t.InputURL = dataurl.Encode(mime, input)
}
