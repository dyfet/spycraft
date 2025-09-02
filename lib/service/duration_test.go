// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 David Sugar <tychosoft@gmail.com>

package service

import (
	"testing"
	"time"
)

func TestNewDuration(t *testing.T) {
	duration := NewDuration(30)
	expected := "30s"
	result := duration.String()
	if result != expected {
		t.Fatalf("Expected %v, but got %v", expected, result)
	}
}

func TestDurationMarshalJSON(t *testing.T) {
	duration := Duration(time.Duration(311 * time.Second))
	expectedJSON := `"5m11s"`

	jsonData, err := duration.MarshalJSON()
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	if string(jsonData) != expectedJSON {
		t.Errorf("Expected %v, but got %v", expectedJSON, string(jsonData))
	}
}
