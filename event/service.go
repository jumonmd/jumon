// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jumonmd/jumon/internal/errors"
	"github.com/jumonmd/jumon/module"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nats.go/micro"
)

const eventSubject = "event"

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

	sub, err := nc.Subscribe(eventSubject+".>", func(msg *nats.Msg) {
		for _, evt := range manageEndpoints {
			if strings.HasPrefix(msg.Subject, eventSubject+"."+evt) {
				return
			}
		}

		ctx := context.Background()

		fwdevt, err := GetForwardEvent(ctx, js, msg.Subject)
		if err != nil && !errors.Is(err, ErrEventNotFound) { // ignore not found error
			slog.Warn("event service", "status", "get forward event failed", "subject", msg.Subject, "error", err)
			return
		}

		// just forward event
		if fwdevt != nil {
			slog.Info("event service", "status", "forward event", "subject", fwdevt.SubscribeSubject, "module", fwdevt.Module)
			if fwdevt.PublishSubject == "" {
				slog.Warn("event service", "status", "forward event to is empty", "subject", fwdevt.PublishSubject, "module", fwdevt.Module)
				return
			}

			formatdata, err := FormatJSON(msg.Data, fwdevt.Template)
			if err != nil {
				slog.Error("event service", "status", "format json failed", "error", err)
				return
			}
			slog.Debug("event service", "status", "forward event", "subject", fwdevt.PublishSubject, "content", string(formatdata))
			err = nc.Publish(fwdevt.PublishSubject, formatdata)
			if err != nil {
				slog.Error("event service", "status", "publish failed", "error", err)
			}
			return
		}

		subevt, err := GetSubscribeEvent(ctx, js, msg.Subject)
		if err != nil {
			slog.Warn("event service", "status", "get subscribe event failed", "subject", msg.Subject, "error", err)
			return
		}

		// subscribe event to run script
		slog.Info("event service", "status", "subscribe event", "subject", subevt.SubscribeSubject, "module", subevt.Module)
		resp, err := trigger(ctx, nc, subevt, msg.Data)
		if err != nil {
			slog.Error("event service", "status", "trigger failed", "error", err)
		}

		pubevt, err := GetPublishEvent(ctx, js, subevt.Module)
		if err != nil {
			slog.Warn("event service", "status", "get publish event failed", "subject", msg.Subject, "error", err)
			return
		}

		// publish event after run script
		if resp != nil && pubevt.Module == subevt.Module {
			if pubevt.PublishSubject == subevt.SubscribeSubject {
				slog.Warn("event service", "status", "publish event subject is same as subscribe event subject", "subject", pubevt.PublishSubject)
				return
			}
			slog.Debug("event service", "status", "publish event", "publish subject", pubevt.PublishSubject, "subscribe subject", subevt.SubscribeSubject, "content", string(resp))
			formatdata, err := FormatJSON(resp, pubevt.Template)
			if err != nil {
				slog.Error("event service", "status", "format json failed", "error", err)
				return
			}
			slog.Debug("event service", "status", "publish event", "subject", pubevt.PublishSubject, "content", string(formatdata))

			err = nc.Publish(pubevt.PublishSubject, formatdata)
			if err != nil {
				slog.Error("event service", "status", "publish failed", "error", err)
			}
		}
	})
	if err != nil {
		slog.Error("event service", "status", "subscribe failed", "error", err)
		return nil, fmt.Errorf("subscribe: %w", err)
	}

	slog.Info("event service", "status", "started")
	return &Service{svc: svc, sub: sub}, nil
}

func trigger(ctx context.Context, nc *nats.Conn, evt *Event, data []byte) (json.RawMessage, error) {
	subject := strings.TrimPrefix(evt.SubscribeSubject, eventSubject+".")
	slog.Info("event", "subject", subject)

	resp, err := module.Run(ctx, nc, evt.Module, data)
	if err != nil {
		return nil, fmt.Errorf("run module: %w", err)
	}

	slog.Info("event", "subject", subject, "response", string(resp))

	return resp, nil
}

func putEventHandler(ctx context.Context, js jetstream.JetStream, r micro.Request) {
	var evt Event
	if err := json.Unmarshal(r.Data(), &evt); err != nil {
		r.Error("403", "invalid payload", nil)
		return
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

	switch input.Type {
	case EventTypeSubscribe:
		evt, err := GetSubscribeEvent(ctx, js, input.SubscribeSubject)
		if err != nil {
			slog.Error("event service", "status", "get event failed", "error", err)
			r.Error("404", "event not found", nil)
			return
		}
		r.RespondJSON(evt)
	case EventTypePublish:
		evt, err := GetPublishEvent(ctx, js, input.Module)
		if err != nil {
			slog.Error("event service", "status", "get event failed", "error", err)
			r.Error("404", "event not found", nil)
			return
		}
		r.RespondJSON(evt)
	case EventTypeForward:
		evt, err := GetForwardEvent(ctx, js, input.SubscribeSubject)
		if err != nil {
			slog.Error("event service", "status", "get event failed", "error", err)
			r.Error("404", "event not found", nil)
			return
		}
		r.RespondJSON(evt)
	default:
		r.Error("403", "invalid event type", nil)
	}
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
	// TODO: 引数の定義を決める
	err := DeleteEvent(ctx, js, input.Type, input.SubscribeSubject)
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
