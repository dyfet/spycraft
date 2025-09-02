// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 David Sugar <tychosoft@gmail.com>

package byteshark

import (
	"bytes"

	"github.com/google/gopacket/reassembly"
)

type SIPStream struct {
	//	net, Transport gopacket.Flow
	buf     bytes.Buffer
	msgChan chan []byte
}

func (s *SIPStream) ReassembledSG(sg reassembly.ScatterGather, ac reassembly.AssemblerContext) {
	data := sg.Fetch(0)
	s.buf.Write(data)
	for {
		msg, ok := ExtractSIPMessage(s.buf.Bytes())
		if !ok {
			break
		}
		s.msgChan <- msg
		s.buf.Next(len(msg)) // remove processed bytes
	}
}

func ExtractSIPMessage(data []byte) ([]byte, bool) {
	headerEnd := bytes.Index(data, []byte("\r\n\r\n"))
	if headerEnd == -1 {
		return nil, false
	}

	headers := data[:headerEnd+4]
	contentLen := ParseContentLength(headers)
	totalLen := headerEnd + 4 + contentLen
	if len(data) < totalLen {
		return nil, false
	}

	return data[:totalLen], true
}
