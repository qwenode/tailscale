// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

//go:build !windows

package osshare

import (
	"github.com/qwenode/tailscale/types/logger"
)

func SetFileSharingEnabled(enabled bool, logf logger.Logf) {}
