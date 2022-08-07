// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

//go:build !linux || ts_omit_aws

package awsstore

import (
	"fmt"
	"runtime"

	"github.com/qwenode/tailscale/ipn"
	"github.com/qwenode/tailscale/types/logger"
)

func New(logger.Logf, string) (ipn.StateStore, error) {
	return nil, fmt.Errorf("AWS store is not supported on %v", runtime.GOOS)
}
