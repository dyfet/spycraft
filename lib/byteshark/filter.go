// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 David Sugar <tychosoft@gmail.com>

package byteshark

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

func ExtractHostFromBPF(filter string) net.IP {
	var hostRE = regexp.MustCompile(`(?i)(?:src|dst)?\s*host\s+([^\s]+)`)
	match := hostRE.FindStringSubmatch(filter)
	if len(match) > 1 {
		ip := net.ParseIP(strings.TrimSpace(match[1]))
		if ip != nil {
			return ip
		}
	}
	return nil
}

func ExtractPortFromBPF(filter string) uint16 {
	var portRE = regexp.MustCompile(`(?i)(?:udp|tcp)?\s*port\s+(\d+)`)
	match := portRE.FindStringSubmatch(filter)
	if len(match) > 1 {
		if port, err := strconv.Atoi(match[1]); err == nil {
			return uint16(port)
		}
	}
	return 0
}

func GetInterfaceIP(device string) (net.IP, error) {
	iface, err := net.InterfaceByName(device)
	if err != nil {
		return nil, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	var ipv6 net.IP
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP, nil // Prefer IPv4
			}
			if ipnet.IP.To16() != nil && ipv6 == nil {
				ipv6 = ipnet.IP // Save first non-loopback IPv6
			}
		}
	}

	if ipv6 != nil {
		return ipv6, nil
	}

	return nil, fmt.Errorf("no IP address found for interface %s", device)
}

func InjectHostIntoBPF(filter string, ip net.IP) string {
	if host := ExtractHostFromBPF(filter); host != nil {
		return filter // Host already present
	}

	trimmed := strings.TrimSpace(filter)
	if trimmed == "" {
		return fmt.Sprintf("host %s", ip.String())
	}

	return fmt.Sprintf("host %s and (%s)", ip.String(), trimmed)
}

func BuildBPFFilter(ip net.IP, port uint16) string {
	var protoPrefix string
	if ip.To4() != nil {
		protoPrefix = "host"
	} else {
		protoPrefix = "ip6 host"
	}

	if port > 0 {
		return fmt.Sprintf("%s %s and port %d", protoPrefix, ip.String(), port)
	}
	return fmt.Sprintf("%s %s", protoPrefix, ip.String())
}
