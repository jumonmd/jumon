// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package module

import (
	"fmt"

	"github.com/jumonmd/jumon/script"
	"github.com/jumonmd/jumon/tool"
)

const (
	JumonVersion = "0.1"
)

// Module defines a jumon module with its resources.
type Module struct {
	// JumonVersion specifies the compatibility version.
	JumonVersion string `json:"jumon"`
	// Name is the module's unique identifier in package path format.
	Name    string           `json:"module"`
	Scripts []*script.Script `json:"scripts"`
	Tools   []tool.Tool      `json:"tools,omitempty"`
}

func (m *Module) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("module name is required")
	}
	if len(m.Scripts) == 0 {
		return fmt.Errorf("scripts are required")
	}
	for _, s := range m.Scripts {
		if s.Name == "" {
			return fmt.Errorf("script name is required")
		}
	}
	return nil
}

// GetScript returns the script with the given name or the main script if name is empty.
func (m *Module) GetScript(name string) *script.Script {
	if name == "" {
		name = "main"
	}
	for _, s := range m.Scripts {
		if s.Name == name {
			return s
		}
	}
	return nil
}
