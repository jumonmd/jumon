// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/nats-io/nats.go"
)

const minInterval = 10 * time.Minute

// MetricsClient is a metrics client push to telemetry endpoint as prometheus text format using NATS.
type Client struct {
	endpoint string
	metadata map[string]string
	interval time.Duration
	sub      *nats.Subscription
}

type countMetrics struct {
	Name       string `json:"name"`
	StatusCode int    `json:"statusCode"`
}

func NewClient(endpoint string, metadata map[string]string, interval time.Duration) *Client {
	c := &Client{endpoint: endpoint, metadata: metadata, interval: interval}
	return c
}

// Close pushes the remaining metrics and closes the subscription.
func (c *Client) Close() error {
	if c.sub != nil {
		c.sub.Unsubscribe()
	}
	return c.PushMetrics()
}

// Subscribe subscribes subject for collecting metrics.
func (c *Client) Subscribe(nc *nats.Conn, subject string) error {
	sub, err := nc.Subscribe(subject, func(msg *nats.Msg) {
		var metrics countMetrics
		err := json.Unmarshal(msg.Data, &metrics)
		if err != nil {
			slog.Debug("metrics unmarshall", "error", err)
			return
		}

		metrics.Name = strings.TrimSuffix(metrics.Name, ".request")
		metrics.Name = strings.TrimSuffix(metrics.Name, ".response")
		metrics.Name = strings.TrimSuffix(metrics.Name, ".error")

		c.CountResponse(metrics.Name, metrics.StatusCode)
	})
	if err != nil {
		return err
	}
	c.sub = sub
	return nil
}

// InitPushMetrics starts periodically pushing metrics.
func (c *Client) InitPushMetrics() error {
	if c.interval < minInterval {
		c.interval = minInterval
	}
	return metrics.InitPushWithOptions(context.Background(), c.endpoint, c.interval, false, nil)
}

// PushMetrics pushes the metrics.
func (c *Client) PushMetrics() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	slog.Debug("metrics push", " endpoint", c.endpoint)
	return metrics.PushMetrics(ctx, c.endpoint, false, nil)
}

// CountResponse increments the metrics of the response.
// name is the event name. client metadata is automatically added to the metrics.
func (c *Client) CountResponse(name string, statusCode int) *metrics.Counter {
	os := runtime.GOOS
	arch := runtime.GOARCH
	metadata := ""
	if len(c.metadata) > 0 {
		kv := make([]string, 0, len(c.metadata))
		for k, v := range c.metadata {
			kv = append(kv, fmt.Sprintf(`%s="%s"`, k, v))
		}
		metadata = "," + strings.Join(kv, ",")
	}

	total := metrics.GetOrCreateCounter(fmt.Sprintf(`response{name="%s",status_code="%d",os="%s",arch="%s"%s}`, name, statusCode, os, arch, metadata))
	total.Inc()
	return total
}
