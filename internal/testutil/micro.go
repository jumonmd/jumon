// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package testutil

import (
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

// NewMicroServer creates a test NATS micro echo server.
// It echoes the response to the request.
func NewMicroServer(nc *nats.Conn, subject string, resp []byte) (micro.Service, error) {
	return micro.AddService(nc, micro.Config{
		Name:        "test-service",
		Version:     "0.1.0",
		Description: "test service",
		Endpoint: &micro.EndpointConfig{
			Subject: subject,
			Handler: micro.HandlerFunc(func(r micro.Request) {
				slog.Debug("test-service received request", "subject", r.Subject())
				r.Respond(resp)
			}),
		},
	})
}
