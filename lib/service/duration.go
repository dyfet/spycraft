// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 David Sugar <tychosoft@gmail.com>

package service

import (
	"encoding/json"
	"errors"
	"time"
)

// service.Duration allows json marshalling and prefers second intervals
type Duration time.Duration

// We convert from int to duration by seconds...
func NewDuration(seconds int) Duration {
	return Duration(time.Second * time.Duration(seconds))
}

// Marshall as duration string
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).Truncate(time.Second).String())
}

// Parse duration
func (d Duration) Parse(text string) (Duration, error) {
	tmp, err := time.ParseDuration(text)
	return Duration(tmp.Truncate(time.Second)), err
}

// Duration as string
func (d Duration) String() string {
	return time.Duration(d).Truncate(time.Second).String()
}

// Unmarshal from float or string form
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value).Truncate(time.Second))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp.Truncate(time.Second))
		return nil
	default:
		return errors.New("invalid duration")
	}
}
