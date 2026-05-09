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

// TestNewTradeListRowsProjectsCompactTradeFields verifies default trade output keeps analysis fields only.
func TestNewTradeListRowsProjectsCompactTradeFields(t *testing.T) {
	t.Parallel()

	industry := "Semiconductors"
	fullDateTime := "2026-05-08T14:30:00Z"
	trades := []Trade{
		{
			Date:                   AspNetDate{Time: time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC), Valid: true},
			FullDateTime:           &fullDateTime,
			Ticker:                 "NVDA",
			Name:                   "NVIDIA Corporation",
			Sector:                 "Technology",
			Industry:               &industry,
			Price:                  900.50,
			Volume:                 125000,
			Dollars:                112562500,
			DollarsMultiplier:      4.56789,
			PercentDailyVolume:     12.5,
			RelativeSize:           6.25,
			CumulativeDistribution: 99.1,
			TradeRank:              3,
			TradeRankSnapshot:      10,
			DarkPool:               true,
			Sweep:                  true,
			LatePrint:              false,
			SignaturePrint:         true,
			OpeningTrade:           false,
			ClosingTrade:           true,
			PhantomPrint:           false,
			TradeID:                123456,
			SecurityKey:            789,
		},
		{Ticker: "AAPL", Name: "Apple Inc."},
	}

	got := NewTradeListRows(trades)
	if len(got) != len(trades) {
		t.Fatalf("NewTradeListRows(%d trades) length = %d, want %d", len(trades), len(got), len(trades))
	}
	first := got[0]
	if first.Ticker != "NVDA" || first.Name != "NVIDIA Corporation" || first.Sector != "Technology" {
		t.Fatalf("NewTradeListRows(first identity) = %#v, want NVDA technology row", first)
	}
	if first.Industry == nil || *first.Industry != industry {
		t.Fatalf("NewTradeListRows(first).Industry = %v, want %q", first.Industry, industry)
	}
	if first.FullDateTime == nil || *first.FullDateTime != fullDateTime {
		t.Fatalf("NewTradeListRows(first).FullDateTime = %#v, want %q", first.FullDateTime, fullDateTime)
	}
	if first.Price != 900.50 || first.Volume != 125000 || first.Dollars != 112562500 {
		t.Fatalf("NewTradeListRows(first) price/size = %#v, want source price and volume fields", first)
	}
	if first.DollarsMultiplier != 4.57 {
		t.Fatalf("NewTradeListRows(first).DollarsMultiplier = %.5f, want 4.57", first.DollarsMultiplier)
	}
	if first.PercentDailyVolume != 12.5 || first.RelativeSize != 6.25 || first.CumulativeDistribution != 99.1 {
		t.Fatalf("NewTradeListRows(first) relative metrics = %#v, want source metrics", first)
	}
	if first.TradeRank != 3 {
		t.Fatalf("NewTradeListRows(first).TradeRank = %d, want 3", first.TradeRank)
	}
	if !bool(first.DarkPool) || !bool(first.Sweep) || bool(first.LatePrint) || !bool(first.SignaturePrint) || bool(first.OpeningTrade) || !bool(first.ClosingTrade) || bool(first.PhantomPrint) {
		t.Fatalf("NewTradeListRows(first) flags = %#v, want source boolean flags", first)
	}
	if got[1].Ticker != "AAPL" || got[1].Industry != nil {
		t.Fatalf("NewTradeListRows(second) = %#v, want sparse AAPL row with nil optional fields", got[1])
	}
}
