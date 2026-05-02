package common

import "testing"

func TestParseSnapshotString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected map[string]float64
	}{
		{name: "empty string", input: "", expected: map[string]float64{}},
		{name: "single ticker", input: "AAPL:255.30", expected: map[string]float64{"AAPL": 255.30}},
		{name: "multiple tickers", input: "A:114.73;AA:70.96;AAPL:255.30", expected: map[string]float64{"A": 114.73, "AA": 70.96, "AAPL": 255.30}},
		{name: "trailing semicolon", input: "AAPL:255.30;MSFT:420.50;", expected: map[string]float64{"AAPL": 255.30, "MSFT": 420.50}},
		{name: "skips malformed entries", input: "AAPL:255.30;BADENTRY;MSFT:420.50", expected: map[string]float64{"AAPL": 255.30, "MSFT": 420.50}},
		{name: "skips non-numeric prices", input: "AAPL:255.30;BAD:notanumber;MSFT:420.50", expected: map[string]float64{"AAPL": 255.30, "MSFT": 420.50}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ParseSnapshotString(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d entries, got %d", len(tt.expected), len(result))
			}
			for ticker, expectedPrice := range tt.expected {
				got, ok := result[ticker]
				if !ok {
					t.Errorf("missing ticker %s", ticker)
					continue
				}
				if got != expectedPrice {
					t.Errorf("ticker %s: expected %f, got %f", ticker, expectedPrice, got)
				}
			}
		})
	}
}

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
