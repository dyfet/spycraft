// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 David Sugar <tychosoft@gmail.com>

package main

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"

	"spycraft/lib/byteshark"
)

type SIPMessage struct {
	Data       []byte
	RemoteIP   net.IP
	RemotePort uint16
	Incoming   bool
	Timestamp  time.Time
}

func Messages(wg *sync.WaitGroup) {
	stacks := make(map[string]int)
	defer wg.Done()

	var parts_store [4][]byte
	var fields_store [4][]byte
	var headers_store [64][]byte
	var key, value, callid, version []byte
	for {
		message := <-messages
		if message == nil {
			return
		}

		parts := parts_store[:0]
		count := byteshark.SplitSections(message.Data, []byte("\r\n\r\n"), &parts)
		if count < 1 || count > 2 {
			continue
		}

		headers := headers_store[:0]
		count = byteshark.SplitSections(parts[0], []byte("\r\n"), &headers)
		if count < 2 || count >= cap(headers) {
			continue
		}

		fields := fields_store[:0]
		byteshark.SplitSections(headers[0], []byte(" "), &fields)
		if len(fields) != 3 {
			continue
		}

		if !bytes.HasPrefix(fields[0], []byte("SIP/")) {
			// Check Request
			version = fields[2]
			if !bytes.HasPrefix(version, []byte("SIP/")) {
				continue
			}
		}

		agent := []byte("")
		mode := []byte("Unknown")
		for i := 1; i < len(headers); i++ {
			key, value = byteshark.SplitKeypair(headers[i], ':')
			if byteshark.MatchKeyword(key, []byte("call-id")) {
				callid = value
				continue
			}
			if byteshark.MatchKeyword(key, []byte("user-agent")) {
				agent = value
				mode = []byte("Agent")
				continue
			}
			if byteshark.MatchKeyword(key, []byte("server")) {
				agent = value
				mode = []byte("Server")
				continue
			}
		}
		if len(callid) == 0 {
			continue // lets skip non-call sip traffic
		}

		stack := fmt.Sprintf("%v:%v", message.RemoteIP, message.RemotePort)
		stacks[stack]++
		if stacks[stack] == 1 {
			fmt.Printf("%s %v:%v %s %s\n", version, message.RemoteIP, message.RemotePort, mode, agent)
		}
	}
}
