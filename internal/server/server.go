// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jumonmd/jumon/internal/logger"
	"github.com/jumonmd/jumon/internal/metrics"
	"github.com/jumonmd/jumon/internal/version"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

const telemetryEndpoint = "https://telemetry.jumon.md"

func Serve(disableTelemetry bool) error {
	// Check if server is already running
	if running, _ := IsRunning(); running {
		return fmt.Errorf("server is already running")
	}

	ctx := context.Background()

	err := createConfig()
	if err != nil {
		return fmt.Errorf("create config: %w", err)
	}

	natscfg, err := server.ProcessConfigFile(configPath())
	if err != nil {
		return fmt.Errorf("nats processconfig: %w", err)
	}

	isDebug := os.Getenv("JUMON_DEBUG") == "1"

	// Setup NATS server
	ns, err := setupNatsServer(natscfg)
	if err != nil {
		return fmt.Errorf("nats server: %w", err)
	}
	defer shutdownNatsServer(ns)

	// Setup NATS client
	nc, js, err := SetupNatsClient("nats://localhost:" + strconv.Itoa(natscfg.Port))
	if err != nil {
		return fmt.Errorf("nats client: %w", err)
	}
	defer cleanupNatsClient(nc)

	// Setup KV
	err = setupKV(ctx, js)
	if err != nil {
		return fmt.Errorf("kv setup: %w", err)
	}

	// Setup Cache
	obs, err := SetupCache(ctx, js)
	if err != nil {
		return fmt.Errorf("cache setup: %w", err)
	}

	// Setup logger
	setupLogger(nc, isDebug)

	// Setup metrics
	mc, err := setupMetrics(nc, disableTelemetry)
	if err != nil {
		slog.Warn("setup metrics", "error", err)
	}
	defer cleanupMetrics(mc)

	// Setup services
	svcs, err := setupServices(nc, js, obs)
	if err != nil {
		return fmt.Errorf("setup services: %w", err)
	}
	defer stopServices(svcs)

	slog.Debug("server", "isDebug", isDebug, "disableTelemetry", disableTelemetry)

	slog.Info("server is ready")

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	return nil
}

func shutdownNatsServer(ns *server.Server) {
	log.Println("server shutting down")
	ns.Shutdown()
	ns.WaitForShutdown()
	log.Println("server shutdown complete")
}

// cleanupNatsClient gracefully drains and closes the NATS connection.
func cleanupNatsClient(nc *nats.Conn) {
	_ = nc.Drain()
	for !nc.IsClosed() {
		time.Sleep(time.Millisecond * 25)
	}
}

func setupLogger(nc *nats.Conn, isDebug bool) {
	logger := logger.New(nc, isDebug, false)
	slog.SetDefault(slog.New(logger))
}

func setupMetrics(nc *nats.Conn, disableTelemetry bool) (*metrics.Client, error) {
	var mc *metrics.Client

	if !disableTelemetry {
		mc = metrics.NewClient(
			telemetryEndpoint,
			map[string]string{"version": version.Version},
			60*time.Minute,
		)
		err := mc.Subscribe(nc, "trace.>")
		if err != nil {
			return nil, fmt.Errorf("failed to subscribe to trace.>: %w", err)
		}
	}

	return mc, nil
}

func cleanupMetrics(mc *metrics.Client) {
	if mc != nil {
		mc.Close()
	}
}

// stopServices stops all services in reverse order.
func stopServices(svcs []micro.Service) {
	for _, svc := range svcs {
		slog.Info("stopping service", "name", svc.Info().Name)
		svc.Stop()
	}
}

func Quit() error {
	pidpath := pidPath()
	pid, err := os.ReadFile(pidpath)
	if err != nil {
		return fmt.Errorf("failed to read pid file: %w", err)
	}
	slog.Info("quitting server", "pid", string(pid))
	return server.ProcessSignal(server.CommandQuit, string(pid))
}

func IsRunning() (bool, error) {
	pidpath := pidPath()
	pid, err := os.ReadFile(pidpath)
	if err != nil {
		return false, fmt.Errorf("failed to read pid file: %w", err)
	}

	pidint, err := strconv.Atoi(string(pid))
	if err != nil {
		return false, fmt.Errorf("failed to convert pid to int: %w", err)
	}

	process, err := os.FindProcess(pidint)
	if err == nil {
		err = process.Kill()
	}
	return err == nil, nil
}
