// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
)

// FormatJSON parses the input data as JSON, and if a template is provided,
// it unmarshals the JSON into a map[string]any and generates new JSON from the template using that map as context.
func FormatJSON(data []byte, tmpl string) (json.RawMessage, error) {
	data = bytes.TrimSpace(data)
	data = bytes.ReplaceAll(data, []byte("```json"), []byte(""))
	data = bytes.ReplaceAll(data, []byte("```"), []byte(""))
	data = bytes.TrimSpace(data)

	if tmpl == "" {
		return json.RawMessage(string(data)), nil
	}

	var mapjson map[string]any
	err := json.Unmarshal(data, &mapjson)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json: %w", err)
	}
	t, err := template.New("tmpl").Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w > %s", err, tmpl)
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, mapjson)
	if err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}
	return json.RawMessage(buf.Bytes()), nil
}
