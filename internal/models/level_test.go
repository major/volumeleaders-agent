package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTradeLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, level TradeLevel)
	}{
		{
			name: "full trade level with pointer fields",
			input: `{
				"Ticker": "NVDA",
				"Name": "NVIDIA Corporation",
				"Price": 875.50,
				"Dollars": 1500000.00,
				"Volume": 2000,
				"Trades": 5,
				"RelativeSize": 1.25,
				"CumulativeDistribution": 0.85,
				"TradeLevelRank": 1,
				"MinDate": "/Date(1745366400000)/",
				"MaxDate": "/Date(1745452800000)/",
				"Dates": "2025-04-23,2025-04-24"
			}`,
			check: func(t *testing.T, level TradeLevel) {
				if level.Ticker == nil || *level.Ticker != "NVDA" {
					t.Errorf("Ticker: expected NVDA, got %v", level.Ticker)
				}
				if level.Name == nil || *level.Name != "NVIDIA Corporation" {
					t.Errorf("Name: expected NVIDIA Corporation, got %v", level.Name)
				}
				if level.Price != 875.50 {
					t.Errorf("Price: expected 875.50, got %v", level.Price)
				}
				if !level.MinDate.Valid {
					t.Error("MinDate should be valid")
				}
				if !level.MaxDate.Valid {
					t.Error("MaxDate should be valid")
				}
			},
		},
		{
			name: "trade level with null pointer fields",
			input: `{
				"Ticker": null,
				"Name": null,
				"Price": 100.00,
				"Dollars": 500000.00,
				"Volume": 1000,
				"Trades": 3,
				"RelativeSize": 0.50,
				"CumulativeDistribution": 0.50,
				"TradeLevelRank": 5,
				"MinDate": "/Date(1745366400000)/",
				"MaxDate": "/Date(1745452800000)/",
				"Dates": "2025-04-23"
			}`,
			check: func(t *testing.T, level TradeLevel) {
				if level.Ticker != nil {
					t.Errorf("Ticker: expected nil, got %v", level.Ticker)
				}
				if level.Name != nil {
					t.Errorf("Name: expected nil, got %v", level.Name)
				}
			},
		},
		{
			name: "trade level with invalid dates",
			input: `{
				"Ticker": "AAPL",
				"Name": "Apple Inc.",
				"Price": 150.00,
				"Dollars": 750000.00,
				"Volume": 5000,
				"Trades": 10,
				"RelativeSize": 1.00,
				"CumulativeDistribution": 0.75,
				"TradeLevelRank": 2,
				"MinDate": null,
				"MaxDate": null,
				"Dates": ""
			}`,
			check: func(t *testing.T, level TradeLevel) {
				if level.MinDate.Valid {
					t.Error("MinDate should be invalid")
				}
				if level.MaxDate.Valid {
					t.Error("MaxDate should be invalid")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var level TradeLevel
			err := json.Unmarshal([]byte(tt.input), &level)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected unmarshal error: %v", err)
			}
			tt.check(t, level)
		})
	}
}

func TestTradeLevelRow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, row TradeLevelRow)
	}{
		{
			name: "compact trade level row",
			input: `{
				"Price": 875.50,
				"Dollars": 1500000.00,
				"Volume": 2000,
				"Trades": 5,
				"RelativeSize": 1.25,
				"CumulativeDistribution": 0.85,
				"TradeLevelRank": 1,
				"MinDate": "/Date(1745366400000)/",
				"MaxDate": "/Date(1745452800000)/"
			}`,
			check: func(t *testing.T, row TradeLevelRow) {
				if row.Price != 875.50 {
					t.Errorf("Price: expected 875.50, got %v", row.Price)
				}
				if row.Volume != 2000 {
					t.Errorf("Volume: expected 2000, got %v", row.Volume)
				}
				if !row.MinDate.Valid {
					t.Error("MinDate should be valid")
				}
				if !row.MaxDate.Valid {
					t.Error("MaxDate should be valid")
				}
			},
		},
		{
			name: "trade level row with invalid dates",
			input: `{
				"Price": 100.00,
				"Dollars": 500000.00,
				"Volume": 1000,
				"Trades": 3,
				"RelativeSize": 0.50,
				"CumulativeDistribution": 0.50,
				"TradeLevelRank": 5,
				"MinDate": null,
				"MaxDate": null
			}`,
			check: func(t *testing.T, row TradeLevelRow) {
				if row.MinDate.Valid {
					t.Error("MinDate should be invalid")
				}
				if row.MaxDate.Valid {
					t.Error("MaxDate should be invalid")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var row TradeLevelRow
			err := json.Unmarshal([]byte(tt.input), &row)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected unmarshal error: %v", err)
			}
			tt.check(t, row)
		})
	}
}

