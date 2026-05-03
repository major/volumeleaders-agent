package models

import (
	"encoding/json"
	"testing"
)

func TestTradeCluster(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, cluster TradeCluster)
	}{
		{
			name: "full trade cluster with flex bools and pointer fields",
			input: `{
				"Date": "/Date(1745366400000)/",
				"DateKey": 20250423,
				"SecurityKey": 12345,
				"Ticker": "NVDA",
				"Sector": "Technology",
				"Industry": "Semiconductors",
				"Name": "NVIDIA Corporation",
				"MinFullDateTime": "2025-04-23 09:30:00",
				"MaxFullDateTime": "2025-04-23 16:00:00",
				"MinFullTimeString24": "09:30:00",
				"MaxFullTimeString24": "16:00:00",
				"ClosePrice": 875.50,
				"Price": 870.00,
				"Dollars": 5000000.00,
				"AverageBlockSizeShares": 50000,
				"AverageBlockSizeDollars": 43500000.00,
				"Volume": 100000,
				"TradeCount": 10,
				"IPODate": "/Date(1234567890000)/",
				"DollarsMultiplier": 2.5,
				"CumulativeDistribution": 0.90,
				"AverageDailyVolume": 50000000,
				"EOM": 1,
				"EOQ": 0,
				"EOY": 1,
				"OPEX": 0,
				"VOLEX": 1,
				"InsideBar": 0,
				"DoubleInsideBar": 1,
				"LastComparibleTradeClusterDate": "/Date(1745280000000)/",
				"TradeClusterRank": 1,
				"TotalRows": 100,
				"ExternalFeed": 0
			}`,
			check: func(t *testing.T, cluster TradeCluster) {
				if cluster.Ticker != "NVDA" {
					t.Errorf("Ticker: expected NVDA, got %v", cluster.Ticker)
				}
				if cluster.Industry == nil || *cluster.Industry != "Semiconductors" {
					t.Errorf("Industry: expected Semiconductors, got %v", cluster.Industry)
				}
				if !cluster.Date.Valid {
					t.Error("Date should be valid")
				}
				if cluster.EOM != true {
					t.Errorf("EOM: expected true, got %v", cluster.EOM)
				}
				if cluster.EOQ != false {
					t.Errorf("EOQ: expected false, got %v", cluster.EOQ)
				}
				if cluster.VOLEX != true {
					t.Errorf("VOLEX: expected true, got %v", cluster.VOLEX)
				}
				if cluster.DoubleInsideBar != true {
					t.Errorf("DoubleInsideBar: expected true, got %v", cluster.DoubleInsideBar)
				}
			},
		},
		{
			name: "trade cluster with null industry",
			input: `{
				"Date": "/Date(1745366400000)/",
				"DateKey": 20250423,
				"SecurityKey": 12345,
				"Ticker": "AAPL",
				"Sector": "Technology",
				"Industry": null,
				"Name": "Apple Inc.",
				"MinFullDateTime": "2025-04-23 09:30:00",
				"MaxFullDateTime": "2025-04-23 16:00:00",
				"MinFullTimeString24": "09:30:00",
				"MaxFullTimeString24": "16:00:00",
				"ClosePrice": 150.00,
				"Price": 149.50,
				"Dollars": 2000000.00,
				"AverageBlockSizeShares": 10000,
				"AverageBlockSizeDollars": 1500000.00,
				"Volume": 50000,
				"TradeCount": 5,
				"IPODate": "/Date(1234567890000)/",
				"DollarsMultiplier": 1.5,
				"CumulativeDistribution": 0.75,
				"AverageDailyVolume": 100000000,
				"EOM": 0,
				"EOQ": 0,
				"EOY": 0,
				"OPEX": 0,
				"VOLEX": 0,
				"InsideBar": 0,
				"DoubleInsideBar": 0,
				"LastComparibleTradeClusterDate": "/Date(1745280000000)/",
				"TradeClusterRank": 5,
				"TotalRows": 50,
				"ExternalFeed": 0
			}`,
			check: func(t *testing.T, cluster TradeCluster) {
				if cluster.Industry != nil {
					t.Errorf("Industry: expected nil, got %v", cluster.Industry)
				}
			},
		},
		{
			name: "trade cluster with invalid dates",
			input: `{
				"Date": null,
				"DateKey": 20250423,
				"SecurityKey": 12345,
				"Ticker": "MSFT",
				"Sector": "Technology",
				"Industry": null,
				"Name": "Microsoft Corporation",
				"MinFullDateTime": "2025-04-23 09:30:00",
				"MaxFullDateTime": "2025-04-23 16:00:00",
				"MinFullTimeString24": "09:30:00",
				"MaxFullTimeString24": "16:00:00",
				"ClosePrice": 300.00,
				"Price": 299.50,
				"Dollars": 3000000.00,
				"AverageBlockSizeShares": 20000,
				"AverageBlockSizeDollars": 6000000.00,
				"Volume": 75000,
				"TradeCount": 8,
				"IPODate": null,
				"DollarsMultiplier": 2.0,
				"CumulativeDistribution": 0.80,
				"AverageDailyVolume": 75000000,
				"EOM": 0,
				"EOQ": 1,
				"EOY": 0,
				"OPEX": 1,
				"VOLEX": 0,
				"InsideBar": 1,
				"DoubleInsideBar": 0,
				"LastComparibleTradeClusterDate": null,
				"TradeClusterRank": 3,
				"TotalRows": 75,
				"ExternalFeed": 1
			}`,
			check: func(t *testing.T, cluster TradeCluster) {
				if cluster.Date.Valid {
					t.Error("Date should be invalid")
				}
				if cluster.IPODate.Valid {
					t.Error("IPODate should be invalid")
				}
				if cluster.LastComparibleTradeClusterDate.Valid {
					t.Error("LastComparibleTradeClusterDate should be invalid")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var cluster TradeCluster
			err := json.Unmarshal([]byte(tt.input), &cluster)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected unmarshal error: %v", err)
			}
			tt.check(t, cluster)
		})
	}
}

