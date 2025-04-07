// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jumonmd/gengo"
	"github.com/jumonmd/gengo/chat"
	"github.com/jumonmd/jumon/internal/config"
	"github.com/jumonmd/jumon/internal/errors"
	"github.com/jumonmd/jumon/internal/tracer"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

var (
	ErrBadRequest   = errors.New(400100, "bad request")
	ErrGeneration   = errors.New(500100, "chat generation failed")
	ErrVerify       = errors.New(500101, "verify operation failed")
	ErrVerifyFailed = errors.New(500102, "verify failed")
)

// NewService creates a chat service for NATS micro service.
// subject: chat.generate
func NewService(nc *nats.Conn) (micro.Service, error) {
	svc, err := micro.AddService(nc, micro.Config{
		Name:        "jumon_chat",
		Version:     "0.1.0",
		Description: "jumon chat service",
		QueueGroup:  "chat",
	})
	if err != nil {
		slog.Error("chat service", "status", "failed", "error", err)
		return nil, fmt.Errorf("create chat service: %w", err)
	}

	chtgroup := svc.AddGroup("chat")
	chtgroup.AddEndpoint("generate", micro.HandlerFunc(func(r micro.Request) {
		go generateHandler(nc, r)
	}))

	slog.Info("chat service", "status", "started")
	return svc, nil
}

// generateHandler handles the chat generate request.
func generateHandler(nc *nats.Conn, r micro.Request) {
	slog.Info("chat generate", "status", "started")

	slog.Debug("chat generate", "headers", r.Headers())
	ctx, span := tracer.Start(tracer.NewContext(r.Headers()), nc, "chat.generate")
	defer span.End()

	req := &chat.Request{}
	err := json.Unmarshal(r.Data(), req)
	if err != nil {
		span.SetError(ErrBadRequest.Wrap(err))
		r.Error(ErrBadRequest.ServiceError(err))
		return
	}

	var streamer chat.Streamer
	streamTo := r.Headers().Get("stream-to")
	if streamTo != "" {
		slog.Debug("chat generate", "stream-to", streamTo)
		streamer = func(resp *chat.StreamResponse) error {
			data, err := json.Marshal(resp)
			if err != nil {
				return fmt.Errorf("marshal response: %w", err)
			}
			err = nc.PublishMsg(&nats.Msg{
				Subject: streamTo,
				Data:    data,
			})
			if err != nil {
				return fmt.Errorf("publish stream: %w", err)
			}
			return nil
		}
	}
	span.SetRequest(req)

	// extract the checks from the message and separate them from the normal content parts.
	// the check parts are processed separately from the chat context.
	checks := extractCurrentChecks(req)
	removeChecks(req)

	opts := []chat.Option{chat.WithStream(streamer)}
	baseURL := r.Headers().Get("baseurl")
	if baseURL != "" {
		opts = append(opts, chat.WithBaseURL(baseURL))
	}

	resp, err := gengo.Generate(ctx, req, opts...)
	if err != nil {
		slog.Info("chat generate", "status", "completion failed", "err", err)
		span.SetError(ErrGeneration.Wrap(err))
		r.Error(ErrGeneration.ServiceError(err))
		return
	}

	slog.Debug("chat generate", "response", resp)
	span.SetResponse(resp)

	// check response with checks directive
	if checks != "" {
		handleVerify(ctx, nc, r, checks, req, resp)
	}

	r.RespondJSON(resp, micro.WithHeaders(span.Headers()))
}

// handleVerify handles the verify request using AI.
func handleVerify(ctx context.Context, nc *nats.Conn, r micro.Request, checks string, req *chat.Request, resp *chat.Response) {
	slog.Info("chat verify", "status", "started", "checks", checks)
	defaultVerifyModel, err := config.Get(ctx, nc, config.DefaultVerifyModel)
	if err != nil {
		slog.Warn("chat generate", "status", "get default verify model from configfailed", "error", err)
	}
	if defaultVerifyModel != "" {
		req.Model = defaultVerifyModel
	}

	ctx, cspan := tracer.Start(ctx, nc, "chat.verify")
	defer cspan.End()

	cspan.SetRequest(struct {
		Response *chat.Response
		Checks   string
	}{
		Response: resp,
		Checks:   checks,
	})
	passed, err := VerifyResponse(ctx, req, resp, checks)
	if err != nil {
		slog.Error("chat verify", "status", "verify error", "error", err)
		cspan.SetError(ErrVerify.Wrap(err))
		r.Error(ErrVerify.ServiceError(err))
		return
	}
	cspan.SetResponse(passed)

	if !passed {
		err := fmt.Errorf("checks: %s response: %s", checks, resp.String())
		slog.Info("chat verify", "status", "verify failed", "error", err)
		cspan.SetError(ErrVerifyFailed.Wrap(err))
		r.Error(ErrVerifyFailed.ServiceError(err))
		return
	}
}

// removeChecks removes custom check message content part from the request.
func removeChecks(req *chat.Request) {
	for i, msg := range req.Messages {
		if msg.Role != chat.MessageRoleHuman {
			continue
		}
		parts := []chat.ContentPart{}
		for _, part := range msg.Content {
			if part.Type == "check" {
				continue
			}
			parts = append(parts, part)
		}
		req.Messages[i].Content = parts
	}
}

// extractCurrentChecks returns the current message checks directive content part.
func extractCurrentChecks(req *chat.Request) string {
	msg := chat.Message{}
	for _, m := range req.Messages {
		if m.Role != chat.MessageRoleHuman {
			continue
		}
		msg = m
	}

	for _, part := range msg.Content {
		if part.Type == "check" {
			return part.Text
		}
	}
	return ""
}
