// Copyright (c) 2022 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package iosdeps is a just a list of the packages we import on iOS, to let us
// test that our transitive closure of dependencies on iOS doesn't accidentally
// grow too large, as we've historically been memory constrained there.
package iosdeps

import (
	_ "bufio"
	_ "bytes"
	_ "context"
	_ "crypto/rand"
	_ "crypto/sha256"
	_ "encoding/json"
	_ "errors"
	_ "fmt"
	_ "io"
	_ "io/fs"
	_ "io/ioutil"
	_ "log"
	_ "math"
	_ "net"
	_ "net/http"
	_ "os"
	_ "os/signal"
	_ "path/filepath"
	_ "runtime"
	_ "runtime/debug"
	_ "strings"
	_ "sync"
	_ "sync/atomic"
	_ "syscall"
	_ "time"
	_ "unsafe"

	_ "github.com/qwenode/tailscale/hostinfo"
	_ "github.com/qwenode/tailscale/ipn"
	_ "github.com/qwenode/tailscale/ipn/ipnlocal"
	_ "github.com/qwenode/tailscale/ipn/localapi"
	_ "github.com/qwenode/tailscale/log/logheap"
	_ "github.com/qwenode/tailscale/logtail"
	_ "github.com/qwenode/tailscale/logtail/filch"
	_ "github.com/qwenode/tailscale/net/dns"
	_ "github.com/qwenode/tailscale/net/netaddr"
	_ "github.com/qwenode/tailscale/net/tsdial"
	_ "github.com/qwenode/tailscale/net/tstun"
	_ "github.com/qwenode/tailscale/paths"
	_ "github.com/qwenode/tailscale/tempfork/pprof"
	_ "github.com/qwenode/tailscale/types/empty"
	_ "github.com/qwenode/tailscale/types/logger"
	_ "github.com/qwenode/tailscale/util/clientmetric"
	_ "github.com/qwenode/tailscale/util/dnsname"
	_ "github.com/qwenode/tailscale/version"
	_ "github.com/qwenode/tailscale/wgengine"
	_ "github.com/qwenode/tailscale/wgengine/router"
	_ "go4.org/mem"
	_ "golang.org/x/sys/unix"
	_ "golang.zx2c4.com/wireguard/device"
	_ "golang.zx2c4.com/wireguard/tun"
)
