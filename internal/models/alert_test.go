package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAlertConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, ac *AlertConfig)
	}{
		{
			name: "full alert config with all fields",
			input: `{
				"AlertConfigKey": 123,
				"UserKey": 456,
				"Name": "My Alert",
				"Tickers": "AAPL,MSFT",
				"TradeRankLTE": 10,
				"TradeVCDGTE": 0.5,
				"TradeMultGTE": 2.0,
				"TradeVolumeGTE": 100000,
				"TradeDollarsGTE": 1000000.50,
				"TradeConditions": "F",
				"TradeClusterRankLTE": 5,
				"TradeClusterVCDGTE": 0.6,
				"TradeClusterMultGTE": 1.5,
				"TradeClusterVolumeGTE": 500000,
				"TradeClusterDollarsGTE": 5000000.75,
				"TotalRankLTE": 20,
				"TotalVolumeGTE": 1000000,
				"TotalDollarsGTE": 10000000.00,
				"AHRankLTE": 15,
				"AHVolumeGTE": 500000,
				"AHDollarsGTE": 5000000.00,
				"ClosingTradeRankLTE": 8,
				"ClosingTradeVCDGTE": 0.7,
				"ClosingTradeMultGTE": 1.8,
				"ClosingTradeVolumeGTE": 200000,
				"ClosingTradeDollarsGTE": 2000000.00,
				"ClosingTradeConditions": "C",
				"OffsettingPrint": true,
				"PhantomPrint": false,
				"Sweep": true,
				"DarkPool": false
			}`,
			wantErr: false,
			check: func(t *testing.T, ac *AlertConfig) {
				if ac.AlertConfigKey != 123 {
					t.Errorf("AlertConfigKey: expected 123, got %d", ac.AlertConfigKey)
				}
				if ac.UserKey != 456 {
					t.Errorf("UserKey: expected 456, got %d", ac.UserKey)
				}
				if ac.Name != "My Alert" {
					t.Errorf("Name: expected 'My Alert', got %q", ac.Name)
				}
				if ac.Tickers != "AAPL,MSFT" {
					t.Errorf("Tickers: expected 'AAPL,MSFT', got %q", ac.Tickers)
				}
				if ac.TradeRankLTE == nil || *ac.TradeRankLTE != 10 {
					t.Errorf("TradeRankLTE: expected 10, got %v", ac.TradeRankLTE)
				}
				if ac.TradeVCDGTE == nil || *ac.TradeVCDGTE != 0.5 {
					t.Errorf("TradeVCDGTE: expected 0.5, got %v", ac.TradeVCDGTE)
				}
				if ac.OffsettingPrint != true {
					t.Errorf("OffsettingPrint: expected true, got %v", ac.OffsettingPrint)
				}
				if ac.PhantomPrint != false {
					t.Errorf("PhantomPrint: expected false, got %v", ac.PhantomPrint)
				}
			},
		},
		{
			name: "alert config with null pointer fields",
			input: `{
				"AlertConfigKey": 1,
				"UserKey": 2,
				"Name": "Minimal Alert",
				"Tickers": "NVDA",
				"TradeRankLTE": null,
				"TradeVCDGTE": null,
				"TradeMultGTE": null,
				"TradeVolumeGTE": null,
				"TradeDollarsGTE": null,
				"TradeConditions": null,
				"TradeClusterRankLTE": null,
				"TradeClusterVCDGTE": null,
				"TradeClusterMultGTE": null,
				"TradeClusterVolumeGTE": null,
				"TradeClusterDollarsGTE": null,
				"TotalRankLTE": null,
				"TotalVolumeGTE": null,
				"TotalDollarsGTE": null,
				"AHRankLTE": null,
				"AHVolumeGTE": null,
				"AHDollarsGTE": null,
				"ClosingTradeRankLTE": null,
				"ClosingTradeVCDGTE": null,
				"ClosingTradeMultGTE": null,
				"ClosingTradeVolumeGTE": null,
				"ClosingTradeDollarsGTE": null,
				"ClosingTradeConditions": null,
				"OffsettingPrint": false,
				"PhantomPrint": false,
				"Sweep": false,
				"DarkPool": false
			}`,
			wantErr: false,
			check: func(t *testing.T, ac *AlertConfig) {
				if ac.AlertConfigKey != 1 {
					t.Errorf("AlertConfigKey: expected 1, got %d", ac.AlertConfigKey)
				}
				if ac.TradeRankLTE != nil {
					t.Errorf("TradeRankLTE: expected nil, got %v", ac.TradeRankLTE)
				}
				if ac.TradeVCDGTE != nil {
					t.Errorf("TradeVCDGTE: expected nil, got %v", ac.TradeVCDGTE)
				}
				if ac.TradeConditions != nil {
					t.Errorf("TradeConditions: expected nil, got %v", ac.TradeConditions)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var ac AlertConfig
			err := json.Unmarshal([]byte(tt.input), &ac)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected unmarshal error: %v", err)
			}
			tt.check(t, &ac)
		})
	}
}

