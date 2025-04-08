// SPDX-FileCopyrightText: 2024 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package script

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jumonmd/gengo/chat"
	chatsvc "github.com/jumonmd/jumon/chat"
	"github.com/jumonmd/jumon/internal/dataurl"
	"github.com/jumonmd/jumon/internal/tracer"
	"github.com/jumonmd/jumon/tool"
	"github.com/nats-io/nats.go"
)

// Run runs the given jumon script and returns the final output.
func Run(ctx context.Context, nc *nats.Conn, scr *Script) (json.RawMessage, error) {
	slog.Info("run script", "name", scr.Name)

	ctx, span := tracer.Start(ctx, nc, "script.run")
	defer span.End()

	slog.Debug("parse steps", "script", scr.Content)
	steps, preface, err := scr.Steps()
	if err != nil {
		span.SetError(fmt.Errorf("parse steps: %w", err))
		return nil, fmt.Errorf("parse steps: %w", err)
	}

	history := &chat.Request{
		Messages: []chat.Message{},
	}

	// construct initial prompt
	initialPrompt, err := initialPrompt(preface, scr.InputURL)
	if err != nil {
		span.SetError(fmt.Errorf("construct initial prompt: %w", err))
		return nil, fmt.Errorf("construct initial prompt: %w", err)
	}

	if initialPrompt != "" {
		history.Messages = append(history.Messages, chat.NewTextMessage(chat.MessageRoleHuman, initialPrompt))
	}

	// if no steps, create a single step with the initial prompt
	if len(steps) == 0 {
		steps = []*Step{{Level: 1, Content: initialPrompt}}
	}

	slog.Debug("initial prompt", "prompt", initialPrompt, "steps", len(steps))

	for i, step := range steps {
		ctx, sspan := tracer.Start(ctx, nc, "script.step.run")
		defer sspan.End()
		slog.Debug("run step", "index", i+1, "step", step.Content)

		req := stepRequest(scr, step, history)
		sspan.SetRequest(req)

		history.Messages = append(history.Messages, req.Messages[len(req.Messages)-1])

		// run step
		slog.Debug("run step", "step", step.Markdown())
		resp, err := runStep(ctx, nc, req, scr.Tools)
		if err != nil {
			sspan.SetError(fmt.Errorf("run step: %w", err))
			return nil, fmt.Errorf("run step: %w", err)
		}

		history.Messages = append(history.Messages, resp.Messages...)
		sspan.SetResponse(resp)
	}

	output, err := finalOutput(history.Messages[len(history.Messages)-1])
	if err != nil {
		span.SetError(fmt.Errorf("final output: %w", err))
		return nil, fmt.Errorf("final output: %w", err)
	}
	span.SetResponse(output)

	return output, nil
}

// initialPrompt constructs the initial prompt from preface and input URL.
func initialPrompt(preface, inputURL string) (string, error) {
	initialPrompt := preface
	if inputURL != "" {
		input, mimetype, err := dataurl.Decode(inputURL)
		if err != nil {
			return "", fmt.Errorf("decode input: %w", err)
		}
		if strings.HasPrefix(mimetype, "text/") || strings.HasPrefix(mimetype, "application/") {
			initialPrompt = fmt.Sprintf("%s\n\nINPUT:\n%s", initialPrompt, string(input))
		}
	}
	return initialPrompt, nil
}

// stepRequest prepares the chat request message for a script step.
func stepRequest(scr *Script, step *Step, history *chat.Request) *chat.Request {
	req := &chat.Request{
		Model:    scr.Model,
		Messages: history.Messages,
	}
	if scr.ModelConfig != nil {
		req.Config = *scr.ModelConfig
	}
	for _, tl := range scr.Tools {
		req.Tools = append(req.Tools, tl.ChatTool())
	}

	// msg is divided into a special check part and a normal content part.
	msg := chat.NewTextMessage(chat.MessageRoleHuman, removeChecks(step.Markdown()))
	if checks := parseChecks(step.Markdown()); checks != "" {
		msg.Content = append(msg.Content, chat.ContentPart{
			Type: "check",
			Text: checks,
		})
	}
	req.Messages = append(req.Messages, msg)

	return req
}

func runStep(ctx context.Context, nc *nats.Conn, req *chat.Request, tools []tool.Tool) (*chat.Response, error) {
	resp, err := chatsvc.Generate(ctx, nc, req)
	if err != nil {
		return nil, fmt.Errorf("chat generate: %w", err)
	}

	req.Messages = append(req.Messages, resp.Messages...)
	for _, msg := range resp.ToolCalls() {
		slog.Info("tool call", "tool", msg.ToolCall.Name, "args", msg.ToolCall.Arguments)
		tl := tool.Tool{}
		for _, t := range tools {
			if t.Name != msg.ToolCall.Name {
				continue
			}
			tl = t
		}
		if tl.Name == "" {
			return nil, fmt.Errorf("tool not found: %s", msg.ToolCall.Name)
		}
		tl.SetInput([]byte(msg.ToolCall.Arguments))
		output, err := tool.Run(ctx, nc, tl)
		if err != nil {
			return nil, fmt.Errorf("tool execute: %w", err)
		}

		slog.Debug("tool call response", "call", msg.ToolCall, "response", string(output))
		resp.Messages = append(resp.Messages, chat.NewToolResponseMessage(msg.ToolCall.Name, msg.ToolCall.ID, string(output)))
	}

	return resp, nil
}

func finalOutput(resp chat.Message) (json.RawMessage, error) {
	if resp.Role != chat.MessageRoleAI {
		slog.Error("final output is not an AI message", "role", resp.Role, "message", resp.ContentString())
		return nil, fmt.Errorf("last message is not an AI message")
	}

	// valid json return as is
	if json.Valid([]byte(resp.ContentString())) {
		return json.RawMessage(resp.ContentString()), nil
	}

	// json marshal if not valid json
	content, err := json.Marshal(resp.ContentString())
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	return content, nil
}
