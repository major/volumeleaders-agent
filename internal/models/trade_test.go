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
		wantErr   bool
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
		{
			name:    "invalid format",
			input:   `"not-a-date"`,
			wantErr: true,
		},
		{
			name:    "non-string value",
			input:   `12345`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var d AspNetDate
			err := json.Unmarshal([]byte(tt.input), &d)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
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
		wantErr  bool
		expected FlexBool
	}{
		{"true", "true", false, true},
		{"false", "false", false, false},
		{"1", "1", false, true},
		{"0", "0", false, false},
		{"null", "null", false, false},
		{"invalid string", `"invalid"`, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var b FlexBool
			err := json.Unmarshal([]byte(tt.input), &b)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
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


