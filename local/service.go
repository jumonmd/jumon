// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package local

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

type Config struct {
	WorkingDir      string
	ReadOnly        bool
	CreateOnly      bool
	AllowedCommands []string
}

// NewService creates a new local service.
// subject: local.file.*
// subject: local.exec
func NewService(nc *nats.Conn, config *Config) (micro.Service, error) {
	if err := checkWorkingDir(config); err != nil {
		return nil, fmt.Errorf("working directory: %w", err)
	}

	svc, err := micro.AddService(nc, micro.Config{
		Name:        "jumon_local",
		Version:     "0.1.0",
		Description: `jumon local service`,
	})
	if err != nil {
		slog.Error("local service", "status", "create service failed", "error", err)
		return nil, fmt.Errorf("create local service: %w", err)
	}
	g := svc.AddGroup("local")
	file := g.AddGroup("file")
	file.AddEndpoint("read", micro.HandlerFunc(func(r micro.Request) {
		readHandler(config, r)
	}))
	file.AddEndpoint("write", micro.HandlerFunc(func(r micro.Request) {
		writeHandler(config, r)
	}))
	file.AddEndpoint("list", micro.HandlerFunc(func(r micro.Request) {
		listHandler(config, r)
	}))

	g.AddEndpoint("exec", micro.HandlerFunc(func(r micro.Request) {
		commandHandler(config, r)
	}))

	slog.Info("local service", "status", "started")
	return svc, nil
}

func checkWorkingDir(config *Config) error {
	if config == nil || config.WorkingDir == "" {
		return fmt.Errorf("working directory is required")
	}
	if config.WorkingDir == "/" {
		return fmt.Errorf("working directory cannot be root")
	}
	stat, err := os.Stat(config.WorkingDir)
	if err != nil {
		return fmt.Errorf("working directory does not exist: %s", config.WorkingDir)
	}
	if !stat.IsDir() {
		return fmt.Errorf("working directory is not a directory: %s", config.WorkingDir)
	}
	return nil
}
