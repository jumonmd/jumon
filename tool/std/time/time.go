// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package time

import (
	"strconv"
	"time"

	"github.com/nats-io/nats.go/micro"
)

func New(g micro.Group) {
	timeg := g.AddGroup("time")
	timeg.AddEndpoint("now", micro.HandlerFunc(handleNow))
	timeg.AddEndpoint("sleep", micro.HandlerFunc(handleSleep))
}

func handleNow(r micro.Request) {
	r.RespondJSON(time.Now().Format(time.RFC3339))
}

func handleSleep(r micro.Request) {
	sec, _ := strconv.Atoi(string(r.Data()))
	if sec <= 1 {
		sec = 1
	}
	time.Sleep(time.Second * time.Duration(sec))
	r.RespondJSON(sec)
}
