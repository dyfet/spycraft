// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 David Sugar <tychosoft@gmail.com>
//go:build !systemd

package service

import (
	"fmt"
	"os"
)

var stopping = false

func Reload(args ...interface{}) error {
	msg := fmt.Sprint(args...)
	if len(msg) > 0 {
		Info(msg)
	}
	return nil
}

func Live(args ...interface{}) error {
	if stopping {
		return nil
	}
	msg := fmt.Sprint(args...)
	if len(msg) > 0 {
		Info(msg)
	}
	return nil
}

func Status(string) error {
	if stopping {
		return fmt.Errorf("already exiting")
	}
	return nil
}

func Stop(args ...interface{}) error {
	if stopping {
		return nil
	}
	stopping = true
	msg := fmt.Sprint(args...)
	if len(msg) > 0 {
		Info(msg)
	}
	return nil
}

func Watchdog() error {
	return nil
}

func IsService() bool {
	return os.Geteuid() == 0 || os.Getpid() == 1 || os.Getppid() == 1
}
