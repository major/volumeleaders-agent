package models

import (
	"encoding/json"
	"testing"
)

func TestWatchListTicker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, ticker WatchListTicker)
	}{
		{
			name: "full watchlist ticker with all fields",
			input: `{
				"Ticker": "NVDA",
				"Price": 875.50,
				"NearestTop10TradeDate": "/Date(1745366400000)/",
				"NearestTop10TradeClusterDate": "/Date(1745452800000)/",
				"NearestTop10TradeLevel": 870.00
			}`,
			check: func(t *testing.T, ticker WatchListTicker) {
				if ticker.Ticker != "NVDA" {
					t.Errorf("Ticker: expected NVDA, got %v", ticker.Ticker)
				}
				if ticker.Price != 875.50 {
					t.Errorf("Price: expected 875.50, got %v", ticker.Price)
				}
				if !ticker.NearestTop10TradeDate.Valid {
					t.Error("NearestTop10TradeDate should be valid")
				}
				if !ticker.NearestTop10TradeClusterDate.Valid {
					t.Error("NearestTop10TradeClusterDate should be valid")
				}
				if ticker.NearestTop10TradeLevel == nil || *ticker.NearestTop10TradeLevel != 870.00 {
					t.Errorf("NearestTop10TradeLevel: expected 870.00, got %v", ticker.NearestTop10TradeLevel)
				}
			},
		},
		{
			name: "watchlist ticker with null trade level",
			input: `{
				"Ticker": "AAPL",
				"Price": 150.00,
				"NearestTop10TradeDate": "/Date(1745366400000)/",
				"NearestTop10TradeClusterDate": "/Date(1745452800000)/",
				"NearestTop10TradeLevel": null
			}`,
			check: func(t *testing.T, ticker WatchListTicker) {
				if ticker.Ticker != "AAPL" {
					t.Errorf("Ticker: expected AAPL, got %v", ticker.Ticker)
				}
				if ticker.NearestTop10TradeLevel != nil {
					t.Errorf("NearestTop10TradeLevel: expected nil, got %v", ticker.NearestTop10TradeLevel)
				}
			},
		},
		{
			name: "watchlist ticker with invalid dates",
			input: `{
				"Ticker": "MSFT",
				"Price": 300.00,
				"NearestTop10TradeDate": null,
				"NearestTop10TradeClusterDate": null,
				"NearestTop10TradeLevel": 295.00
			}`,
			check: func(t *testing.T, ticker WatchListTicker) {
				if ticker.NearestTop10TradeDate.Valid {
					t.Error("NearestTop10TradeDate should be invalid")
				}
				if ticker.NearestTop10TradeClusterDate.Valid {
					t.Error("NearestTop10TradeClusterDate should be invalid")
				}
				if ticker.NearestTop10TradeLevel == nil || *ticker.NearestTop10TradeLevel != 295.00 {
					t.Errorf("NearestTop10TradeLevel: expected 295.00, got %v", ticker.NearestTop10TradeLevel)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var ticker WatchListTicker
			err := json.Unmarshal([]byte(tt.input), &ticker)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected unmarshal error: %v", err)
			}
			tt.check(t, ticker)
		})
	}
}

func TestWatchListConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, config WatchListConfig)
	}{
		{
			name: "full watchlist config with all fields",
			input: `{
				"SearchTemplateKey": 1,
				"UserKey": 100,
				"SearchTemplateTypeKey": 2,
				"Name": "Tech Stocks",
				"Tickers": "NVDA,AAPL,MSFT",
				"SortOrder": 1,
				"MinVolume": 10000,
				"MaxVolume": 1000000,
				"MinDollars": 100000.00,
				"MaxDollars": 10000000.00,
				"MinPrice": 50.00,
				"MaxPrice": 500.00,
				"RSIOverboughtHourly": 70,
				"RSIOverboughtDaily": 75,
				"RSIOversoldHourly": 30,
				"RSIOversoldDaily": 25,
				"Conditions": "EOM,EOQ",
				"RSIOverboughtHourlySelected": true,
				"RSIOverboughtDailySelected": false,
				"RSIOversoldHourlySelected": true,
				"RSIOversoldDailySelected": false,
				"MinRelativeSize": 1,
				"MinRelativeSizeSelected": true,
				"MaxTradeRank": 10,
				"SecurityTypeKey": 1,
				"SecurityType": "Stock",
				"MaxTradeRankSelected": true,
				"MinVCD": 0.50,
				"NormalPrints": true,
				"SignaturePrints": true,
				"LatePrints": false,
				"TimelyPrints": true,
				"DarkPools": true,
				"LitExchanges": true,
				"Sweeps": false,
				"Blocks": true,
				"PremarketTrades": true,
				"RTHTrades": true,
				"AHTrades": false,
				"OpeningTrades": true,
				"ClosingTrades": true,
				"PhantomTrades": false,
				"OffsettingTrades": false,
				"NormalPrintsSelected": true,
				"SignaturePrintsSelected": true,
				"LatePrintsSelected": false,
				"TimelyPrintsSelected": true,
				"DarkPoolsSelected": true,
				"LitExchangesSelected": true,
				"SweepsSelected": false,
				"BlocksSelected": true,
				"PremarketTradesSelected": true,
				"RTHTradesSelected": true,
				"AHTradesSelected": false,
				"OpeningTradesSelected": true,
				"ClosingTradesSelected": true,
				"PhantomTradesSelected": false,
				"OffsettingTradesSelected": false,
				"SectorIndustry": "Technology",
				"APIKey": "test-key-123"
			}`,
			check: func(t *testing.T, config WatchListConfig) {
				if config.Name != "Tech Stocks" {
					t.Errorf("Name: expected Tech Stocks, got %v", config.Name)
				}
				if config.Tickers != "NVDA,AAPL,MSFT" {
					t.Errorf("Tickers: expected NVDA,AAPL,MSFT, got %v", config.Tickers)
				}
				if config.SortOrder == nil || *config.SortOrder != 1 {
					t.Errorf("SortOrder: expected 1, got %v", config.SortOrder)
				}
				if config.MinVolume != 10000 {
					t.Errorf("MinVolume: expected 10000, got %v", config.MinVolume)
				}
				if config.NormalPrints != true {
					t.Errorf("NormalPrints: expected true, got %v", config.NormalPrints)
				}
				if config.SectorIndustry == nil || *config.SectorIndustry != "Technology" {
					t.Errorf("SectorIndustry: expected Technology, got %v", config.SectorIndustry)
				}
				if config.APIKey == nil || *config.APIKey != "test-key-123" {
					t.Errorf("APIKey: expected test-key-123, got %v", config.APIKey)
				}
			},
		},
		{
			name: "watchlist config with null pointer fields",
			input: `{
				"SearchTemplateKey": 2,
				"UserKey": 200,
				"SearchTemplateTypeKey": 3,
				"Name": "Default",
				"Tickers": "",
				"SortOrder": null,
				"MinVolume": 0,
				"MaxVolume": 0,
				"MinDollars": 0.00,
				"MaxDollars": 0.00,
				"MinPrice": 0.00,
				"MaxPrice": 0.00,
				"RSIOverboughtHourly": null,
				"RSIOverboughtDaily": null,
				"RSIOversoldHourly": null,
				"RSIOversoldDaily": null,
				"Conditions": "",
				"RSIOverboughtHourlySelected": null,
				"RSIOverboughtDailySelected": null,
				"RSIOversoldHourlySelected": null,
				"RSIOversoldDailySelected": null,
				"MinRelativeSize": 0,
				"MinRelativeSizeSelected": null,
				"MaxTradeRank": 0,
				"SecurityTypeKey": 0,
				"SecurityType": null,
				"MaxTradeRankSelected": null,
				"MinVCD": 0.00,
				"NormalPrints": false,
				"SignaturePrints": false,
				"LatePrints": false,
				"TimelyPrints": false,
				"DarkPools": false,
				"LitExchanges": false,
				"Sweeps": false,
				"Blocks": false,
				"PremarketTrades": false,
				"RTHTrades": false,
				"AHTrades": false,
				"OpeningTrades": false,
				"ClosingTrades": false,
				"PhantomTrades": false,
				"OffsettingTrades": false,
				"NormalPrintsSelected": false,
				"SignaturePrintsSelected": false,
				"LatePrintsSelected": false,
				"TimelyPrintsSelected": false,
				"DarkPoolsSelected": false,
				"LitExchangesSelected": false,
				"SweepsSelected": false,
				"BlocksSelected": false,
				"PremarketTradesSelected": false,
				"RTHTradesSelected": false,
				"AHTradesSelected": false,
				"OpeningTradesSelected": false,
				"ClosingTradesSelected": false,
				"PhantomTradesSelected": false,
				"OffsettingTradesSelected": false,
				"SectorIndustry": null,
				"APIKey": null
			}`,
			check: func(t *testing.T, config WatchListConfig) {
				if config.SortOrder != nil {
					t.Errorf("SortOrder: expected nil, got %v", config.SortOrder)
				}
				if config.SecurityType != nil {
					t.Errorf("SecurityType: expected nil, got %v", config.SecurityType)
				}
				if config.SectorIndustry != nil {
					t.Errorf("SectorIndustry: expected nil, got %v", config.SectorIndustry)
				}
				if config.APIKey != nil {
					t.Errorf("APIKey: expected nil, got %v", config.APIKey)
				}
			},
		},
		{
			name: "watchlist config with mixed bool and pointer bool fields",
			input: `{
				"SearchTemplateKey": 3,
				"UserKey": 300,
				"SearchTemplateTypeKey": 1,
				"Name": "Mixed",
				"Tickers": "TEST",
				"SortOrder": 2,
				"MinVolume": 5000,
				"MaxVolume": 500000,
				"MinDollars": 50000.00,
				"MaxDollars": 5000000.00,
				"MinPrice": 25.00,
				"MaxPrice": 250.00,
				"RSIOverboughtHourly": 70,
				"RSIOverboughtDaily": null,
				"RSIOversoldHourly": null,
				"RSIOversoldDaily": 25,
				"Conditions": "EOY",
				"RSIOverboughtHourlySelected": true,
				"RSIOverboughtDailySelected": null,
				"RSIOversoldHourlySelected": null,
				"RSIOversoldDailySelected": false,
				"MinRelativeSize": 2,
				"MinRelativeSizeSelected": true,
				"MaxTradeRank": 5,
				"SecurityTypeKey": 2,
				"SecurityType": "ETF",
				"MaxTradeRankSelected": null,
				"MinVCD": 0.75,
				"NormalPrints": true,
				"SignaturePrints": false,
				"LatePrints": true,
				"TimelyPrints": false,
				"DarkPools": true,
				"LitExchanges": false,
				"Sweeps": true,
				"Blocks": false,
				"PremarketTrades": true,
				"RTHTrades": false,
				"AHTrades": true,
				"OpeningTrades": false,
				"ClosingTrades": true,
				"PhantomTrades": false,
				"OffsettingTrades": true,
				"NormalPrintsSelected": true,
				"SignaturePrintsSelected": false,
				"LatePrintsSelected": true,
				"TimelyPrintsSelected": false,
				"DarkPoolsSelected": true,
				"LitExchangesSelected": false,
				"SweepsSelected": true,
				"BlocksSelected": false,
				"PremarketTradesSelected": true,
				"RTHTradesSelected": false,
				"AHTradesSelected": true,
				"OpeningTradesSelected": false,
				"ClosingTradesSelected": true,
				"PhantomTradesSelected": false,
				"OffsettingTradesSelected": true,
				"SectorIndustry": "Healthcare",
				"APIKey": null
			}`,
			check: func(t *testing.T, config WatchListConfig) {
				if config.Name != "Mixed" {
					t.Errorf("Name: expected Mixed, got %v", config.Name)
				}
				if config.RSIOverboughtDaily != nil {
					t.Errorf("RSIOverboughtDaily: expected nil, got %v", config.RSIOverboughtDaily)
				}
				if config.RSIOversoldHourly != nil {
					t.Errorf("RSIOversoldHourly: expected nil, got %v", config.RSIOversoldHourly)
				}
				if config.MaxTradeRankSelected != nil {
					t.Errorf("MaxTradeRankSelected: expected nil, got %v", config.MaxTradeRankSelected)
				}
				if config.NormalPrints != true {
					t.Errorf("NormalPrints: expected true, got %v", config.NormalPrints)
				}
				if config.SignaturePrints != false {
					t.Errorf("SignaturePrints: expected false, got %v", config.SignaturePrints)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var config WatchListConfig
			err := json.Unmarshal([]byte(tt.input), &config)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected unmarshal error: %v", err)
			}
			tt.check(t, config)
		})
	}
}
