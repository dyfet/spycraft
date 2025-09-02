// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 David Sugar <tychosoft@gmail.com>
//go:build !debug

package service

func Logger(level int, path string) {
	openLogger(level, path)
}

func Debug(level int, args ...interface{}) {
}

func Debugf(level int, format string, args ...interface{}) {
}

func IsDebug() bool {
	return false
}
