// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package local

import (
	"fmt"
	"io"
	"os"

	"github.com/nats-io/nats.go/micro"
)

func listHandler(config *Config, r micro.Request) {
	path := r.Headers().Get("path")

	root, err := os.OpenRoot(config.WorkingDir)
	if err != nil {
		r.Error("500", fmt.Sprintf("open working directory: %s", err), nil)
		return
	}

	d, err := root.Open(path)
	if err != nil {
		r.Error("500", fmt.Sprintf("failed to read directory: %s", err), nil)
		return
	}
	defer d.Close()

	files, err := d.ReadDir(-1)
	if err != nil {
		r.Error("500", fmt.Sprintf("failed to read directory: %s", err), nil)
		return
	}

	names := make([]string, len(files))
	for i, file := range files {
		names[i] = file.Name()
	}
	r.RespondJSON(names)
}

func readHandler(config *Config, r micro.Request) {
	path := r.Headers().Get("path")

	root, err := os.OpenRoot(config.WorkingDir)
	if err != nil {
		r.Error("500", fmt.Sprintf("open working directory: %s", err), nil)
		return
	}

	f, err := root.Open(path)
	if err != nil {
		r.Error("500", fmt.Sprintf("failed to read file: %s", err), nil)
		return
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		r.Error("500", fmt.Sprintf("failed to read file: %s", err), nil)
		return
	}
	r.Respond(data)
}

func writeHandler(config *Config, r micro.Request) {
	if config.ReadOnly {
		r.Error("403", "read only mode", nil)
		return
	}
	path := r.Headers().Get("path")
	root, err := os.OpenRoot(config.WorkingDir)
	if err != nil {
		r.Error("500", fmt.Sprintf("open working directory: %s", err), nil)
		return
	}

	if _, err := root.Stat(path); err != nil && config.CreateOnly {
		r.Error("403", fmt.Sprintf("append only mode: %s", path), nil)
		return
	}

	f, err := root.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		r.Error("500", fmt.Sprintf("failed to read file: %s", err), nil)
		return
	}
	defer f.Close()
	if _, err := f.Write(r.Data()); err != nil {
		r.Error("500", fmt.Sprintf("failed to write file: %s", err), nil)
		return
	}
	r.Respond(nil)
}
