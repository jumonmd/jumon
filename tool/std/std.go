// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package std

import (
	"github.com/jumonmd/jumon/tool/std/time"
	"github.com/nats-io/nats.go/micro"
)

func New(g micro.Group) {
	stdg := g.AddGroup("std")
	time.New(stdg)
}
