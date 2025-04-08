// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package tool

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/jumonmd/jumon/internal/cachefetch"
	"github.com/nats-io/nats.go/jetstream"
)

// Resource is a plugin resource of the tool.
type Resource struct {
	Name string `json:"name" yaml:"name" toml:"name"`
	URL  string `json:"url" yaml:"url" toml:"url"`
	// Hash is the SHA-256 hex hash of the resource.
	Hash string `json:"hash" yaml:"hash" toml:"hash"`
	// Size is the data size of the resource.
	Size uint64 `json:"size" yaml:"size" toml:"size"`
	// reader for the resource.
	reader io.ReadCloser `json:"-" yaml:"-" toml:"-"`
}

func NewResource(name, url string, reader io.ReadCloser) *Resource {
	return &Resource{
		Name:   name,
		URL:    url,
		reader: reader,
	}
}

func (r *Resource) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if r.URL == "" {
		return fmt.Errorf("url is required")
	}
	return nil
}

// GetResource returns the resource with the given name.
func (t *Tool) GetResource(name string) (*Resource, error) {
	for _, r := range t.Resources {
		if r != nil && r.Name == name {
			return r, nil
		}
	}
	return nil, fmt.Errorf("resource %q not found", name)
}

// LoadResources loads all resources in the tool.
// obs is used to cache the resources.
func (t *Tool) LoadResources(ctx context.Context, obs jetstream.ObjectStore) error {
	for i, resource := range t.Resources {
		if resource == nil {
			return fmt.Errorf("resource at index %d is nil", i)
		}

		if err := resource.Validate(); err != nil {
			return fmt.Errorf("validate resource %q: %w", resource.Name, err)
		}

		if err := resource.load(ctx, obs); err != nil {
			return fmt.Errorf("load resource %q: %w", resource.Name, err)
		}
	}
	return nil
}

// CloseResources closes all resources in the tool.
func (t *Tool) CloseResources() error {
	var errs []string
	for _, resource := range t.Resources {
		if resource == nil {
			continue
		}

		if err := resource.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("close resource %q: %v", resource.Name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing resources: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (r *Resource) Close() error {
	if r.reader == nil {
		return nil
	}
	return r.reader.Close()
}

// load downloads the resource URL and sets data to the reader.
// obs is used to cache the resource.
func (r *Resource) load(ctx context.Context, obs jetstream.ObjectStore) error {
	rr, err := cachefetch.Open(ctx, r.URL, obs)
	if err != nil {
		return fmt.Errorf("get resource %s: %w", r.URL, err)
	}

	r.reader = rr
	return nil
}
