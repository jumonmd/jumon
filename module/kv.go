// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package module

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func getModule(ctx context.Context, nc *nats.Conn, name string) (*Module, error) {
	kv, err := keyvalue(ctx, nc)
	if err != nil {
		return nil, fmt.Errorf("get keyvalue: %w", err)
	}
	moddata, err := kv.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get module: %w", err)
	}
	mod, err := ParseMarkdown(moddata.Value())
	if err != nil {
		return nil, fmt.Errorf("parse module: %w", err)
	}
	return mod, nil
}

func keyvalue(ctx context.Context, nc *nats.Conn) (jetstream.KeyValue, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("get jetstream: %w", err)
	}
	kv, err := js.KeyValue(ctx, "module")
	if err != nil {
		return nil, fmt.Errorf("get keyvalue: %w", err)
	}
	return kv, nil
}
