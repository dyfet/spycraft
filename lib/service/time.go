// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 David Sugar <tychosoft@gmail.com>

package service

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

// service.Time reports utc and truncates to nearest second
type Time time.Time

// Get current time
func Now() Time {
	return Time(time.Now().UTC().Truncate(time.Second))
}

// Marshall as time string (zulu time) nearest second
func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.Quote(time.Time(t).UTC().Truncate(time.Second).String()))
}

// Time as zulu time string
func (t Time) String() string {
	return time.Time(t).UTC().Truncate(time.Second).String()
}

// Local time format such as for logs
func (t Time) Local() string {
	return time.Time(t).Local().Truncate(time.Second).Format("2001-03-28 15:03:04")
}

// Return timestamp as seconds
func (t Time) Seconds() int64 {
	return time.Time(t).Unix()
}

// Parse Time
func (t *Time) Parse(text string) (Time, error) {
	tmp, err := time.Parse(time.RFC3339, text)
	return Time(tmp.Truncate(time.Second)), err
}

// Unmarshal from float or string form
func (t *Time) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		tmp, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return err
		}
		*t = Time(tmp.Truncate(time.Second))
		return nil
	default:
		return errors.New("invalid duration")
	}
}
