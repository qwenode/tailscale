// Copyright (c) 2020 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ipnserver_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qwenode/tailscale/ipn"
	"github.com/qwenode/tailscale/ipn/ipnserver"
	"github.com/qwenode/tailscale/ipn/store/mem"
	"github.com/qwenode/tailscale/net/tsdial"
	"github.com/qwenode/tailscale/safesocket"
	"github.com/qwenode/tailscale/wgengine"
	"github.com/qwenode/tailscale/wgengine/netstack"
)

func TestRunMultipleAccepts(t *testing.T) {
	t.Skipf("TODO(bradfitz): finish this test, once other fires are out")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	td := t.TempDir()
	socketPath := filepath.Join(td, "tailscale.sock")

	logf := func(format string, args ...any) {
		format = strings.TrimRight(format, "\n")
		println(fmt.Sprintf(format, args...))
		t.Logf(format, args...)
	}

	s := safesocket.DefaultConnectionStrategy(socketPath)
	connect := func() {
		for i := 1; i <= 2; i++ {
			logf("connect %d ...", i)
			c, err := safesocket.Connect(s)
			if err != nil {
				t.Fatalf("safesocket.Connect: %v\n", err)
			}
			clientToServer := func(b []byte) {
				ipn.WriteMsg(c, b)
			}
			bc := ipn.NewBackendClient(logf, clientToServer)
			prefs := ipn.NewPrefs()
			bc.SetPrefs(prefs)
			c.Close()
		}
	}

	logTriggerTestf := func(format string, args ...any) {
		logf(format, args...)
		if strings.HasPrefix(format, "Listening on ") {
			connect()
		}
	}

	eng, err := wgengine.NewFakeUserspaceEngine(logf, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(eng.Close)

	opts := ipnserver.Options{}
	t.Logf("pre-Run")
	store := new(mem.Store)

	ln, _, err := safesocket.Listen(socketPath, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	err = ipnserver.Run(ctx, logTriggerTestf, ln, store, nil /* mon */, new(tsdial.Dialer), "dummy_logid", FixedEngine(eng), opts)
	t.Logf("ipnserver.Run = %v", err)
}

// FixedEngine returns a func that returns eng and a nil error.
func FixedEngine(eng wgengine.Engine) func() (wgengine.Engine, *netstack.Impl, error) {
	return func() (wgengine.Engine, *netstack.Impl, error) { return eng, nil, nil }
}
