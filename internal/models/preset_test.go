package models

import (
	"encoding/json"
	"testing"
)

func TestPresetInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, preset PresetInfo)
	}{
		{
			name: "preset info with filters",
			input: `{
				"Name": "Large Blocks",
				"Group": "Trade Filters",
				"Filters": {
					"MinDollars": "1000000",
					"MinVolume": "10000",
					"MaxPrice": "500"
				}
			}`,
			check: func(t *testing.T, preset PresetInfo) {
				if preset.Name != "Large Blocks" {
					t.Errorf("Name: expected Large Blocks, got %v", preset.Name)
				}
				if preset.Group != "Trade Filters" {
					t.Errorf("Group: expected Trade Filters, got %v", preset.Group)
				}
				if len(preset.Filters) != 3 {
					t.Errorf("Filters: expected 3 entries, got %d", len(preset.Filters))
				}
				if preset.Filters["MinDollars"] != "1000000" {
					t.Errorf("MinDollars filter: expected 1000000, got %v", preset.Filters["MinDollars"])
				}
			},
		},
		{
			name: "preset info with empty filters",
			input: `{
				"Name": "All Trades",
				"Group": "Default",
				"Filters": {}
			}`,
			check: func(t *testing.T, preset PresetInfo) {
				if preset.Name != "All Trades" {
					t.Errorf("Name: expected All Trades, got %v", preset.Name)
				}
				if len(preset.Filters) != 0 {
					t.Errorf("Filters: expected 0 entries, got %d", len(preset.Filters))
				}
			},
		},
		{
			name: "preset info with null filters",
			input: `{
				"Name": "Test Preset",
				"Group": "Test",
				"Filters": null
			}`,
			check: func(t *testing.T, preset PresetInfo) {
				if preset.Name != "Test Preset" {
					t.Errorf("Name: expected Test Preset, got %v", preset.Name)
				}
				if preset.Filters != nil {
					t.Errorf("Filters: expected nil, got %v", preset.Filters)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var preset PresetInfo
			err := json.Unmarshal([]byte(tt.input), &preset)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected unmarshal error: %v", err)
			}
			tt.check(t, preset)
		})
	}
}

func TestPresetTickersInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, preset PresetTickersInfo)
	}{
		{
			name: "preset with explicit tickers",
			input: `{
				"Preset": "Tech Stocks",
				"Group": "Sector",
				"Type": "tickers",
				"Tickers": ["NVDA", "AAPL", "MSFT"],
				"SectorIndustry": ""
			}`,
			check: func(t *testing.T, preset PresetTickersInfo) {
				if preset.Preset != "Tech Stocks" {
					t.Errorf("Preset: expected Tech Stocks, got %v", preset.Preset)
				}
				if preset.Type != "tickers" {
					t.Errorf("Type: expected tickers, got %v", preset.Type)
				}
				if len(preset.Tickers) != 3 {
					t.Errorf("Tickers: expected 3 items, got %d", len(preset.Tickers))
				}
				if preset.Tickers[0] != "NVDA" {
					t.Errorf("First ticker: expected NVDA, got %v", preset.Tickers[0])
				}
			},
		},
		{
			name: "preset with sector filter",
			input: `{
				"Preset": "Healthcare Sector",
				"Group": "Sector",
				"Type": "sector-filter",
				"SectorIndustry": "Healthcare"
			}`,
			check: func(t *testing.T, preset PresetTickersInfo) {
				if preset.Preset != "Healthcare Sector" {
					t.Errorf("Preset: expected Healthcare Sector, got %v", preset.Preset)
				}
				if preset.Type != "sector-filter" {
					t.Errorf("Type: expected sector-filter, got %v", preset.Type)
				}
				if preset.SectorIndustry != "Healthcare" {
					t.Errorf("SectorIndustry: expected Healthcare, got %v", preset.SectorIndustry)
				}
				if len(preset.Tickers) != 0 {
					t.Errorf("Tickers: expected empty, got %d items", len(preset.Tickers))
				}
			},
		},
		{
			name: "preset with omitempty fields missing",
			input: `{
				"Preset": "Unfiltered",
				"Group": "Default",
				"Type": "unfiltered"
			}`,
			check: func(t *testing.T, preset PresetTickersInfo) {
				if preset.Preset != "Unfiltered" {
					t.Errorf("Preset: expected Unfiltered, got %v", preset.Preset)
				}
				if preset.Type != "unfiltered" {
					t.Errorf("Type: expected unfiltered, got %v", preset.Type)
				}
				if len(preset.Tickers) != 0 {
					t.Errorf("Tickers: expected empty slice, got %d items", len(preset.Tickers))
				}
				if preset.SectorIndustry != "" {
					t.Errorf("SectorIndustry: expected empty string, got %v", preset.SectorIndustry)
				}
			},
		},
		{
			name: "preset with empty tickers array",
			input: `{
				"Preset": "Empty Tickers",
				"Group": "Test",
				"Type": "tickers",
				"Tickers": [],
				"SectorIndustry": ""
			}`,
			check: func(t *testing.T, preset PresetTickersInfo) {
				if len(preset.Tickers) != 0 {
					t.Errorf("Tickers: expected empty, got %d items", len(preset.Tickers))
				}
			},
		},
		{
			name: "preset with null tickers",
			input: `{
				"Preset": "Null Tickers",
				"Group": "Test",
				"Type": "tickers",
				"Tickers": null,
				"SectorIndustry": ""
			}`,
			check: func(t *testing.T, preset PresetTickersInfo) {
				if preset.Tickers != nil {
					t.Errorf("Tickers: expected nil, got %v", preset.Tickers)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var preset PresetTickersInfo
			err := json.Unmarshal([]byte(tt.input), &preset)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected unmarshal error: %v", err)
			}
			tt.check(t, preset)
		})
	}
}

func TestPresetTickersInfoMarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		preset   PresetTickersInfo
		checkOut func(t *testing.T, data []byte)
	}{
		{
			name: "marshal with tickers omitted when empty",
			preset: PresetTickersInfo{
				Preset:         "Test",
				Group:          "Group",
				Type:           "unfiltered",
				Tickers:        []string{},
				SectorIndustry: "",
			},
			checkOut: func(t *testing.T, data []byte) {
				var m map[string]interface{}
				if err := json.Unmarshal(data, &m); err != nil {
					t.Fatalf("unmarshal output: %v", err)
				}
				if _, hasTickers := m["Tickers"]; hasTickers {
					t.Error("Tickers should be omitted when empty")
				}
				if _, hasSector := m["SectorIndustry"]; hasSector {
					t.Error("SectorIndustry should be omitted when empty")
				}
			},
		},
		{
			name: "marshal with tickers included when non-empty",
			preset: PresetTickersInfo{
				Preset:         "Tech",
				Group:          "Sector",
				Type:           "tickers",
				Tickers:        []string{"NVDA", "AAPL"},
				SectorIndustry: "",
			},
			checkOut: func(t *testing.T, data []byte) {
				var m map[string]interface{}
				if err := json.Unmarshal(data, &m); err != nil {
					t.Fatalf("unmarshal output: %v", err)
				}
				if _, hasTickers := m["Tickers"]; !hasTickers {
					t.Error("Tickers should be included when non-empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.preset)
			if err != nil {
				t.Fatalf("unexpected marshal error: %v", err)
			}
			tt.checkOut(t, data)
		})
	}
}
