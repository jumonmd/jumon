// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/nats-io/nats.go"
)

func WaitServer(execPath, url string) error {
	running := isNatsServerRunning(url)
	if running {
		return nil
	}
	log.Println("starting server")

	cmd := exec.Command(execPath, "serve")
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("server start failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for {
		nc, err := nats.Connect(url)
		if err == nil {
			nc.Close()
			break
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for server connection")
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}

	return nil
}

func isNatsServerRunning(url string) bool {
	nc, err := nats.Connect(url)
	if err != nil {
		return false
	}
	defer nc.Close()
	return nc.IsConnected()
}
