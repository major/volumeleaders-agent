package commands

import (
	"reflect"
	"testing"
)

func TestBuildTradeFiltersPreservesAPIKeys(t *testing.T) {
	t.Parallel()

	filters := buildTradeFilters(&tradesOptions{
		tickers:      "AAPL,NVDA",
		startDate:    "2026-04-01",
		endDate:      "2026-04-24",
		minVolume:    100,
		maxVolume:    200,
		minPrice:     1.5,
		maxPrice:     99.25,
		minDollars:   500000,
		maxDollars:   30000000000,
		conditions:   -1,
		vcd:          0,
		securityType: -1,
		relativeSize: 5,
		darkPools:    1,
		sweeps:       0,
		latePrints:   -1,
		sigPrints:    1,
		evenShared:   -1,
		tradeRank:    10,
		rankSnapshot: 3,
		marketCap:    0,
		premarket:    1,
		rth:          1,
		ah:           0,
		opening:      1,
		closing:      1,
		phantom:      0,
		offsetting:   1,
		sector:       "Technology",
	})

	expected := map[string]string{
		"Tickers":           "AAPL,NVDA",
		"StartDate":         "2026-04-01",
		"EndDate":           "2026-04-24",
		"MinVolume":         "100",
		"MaxVolume":         "200",
		"MinPrice":          "1.5",
		"MaxPrice":          "99.25",
		"MinDollars":        "500000",
		"MaxDollars":        "30000000000",
		"Conditions":        "-1",
		"VCD":               "0",
		"SecurityTypeKey":   "-1",
		"RelativeSize":      "5",
		"DarkPools":         "1",
		"Sweeps":            "0",
		"LatePrints":        "-1",
		"SignaturePrints":   "1",
		"EvenShared":        "-1",
		"TradeRank":         "10",
		"TradeRankSnapshot": "3",
		"MarketCap":         "0",
		"IncludePremarket":  "1",
		"IncludeRTH":        "1",
		"IncludeAH":         "0",
		"IncludeOpening":    "1",
		"IncludeClosing":    "1",
		"IncludePhantom":    "0",
		"IncludeOffsetting": "1",
		"SectorIndustry":    "Technology",
	}
	if !reflect.DeepEqual(filters, expected) {
		t.Fatalf("filters mismatch\nexpected: %#v\ngot:      %#v", expected, filters)
	}
}

func TestBuildTradeLevelFiltersUseLevelDateKeys(t *testing.T) {
	t.Parallel()

	filters := buildTradeLevelFilters(&tradeLevelOptions{
		ticker:          "AAPL",
		startDate:       "2026-04-01",
		endDate:         "2026-04-24",
		minVolume:       100,
		maxVolume:       200,
		minPrice:        1.5,
		maxPrice:        99.25,
		minDollars:      500000,
		maxDollars:      30000000000,
		vcd:             99,
		relativeSize:    10,
		tradeLevelRank:  5,
		tradeLevelCount: 20,
	})

	expected := map[string]string{
		"Ticker":          "AAPL",
		"MinVolume":       "100",
		"MaxVolume":       "200",
		"MinPrice":        "1.5",
		"MaxPrice":        "99.25",
		"MinDollars":      "500000",
		"MaxDollars":      "30000000000",
		"VCD":             "99",
		"RelativeSize":    "10",
		"MinDate":         "2026-04-01",
		"MaxDate":         "2026-04-24",
		"TradeLevelRank":  "5",
		"TradeLevelCount": "20",
	}
	if !reflect.DeepEqual(filters, expected) {
		t.Fatalf("filters mismatch\nexpected: %#v\ngot:      %#v", expected, filters)
	}
}
