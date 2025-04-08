// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type key string

const (
	// DefaultModel is the default model to use.
	DefaultModel key = "DefaultModel"
	// DefaultVerifyModel is the default verify model to use.
	DefaultVerifyModel key = "DefaultVerifyModel"
)

func Get(ctx context.Context, nc *nats.Conn, key key) (string, error) {
	kv, err := keyvalue(ctx, nc)
	if err != nil {
		return "", fmt.Errorf("key value store: %w", err)
	}

	val, err := kv.Get(ctx, string(key))
	if errors.Is(err, jetstream.ErrKeyNotFound) {
		return defaultConfig(key), nil
	}
	if err != nil {
		return "", fmt.Errorf("get config: %w", err)
	}

	return string(val.Value()), nil
}

func Set(ctx context.Context, nc *nats.Conn, key key, value string) error {
	kv, err := keyvalue(ctx, nc)
	if err != nil {
		return fmt.Errorf("key value store: %w", err)
	}

	_, err = kv.Put(ctx, string(key), []byte(value))
	if err != nil {
		return fmt.Errorf("set config: %w", err)
	}
	return nil
}

func defaultConfig(key key) string {
	switch key {
	case DefaultModel:
		return "gpt-4o"
	case DefaultVerifyModel:
		return "gpt-4o-mini"
	default:
		return ""
	}
}

func keyvalue(ctx context.Context, nc *nats.Conn) (jetstream.KeyValue, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("create stream: %w", err)
	}
	kv, err := js.KeyValue(ctx, "config")
	if err != nil {
		return nil, fmt.Errorf("key value store: %w", err)
	}
	return kv, nil
}
