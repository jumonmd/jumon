// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jumonmd/jumon/internal/logger"
	"github.com/jumonmd/jumon/internal/server"
	"github.com/jumonmd/jumon/internal/tracer"
	"github.com/jumonmd/jumon/local"
	"github.com/jumonmd/jumon/module"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nats.go/micro"
)

// Run is the main entry point for running a module.
func Run(name string, input []byte) error {
	cfg, err := LoadConfig(DefaultConfigPath())
	if err != nil {
		return fmt.Errorf("load client config: %w", err)
	}

	isDebug := os.Getenv("JUMON_DEBUG") == "1"

	// Setup services and dependencies
	nc, js, _, cleanup, err := setupServices(cfg, isDebug)
	if err != nil {
		return err
	}
	defer cleanup()

	// Prepare context with timeout
	ctx, cancel := notifyContext(cfg)
	defer cancel()

	// Setup notification
	err = subscribeNotification(ctx, nc)
	if err != nil {
		slog.Warn("setup notification", "error", err)
	}

	// Get module to the system
	mod, err := getModule(ctx, js, name)
	if err != nil {
		return fmt.Errorf("get module: %w", err)
	}

	// Run the module
	_, err = module.Run(ctx, nc, mod.Name, input)
	if err != nil {
		return fmt.Errorf("run module: %w", err)
	}

	return nil
}

// setupServices initializes all required services (NATS, logger, local service)
// and returns a cleanup function.
func setupServices(cfg *Config, isDebug bool) (nc *nats.Conn, js jetstream.JetStream, localSvc micro.Service, cleanup func(), err error) {
	// Setup NATS client
	nc, js, err = server.SetupNatsClient(cfg.ServerURL)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("nats client setup: %w", err)
	}

	// Setup a cleanup function that ensures NATS connection is properly closed
	cleanup = func() {
		if localSvc != nil {
			localSvc.Stop()
		}
		_ = nc.Drain()
		for !nc.IsClosed() {
			time.Sleep(time.Millisecond * 25)
		}
	}

	// Setup logger
	l := logger.New(nc, isDebug, true)
	slog.SetDefault(slog.New(l))

	// Setup local service
	localSvc, err = local.NewService(nc, &local.Config{
		WorkingDir: ".",
		ReadOnly:   true,
	})
	if err != nil {
		return nil, nil, nil, cleanup, fmt.Errorf("local service setup: %w", err)
	}

	// Log that local service is setup
	slog.Debug("local service setup complete")

	return nc, js, localSvc, cleanup, nil
}

// notifyContext creates and returns a context with timeout based on configuration.
func notifyContext(cfg *Config) (context.Context, context.CancelFunc) {
	ctx, notifyTo := tracer.CreateNotifyContext()
	slog.Debug("created notification context", "notifyTo", notifyTo)
	ctx, cancel := context.WithTimeout(ctx, cfg.RunTimeoutDuration())
	return ctx, cancel
}

// subscribeNotification configures notification subscription.
func subscribeNotification(ctx context.Context, nc *nats.Conn) error {
	notifyTo := tracer.ContextValueNotifyTo(ctx)
	if notifyTo == "" {
		slog.Warn("empty notification ID")
		return fmt.Errorf("empty notification ID")
	}

	slog.Info("notify to", "notifyTo", notifyTo)

	sub, err := PrintNotifications(nc, "notification."+notifyTo, os.Stdout)
	if err != nil {
		return err
	}

	// Ensure subscription is cleaned up when context is done
	go func() {
		<-ctx.Done()
		sub.Unsubscribe()
	}()

	return nil
}

// getModule resolves and retrieves a module either from local directory or git.
func getModule(ctx context.Context, js jetstream.JetStream, name string) (*module.Module, error) {
	// Get module key-value store
	modkv, err := js.KeyValue(ctx, "module")
	if err != nil {
		return nil, fmt.Errorf("kv create: %w", err)
	}

	// Resolve module by name (either local path or git repository)
	var mod *module.Module
	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, ".") {
		// Local module
		mod, err = module.GetByDir(ctx, modkv, name)
		if err != nil {
			return nil, fmt.Errorf("get dir failed: %w", err)
		}
	} else {
		// Remote module
		mod, err = module.GetByGit(ctx, modkv, name)
		if err != nil {
			return nil, fmt.Errorf("get git failed: %w", err)
		}
	}

	return mod, nil
}

// executeModule runs the specified module and handles the response.
func executeModule(ctx context.Context, nc *nats.Conn, mod *module.Module, input []byte) error {
	// Run module and wait for response
	resp, err := nc.RequestMsgWithContext(ctx, &nats.Msg{
		Subject: "module.run." + mod.Name,
		Data:    input,
		Header:  tracer.HeadersFromContext(ctx),
	})
	if err != nil {
		return fmt.Errorf("request module failed: %w", err)
	}

	// Check for errors in response
	if errorCode := resp.Header.Get("Nats-Service-Error-Code"); errorCode != "" {
		errorMessage := resp.Header.Get("Nats-Service-Error")
		return fmt.Errorf("module run error: %s: %s", errorCode, errorMessage)
	}

	// Log success
	slog.Info("count run", "status_code", 200, "subject", "module.run")
	slog.Info("run module", "response", string(resp.Data))
	return nil
}
