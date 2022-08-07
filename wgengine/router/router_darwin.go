// Copyright (c) 2020 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package router

import (
	"github.com/qwenode/tailscale/types/logger"
	"github.com/qwenode/tailscale/wgengine/monitor"
	"golang.zx2c4.com/wireguard/tun"
)

func newUserspaceRouter(logf logger.Logf, tundev tun.Device, linkMon *monitor.Mon) (Router, error) {
	return newUserspaceBSDRouter(logf, tundev, linkMon)
}

func cleanup(logger.Logf, string) {
	// Nothing to do.
}
