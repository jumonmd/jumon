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
)

type Event struct {
	Type    Type   `json:"type"`
	Subject string `json:"subject"`
	Module  string `json:"module"`
}

// PutEvent puts an event into the key value store.
func PutEvent(ctx context.Context, js jetstream.JetStream, evt *Event) error {
	key := fmt.Sprintf("%s.%s", evt.Type, evt.Subject)

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

// GetEvent gets an event from the key value store.
func GetEvent(ctx context.Context, js jetstream.JetStream, typ Type, subject string) (*Event, error) {
	key := fmt.Sprintf("%s.%s", typ, subject)

	kv, err := js.KeyValue(ctx, "event")
	if err != nil {
		return nil, fmt.Errorf("create key value store: %w", err)
	}

	evtdata, err := kv.Get(ctx, key)
	if err != nil {
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
