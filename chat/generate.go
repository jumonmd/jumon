// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package chat

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jumonmd/gengo/chat"
	"github.com/jumonmd/jumon/internal/tracer"
	"github.com/nats-io/nats.go"
)

// Generate chat response using NATS service.
func Generate(ctx context.Context, nc *nats.Conn, req *chat.Request, opts ...chat.Option) (*chat.Response, error) {
	chatdata, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal chat: %w", err)
	}

	opt := chat.NewOptions(opts...)

	headers := tracer.HeadersFromContext(ctx)
	if streamto, ok := ctx.Value("stream-to").(string); ok {
		headers.Set("stream-to", streamto)
	}
	headers.Set("baseurl", opt.BaseURL)

	resp, err := nc.RequestMsgWithContext(ctx, &nats.Msg{
		Subject: "chat.generate",
		Data:    chatdata,
		Header:  headers,
	})
	if err != nil {
		return nil, fmt.Errorf("nats: %w", err)
	}
	if errorCode := resp.Header.Get("Nats-Service-Error-Code"); errorCode != "" {
		errorMessage := resp.Header.Get("Nats-Service-Error")
		return nil, fmt.Errorf("%s: %s", errorCode, errorMessage)
	}

	chatresp := &chat.Response{}
	err = json.Unmarshal(resp.Data, chatresp)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return chatresp, nil
}
