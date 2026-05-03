package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEarnings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, e *Earnings)
	}{
		{
			name: "earnings with all fields populated",
			input: `{
				"Ticker": "AAPL",
				"Name": "Apple Inc.",
				"Sector": "Technology",
				"Industry": "Consumer Electronics",
				"EarningsDate": "/Date(1745366400000)/",
				"AfterMarketClose": true,
				"TradeCount": 150,
				"TradeClusterCount": 25,
				"TradeClusterBombCount": 3
			}`,
			wantErr: false,
			check: func(t *testing.T, e *Earnings) {
				if e.Ticker != "AAPL" {
					t.Errorf("Ticker: expected 'AAPL', got %q", e.Ticker)
				}
				if e.Name != "Apple Inc." {
					t.Errorf("Name: expected 'Apple Inc.', got %q", e.Name)
				}
				if e.Sector == nil || *e.Sector != "Technology" {
					t.Errorf("Sector: expected 'Technology', got %v", e.Sector)
				}
				if e.Industry == nil || *e.Industry != "Consumer Electronics" {
					t.Errorf("Industry: expected 'Consumer Electronics', got %v", e.Industry)
				}
				if !e.EarningsDate.Valid {
					t.Errorf("EarningsDate.Valid: expected true, got false")
				}
				if !e.EarningsDate.Equal(time.Date(2025, 4, 23, 0, 0, 0, 0, time.UTC)) {
					t.Errorf("EarningsDate.Time: expected 2025-04-23, got %v", e.EarningsDate.Time)
				}
				if e.AfterMarketClose != true {
					t.Errorf("AfterMarketClose: expected true, got %v", e.AfterMarketClose)
				}
				if e.TradeCount != 150 {
					t.Errorf("TradeCount: expected 150, got %d", e.TradeCount)
				}
				if e.TradeClusterCount != 25 {
					t.Errorf("TradeClusterCount: expected 25, got %d", e.TradeClusterCount)
				}
				if e.TradeClusterBombCount != 3 {
					t.Errorf("TradeClusterBombCount: expected 3, got %d", e.TradeClusterBombCount)
				}
			},
		},
		{
			name: "earnings with null sector and industry",
			input: `{
				"Ticker": "MSFT",
				"Name": "Microsoft Corporation",
				"Sector": null,
				"Industry": null,
				"EarningsDate": "/Date(1745452800000)/",
				"AfterMarketClose": false,
				"TradeCount": 200,
				"TradeClusterCount": 30,
				"TradeClusterBombCount": 5
			}`,
			wantErr: false,
			check: func(t *testing.T, e *Earnings) {
				if e.Ticker != "MSFT" {
					t.Errorf("Ticker: expected 'MSFT', got %q", e.Ticker)
				}
				if e.Sector != nil {
					t.Errorf("Sector: expected nil, got %v", e.Sector)
				}
				if e.Industry != nil {
					t.Errorf("Industry: expected nil, got %v", e.Industry)
				}
				if e.AfterMarketClose != false {
					t.Errorf("AfterMarketClose: expected false, got %v", e.AfterMarketClose)
				}
			},
		},
		{
			name: "earnings with null earnings date",
			input: `{
				"Ticker": "GOOGL",
				"Name": "Alphabet Inc.",
				"Sector": "Technology",
				"Industry": "Internet Services",
				"EarningsDate": null,
				"AfterMarketClose": true,
				"TradeCount": 100,
				"TradeClusterCount": 15,
				"TradeClusterBombCount": 1
			}`,
			wantErr: false,
			check: func(t *testing.T, e *Earnings) {
				if e.Ticker != "GOOGL" {
					t.Errorf("Ticker: expected 'GOOGL', got %q", e.Ticker)
				}
				if e.EarningsDate.Valid {
					t.Errorf("EarningsDate.Valid: expected false, got true")
				}
				if e.TradeCount != 100 {
					t.Errorf("TradeCount: expected 100, got %d", e.TradeCount)
				}
			},
		},
		{
			name: "earnings with zero trade counts",
			input: `{
				"Ticker": "TSLA",
				"Name": "Tesla Inc.",
				"Sector": "Automotive",
				"Industry": "Auto Manufacturers",
				"EarningsDate": "/Date(1745280000000)/",
				"AfterMarketClose": false,
				"TradeCount": 0,
				"TradeClusterCount": 0,
				"TradeClusterBombCount": 0
			}`,
			wantErr: false,
			check: func(t *testing.T, e *Earnings) {
				if e.TradeCount != 0 {
					t.Errorf("TradeCount: expected 0, got %d", e.TradeCount)
				}
				if e.TradeClusterCount != 0 {
					t.Errorf("TradeClusterCount: expected 0, got %d", e.TradeClusterCount)
				}
				if e.TradeClusterBombCount != 0 {
					t.Errorf("TradeClusterBombCount: expected 0, got %d", e.TradeClusterBombCount)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var e Earnings
			err := json.Unmarshal([]byte(tt.input), &e)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected unmarshal error: %v", err)
			}
			tt.check(t, &e)
		})
	}
}
