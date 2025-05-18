// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package server

import (
	"context"
	"fmt"
	"time"

	"github.com/jumonmd/jumon/chat"
	"github.com/jumonmd/jumon/event"
	"github.com/jumonmd/jumon/internal/version"
	"github.com/jumonmd/jumon/module"
	"github.com/jumonmd/jumon/script"
	"github.com/jumonmd/jumon/tool"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nats.go/micro"
)

func setupNatsServer(opts *natsserver.Options) (ns *natsserver.Server, err error) {
	opts.ServerName = "jumon(" + version.Version + ")"
	// signal handle must be disabled
	opts.NoSigs = true
	if opts.PidFile == "" {
		opts.PidFile = pidPath()
	}

	ns, err = natsserver.NewServer(opts)
	if err != nil {
		return nil, fmt.Errorf("nats server create error: %w", err)
	}
	ns.Start()

	// wait for nats server to be ready
	for !ns.Running() {
		time.Sleep(50 * time.Millisecond)
	}

	return ns, nil
}

func SetupNatsClient(url string) (nc *nats.Conn, js jetstream.JetStream, err error) {
	nc, err = nats.Connect(url)
	if err != nil {
		return nil, nil, fmt.Errorf("nats connect failed: %w", err)
	}
	js, err = jetstream.New(nc)
	if err != nil {
		return nil, nil, fmt.Errorf("jetstream create error: %w", err)
	}
	return nc, js, nil
}

func SetupCache(ctx context.Context, js jetstream.JetStream) (jetstream.ObjectStore, error) {
	obs, err := js.CreateOrUpdateObjectStore(ctx, jetstream.ObjectStoreConfig{
		Bucket:      "cache",
		Description: "cache for jumon",
		TTL:         24 * time.Hour * 30,
	})
	if err != nil {
		return nil, fmt.Errorf("object store create error: %w", err)
	}
	return obs, nil
}

func setupServices(nc *nats.Conn, js jetstream.JetStream, obs jetstream.ObjectStore) ([]micro.Service, error) {
	services := []micro.Service{}

	chatsvc, err := chat.NewService(nc)
	if err != nil {
		return nil, fmt.Errorf("chat service create error: %w", err)
	}
	services = append(services, chatsvc)

	toolsvc, err := tool.NewService(nc, obs)
	if err != nil {
		return nil, fmt.Errorf("tool service create error: %w", err)
	}
	services = append(services, toolsvc)

	scriptsvc, err := script.NewService(nc)
	if err != nil {
		return nil, fmt.Errorf("script service create error: %w", err)
	}
	services = append(services, scriptsvc)

	modsvc, err := module.NewService(nc)
	if err != nil {
		return nil, fmt.Errorf("module service create error: %w", err)
	}
	services = append(services, modsvc)

	eventsvc, err := event.NewService(nc, js)
	if err != nil {
		return nil, fmt.Errorf("event service create error: %w", err)
	}
	services = append(services, eventsvc)

	return services, nil
}

func setupKV(ctx context.Context, js jetstream.JetStream) error {
	_, err := js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:      "module",
		Description: "modules for jumon",
	})
	if err != nil {
		return fmt.Errorf("module kv create error: %w", err)
	}
	_, err = js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:      "config",
		Description: "config for jumon",
	})
	if err != nil {
		return fmt.Errorf("config kv create error: %w", err)
	}
	_, err = js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:      "event",
		Description: "events for jumon",
	})
	if err != nil {
		return fmt.Errorf("event kv create error: %w", err)
	}

	return nil
}
