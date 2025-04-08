// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package frontmatter

import (
	"bytes"

	"github.com/BurntSushi/toml"
	"github.com/goccy/go-yaml"
)

var yamldelim = []byte("---")

var tomldelim = []byte("+++")

// Unmarshal parses YAML frontmatter and returns the content. When no
// frontmatter delimiters are present the original content is returned.
func Unmarshal(b []byte, v interface{}) (content []byte, err error) {
	b = bytes.TrimSpace(b)
	if bytes.HasPrefix(b, yamldelim) {
		parts := bytes.SplitN(b, yamldelim, 3)
		content = parts[2]
		err = yaml.Unmarshal(parts[1], v)
		return
	} else if bytes.HasPrefix(b, tomldelim) {
		parts := bytes.SplitN(b, tomldelim, 3)
		content = parts[2]
		err = toml.Unmarshal(parts[1], v)
		return
	}
	return b, nil
}
