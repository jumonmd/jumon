// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/jumonmd/gengo/chat"
	"github.com/jumonmd/jumon/internal/tracer"
	"github.com/nats-io/nats.go"
)

// PrintNotifications subscribes to the notification subject and prints the notifications.
func PrintNotifications(nc *nats.Conn, subject string, w io.Writer) (*nats.Subscription, error) {
	return nc.Subscribe(subject, func(msg *nats.Msg) {
		n := &tracer.Notification{}
		err := json.Unmarshal(msg.Data, n)
		if err != nil {
			return
		}

		if n.Name == "script.step.run" {
			switch n.On {
			case "request":
				printStepRequest(w, n.Content)
			case "response":
				printStepResponse(w, n.Content)
			case "error":
				printError(w, n.Content)
			}
		}
		if n.Name == "script.run" {
			switch n.On {
			case "response":
				printScriptResponse(w, n.Content)
			case "error":
				printError(w, n.Content)
			}
		}
	})
}

func printScriptResponse(w io.Writer, s string) {
	fmt.Fprintln(w, "\nFinal Output:", s)
}

func printStepRequest(w io.Writer, s string) {
	req := &chat.Request{}
	err := json.Unmarshal([]byte(s), req)
	if err != nil {
		return
	}

	if len(req.Messages) == 0 {
		return
	}

	last := req.Messages[len(req.Messages)-1]
	if last.Role == chat.MessageRoleHuman {
		fmt.Fprintf(w, "  %s\n", last.String())
	}
}

func printStepResponse(w io.Writer, s string) {
	res := &chat.Response{}
	err := json.Unmarshal([]byte(s), res)
	if err != nil {
		return
	}

	fmt.Fprintf(w, "  %s\n", res.String())
}

func printError(w io.Writer, s string) {
	fmt.Fprintln(w, "Error: ", s)
}
