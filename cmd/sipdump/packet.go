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
			return
		}

		var sourceIP net.IP
		var targetIP net.IP
		if config.Host.To4() != nil {
			ipLayer := packet.Layer(layers.LayerTypeIPv4)
			if ipLayer == nil {
				return
			}
			ip, _ := ipLayer.(*layers.IPv4)
			sourceIP = ip.SrcIP
			targetIP = ip.DstIP
		} else if config.Host.To16() != nil {
			ipLayer := packet.Layer(layers.LayerTypeIPv6)
			if ipLayer == nil {
				return
			}
			ip, _ := ipLayer.(*layers.IPv6)
			sourceIP = ip.SrcIP
			targetIP = ip.DstIP
		}
		if !sourceIP.Equal(config.Host) && !targetIP.Equal(config.Host) {
			return
		}

		var sourcePort uint16
		var targetPort uint16
		udpLayer := packet.Layer(layers.LayerTypeUDP)
		if udpLayer != nil {
			udp, _ := udpLayer.(*layers.UDP)
			sourcePort = uint16(udp.SrcPort)
			targetPort = uint16(udp.DstPort)
			//timestamp := packet.Metadata().Timestamp
			if sourcePort != config.Port && targetPort != config.Port {
				return
			}

			fmt.Printf("--- UDP %v/%v to %v/%v\r\n", sourceIP, sourcePort, targetIP, targetPort)
			os.Stdout.Write(udp.Payload)
		}
	}
}

func Scan(handle *pcap.Handle) {
	fmt.Printf("Scanning for %v/%v\n", config.Host, config.Port)
	defer handle.Close()
	source := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range source.Packets() {
		if packet != nil {
			packets <- packet
		}
	}
}

func Capture(ctx context.Context, handle *pcap.Handle, wg *sync.WaitGroup) {
	fmt.Printf("starting capture from %s for %v/%v\n", config.Device, config.Host, config.Port)
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
