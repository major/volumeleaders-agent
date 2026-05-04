package common

import "testing"

func TestIntStr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{42, "42"},
		{-1, "-1"},
		{2000000000, "2000000000"},
	}
	for _, tt := range tests {
		if got := IntStr(tt.input); got != tt.expected {
			t.Errorf("IntStr(%d) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFormatFloat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0"},
		{1.5, "1.5"},
		{100000, "100000"},
		{99.25, "99.25"},
		{30000000000, "30000000000"},
		{0.001, "0.001"},
	}
	for _, tt := range tests {
		if got := FormatFloat(tt.input); got != tt.expected {
			t.Errorf("FormatFloat(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestBoolString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    bool
		expected string
	}{
		{true, "true"},
		{false, "false"},
	}
	for _, tt := range tests {
		if got := BoolString(tt.input); got != tt.expected {
			t.Errorf("BoolString(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestToDateKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input, expected string
	}{
		{"2025-01-15", "20250115"},
		{"2026-04-24", "20260424"},
		{"20250115", "20250115"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := ToDateKey(tt.input); got != tt.expected {
			t.Errorf("ToDateKey(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
