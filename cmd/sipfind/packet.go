// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 David Sugar <tychosoft@gmail.com>

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func Process(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		packet := <-packets
		if packet == nil { // end of input marker...
			messages <- nil
			return
		}

		var sourceIP net.IP
		ip4Layer := packet.Layer(layers.LayerTypeIPv4)
		ip6Layer := packet.Layer(layers.LayerTypeIPv6)
		if ip4Layer == nil && ip6Layer == nil {
			continue
		}
		if ip4Layer != nil {
			ip, _ := ip4Layer.(*layers.IPv4)
			sourceIP = ip.SrcIP
		} else if ip6Layer != nil {
			ip, _ := ip6Layer.(*layers.IPv6)
			sourceIP = ip.SrcIP
		}

		var sourcePort uint16
		udpLayer := packet.Layer(layers.LayerTypeUDP)
		if udpLayer != nil {
			udp, _ := udpLayer.(*layers.UDP)
			sourcePort = uint16(udp.SrcPort)
			msg := &SIPMessage{
				Data:       udp.Payload,
				RemoteIP:   sourceIP,
				RemotePort: sourcePort,
				Timestamp:  packet.Metadata().Timestamp,
			}
			messages <- msg
		}
	}
}

func Scan(handle *pcap.Handle) {
	fmt.Printf("Searching %s\n", config.Path)
	defer handle.Close()
	source := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range source.Packets() {
		if packet != nil {
			packets <- packet
		}
	}
}

func Capture(ctx context.Context, handle *pcap.Handle, wg *sync.WaitGroup) {
	defer handle.Close()
	source := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := source.Packets()
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case packet, ok := <-packets:
			if !ok {
				fmt.Fprintf(os.Stderr, "*** packet source closed\n")
				return
			}
			packets <- packet
		}
	}
}
