// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 David Sugar <tychosoft@gmail.com>

package service

import (
	"fmt"
	"io"
	"log"
	"log/syslog"
	"os"
	"path"
)

var (
	logger  *syslog.Writer = nil
	logfile *os.File       = nil
	logpath                = ""
	console                = log.New(io.Discard, "", log.LstdFlags)
	verbose                = 0
	argv0                  = path.Base(os.Args[0])
)

// internal specify logging level and path
func openLogger(level int, path string) {
	var err error
	verbose = level
	logpath = path
	LoggerRestart()
	logger, err = syslog.New(syslog.LOG_SYSLOG, argv0)
	if err != nil {
		log.Println(err)
		logger = nil
	}
}

// Reset Logger such as from sighup
func LoggerRestart() {
	var err error
	if logfile != nil {
		logfile.Close()
		logfile = nil
	}
	if len(logpath) > 0 && logpath != "none" && logpath != "no" && logpath != "/dev/nul" {
		logfile, err = os.OpenFile(logpath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0660)
		if err != nil {
			Error(err)
			return
		}
		log.SetOutput(logfile)
		console.SetOutput(os.Stderr)
		console.SetFlags(0) // log.Ltime?
		Notice("logger restart")
	}
}

// Log errors
func Error(args ...interface{}) {
	msg := fmt.Sprint(args...)
	if logger != nil {
		logger.Err(msg)
	}
	if verbose > 0 {
		console.Println("error:", msg)
	}
	log.Println(msg)
}

func Errorf(format string, args ...interface{}) {
	Error(fmt.Sprintf(format, args...))
}

// Log failure and exit
func Fail(code int, args ...interface{}) {
	msg := fmt.Sprint(args...)
	if logger != nil {
		logger.Crit(msg)
	}
	if verbose > 0 {
		console.Println("fail:", msg)
	}
	log.Println(msg)
	os.Exit(code)
}

func Failf(code int, format string, args ...interface{}) {
	Fail(code, fmt.Sprintf(format, args...))
}

// Log warnings
func Warn(args ...interface{}) {
	msg := fmt.Sprint(args...)
	if logger != nil {
		logger.Warning(msg)
	}
	if verbose > 0 {
		console.Println("warn:", msg)
	}
	log.Println(msg)
}

func Warnf(format string, args ...interface{}) {
	Warn(fmt.Sprintf(format, args...))
}

// Log notices
func Notice(args ...interface{}) {
	msg := fmt.Sprint(args...)
	if logger != nil {
		logger.Notice(msg)
	}
	if verbose > 1 {
		console.Println("notice:", msg)
	}
	log.Println(msg)
}

func Noticef(format string, args ...interface{}) {
	Notice(fmt.Sprintf(format, args...))
}

// Log info
func Info(args ...interface{}) {
	msg := fmt.Sprint(args...)
	if logger != nil {
		logger.Info(msg)
	}
	if verbose > 1 {
		console.Println("info:", msg)
	}
	log.Println(msg)
}

func Infof(format string, args ...interface{}) {
	Info(fmt.Sprintf(format, args...))
}

// Verbose output
func Output(level int, args ...interface{}) {
	if level > verbose {
		return
	}

	msg := fmt.Sprint(args...)
	console.Println("debug:", msg)
	log.Println(msg)
}

func Outputf(level int, format string, args ...interface{}) {
	Output(level, fmt.Sprintf(format, args...))
}
