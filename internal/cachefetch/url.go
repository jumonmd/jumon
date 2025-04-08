// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package cachefetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/jumonmd/jumon/internal/subject"
	"github.com/nats-io/nats.go/jetstream"
)

var getURLTimeout = 10 * time.Minute

// Open fetches a resource from a URL and caches it in the object store.
func Open(ctx context.Context, url string, obs jetstream.ObjectStore) (io.ReadCloser, error) {
	key := subject.Escape(url)

	if obs == nil {
		return getHTTP(ctx, url)
	}

	// cache hit or relative file
	r, err := obs.Get(ctx, key)
	if err != nil && !errors.Is(err, jetstream.ErrObjectNotFound) {
		return nil, fmt.Errorf("open cache: %w", err)
	} else if err == nil {
		return r, nil
	}

	// cache miss
	body, err := getHTTP(ctx, url)
	if err != nil {
		slog.Error("cachefetch", "status", "get http failed", "url", url, "error", err)
		return nil, fmt.Errorf("get http: %w", err)
	}
	defer body.Close()
	_, err = obs.Put(ctx, jetstream.ObjectMeta{Name: key}, body)
	if err != nil {
		slog.Error("cachefetch", "status", "put cache failed", "url", url, "error", err)
		return nil, fmt.Errorf("put cache: %w", err)
	}
	slog.Debug("cachefetch", "status", "body cached", "url", url)
	return obs.Get(ctx, key)
}

// getHTTP fetches a resource from a URL and returns the body as an io.Reader.
func getHTTP(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	client := &http.Client{
		Timeout: getURLTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get resource: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("get resource: status code %d", resp.StatusCode)
	}
	return resp.Body, nil
}
