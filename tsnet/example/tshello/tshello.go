// Copyright (c) 2021 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The tshello server demonstrates how to use Tailscale as a library.
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"

	"github.com/qwenode/tailscale/client/tailscale"
	"github.com/qwenode/tailscale/tsnet"
)

var (
	addr = flag.String("addr", ":80", "address to listen on")
)

func main() {
	flag.Parse()
	s := new(tsnet.Server)
	ln, err := s.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	if *addr == ":443" {
		ln = tls.NewListener(ln, &tls.Config{
			GetCertificate: tailscale.GetCertificate,
		})
	}
	log.Fatal(http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		who, err := tailscale.WhoIs(r.Context(), r.RemoteAddr)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, "<html><body><h1>Hello, world!</h1>\n")
		fmt.Fprintf(w, "<p>You are <b>%s</b> from <b>%s</b> (%s)</p>",
			html.EscapeString(who.UserProfile.LoginName),
			html.EscapeString(firstLabel(who.Node.ComputedName)),
			r.RemoteAddr)
	})))
}

func firstLabel(s string) string {
	if i := strings.Index(s, "."); i != -1 {
		return s[:i]
	}
	return s
}
