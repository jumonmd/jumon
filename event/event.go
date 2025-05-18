// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
)

type Type string

const (
	EventTypeSubscribe Type = "subscribe"
	EventTypePublish   Type = "publish"
	EventTypeConsume   Type = "consume"
	EventTypeForward   Type = "forward"
)

type Event struct {
	Type             Type   `json:"type"`
	SubscribeSubject string `json:"subscribe_subject"`
	PublishSubject   string `json:"publish_subject"`
	Consumer         string `json:"consumer"`
	Module           string `json:"module"`
	Template         string `json:"template"`
}

// PutEvent puts an event into the key value store.
func PutEvent(ctx context.Context, js jetstream.JetStream, evt *Event) error {
	var key string
	switch evt.Type {
	case EventTypeSubscribe, EventTypeForward:
		key = fmt.Sprintf("%s.%s", evt.Type, evt.SubscribeSubject)
	case EventTypePublish:
		key = fmt.Sprintf("%s.%s", evt.Type, evt.Module)
	}

	evtdata, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	kv, err := js.KeyValue(ctx, "event")
	if err != nil {
		return fmt.Errorf("create key value store: %w", err)
	}

	slog.Debug("put event", "key", key, "data", string(evtdata))

	_, err = kv.Put(ctx, key, evtdata)
	if err != nil {
		return fmt.Errorf("failed to put event: %w", err)
	}
	return nil
}

func GetConsumerEvent(ctx context.Context, js jetstream.JetStream, subject string) (*Event, error) {
	key := fmt.Sprintf("%s.%s", EventTypeConsume, subject)

	kv, err := js.KeyValue(ctx, "event")
	if err != nil {
		return nil, fmt.Errorf("create key value store: %w", err)
	}

	evtdata, err := kv.Get(ctx, key)
	if err != nil && errors.Is(err, jetstream.ErrKeyNotFound) {
		return nil, ErrEventNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	var evt Event
	err = json.Unmarshal(evtdata.Value(), &evt)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return &evt, nil
}

// GetSubscribeEvent gets an event from the key value store.
func GetSubscribeEvent(ctx context.Context, js jetstream.JetStream, subject string) (*Event, error) {
	key := fmt.Sprintf("%s.%s", EventTypeSubscribe, subject)

	kv, err := js.KeyValue(ctx, "event")
	if err != nil {
		return nil, fmt.Errorf("create key value store: %w", err)
	}

	evtdata, err := kv.Get(ctx, key)
	if err != nil && errors.Is(err, jetstream.ErrKeyNotFound) {
		return nil, ErrEventNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	var evt Event
	err = json.Unmarshal(evtdata.Value(), &evt)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return &evt, nil
}

// GetForwardEvent gets a forward event from the key value store.
func GetForwardEvent(ctx context.Context, js jetstream.JetStream, subject string) (*Event, error) {
	key := fmt.Sprintf("%s.%s", EventTypeForward, subject)

	kv, err := js.KeyValue(ctx, "event")
	if err != nil {
		return nil, fmt.Errorf("create key value store: %w", err)
	}

	evtdata, err := kv.Get(ctx, key)
	if err != nil && errors.Is(err, jetstream.ErrKeyNotFound) {
		return nil, ErrEventNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	var evt Event
	err = json.Unmarshal(evtdata.Value(), &evt)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return &evt, nil
}

func GetPublishEvent(ctx context.Context, js jetstream.JetStream, module string) (*Event, error) {
	key := fmt.Sprintf("%s.%s", EventTypePublish, module)

	kv, err := js.KeyValue(ctx, "event")
	if err != nil {
		return nil, fmt.Errorf("create key value store: %w", err)
	}

	evtdata, err := kv.Get(ctx, key)
	if err != nil && errors.Is(err, jetstream.ErrKeyNotFound) {
		return nil, ErrEventNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	var evt Event
	err = json.Unmarshal(evtdata.Value(), &evt)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return &evt, nil
}

// DeleteEvent deletes an event from the key value store.
func DeleteEvent(ctx context.Context, js jetstream.JetStream, typ Type, subject string) error {
	key := fmt.Sprintf("%s.%s", typ, subject)

	kv, err := js.KeyValue(ctx, "event")
	if err != nil {
		return fmt.Errorf("create key value store: %w", err)
	}

	err = kv.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}

// ListEvents lists all events from the key value store.
func ListEvents(ctx context.Context, js jetstream.JetStream) ([]Event, error) {
	kv, err := js.KeyValue(ctx, "event")
	if err != nil {
		return nil, fmt.Errorf("create key value store: %w", err)
	}

	events := []Event{}

	keys, err := kv.Keys(ctx, nil)
	if err != nil && errors.Is(err, jetstream.ErrNoKeysFound) {
		return events, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	for _, key := range keys {
		evtdata, err := kv.Get(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get event: %w", err)
		}
		var evt Event
		err = json.Unmarshal(evtdata.Value(), &evt)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal event: %w", err)
		}
		events = append(events, evt)
	}

	return events, nil
}
