// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 David Sugar <tychosoft@gmail.com>
//go:build debug

package service

import (
	"os"
)

func Logger(level int, path string) {
	os.Remove(path)
	openLogger(level, path)
}

// Debug output
func Debug(level int, args ...interface{}) {
	Output(level, args...)
}

func Debugf(level int, format string, args ...interface{}) {
	Outputf(level, format, args...)
}

func IsDebug() bool {
	return true
}