func TestTradeClusterBomb(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, bomb TradeClusterBomb)
	}{
		{
			name: "full trade cluster bomb",
			input: `{
				"Date": "/Date(1745366400000)/",
				"DateKey": 20250423,
				"SecurityKey": 12345,
				"Ticker": "NVDA",
				"Sector": "Technology",
				"Industry": "Semiconductors",
				"Name": "NVIDIA Corporation",
				"MinFullDateTime": "2025-04-23 09:30:00",
				"MaxFullDateTime": "2025-04-23 16:00:00",
				"MinFullTimeString24": "09:30:00",
				"MaxFullTimeString24": "16:00:00",
				"ClosePrice": 875.50,
				"Dollars": 10000000.00,
				"AverageBlockSizeShares": 100000,
				"AverageBlockSizeDollars": 87500000.00,
				"Volume": 200000,
				"TradeCount": 20,
				"IPODate": "/Date(1234567890000)/",
				"DollarsMultiplier": 3.0,
				"CumulativeDistribution": 0.95,
				"AverageDailyVolume": 50000000,
				"EOM": 1,
				"EOQ": 1,
				"EOY": 0,
				"OPEX": 1,
				"VOLEX": 1,
				"InsideBar": 0,
				"DoubleInsideBar": 1,
				"LastComparableTradeClusterBombDate": "/Date(1745280000000)/",
				"TradeClusterBombRank": 1,
				"TotalRows": 100,
				"ExternalFeed": 0
			}`,
			check: func(t *testing.T, bomb TradeClusterBomb) {
				if bomb.Ticker != "NVDA" {
					t.Errorf("Ticker: expected NVDA, got %v", bomb.Ticker)
				}
				if bomb.Industry == nil || *bomb.Industry != "Semiconductors" {
					t.Errorf("Industry: expected Semiconductors, got %v", bomb.Industry)
				}
				if !bomb.Date.Valid {
					t.Error("Date should be valid")
				}
				if bomb.EOM != true {
					t.Errorf("EOM: expected true, got %v", bomb.EOM)
				}
				if bomb.EOQ != true {
					t.Errorf("EOQ: expected true, got %v", bomb.EOQ)
				}
				if bomb.TradeClusterBombRank != 1 {
					t.Errorf("TradeClusterBombRank: expected 1, got %v", bomb.TradeClusterBombRank)
				}
			},
		},
		{
			name: "trade cluster bomb with null industry",
			input: `{
				"Date": "/Date(1745366400000)/",
				"DateKey": 20250423,
				"SecurityKey": 12345,
				"Ticker": "AAPL",
				"Sector": "Technology",
				"Industry": null,
				"Name": "Apple Inc.",
				"MinFullDateTime": "2025-04-23 09:30:00",
				"MaxFullDateTime": "2025-04-23 16:00:00",
				"MinFullTimeString24": "09:30:00",
				"MaxFullTimeString24": "16:00:00",
				"ClosePrice": 150.00,
				"Dollars": 5000000.00,
				"AverageBlockSizeShares": 50000,
				"AverageBlockSizeDollars": 7500000.00,
				"Volume": 100000,
				"TradeCount": 10,
				"IPODate": "/Date(1234567890000)/",
				"DollarsMultiplier": 2.0,
				"CumulativeDistribution": 0.85,
				"AverageDailyVolume": 100000000,
				"EOM": 0,
				"EOQ": 0,
				"EOY": 0,
				"OPEX": 0,
				"VOLEX": 0,
				"InsideBar": 0,
				"DoubleInsideBar": 0,
				"LastComparableTradeClusterBombDate": "/Date(1745280000000)/",
				"TradeClusterBombRank": 5,
				"TotalRows": 50,
				"ExternalFeed": 0
			}`,
			check: func(t *testing.T, bomb TradeClusterBomb) {
				if bomb.Industry != nil {
					t.Errorf("Industry: expected nil, got %v", bomb.Industry)
				}
			},
		},
		{
			name: "trade cluster bomb with invalid dates",
			input: `{
				"Date": null,
				"DateKey": 20250423,
				"SecurityKey": 12345,
				"Ticker": "MSFT",
				"Sector": "Technology",
				"Industry": null,
				"Name": "Microsoft Corporation",
				"MinFullDateTime": "2025-04-23 09:30:00",
				"MaxFullDateTime": "2025-04-23 16:00:00",
				"MinFullTimeString24": "09:30:00",
				"MaxFullTimeString24": "16:00:00",
				"ClosePrice": 300.00,
				"Dollars": 6000000.00,
				"AverageBlockSizeShares": 40000,
				"AverageBlockSizeDollars": 12000000.00,
				"Volume": 150000,
				"TradeCount": 15,
				"IPODate": null,
				"DollarsMultiplier": 2.5,
				"CumulativeDistribution": 0.90,
				"AverageDailyVolume": 75000000,
				"EOM": 0,
				"EOQ": 0,
				"EOY": 1,
				"OPEX": 0,
				"VOLEX": 1,
				"InsideBar": 1,
				"DoubleInsideBar": 0,
				"LastComparableTradeClusterBombDate": null,
				"TradeClusterBombRank": 3,
				"TotalRows": 75,
				"ExternalFeed": 1
			}`,
			check: func(t *testing.T, bomb TradeClusterBomb) {
				if bomb.Date.Valid {
					t.Error("Date should be invalid")
				}
				if bomb.IPODate.Valid {
					t.Error("IPODate should be invalid")
				}
				if bomb.LastComparableTradeClusterBombDate.Valid {
					t.Error("LastComparableTradeClusterBombDate should be invalid")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var bomb TradeClusterBomb
			err := json.Unmarshal([]byte(tt.input), &bomb)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected unmarshal error: %v", err)
			}
			tt.check(t, bomb)
		})
	}
}

