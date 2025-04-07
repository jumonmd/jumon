// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package script

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jumonmd/gengo/jsonschema"
	"github.com/jumonmd/jumon/tool"
)

func TestScriptValidate(t *testing.T) {
	tests := []struct {
		name    string
		script  Script
		wantErr bool
	}{
		{
			name: "valid script",
			script: Script{
				Name:  "test script",
				Model: "test model",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			script: Script{
				Model: "test model",
			},
			wantErr: true,
		},
		{
			name: "missing model",
			script: Script{
				Name: "test script",
			},
			wantErr: true,
		},
		{
			name:    "empty script",
			script:  Script{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.script.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Script.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScriptSteps(t *testing.T) {
	tests := []struct {
		name    string
		script  Script
		want    []*Step
		wantErr bool
	}{
		{
			name: "valid script",
			script: Script{
				Name: "test script",
				Content: `
- step1
- step2
`,
			},
			want: []*Step{
				{
					Level:    1,
					Content:  "- step1",
					Children: []*Step{},
				},
				{
					Level:    1,
					Content:  "- step2",
					Children: []*Step{},
				},
			},
			wantErr: false,
		},
		{
			name: "nested steps",
			script: Script{
				Name: "test script",
				Content: `
- step1
  - step1.1
  - step1.2
- step2
  - step2.1
  - step2.2
`,
			},
			want: []*Step{
				{
					Level:   1,
					Content: "- step1",
					Children: []*Step{
						{
							Level:    2,
							Content:  "- step1.1",
							Children: []*Step{},
						},
						{
							Level:    2,
							Content:  "- step1.2",
							Children: []*Step{},
						},
					},
				},
				{
					Level:   1,
					Content: "- step2",
					Children: []*Step{
						{
							Level:    2,
							Content:  "- step2.1",
							Children: []*Step{},
						},
						{
							Level:    2,
							Content:  "- step2.2",
							Children: []*Step{},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := tt.script.Steps()
			if (err != nil) != tt.wantErr {
				t.Errorf("steps error = %v, wantErr %v", err, tt.wantErr)
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("steps differ: %v", cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestScriptAsTool(t *testing.T) {
	tests := []struct {
		name    string
		script  Script
		want    tool.Tool
		wantErr bool
	}{
		{
			name: "valid script",
			script: Script{
				Name:         "test script",
				Description:  "test description",
				Model:        "test model",
				InputSchema:  jsonschema.Schema{"type": "object"},
				OutputSchema: jsonschema.Schema{"type": "string"},
			},
			want: tool.Tool{
				Name:         "test script",
				Description:  "test description",
				Type:         "script",
				InputSchema:  jsonschema.Schema{"type": "object"},
				OutputSchema: jsonschema.Schema{"type": "string"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.script.AsTool()
			if (err != nil) != tt.wantErr {
				t.Errorf("Script.AsTool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			scrdata, ok := got.Arguments["script"].(string)
			if !ok {
				t.Errorf("script is not a string")
				return
			}

			// Verify the argument is a valid JSON representation of the script
			var unmarshaled Script
			if err := json.Unmarshal([]byte(scrdata), &unmarshaled); err != nil {
				t.Errorf("Failed to unmarshal script from argument: %v", err)
				return
			}

			// Compare the script with the unmarshaled one
			if !cmp.Equal(tt.script, unmarshaled) {
				t.Errorf("Unmarshaled script differs from original: %v", cmp.Diff(tt.script, unmarshaled))
			}
		})
	}
}
