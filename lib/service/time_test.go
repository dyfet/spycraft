// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 David Sugar <tychosoft@gmail.com>

package service

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimeMarshalJSON(t *testing.T) {
	testTime := Time(time.Date(2001, time.March, 5, 12, 30, 45, 123456789, time.UTC))

	expectedJSON := `"2001-03-05 12:30:45 +0000 UTC"`
	jsonData, err := testTime.MarshalJSON()
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	var result string
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Expected valid JSON, but got error %v", err)
	}

	if result != expectedJSON {
		t.Errorf("Expected %v, but got %v", expectedJSON, result)
	}
}

func TestTimeUnmarshalJSON_Valid(t *testing.T) {
	jsonStr := `"2001-03-05T12:30:45Z"`
	var tm Time

	err := json.Unmarshal([]byte(jsonStr), &tm)
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	expected := Time(time.Date(2001, time.March, 5, 12, 30, 45, 0, time.UTC))
	if tm != expected {
		t.Errorf("Expected %v, but got %v", expected, tm)
	}
}

func TestTimeUnmarshalJSON_InvalidFormat(t *testing.T) {
	jsonStr := `"invalid time format"`
	var tm Time

	err := json.Unmarshal([]byte(jsonStr), &tm)
	if err == nil {
		t.Fatalf("Expected an error, but got none")
	}
}

func TestTimeUnmarshalJSON_InvalidType(t *testing.T) {
	jsonStr := `12345`
	var tm Time

	err := json.Unmarshal([]byte(jsonStr), &tm)
	if err == nil {
		t.Fatalf("Expected an error, but got none")
	}

	if err.Error() != "invalid duration" {
		t.Errorf("Expected 'invalid duration', but got %v", err)
	}
}
