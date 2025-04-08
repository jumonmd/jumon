// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/jumonmd/jumon/internal/errors"
	"github.com/jumonmd/jumon/module"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nats.go/micro"
)

const subscribeSubject = "event"

var (
	// ErrValidateEvent is returned when event validation fails.
	ErrValidateEvent = errors.New(400500, "validate event failed")
	// ErrEventNotFound is returned when a requested event doesn't exist.
	ErrEventNotFound = errors.New(404500, "event not found")
)

var manageEndpoints = []string{
	"put",
	"get",
	"list",
	"delete",
}

type Service struct {
	svc micro.Service
	sub *nats.Subscription
}

func (s *Service) AddEndpoint(name string, handler micro.Handler, opts ...micro.EndpointOpt) error {
	return s.svc.AddEndpoint(name, handler, opts...)
}

func (s *Service) AddGroup(name string, opts ...micro.GroupOpt) micro.Group {
	return s.svc.AddGroup(name, opts...)
}

func (s *Service) Info() micro.Info {
	return s.svc.Info()
}

func (s *Service) Stats() micro.Stats {
	return s.svc.Stats()
}

func (s *Service) Reset() {
	s.svc.Reset()
}

func (s *Service) Stop() error {
	if s.sub != nil {
		s.sub.Unsubscribe()
	}
	return s.svc.Stop()
}

func (s *Service) Stopped() bool {
	return s.svc.Stopped()
}

// NewService creates a new event service.
// subject: event.>
func NewService(nc *nats.Conn, js jetstream.JetStream) (micro.Service, error) {
	svc, err := micro.AddService(nc, micro.Config{
		Name:        "jumon_event",
		Version:     "0.1.0",
		Description: `jumon event service`,
	})
	if err != nil {
		slog.Error("event service", "status", "create service failed", "error", err)
		return nil, fmt.Errorf("create event service: %w", err)
	}

	_, err = setupKV(js)
	if err != nil {
		slog.Error("event service", "status", "setup kv failed", "error", err)
		return nil, fmt.Errorf("setup kv: %w", err)
	}

	g := svc.AddGroup("event")
	g.AddEndpoint("put", micro.HandlerFunc(func(r micro.Request) {
		ctx := context.Background()
		putEventHandler(ctx, js, r)
	}))

	g.AddEndpoint("get", micro.HandlerFunc(func(r micro.Request) {
		ctx := context.Background()
		getEventHandler(ctx, js, r)
	}))

	g.AddEndpoint("list", micro.HandlerFunc(func(r micro.Request) {
		ctx := context.Background()
		listEventHandler(ctx, js, r)
	}))

	g.AddEndpoint("delete", micro.HandlerFunc(func(r micro.Request) {
		ctx := context.Background()
		deleteEventHandler(ctx, js, r)
	}))

	sub, err := nc.Subscribe(subscribeSubject+".>", func(msg *nats.Msg) {
		if slices.Contains(manageEndpoints, msg.Subject) {
			return
		}

		ctx := context.Background()
		err := subscribeMessage(ctx, nc, js, msg)
		if err != nil {
			slog.Error("event service", "status", "subscribe failed", "error", err)
		}
	})
	if err != nil {
		slog.Error("event service", "status", "subscribe failed", "error", err)
		return nil, fmt.Errorf("subscribe: %w", err)
	}

	slog.Info("event service", "status", "started")
	return &Service{svc: svc, sub: sub}, nil
}

func subscribeMessage(ctx context.Context, nc *nats.Conn, js jetstream.JetStream, msg *nats.Msg) error {
	subject := strings.TrimPrefix(msg.Subject, subscribeSubject+".")
	slog.Info("event", "subject", subject)

	evt, err := GetEvent(ctx, js, EventTypeSubscribe, subject)
	if err != nil {
		return fmt.Errorf("get event: %w", err)
	}

	resp, err := module.Run(ctx, nc, evt.Module, msg.Data)
	if err != nil {
		return fmt.Errorf("run module: %w", err)
	}

	slog.Info("event", "subject", subject, "response", string(resp))

	return nil
}

func putEventHandler(ctx context.Context, js jetstream.JetStream, r micro.Request) {
	var input Event
	if err := json.Unmarshal(r.Data(), &input); err != nil {
		r.Error("403", "invalid payload", nil)
		return
	}

	evt := Event{
		Type:    EventTypeSubscribe,
		Subject: input.Subject,
		Module:  input.Module,
	}
	err := PutEvent(ctx, js, &evt)
	if err != nil {
		slog.Error("event service", "status", "put event failed", "error", err)
		r.Error("500", "failed to put event", nil)
		return
	}
	r.Respond(nil)
}

func getEventHandler(ctx context.Context, js jetstream.JetStream, r micro.Request) {
	var input Event
	if err := json.Unmarshal(r.Data(), &input); err != nil {
		r.Error("403", "invalid payload", nil)
		return
	}

	evt, err := GetEvent(ctx, js, EventTypeSubscribe, input.Subject)
	if err != nil {
		slog.Error("event service", "status", "get event failed", "error", err)
		r.Error("404", "event not found", nil)
		return
	}
	r.RespondJSON(evt)
}

func listEventHandler(ctx context.Context, js jetstream.JetStream, r micro.Request) {
	events, err := ListEvents(ctx, js)
	if err != nil {
		slog.Error("event service", "status", "list events failed", "error", err)
		r.Error("500", "failed to list events", nil)
		return
	}
	r.RespondJSON(events)
}

func deleteEventHandler(ctx context.Context, js jetstream.JetStream, r micro.Request) {
	var input Event
	if err := json.Unmarshal(r.Data(), &input); err != nil {
		r.Error("403", "invalid payload", nil)
		return
	}
	err := DeleteEvent(ctx, js, EventTypeSubscribe, input.Subject)
	if err != nil {
		slog.Error("event service", "status", "delete event failed", "error", err)
		r.Error("500", "failed to delete event", nil)
		return
	}
	r.Respond(nil)
}

func setupKV(js jetstream.JetStream) (jetstream.KeyValue, error) {
	return js.CreateOrUpdateKeyValue(context.Background(), jetstream.KeyValueConfig{
		Bucket:      "event",
		Description: "event key value store",
	})
}
