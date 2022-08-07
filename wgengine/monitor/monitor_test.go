// Copyright (c) 2021 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package monitor

import (
	"flag"
	"testing"
	"time"

	"github.com/qwenode/tailscale/net/interfaces"
)

func TestMonitorStartClose(t *testing.T) {
	mon, err := New(t.Logf)
	if err != nil {
		t.Fatal(err)
	}
	mon.Start()
	if err := mon.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMonitorJustClose(t *testing.T) {
	mon, err := New(t.Logf)
	if err != nil {
		t.Fatal(err)
	}
	if err := mon.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMonitorInjectEvent(t *testing.T) {
	mon, err := New(t.Logf)
	if err != nil {
		t.Fatal(err)
	}
	defer mon.Close()
	got := make(chan bool, 1)
	mon.RegisterChangeCallback(func(changed bool, state *interfaces.State) {
		select {
		case got <- true:
		default:
		}
	})
	mon.Start()
	mon.InjectEvent()
	select {
	case <-got:
		// Pass.
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for callback")
	}
}

var monitor = flag.String("monitor", "", `go into monitor mode like 'route monitor'; test never terminates. Value can be either "raw" or "callback"`)

func TestMonitorMode(t *testing.T) {
	switch *monitor {
	case "":
		t.Skip("skipping non-test without --monitor")
	case "raw", "callback":
	default:
		t.Skipf(`invalid --monitor value: must be "raw" or "callback"`)
	}
	mon, err := New(t.Logf)
	if err != nil {
		t.Fatal(err)
	}
	switch *monitor {
	case "raw":
		for {
			msg, err := mon.om.Receive()
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("msg: %#v", msg)
		}
	case "callback":
		mon.RegisterChangeCallback(func(changed bool, st *interfaces.State) {
			t.Logf("cb: changed=%v, ifSt=%v", changed, st)
		})
		mon.Start()
		select {}
	}
}
