// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package testutil

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// NewNATSServer creates a new in memory NATS server and returns a connection and JetStream.
func NewNATSServer() (nc *nats.Conn, js jetstream.JetStream, obs jetstream.ObjectStore, cleanup func(), err error) {
	tmpdir, err := os.MkdirTemp("", "nats_test")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create temp directory for NATS storage: %w", err)
	}
	server, err := natsserver.NewServer(&natsserver.Options{
		DontListen: true,
		JetStream:  true,
		StoreDir:   tmpdir,
	})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create NATS server: %w", err)
	}

	server.Start()
	cleanup = func() {
		server.Shutdown()
		os.RemoveAll(tmpdir)
	}

	if !server.ReadyForConnections(time.Second * 5) {
		return nil, nil, nil, cleanup, errors.New("failed to start server after 5 seconds")
	}

	nc, err = nats.Connect("", nats.InProcessServer(server))
	if err != nil {
		return nil, nil, nil, cleanup, fmt.Errorf("failed to connect to server: %w", err)
	}

	js, err = jetstream.New(nc)
	if err != nil {
		return nil, nil, nil, cleanup, fmt.Errorf("failed to create jetstream: %w", err)
	}

	obs, err = js.CreateObjectStore(context.Background(), jetstream.ObjectStoreConfig{
		Bucket:  "cache",
		Storage: jetstream.MemoryStorage,
	})
	if err != nil {
		return nil, nil, nil, cleanup, fmt.Errorf("failed to create object store: %w", err)
	}

	return nc, js, obs, cleanup, nil
}
