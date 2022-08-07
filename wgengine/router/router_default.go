// Copyright (c) 2020 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !windows && !linux && !darwin && !openbsd && !freebsd
// +build !windows,!linux,!darwin,!openbsd,!freebsd

package router

import (
	"fmt"
	"runtime"

	"github.com/qwenode/tailscale/types/logger"
	"github.com/qwenode/tailscale/wgengine/monitor"
	"golang.zx2c4.com/wireguard/tun"
)

func newUserspaceRouter(logf logger.Logf, tunDev tun.Device, linkMon *monitor.Mon) (Router, error) {
	return nil, fmt.Errorf("unsupported OS %q", runtime.GOOS)
}

func cleanup(logf logger.Logf, interfaceName string) {
	// Nothing to do here.
}
