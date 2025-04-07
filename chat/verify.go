// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package chat

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jumonmd/gengo"
	"github.com/jumonmd/gengo/chat"
)

const checkPromptTemplate = `
You are a helpful assistant that checks the response of the user.
The user will provide a response and a list of checks.
You will check the response against the checks and return a list of results.
Answer only with true or false.

Response:
%s

Checks:
%s
`

// VerifyResponse checks the response against the checks and returns true if the response is accepted.
func VerifyResponse(ctx context.Context, req *chat.Request, resp *chat.Response, checks string) (bool, error) {
	slog.Info("check response", "response", resp.String(), "checks", checks)
	prompt := fmt.Sprintf(checkPromptTemplate, resp.String(), checks)

	r := &chat.Request{
		Model: req.Model,
		Config: chat.ModelConfig{
			Temperature: 0.0001,
		},
		Messages: []chat.Message{chat.NewTextMessage(chat.MessageRoleHuman, prompt)},
	}
	resp, err := gengo.Generate(ctx, r)
	if err != nil {
		return false, fmt.Errorf("generate: %w", err)
	}

	slog.Debug("check response", "response", resp.String())
	if strings.Contains(strings.ToLower(resp.String()), "true") {
		return true, nil
	}

	// failed
	return false, nil
}
