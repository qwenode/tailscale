// Copyright (c) 2020 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (!linux && !windows && !darwin) || (darwin && ts_macext)
// +build !linux,!windows,!darwin darwin,ts_macext

package netns

import (
	"syscall"

	"github.com/qwenode/tailscale/types/logger"
)

func control(logger.Logf) func(network, address string, c syscall.RawConn) error {
	return controlC
}

// controlC does nothing to c.
func controlC(network, address string, c syscall.RawConn) error {
	return nil
}
