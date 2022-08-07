// Copyright (c) 2022 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package tests serves a list of tests for github.com/qwenode/tailscale/cmd/viewer.
package tests

import (
	"fmt"
	"net/netip"
)

//go:generate go run github.com/qwenode/tailscale/cmd/viewer --type=StructWithPtrs,StructWithoutPtrs,Map,StructWithSlices

type StructWithoutPtrs struct {
	Int int
	Pfx netip.Prefix
}

type Map struct {
	Int                 map[string]int
	SliceInt            map[string][]int
	StructPtrWithPtr    map[string]*StructWithPtrs
	StructPtrWithoutPtr map[string]*StructWithoutPtrs
	StructWithoutPtr    map[string]StructWithoutPtrs
	SlicesWithPtrs      map[string][]*StructWithPtrs
	SlicesWithoutPtrs   map[string][]*StructWithoutPtrs
	StructWithoutPtrKey map[StructWithoutPtrs]int `json:"-"`

	// Unsupported views.
	SliceIntPtr      map[string][]*int
	PointerKey       map[*string]int        `json:"-"`
	StructWithPtrKey map[StructWithPtrs]int `json:"-"`
	StructWithPtr    map[string]StructWithPtrs
}

type StructWithPtrs struct {
	Value *StructWithoutPtrs
	Int   *int

	NoCloneValue *StructWithoutPtrs `codegen:"noclone"`
}

func (v *StructWithPtrs) String() string { return fmt.Sprintf("%v", v.Int) }

func (v *StructWithPtrs) Equal(v2 *StructWithPtrs) bool {
	return v.Value == v2.Value
}

type StructWithSlices struct {
	Values         []StructWithoutPtrs
	ValuePointers  []*StructWithoutPtrs
	StructPointers []*StructWithPtrs
	Structs        []StructWithPtrs
	Ints           []*int

	Slice    []string
	Prefixes []netip.Prefix
	Data     []byte
}
