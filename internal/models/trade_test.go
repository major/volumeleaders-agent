package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAspNetDateUnmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantValid bool
		wantTime  time.Time
	}{
		{
			name:      "valid date",
			input:     `"/Date(1745366400000)/"`,
			wantValid: true,
			wantTime:  time.Date(2025, 4, 23, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "null value",
			input:     `null`,
			wantValid: false,
		},
		{
			name:      "empty string",
			input:     `""`,
			wantValid: false,
		},
		{
			name:      "datetime min value sentinel",
			input:     `"/Date(-62135596800000)/"`,
			wantValid: false,
		},
		{
			name:      "1900 sentinel",
			input:     `"/Date(-2208988800000)/"`,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var d AspNetDate
			if err := json.Unmarshal([]byte(tt.input), &d); err != nil {
				t.Fatalf("unexpected unmarshal error: %v", err)
			}
			if d.Valid != tt.wantValid {
				t.Errorf("Valid: expected %v, got %v", tt.wantValid, d.Valid)
			}
			if tt.wantValid && !d.Equal(tt.wantTime) {
				t.Errorf("Time: expected %v, got %v", tt.wantTime, d.Time)
			}
		})
	}
}

func TestAspNetDateMarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		date     AspNetDate
		expected string
	}{
		{
			name:     "invalid date marshals to null",
			date:     AspNetDate{},
			expected: "null",
		},
		{
			name:     "valid date marshals to RFC3339",
			date:     AspNetDate{Time: time.Date(2025, 4, 23, 0, 0, 0, 0, time.UTC), Valid: true},
			expected: `"2025-04-23T00:00:00Z"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.date)
			if err != nil {
				t.Fatalf("unexpected marshal error: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

func TestFlexBoolUnmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected FlexBool
	}{
		{"true", "true", true},
		{"false", "false", false},
		{"1", "1", true},
		{"0", "0", false},
		{"null", "null", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var b FlexBool
			if err := json.Unmarshal([]byte(tt.input), &b); err != nil {
				t.Fatalf("unexpected unmarshal error: %v", err)
			}
			if b != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, b)
			}
		})
	}
}

func TestFlexBoolMarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    FlexBool
		expected string
	}{
		{"true", true, "true"},
		{"false", false, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("unexpected marshal error: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

func TestAspNetDateUnmarshalInvalidFormat(t *testing.T) {
	t.Parallel()

	var d AspNetDate
	err := json.Unmarshal([]byte(`"not-a-date"`), &d)
	if err == nil {
		t.Fatal("expected error for invalid ASP.NET date format")
	}
}

func TestAspNetDateUnmarshalNonString(t *testing.T) {
	t.Parallel()

	var d AspNetDate
	err := json.Unmarshal([]byte(`12345`), &d)
	if err == nil {
		t.Fatal("expected error for non-string value")
	}
}

func TestFlexBoolUnmarshalError(t *testing.T) {
	t.Parallel()

	var b FlexBool
	err := json.Unmarshal([]byte(`"invalid"`), &b)
	if err == nil {
		t.Fatal("expected error for invalid FlexBool value")
	}
}
