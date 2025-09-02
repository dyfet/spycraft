// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 David Sugar <tychosoft@gmail.com>

package main

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"spycraft/lib/service"
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
		service.Debugf(4, "PACKET %v = %v", sourceIP, targetIP)
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
			remotePort := targetPort
			remoteIP := targetIP
			incoming := false
			if sourcePort == config.Port && sourceIP.Equal(config.Host) {
				incoming = true
			} else if targetPort == config.Port && targetIP.Equal(config.Host) {
				remotePort = sourcePort
				remoteIP = sourceIP
			} else {
				continue
			}
			service.Debugf(3, "UDP %v/%v to %v/%v", sourceIP, sourcePort, targetIP, targetPort)
			msg := &SIPMessage{
				Data:       udp.Payload,
				Incoming:   incoming,
				RemoteIP:   remoteIP,
				RemotePort: remotePort,
				Timestamp:  packet.Metadata().Timestamp,
			}
			messages <- msg
		}
	}
}

func Scan(handle *pcap.Handle) {
	service.Infof("Scanning for %v/%v", config.Host, config.Port)
	defer handle.Close()
	source := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range source.Packets() {
		if packet != nil {
			packets <- packet
		}
	}
}

func Capture(ctx context.Context, handle *pcap.Handle, wg *sync.WaitGroup) {
	service.Noticef("starting capture from %s for %v/%v", config.Device, config.Host, config.Port)
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
				err := fmt.Errorf("packet source closed")
				service.Error(err)
				return
			}
			packets <- packet
		}
	}
}
