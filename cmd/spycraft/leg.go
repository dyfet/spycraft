// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 David Sugar <tychosoft@gmail.com>

package main

import (
	"net"
	"time"
)

type CallState int

type State struct {
	Request CallState
	Status  int
	Updated time.Time
}

type LegEvent struct {
	Method    []byte
	Status    int
	Selected  *State
	Timestamp time.Time
	Endpoint  net.IP
	Port      uint16
}

type Leg struct {
	Collated  string // will have CallID if neither end has collation
	Agent     string
	Endpoint  net.IP
	Port      uint16
	Incoming  bool
	Pending   bool     // pending connection?
	Connected bool     // Leg ever connected?
	Final     int      // final status code of leg
	States    [2]State // Local and remote state
	Created   time.Time
	Updated   time.Time
	Finished  time.Time
}

const (
	Invite CallState = iota
	ReInvite
	Joined
	Bye
	Hold
	Xfer
	Ring
	Answer
	Active
	Failed
)

func (leg *Leg) Dummy(event *LegEvent) {
}
