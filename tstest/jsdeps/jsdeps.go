// Copyright (c) 2022 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
	_ "math/rand"
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
