// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package local

import (
	"bytes"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/nats-io/nats.go/micro"
)

type CommandResponse struct {
	ExitCode int    `json:"code"`
	Output   []byte `json:"output"`
	Error    []byte `json:"error"`
}

func commandHandler(config *Config, r micro.Request) {
	command := r.Headers().Get("command")

	for _, prefix := range config.AllowedCommands {
		if !strings.HasPrefix(command, prefix) {
			continue
		}
		slog.Info("command", "run", command)
		cmd := exec.Command("sh", "-c", command)
		var outb, errb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = &errb
		err := cmd.Run()
		exitCode := cmd.ProcessState.ExitCode()
		if err != nil {
			slog.Error("command", "error", err)
			r.Error("500", "command execution failed", []byte(err.Error()))
			return
		}
		response := CommandResponse{
			ExitCode: exitCode,
			Output:   outb.Bytes(),
			Error:    errb.Bytes(),
		}
		r.RespondJSON(response)
		return
	}
	r.Error("400", "command not allowed", nil)
}
