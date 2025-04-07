// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jumonmd/jumon/internal/tracer"
	"github.com/nats-io/nats.go"
)

// PrintNotifications subscribes to the notification subject and prints the notifications.
func PrintNotifications(nc *nats.Conn, subject string) (*nats.Subscription, error) {
	return nc.Subscribe(subject, func(msg *nats.Msg) {
		n := &tracer.Notification{}
		err := json.Unmarshal(msg.Data, n)
		if err != nil {
			return
		}
		fmt.Printf("%s %s\n", n.On, n.Name)
		fristline := strings.Split(n.Content, "\n")[0]
		fmt.Println("  ", fristline)

		// if n.Name == "script.run" && n.On == "request" {

		// if n.Name == "script.run" {
		// 	var raw interface{}
		// 	_ = json.Unmarshal([]byte(n.Content), &raw)
		// 	// if content, ok := raw.(string); ok {
		// 	// 	fmt.Print(content)
		// 	// }
		// }
	})
}