func TestTradeLevelTouch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, touch TradeLevelTouch)
	}{
		{
			name: "full trade level touch with pointer fields",
			input: `{
				"Ticker": "NVDA",
				"Sector": "Technology",
				"Industry": "Semiconductors",
				"Name": "NVIDIA Corporation",
				"Date": "/Date(1745366400000)/",
				"MinDate": "/Date(1745366400000)/",
				"MaxDate": "/Date(1745452800000)/",
				"FullDateTime": "2025-04-23 09:30:00",
				"FullTimeString24": "09:30:00",
				"Dates": "2025-04-23",
				"Price": 875.50,
				"Dollars": 1500000.00,
				"Volume": 2000,
				"Trades": 5,
				"CumulativeDistribution": 0.85,
				"TradeLevelRank": 1,
				"TotalRows": 100,
				"TradeLevelTouches": 10,
				"RelativeSize": 1.25
			}`,
			check: func(t *testing.T, touch TradeLevelTouch) {
				if touch.Ticker != "NVDA" {
					t.Errorf("Ticker: expected NVDA, got %v", touch.Ticker)
				}
				if touch.Sector == nil || *touch.Sector != "Technology" {
					t.Errorf("Sector: expected Technology, got %v", touch.Sector)
				}
				if touch.Industry == nil || *touch.Industry != "Semiconductors" {
					t.Errorf("Industry: expected Semiconductors, got %v", touch.Industry)
				}
				if touch.FullTimeString24 == nil || *touch.FullTimeString24 != "09:30:00" {
					t.Errorf("FullTimeString24: expected 09:30:00, got %v", touch.FullTimeString24)
				}
				if !touch.Date.Valid {
					t.Error("Date should be valid")
				}
			},
		},
		{
			name: "trade level touch with null pointer fields",
			input: `{
				"Ticker": "AAPL",
				"Sector": null,
				"Industry": null,
				"Name": "Apple Inc.",
				"Date": "/Date(1745366400000)/",
				"MinDate": "/Date(1745366400000)/",
				"MaxDate": "/Date(1745452800000)/",
				"FullDateTime": "2025-04-23 10:00:00",
				"FullTimeString24": null,
				"Dates": "2025-04-23",
				"Price": 150.00,
				"Dollars": 750000.00,
				"Volume": 5000,
				"Trades": 10,
				"CumulativeDistribution": 0.75,
				"TradeLevelRank": 2,
				"TotalRows": 50,
				"TradeLevelTouches": 5,
				"RelativeSize": 1.00
			}`,
			check: func(t *testing.T, touch TradeLevelTouch) {
				if touch.Sector != nil {
					t.Errorf("Sector: expected nil, got %v", touch.Sector)
				}
				if touch.Industry != nil {
					t.Errorf("Industry: expected nil, got %v", touch.Industry)
				}
				if touch.FullTimeString24 != nil {
					t.Errorf("FullTimeString24: expected nil, got %v", touch.FullTimeString24)
				}
			},
		},
		{
			name: "trade level touch with invalid dates",
			input: `{
				"Ticker": "MSFT",
				"Sector": null,
				"Industry": null,
				"Name": "Microsoft Corporation",
				"Date": null,
				"MinDate": null,
				"MaxDate": null,
				"FullDateTime": "2025-04-23 11:00:00",
				"FullTimeString24": null,
				"Dates": "",
				"Price": 300.00,
				"Dollars": 1000000.00,
				"Volume": 3000,
				"Trades": 8,
				"CumulativeDistribution": 0.80,
				"TradeLevelRank": 3,
				"TotalRows": 75,
				"TradeLevelTouches": 7,
				"RelativeSize": 1.10
			}`,
			check: func(t *testing.T, touch TradeLevelTouch) {
				if touch.Date.Valid {
					t.Error("Date should be invalid")
				}
				if touch.MinDate.Valid {
					t.Error("MinDate should be invalid")
				}
				if touch.MaxDate.Valid {
					t.Error("MaxDate should be invalid")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var touch TradeLevelTouch
			err := json.Unmarshal([]byte(tt.input), &touch)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected unmarshal error: %v", err)
			}
			tt.check(t, touch)
		})
	}
}

func TestNewTradeLevelRows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []TradeLevel
		expected int
	}{
		{
			name:     "empty slice",
			input:    []TradeLevel{},
			expected: 0,
		},
		{
			name: "single level",
			input: []TradeLevel{
				{
					Price:                  875.50,
					Dollars:                1500000.00,
					Volume:                 2000,
					Trades:                 5,
					RelativeSize:           1.25,
					CumulativeDistribution: 0.85,
					TradeLevelRank:         1,
					MinDate:                AspNetDate{Time: time.Date(2025, 4, 23, 0, 0, 0, 0, time.UTC), Valid: true},
					MaxDate:                AspNetDate{Time: time.Date(2025, 4, 24, 0, 0, 0, 0, time.UTC), Valid: true},
				},
			},
			expected: 1,
		},
		{
			name: "multiple levels",
			input: []TradeLevel{
				{
					Price:                  875.50,
					Dollars:                1500000.00,
					Volume:                 2000,
					Trades:                 5,
					RelativeSize:           1.25,
					CumulativeDistribution: 0.85,
					TradeLevelRank:         1,
					MinDate:                AspNetDate{Time: time.Date(2025, 4, 23, 0, 0, 0, 0, time.UTC), Valid: true},
					MaxDate:                AspNetDate{Time: time.Date(2025, 4, 24, 0, 0, 0, 0, time.UTC), Valid: true},
				},
				{
					Price:                  850.00,
					Dollars:                1200000.00,
					Volume:                 1500,
					Trades:                 4,
					RelativeSize:           1.00,
					CumulativeDistribution: 0.70,
					TradeLevelRank:         2,
					MinDate:                AspNetDate{Time: time.Date(2025, 4, 23, 0, 0, 0, 0, time.UTC), Valid: true},
					MaxDate:                AspNetDate{Time: time.Date(2025, 4, 24, 0, 0, 0, 0, time.UTC), Valid: true},
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rows := NewTradeLevelRows(tt.input)
			if len(rows) != tt.expected {
				t.Errorf("expected %d rows, got %d", tt.expected, len(rows))
			}
			for i, row := range rows {
				if row.Price != tt.input[i].Price {
					t.Errorf("row %d Price: expected %v, got %v", i, tt.input[i].Price, row.Price)
				}
				if row.Volume != tt.input[i].Volume {
					t.Errorf("row %d Volume: expected %v, got %v", i, tt.input[i].Volume, row.Volume)
				}
			}
		})
	}
}