func TestTradeAlert(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, ta *TradeAlert)
	}{
		{
			name: "trade alert with AspNetDate and FlexBool fields",
			input: `{
				"Date": "/Date(1745366400000)/",
				"StartDate": "/Date(1745280000000)/",
				"EndDate": "/Date(1745452800000)/",
				"FullTimeString24": "14:30:00",
				"DateKey": 20250423,
				"SecurityKey": 789,
				"TimeKey": 143000,
				"TradeID": 999,
				"SequenceNumber": 1,
				"UserKey": 456,
				"UserKeys": "456,789",
				"Sent": true,
				"Email": "user@example.com",
				"Emails": "user@example.com,other@example.com",
				"Ticker": "AAPL",
				"Sector": "Technology",
				"Industry": "Consumer Electronics",
				"Name": "Apple Inc.",
				"AlertType": "TradeRank",
				"Price": 175.50,
				"TradeRank": 5,
				"VolumeCumulativeDistribution": 0.75,
				"DollarsMultiplier": 2.5,
				"Volume": 500000,
				"Dollars": 87750000.00,
				"LastComparibleTradeDateKey": 20250422,
				"LastComparibleTradeDate": "/Date(1745280000000)/",
				"OffsettingTradeDate": "/Date(1745366400000)/",
				"PhantomPrintFulfillmentDate": "/Date(1745452800000)/",
				"FullDateTime": "2025-04-23T14:30:00Z",
				"IPODate": "/Date(1104537600000)/",
				"RSIHour": 65.5,
				"RSIDay": 58.3,
				"InProcess": false,
				"Complete": true,
				"Sweep": 1,
				"DarkPool": 0,
				"LatePrint": 1,
				"ClosingTrade": 0,
				"SignaturePrint": 1,
				"PhantomPrint": 0
			}`,
			wantErr: false,
			check: func(t *testing.T, ta *TradeAlert) {
				if ta.DateKey != 20250423 {
					t.Errorf("DateKey: expected 20250423, got %d", ta.DateKey)
				}
				if ta.Ticker != "AAPL" {
					t.Errorf("Ticker: expected 'AAPL', got %q", ta.Ticker)
				}
				if ta.Price != 175.50 {
					t.Errorf("Price: expected 175.50, got %f", ta.Price)
				}
				if !ta.Date.Valid {
					t.Errorf("Date.Valid: expected true, got false")
				}
				if !ta.Date.Equal(time.Date(2025, 4, 23, 0, 0, 0, 0, time.UTC)) {
					t.Errorf("Date.Time: expected 2025-04-23, got %v", ta.Date.Time)
				}
				if ta.Sweep != true {
					t.Errorf("Sweep: expected true (1), got %v", ta.Sweep)
				}
				if ta.DarkPool != false {
					t.Errorf("DarkPool: expected false (0), got %v", ta.DarkPool)
				}
				if ta.LatePrint != true {
					t.Errorf("LatePrint: expected true (1), got %v", ta.LatePrint)
				}
				if ta.ClosingTrade != false {
					t.Errorf("ClosingTrade: expected false (0), got %v", ta.ClosingTrade)
				}
			},
		},
		{
			name: "trade alert with null AspNetDate fields",
			input: `{
				"Date": null,
				"StartDate": null,
				"EndDate": null,
				"FullTimeString24": "14:30:00",
				"DateKey": 20250423,
				"SecurityKey": 789,
				"TimeKey": 143000,
				"TradeID": 999,
				"SequenceNumber": 1,
				"UserKey": 456,
				"UserKeys": null,
				"Sent": false,
				"Email": null,
				"Emails": null,
				"Ticker": "MSFT",
				"Sector": null,
				"Industry": null,
				"Name": "Microsoft Corp.",
				"AlertType": "TradeClusterRank",
				"Price": 420.00,
				"TradeRank": 1,
				"VolumeCumulativeDistribution": 0.95,
				"DollarsMultiplier": 3.0,
				"Volume": 1000000,
				"Dollars": 420000000.00,
				"LastComparibleTradeDateKey": 20250422,
				"LastComparibleTradeDate": null,
				"OffsettingTradeDate": null,
				"PhantomPrintFulfillmentDate": null,
				"FullDateTime": "2025-04-23T14:30:00Z",
				"IPODate": null,
				"RSIHour": 70.0,
				"RSIDay": 65.0,
				"InProcess": false,
				"Complete": true,
				"Sweep": 0,
				"DarkPool": 0,
				"LatePrint": 0,
				"ClosingTrade": 0,
				"SignaturePrint": 0,
				"PhantomPrint": 0
			}`,
			wantErr: false,
			check: func(t *testing.T, ta *TradeAlert) {
				if ta.Date.Valid {
					t.Errorf("Date.Valid: expected false, got true")
				}
				if ta.LastComparibleTradeDate.Valid {
					t.Errorf("LastComparibleTradeDate.Valid: expected false, got true")
				}
				if ta.Sector != nil {
					t.Errorf("Sector: expected nil, got %v", ta.Sector)
				}
				if ta.Email != nil {
					t.Errorf("Email: expected nil, got %v", ta.Email)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var ta TradeAlert
			err := json.Unmarshal([]byte(tt.input), &ta)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected unmarshal error: %v", err)
			}
			tt.check(t, &ta)
		})
	}
}
