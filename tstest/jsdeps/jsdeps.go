// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

// Package jsdeps is a just a list of the packages we import in the
// JavaScript/WASM build, to let us test that our transitive closure of
// dependencies doesn't accidentally grow too large, since binary size
// is more of a concern.
package jsdeps

import (
	_ "bytes"
	_ "context"
	_ "encoding/hex"
	_ "encoding/json"
	_ "fmt"
	_ "log"
	_ "math/rand/v2"
	_ "net"
	_ "strings"
	_ "time"

	_ "github.com/qwenode/tailscale/control/controlclient"
	_ "github.com/qwenode/tailscale/ipn"
	_ "github.com/qwenode/tailscale/ipn/ipnserver"
	_ "github.com/qwenode/tailscale/net/netaddr"
	_ "github.com/qwenode/tailscale/net/netns"
	_ "github.com/qwenode/tailscale/net/tsdial"
	_ "github.com/qwenode/tailscale/safesocket"
	_ "github.com/qwenode/tailscale/tailcfg"
	_ "github.com/qwenode/tailscale/types/logger"
	_ "github.com/qwenode/tailscale/wgengine"
	_ "github.com/qwenode/tailscale/wgengine/netstack"
	_ "github.com/qwenode/tailscale/words"
	_ "golang.org/x/crypto/ssh"
)
