// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 David Sugar <tychosoft@gmail.com>

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"gopkg.in/ini.v1"
)

type Config struct {
	Device      string `ini:"device" arg:"-"`
	Promiscuous bool   `ini:"promiscuous" arg:"-p" help:"promiscuous mode"`
	Snapshot    int32  `ini:"snapshot" arg:"-s,--snapshot" help:"snapshot size"`
	Timeout     int    `ini:"timeout" help:"msec capture timeout"`

	Capture bool   `ini:"-" arg:"-c,--capture" help:"run in capture mode"`
	Path    string `ini:"-" arg:"positional" help:"pcap file or eth device"`
}

type Pipelines struct {
	Capture int `ini:"capture"`
	Scan    int `ini:"scan"`
	Message int `ini:"message"`
}

var (
	// bind Makefile config
	etcPrefix = "/etc"

	config = Config{
		Device: "lo",
		// example:		Filter:     "host 127.0.0.1 and (tcp or udp port 5060)",
		Snapshot: 1600,
		Timeout:  500,
		Capture:  os.Geteuid() == 0 || os.Getpid() == 1 || os.Getppid() == 1,
	}

	pipelines = Pipelines{
		Capture: 32,
		Scan:    128,
	}

	packets  chan gopacket.Packet
	messages chan *SIPMessage
)

func (Config) Description() string {
	return "sipfind - find sip traffic"
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
	if config.Capture {
		if len(config.Path) > 0 {
			config.Device = config.Path
		}
	}

	messages = make(chan *SIPMessage, pipelines.Message)
	var wg sync.WaitGroup
	if config.Capture {
		timeout := time.Duration(config.Timeout) * time.Millisecond
		handle, err := pcap.OpenLive(config.Device, config.Snapshot, config.Promiscuous, timeout)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("searching capture from %s\n", config.Device)
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		wg.Add(3)
		packets = make(chan gopacket.Packet, pipelines.Capture)
		go Messages(&wg)
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
		handle, err := pcap.OpenOffline(config.Path)
		if err != nil {
			log.Fatal(err)
		}

		wg.Add(2)
		packets = make(chan gopacket.Packet, pipelines.Scan)
		go Messages(&wg)
		go Process(&wg)
		Scan(handle)
	}
	packets <- nil
	wg.Wait()
}
