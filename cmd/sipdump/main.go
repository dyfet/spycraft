// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 David Sugar <tychosoft@gmail.com>

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"gopkg.in/ini.v1"

	"spycraft/lib/byteshark"
)

type Config struct {
	Device      string `ini:"device" arg:"-"`
	Filter      string `ini:"filter" arg:"-f,--filter" help:"bpi filter"`
	Promiscuous bool   `ini:"promiscuous" arg:"-p" help:"promiscuous mode"`
	Snapshot    int32  `ini:"snapshot" arg:"-s,--snapshot" help:"snapshot size"`
	Timeout     int    `ini:"timeout" help:"msec capture timeout"`

	Capture bool   `ini:"-" arg:"-c,--capture" help:"run in capture mode"`
	Host    net.IP `ini:"-" arg:"--host" help:"host to reference"`
	Port    uint16 `ini:"-" arg:"--port" help:"port to reference"`
	Path    string `ini:"-" arg:"positional" help:"pcap file or eth device"`
}

type Pipelines struct {
	Capture int `ini:"capture"`
	Scan    int `ini:"scan"`
}

var (
	// bind Makefile config
	etcPrefix = "/etc"

	config = Config{
		Device: "lo",
		// example:		Filter:     "host 127.0.0.1 and (tcp or udp port 5060)",
		Filter:   "udp port 5060",
		Snapshot: 1600,
		Timeout:  500,
		Capture:  os.Geteuid() == 0 || os.Getpid() == 1 || os.Getppid() == 1,
	}

	pipelines = Pipelines{
		Capture: 32,
		Scan:    128,
	}

	packets chan gopacket.Packet
)

func (Config) Description() string {
	return "sipdump - export sip messages"
}

func checkPcapPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("path %s: is directory", path)
	}

	// Check for .pcap suffix (case-insensitive if needed)
	if !strings.HasSuffix(path, ".pcap") {
		return fmt.Errorf("path %s: must be .pcap file", path)
	}
	return nil
}

func main() {
	configs, err := ini.LoadSources(ini.LoadOptions{Loose: true, Insensitive: true}, etcPrefix+"/spycraft.conf")
	if err == nil {
		// map and reset from args if not default
		configs.MapTo(&config)
		configs.Section("server").MapTo(&config)
		configs.Section("pipelines").MapTo(&pipelines)
	} else {
		log.Fatal(err)
	}

	arg.MustParse(&config)
	if config.Capture && config.Port == 0 {
		config.Port = byteshark.ExtractPortFromBPF(config.Filter)
	}
	if config.Capture {
		if len(config.Path) > 0 {
			config.Device = config.Path
		}
		port := byteshark.ExtractPortFromBPF(config.Filter)
		if port != 0 {
			config.Port = port
		}
		host := byteshark.ExtractHostFromBPF(config.Filter)
		if host != nil {
			config.Host = host
		}
		if config.Host == nil {
			config.Host, err = byteshark.GetInterfaceIP(config.Device)
			if err != nil {
				log.Fatal(err)
			}
		}
		if config.Host != nil {
			config.Filter = byteshark.InjectHostIntoBPF(config.Filter, config.Host)
		}
	}
	if config.Host == nil {
		log.Fatal("No host to reference")
	}
	if config.Port == 0 {
		config.Port = 5060
	}

	var wg sync.WaitGroup
	if config.Capture {
		timeout := time.Duration(config.Timeout) * time.Millisecond
		handle, err := pcap.OpenLive(config.Device, config.Snapshot, config.Promiscuous, timeout)
		if err != nil {
			log.Fatal(err)
		}

		if err := handle.SetBPFFilter(config.Filter); err != nil {
			handle.Close()
			log.Fatal(err)
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		wg.Add(2)
		packets = make(chan gopacket.Packet, pipelines.Capture)
		go Process(&wg)
		go Capture(ctx, handle, &wg)
		<-ctx.Done()
	} else {
		if len(config.Path) == 0 {
			log.Fatal("Missing pcap file")
		}
		err = checkPcapPath(config.Path)
		if err != nil {
			log.Fatal(err)
		}
		config.Filter = byteshark.BuildBPFFilter(config.Host, config.Port)
		handle, err := pcap.OpenOffline(config.Path)
		if err != nil {
			log.Fatal(err)
		}

		if err := handle.SetBPFFilter(config.Filter); err != nil {
			handle.Close()
			log.Fatal(err)
		}
		wg.Add(1)
		packets = make(chan gopacket.Packet, pipelines.Capture)
		go Process(&wg)
		Scan(handle)
	}
	packets <- nil
	wg.Wait()
}
