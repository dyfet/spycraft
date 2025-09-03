// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 David Sugar <tychosoft@gmail.com>

package main

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"spycraft/lib/byteshark"
	"spycraft/lib/service"
)

type SIPMessage struct {
	Data       []byte
	RemoteIP   net.IP
	RemotePort uint16
	Incoming   bool
	Timestamp  time.Time
}

func Messages(wg *sync.WaitGroup) {
	defer wg.Done()
	var parts_store [4][]byte
	var fields_store [4][]byte
	var headers_store [64][]byte
	var err error
	for {
		message := <-messages
		if message == nil {
			return
		}

		parts := parts_store[:0]
		count := byteshark.SplitSections(message.Data, []byte("\r\n\r\n"), &parts)
		if count < 1 {
			service.Error("Unable to split sections")
			continue
		}
		service.Debugf(4, "Split into %d sections", count)

		headers := headers_store[:0]
		count = byteshark.SplitSections(parts[0], []byte("\r\n"), &headers)
		if count < 2 {
			service.Error("No sip headers")
			continue
		}
		service.Debugf(5, "Split header into %d lines", count)

		fields := fields_store[:0]
		byteshark.SplitSections(headers[0], []byte(" "), &fields)
		if len(fields) != 3 {
			service.Error("Not a sip packet")
			continue
		}

		var version, status, reason, uri, method []byte
		incoming := message.Incoming // event direction...
		if bytes.HasPrefix(fields[0], []byte("SIP/")) {
			// Response
			version = fields[0]
			status = fields[1]
			reason = fields[2]
			incoming = !incoming // flip direction association on responses
			service.Debugf(3, "Response: %s %s %s", version, status, reason)
		} else {
			// Request
			method = fields[0]
			uri = fields[1]
			version = fields[2]
			service.Debugf(3, "Request: %s %s %s", method, uri, version)
		}

		var key, value, callid, collateid, agent []byte
		for i := 1; i < len(headers); i++ {
			key, value = byteshark.SplitKeypair(headers[i], ':')
			service.Debugf(6, "%s: %s", key, value)
			if byteshark.MatchKeyword(key, []byte("call-id")) {
				callid = value
				continue
			}
			if byteshark.MatchKeyword(key, []byte("x-collateid")) {
				collateid = value
				continue
			}
			// collect from incoming packets...
			if message.Incoming && byteshark.MatchKeyword(key, []byte("user-agent")) {
				agent = value
			}
		}
		if len(callid) == 0 {
			continue
		}
		event := &LegEvent{
			Method:    method,
			Selected:  nil,
			Timestamp: message.Timestamp,
			Endpoint:  message.RemoteIP,
			Port:      message.RemotePort,
		}
		if len(method) == 0 {
			event.Status, err = strconv.Atoi(string(status))
			if err != nil {
				service.Error(err)
				continue
			}
		}

		if event.Status >= 800 {
			continue
		}

		legid := fmt.Sprintf("%v/%v/%s", message.RemoteIP, message.RemotePort, callid)
		leg := legs[legid]
		if leg == nil && event.Status == 0 {
			// we should make sure this is not a re-invite...
			if byteshark.MatchKeyword(method, []byte("invite")) {
				leg = &Leg{
					Incoming: incoming,
					Created:  message.Timestamp,
					Updated:  message.Timestamp,
					Endpoint: message.RemoteIP,
					Port:     message.RemotePort,
				}

				// if we are the inviter, can set collation id immediately
				if !incoming {
					if len(collateid) > 0 {
						leg.Collated = string(collateid)
					} else {
						leg.Collated = string(callid)
					}
					service.Infof("outgoing leg %v/%v on %s", leg.Endpoint, leg.Port, leg.Collated)
				}

				if incoming {
					leg.States[0].Request = Active
					leg.States[1].Request = Invite
					leg.Agent = string(agent)
				} else {
					leg.States[0].Request = Invite
					leg.States[1].Request = Active
				}
				leg.States[0].Updated = message.Timestamp
				leg.States[1].Updated = message.Timestamp
				legs[legid] = leg
				continue
			}
		}

		// we strip events if we didn't see initial invite
		if leg == nil {
			continue
		}

		// collate if we are responding and nothing set
		if incoming && event.Status >= 180 && len(leg.Collated) == 0 {
			if len(collateid) > 0 {
				leg.Collated = string(collateid)
			} else {
				leg.Collated = string(callid)
			}
			service.Infof("incoming leg %v/%v on %s", leg.Endpoint, leg.Port, leg.Collated)
		}

		leg.Updated = message.Timestamp
		if message.Incoming && len(leg.Agent) == 0 {
			leg.Agent = string(agent) // fill from remote endpoint
		}
		if incoming {
			if len(leg.Agent) == 0 {
				leg.Agent = string(agent)
			}
			event.Selected = &leg.States[1]
		} else {
			event.Selected = &leg.States[0]
		}

		service.Debugf(3, "event for leg %s", legid)
		if event.Status >= 700 {
			leg.Dummy(event)
			continue
		}
		if event.Status >= 600 {
			leg.Dummy(event)
			continue
		}
		if event.Status >= 500 {
			leg.Dummy(event)
			continue
		}
		if event.Status >= 400 {
			leg.Dummy(event)
			continue
		}
		if event.Status >= 300 {
			leg.Dummy(event)
			continue
		}
		if event.Status >= 200 {
			leg.Dummy(event)
			continue
		}
		if event.Status >= 100 {
			leg.Dummy(event)
			continue
		}
		if byteshark.MatchKeyword(method, []byte("invite")) {
			leg.Dummy(event)
			continue
		}
		if byteshark.MatchKeyword(method, []byte("bye")) || byteshark.MatchKeyword(method, []byte("cancel")) {
			if len(leg.Collated) > 0 {
				service.Infof("ending leg %v/%v on %s", leg.Endpoint, leg.Port, leg.Collated)
			}
			leg.Finished = message.Timestamp
			leg.Dummy(event)
			continue
		}
		if byteshark.MatchKeyword(method, []byte("ack")) {
			leg.Dummy(event)
			continue
		}
		if byteshark.MatchKeyword(method, []byte("refer")) {
			leg.Dummy(event)
			continue
		}
		leg.Dummy(event)
	}
}
