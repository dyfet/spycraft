// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 David Sugar <tychosoft@gmail.com>

package byteshark

import (
	"bytes"
)

func MatchKeyword(line, text []byte) bool {
	return len(line) >= len(text) && bytes.EqualFold(line[:len(text)], text)
}

func SplitKeypair(line []byte, sep byte) (key, value []byte) {
	for i, b := range line {
		if b == sep {
			key = bytes.TrimSpace(line[:i])
			if i+1 < len(line) {
				value = bytes.TrimSpace(line[i+1:])
			}
			return key, value
		}
	}
	return nil, nil
}

func SplitSections(input, delim []byte, out *[]([]byte)) int {
	dlen := len(delim)
	if dlen == 0 {
		return 0
	}

	start := 0
	count := 0
	for i := 0; i <= len(input)-dlen; {
		if bytes.Equal(input[i:i+dlen], delim) {
			if i == start {
				break // skip empty section
			}
			if len(*out) < cap(*out) {
				*out = append(*out, input[start:i])
				count++
			} else {
				break // capacity reached
			}
			start = i + dlen
			i = start
		} else {
			i++
		}
	}

	// Handle final section
	if start < len(input) && len(*out) < cap(*out) {
		*out = append(*out, input[start:])
		count++
	}
	return count
}

func ParseContentLength(headers []byte) int {
	lower := []byte("content-length:")
	start := 0
	for {
		end := bytes.Index(headers[start:], []byte("\r\n"))
		if end == -1 {
			break
		}

		line := headers[start : start+end]
		start += end + 2 // advance to next line
		if len(line) >= len(lower) && bytes.EqualFold(line[:len(lower)], lower) {
			val := line[len(lower):]
			for len(val) > 0 && (val[0] == ' ' || val[0] == '\t') {
				val = val[1:]
			}

			n := 0
			for _, b := range val {
				if b < '0' || b > '9' {
					break
				}
				n = n*10 + int(b-'0')
			}
			return n
		}
	}
	return 0 // empty if not found or malformed
}
