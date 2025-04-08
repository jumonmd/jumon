// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package module

import (
	"fmt"
	"os"
)

// InitModule initializes a new module with the given name on the current directory.
func InitModule(name string) error {
	path := "./JUMON.md"
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("JUMON.md already exists")
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, initTemplate, name)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

const initTemplate = `
---
module: %s
---
# 

## Scripts

### main

## Tools

`