func TestTradeClusterAlias(t *testing.T) {
	t.Parallel()

	t.Run("TradeClusterAlert is assignable from TradeCluster", func(t *testing.T) {
		t.Parallel()
		// Verify that TradeClusterAlert (type alias) is compatible with TradeCluster
		cluster := TradeCluster{
			Ticker: "NVDA",
			Sector: "Technology",
			Name:   "NVIDIA Corporation",
		}
		// Explicit type annotation proves TradeClusterAlert is a type alias for TradeCluster.
		// The redundant type is intentional: it verifies the alias relationship at compile time.
		var alert TradeClusterAlert = cluster //nolint:staticcheck // explicit type proves alias relationship at compile time
		if alert.Ticker != "NVDA" {
			t.Errorf("alert.Ticker: expected NVDA, got %v", alert.Ticker)
		}
	})

	t.Run("TradeClusterAlert unmarshals as TradeCluster", func(t *testing.T) {
		t.Parallel()
		input := `{
			"Date": "/Date(1745366400000)/",
			"DateKey": 20250423,
			"SecurityKey": 12345,
			"Ticker": "AAPL",
			"Sector": "Technology",
			"Industry": null,
			"Name": "Apple Inc.",
			"MinFullDateTime": "2025-04-23 09:30:00",
			"MaxFullDateTime": "2025-04-23 16:00:00",
			"MinFullTimeString24": "09:30:00",
			"MaxFullTimeString24": "16:00:00",
			"ClosePrice": 150.00,
			"Price": 149.50,
			"Dollars": 2000000.00,
			"AverageBlockSizeShares": 10000,
			"AverageBlockSizeDollars": 1500000.00,
			"Volume": 50000,
			"TradeCount": 5,
			"IPODate": "/Date(1234567890000)/",
			"DollarsMultiplier": 1.5,
			"CumulativeDistribution": 0.75,
			"AverageDailyVolume": 100000000,
			"EOM": 0,
			"EOQ": 0,
			"EOY": 0,
			"OPEX": 0,
			"VOLEX": 0,
			"InsideBar": 0,
			"DoubleInsideBar": 0,
			"LastComparibleTradeClusterDate": "/Date(1745280000000)/",
			"TradeClusterRank": 5,
			"TotalRows": 50,
			"ExternalFeed": 0
		}`
		var alert TradeClusterAlert
		err := json.Unmarshal([]byte(input), &alert)
		if err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		if alert.Ticker != "AAPL" {
			t.Errorf("alert.Ticker: expected AAPL, got %v", alert.Ticker)
		}
		if !alert.Date.Valid {
			t.Error("alert.Date should be valid")
		}
	})
}
