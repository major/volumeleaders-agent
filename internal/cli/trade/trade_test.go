package trade

import (
	"context"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"
	"time"

	"github.com/major/volumeleaders-agent/internal/cli/testutil"
	"github.com/major/volumeleaders-agent/internal/models"
)

func TestNewCmdRegistersTenRunESubcommands(t *testing.T) {
	t.Parallel()

	cmd := NewCmd()
	if len(cmd.Commands()) != 10 {
		t.Fatalf("subcommand count = %d, want 10", len(cmd.Commands()))
	}
	for _, subcmd := range cmd.Commands() {
		if subcmd.RunE == nil {
			t.Fatalf("subcommand %q missing RunE", subcmd.Name())
		}
	}
}

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
	if !maps.Equal(filters, expected) {
		t.Fatalf("filters mismatch\nexpected: %#v\ngot:      %#v", expected, filters)
	}
}

func TestPresetTradeFilterDefaultsPreserveTriStateValues(t *testing.T) {
	t.Parallel()

	var opts tradeFilterFlags
	presetTradeFilterDefaults(&opts, 97)

	if got := opts.DarkPools.Int(); got != -1 {
		t.Errorf("DarkPools default = %d, want -1", got)
	}
	if got := opts.Premarket.Int(); got != 1 {
		t.Errorf("Premarket default = %d, want 1", got)
	}
}

func TestBuildTradeLevelFiltersUseObservedLevelKeys(t *testing.T) {
	t.Parallel()

	filters := buildTradeLevelFilters(&tradeLevelOptions{
		ticker:          "AMD",
		startDate:       "2025-04-29",
		endDate:         "2026-04-29",
		minVolume:       100,
		maxVolume:       200,
		minPrice:        1.5,
		maxPrice:        99.25,
		minDollars:      500000,
		maxDollars:      30000000000,
		vcd:             99,
		relativeSize:    10,
		tradeLevelRank:  5,
		tradeLevelCount: 10,
	})

	expected := map[string]string{
		"Ticker":         "AMD",
		"MinVolume":      "100",
		"MaxVolume":      "200",
		"MinPrice":       "1.5",
		"MaxPrice":       "99.25",
		"MinDollars":     "500000",
		"MaxDollars":     "30000000000",
		"VCD":            "99",
		"RelativeSize":   "10",
		"StartDate":      "2025-04-29",
		"EndDate":        "2026-04-29",
		"TradeLevelRank": "5",
		"Levels":         "10",
	}
	if !maps.Equal(filters, expected) {
		t.Fatalf("filters mismatch\nexpected: %#v\ngot:      %#v", expected, filters)
	}
}

func TestTradeListLengthCappedAtFifty(t *testing.T) {
	t.Parallel()

	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, "http://127.0.0.1")
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "list", "--start-date", "2025-04-21", "--end-date", "2025-04-21", "--length", "100")
	testutil.AssertErrContains(t, err, "--length must be between 1 and 50 for trade retrieval")
}

func TestTradeListAcceptsMaximumLength(t *testing.T) {
	t.Parallel()

	var gotLength string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}
		params, _ := url.ParseQuery(string(body))
		gotLength = params.Get("length")
		_, _ = w.Write([]byte(testutil.DataTablesJSON(`[]`)))
	}))
	t.Cleanup(server.Close)

	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, server.URL)
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "list", "--start-date", "2025-04-21", "--end-date", "2025-04-21", "--length", "50")
	testutil.AssertErrContains(t, err, "")
	if gotLength != "50" {
		t.Fatalf("length = %q, want 50", gotLength)
	}
}

func TestTradeLevelTouchesCapsLengthAndTradeLevelCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "length over max",
			args: []string{"AAPL", "--start-date", "2025-04-21", "--end-date", "2025-04-21", "--length", "100"},
			want: "--length must be between 1 and 50 for trade level touch retrieval",
		},
		{
			name: "trade level count over max",
			args: []string{"AAPL", "--start-date", "2025-04-21", "--end-date", "2025-04-21", "--trade-level-count", "100"},
			want: "--trade-level-count must be between 1 and 50 for trade level retrieval",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCmd()
			ctx := testutil.ContextWithTestClient(t, "http://127.0.0.1")
			args := append([]string{"level-touches"}, tt.args...)
			_, _, err := testutil.ExecuteCommand(t, cmd, ctx, args...)
			testutil.AssertErrContains(t, err, tt.want)
		})
	}
}

func TestTradeLevelsCapsTradeLevelCount(t *testing.T) {
	t.Parallel()

	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, "http://127.0.0.1")
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "levels", "AAPL", "--trade-level-count", "100")
	testutil.AssertErrContains(t, err, "--trade-level-count must be between 1 and 50 for trade level retrieval")
}

func TestTradeLevelTouchesAcceptsMaximumTradeLevelCount(t *testing.T) {
	t.Parallel()

	var got url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}
		got, _ = url.ParseQuery(string(body))
		_, _ = w.Write([]byte(testutil.DataTablesJSON(`[]`)))
	}))
	t.Cleanup(server.Close)

	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, server.URL)
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "level-touches", "AAPL", "--start-date", "2025-04-21", "--end-date", "2025-04-21", "--trade-level-count", "50")
	testutil.AssertErrContains(t, err, "")
	if got.Get("Levels") != "50" {
		t.Fatalf("Levels = %q, want 50", got.Get("Levels"))
	}
}

func TestExecuteTradeCommandWithoutServerDoesNotPanic(t *testing.T) {
	t.Parallel()

	cmd := newTradePresetsCommand()
	_, _, err := testutil.ExecuteCommand(t, cmd, context.Background())
	testutil.AssertErrContains(t, err, "")
}

func floatPtr(f float64) *float64 {
	return &f
}

func TestTradeSentimentRatio(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		bullDollars float64
		bearDollars float64
		want        *float64
	}{
		{
			name:        "bearDollars zero returns nil",
			bullDollars: 100,
			bearDollars: 0,
			want:        nil,
		},
		{
			name:        "bullDollars 100 bearDollars 50 returns 2.0",
			bullDollars: 100,
			bearDollars: 50,
			want:        floatPtr(2.0),
		},
		{
			name:        "bullDollars 0 bearDollars 100 returns 0.0",
			bullDollars: 0,
			bearDollars: 100,
			want:        floatPtr(0.0),
		},
		{
			name:        "bullDollars 50 bearDollars 100 returns 0.5",
			bullDollars: 50,
			bearDollars: 100,
			want:        floatPtr(0.5),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tradeSentimentRatio(tt.bullDollars, tt.bearDollars)
			if tt.want == nil {
				if got != nil {
					t.Fatalf("tradeSentimentRatio(%v, %v) = %v, want nil", tt.bullDollars, tt.bearDollars, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("tradeSentimentRatio(%v, %v) = nil, want %v", tt.bullDollars, tt.bearDollars, *tt.want)
			}
			if *got != *tt.want {
				t.Fatalf("tradeSentimentRatio(%v, %v) = %v, want %v", tt.bullDollars, tt.bearDollars, *got, *tt.want)
			}
		})
	}
}

func TestTradeSentimentSignal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ratio       *float64
		bullDollars float64
		bearDollars float64
		want        models.TradeSentimentSignal
	}{
		{
			name:        "nil ratio with bullDollars > 0 returns ExtremeBull",
			ratio:       nil,
			bullDollars: 100,
			bearDollars: 0,
			want:        models.TradeSentimentExtremeBull,
		},
		{
			name:        "nil ratio with bearDollars > 0 returns ExtremeBear",
			ratio:       nil,
			bullDollars: 0,
			bearDollars: 100,
			want:        models.TradeSentimentExtremeBear,
		},
		{
			name:        "nil ratio with both zero returns Neutral",
			ratio:       nil,
			bullDollars: 0,
			bearDollars: 0,
			want:        models.TradeSentimentNeutral,
		},
		{
			name:        "ratio 0.1 (< 0.2) returns ExtremeBear",
			ratio:       floatPtr(0.1),
			bullDollars: 0,
			bearDollars: 0,
			want:        models.TradeSentimentExtremeBear,
		},
		{
			name:        "ratio 0.2 (boundary, >= 0.2, < 0.5) returns ModerateBear",
			ratio:       floatPtr(0.2),
			bullDollars: 0,
			bearDollars: 0,
			want:        models.TradeSentimentModerateBear,
		},
		{
			name:        "ratio 0.3 (>= 0.2, < 0.5) returns ModerateBear",
			ratio:       floatPtr(0.3),
			bullDollars: 0,
			bearDollars: 0,
			want:        models.TradeSentimentModerateBear,
		},
		{
			name:        "ratio 0.5 (boundary, >= 0.5, <= 2.0) returns Neutral",
			ratio:       floatPtr(0.5),
			bullDollars: 0,
			bearDollars: 0,
			want:        models.TradeSentimentNeutral,
		},
		{
			name:        "ratio 1.0 (>= 0.5, <= 2.0) returns Neutral",
			ratio:       floatPtr(1.0),
			bullDollars: 0,
			bearDollars: 0,
			want:        models.TradeSentimentNeutral,
		},
		{
			name:        "ratio 2.0 (boundary, >= 0.5, <= 2.0) returns Neutral",
			ratio:       floatPtr(2.0),
			bullDollars: 0,
			bearDollars: 0,
			want:        models.TradeSentimentNeutral,
		},
		{
			name:        "ratio 3.0 (> 2.0, <= 5.0) returns ModerateBull",
			ratio:       floatPtr(3.0),
			bullDollars: 0,
			bearDollars: 0,
			want:        models.TradeSentimentModerateBull,
		},
		{
			name:        "ratio 5.0 (boundary, > 2.0, <= 5.0) returns ModerateBull",
			ratio:       floatPtr(5.0),
			bullDollars: 0,
			bearDollars: 0,
			want:        models.TradeSentimentModerateBull,
		},
		{
			name:        "ratio 6.0 (> 5.0) returns ExtremeBull",
			ratio:       floatPtr(6.0),
			bullDollars: 0,
			bearDollars: 0,
			want:        models.TradeSentimentExtremeBull,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tradeSentimentSignal(tt.ratio, tt.bullDollars, tt.bearDollars)
			if got != tt.want {
				t.Fatalf("tradeSentimentSignal(%v, %v, %v) = %q, want %q", tt.ratio, tt.bullDollars, tt.bearDollars, got, tt.want)
			}
		})
	}
}

func TestLeveragedETFDirection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		ticker string
		want   string
	}{
		{
			name:   "SQQQ is bear",
			ticker: "SQQQ",
			want:   "bear",
		},
		{
			name:   "SPXS is bear",
			ticker: "SPXS",
			want:   "bear",
		},
		{
			name:   "SPXU is bear",
			ticker: "SPXU",
			want:   "bear",
		},
		{
			name:   "SDOW is bear",
			ticker: "SDOW",
			want:   "bear",
		},
		{
			name:   "TZA is bear",
			ticker: "TZA",
			want:   "bear",
		},
		{
			name:   "TQQQ is bull",
			ticker: "TQQQ",
			want:   "bull",
		},
		{
			name:   "SPXL is bull",
			ticker: "SPXL",
			want:   "bull",
		},
		{
			name:   "SSO is bull",
			ticker: "SSO",
			want:   "bull",
		},
		{
			name:   "QLD is bull",
			ticker: "QLD",
			want:   "bull",
		},
		{
			name:   "AAPL is unknown",
			ticker: "AAPL",
			want:   "",
		},
		{
			name:   "sqqq lowercase is bear",
			ticker: "sqqq",
			want:   "bear",
		},
		{
			name:   "tqqq lowercase is bull",
			ticker: "tqqq",
			want:   "bull",
		},
		{
			name:   "SQQQ with leading space is bear",
			ticker: " SQQQ",
			want:   "bear",
		},
		{
			name:   "SQQQ with trailing space is bear",
			ticker: "SQQQ ",
			want:   "bear",
		},
		{
			name:   "SQQQ with both spaces is bear",
			ticker: " SQQQ ",
			want:   "bear",
		},
		{
			name:   "empty string is unknown",
			ticker: "",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := leveragedETFDirection(tt.ticker)
			if got != tt.want {
				t.Fatalf("leveragedETFDirection(%q) = %q, want %q", tt.ticker, got, tt.want)
			}
		})
	}
}

func TestSummarizeTrades(t *testing.T) {
	t.Parallel()

	jan15 := models.AspNetDate{Time: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Valid: true}
	jan16 := models.AspNetDate{Time: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC), Valid: true}

	tests := []struct {
		name             string
		trades           []models.Trade
		group            tradeSummaryGroup
		startDate        string
		endDate          string
		wantTotalTrades  int
		wantTotalDollars float64
		wantByTicker     map[string]models.TradeGroupSummary
		wantByDay        map[string]models.TradeGroupSummary
		wantByTickerDay  map[string]models.TradeGroupSummary
	}{
		{
			name:             "empty trade slice returns zero totals and empty ticker map",
			group:            tradeSummaryGroupTicker,
			startDate:        "2024-01-01",
			endDate:          "2024-01-31",
			wantTotalTrades:  0,
			wantTotalDollars: 0,
			wantByTicker:     map[string]models.TradeGroupSummary{},
		},
		{
			name:             "single trade grouped by ticker",
			trades:           []models.Trade{summaryTrade("AAPL", jan15, 100, 2, true, false, 0.75)},
			group:            tradeSummaryGroupTicker,
			startDate:        "2024-01-15",
			endDate:          "2024-01-15",
			wantTotalTrades:  1,
			wantTotalDollars: 100,
			wantByTicker: map[string]models.TradeGroupSummary{
				"AAPL": {Trades: 1, Dollars: 100, AvgDollarsMultiplier: 2, PctDarkPool: 100, PctSweep: 0, AvgCumulativeDistribution: 0.75},
			},
		},
		{
			name:             "single trade grouped by day",
			trades:           []models.Trade{summaryTrade("AAPL", jan15, 100, 2, true, false, 0.75)},
			group:            tradeSummaryGroupDay,
			startDate:        "2024-01-15",
			endDate:          "2024-01-15",
			wantTotalTrades:  1,
			wantTotalDollars: 100,
			wantByDay: map[string]models.TradeGroupSummary{
				"2024-01-15": {Trades: 1, Dollars: 100, AvgDollarsMultiplier: 2, PctDarkPool: 100, PctSweep: 0, AvgCumulativeDistribution: 0.75},
			},
		},
		{
			name:             "single trade grouped by ticker and day",
			trades:           []models.Trade{summaryTrade("AAPL", jan15, 100, 2, true, false, 0.75)},
			group:            tradeSummaryGroupTickerDay,
			startDate:        "2024-01-15",
			endDate:          "2024-01-15",
			wantTotalTrades:  1,
			wantTotalDollars: 100,
			wantByTickerDay: map[string]models.TradeGroupSummary{
				"AAPL|2024-01-15": {Trades: 1, Dollars: 100, AvgDollarsMultiplier: 2, PctDarkPool: 100, PctSweep: 0, AvgCumulativeDistribution: 0.75},
			},
		},
		{
			name: "multiple trades same ticker aggregate across days",
			trades: []models.Trade{
				summaryTrade("AAPL", jan15, 100, 2, true, false, 0.75),
				summaryTrade("AAPL", jan16, 300, 4, false, true, 0.25),
			},
			group:            tradeSummaryGroupTicker,
			startDate:        "2024-01-15",
			endDate:          "2024-01-16",
			wantTotalTrades:  2,
			wantTotalDollars: 400,
			wantByTicker: map[string]models.TradeGroupSummary{
				"AAPL": {Trades: 2, Dollars: 400, AvgDollarsMultiplier: 3, PctDarkPool: 50, PctSweep: 50, AvgCumulativeDistribution: 0.5},
			},
		},
		{
			name:             "invalid date groups under unknown day key",
			trades:           []models.Trade{summaryTrade("AAPL", models.AspNetDate{}, 100, 2, true, false, 0.75)},
			group:            tradeSummaryGroupDay,
			startDate:        "2024-01-15",
			endDate:          "2024-01-15",
			wantTotalTrades:  1,
			wantTotalDollars: 100,
			wantByDay: map[string]models.TradeGroupSummary{
				"unknown": {Trades: 1, Dollars: 100, AvgDollarsMultiplier: 2, PctDarkPool: 100, PctSweep: 0, AvgCumulativeDistribution: 0.75},
			},
		},
		{
			name:             "date range is copied from parameters",
			trades:           []models.Trade{summaryTrade("AAPL", jan15, 100, 2, true, false, 0.75)},
			group:            tradeSummaryGroupTicker,
			startDate:        "2024-01-01",
			endDate:          "2024-01-31",
			wantTotalTrades:  1,
			wantTotalDollars: 100,
			wantByTicker: map[string]models.TradeGroupSummary{
				"AAPL": {Trades: 1, Dollars: 100, AvgDollarsMultiplier: 2, PctDarkPool: 100, PctSweep: 0, AvgCumulativeDistribution: 0.75},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := summarizeTrades(tt.trades, tt.group, tt.startDate, tt.endDate)
			if got.TotalTrades != tt.wantTotalTrades {
				t.Fatalf("TotalTrades = %d, want %d", got.TotalTrades, tt.wantTotalTrades)
			}
			if got.TotalDollars != tt.wantTotalDollars {
				t.Fatalf("TotalDollars = %v, want %v", got.TotalDollars, tt.wantTotalDollars)
			}
			if got.DateRange.Start != tt.startDate || got.DateRange.End != tt.endDate {
				t.Fatalf("DateRange = {%q %q}, want {%q %q}", got.DateRange.Start, got.DateRange.End, tt.startDate, tt.endDate)
			}
			assertTradeGroupSummaries(t, got.ByTicker, tt.wantByTicker)
			assertTradeGroupSummaries(t, got.ByDay, tt.wantByDay)
			assertTradeGroupSummaries(t, got.ByTickerDay, tt.wantByTickerDay)
		})
	}
}

func TestTopTradeSentimentTickers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		tickerDollars map[string]float64
		limit         int
		want          []string
	}{
		{
			name:          "empty map returns empty slice",
			tickerDollars: map[string]float64{},
			limit:         3,
			want:          []string{},
		},
		{
			name:          "single ticker returns that ticker",
			tickerDollars: map[string]float64{"TQQQ": 100},
			limit:         3,
			want:          []string{"TQQQ"},
		},
		{
			name:          "multiple tickers sort by dollars descending",
			tickerDollars: map[string]float64{"SPXL": 200, "TQQQ": 300, "SSO": 100},
			limit:         3,
			want:          []string{"TQQQ", "SPXL", "SSO"},
		},
		{
			name:          "same dollars sort by ticker ascending",
			tickerDollars: map[string]float64{"TQQQ": 100, "SPXL": 100, "SSO": 100},
			limit:         3,
			want:          []string{"SPXL", "SSO", "TQQQ"},
		},
		{
			name:          "limit two returns top two tickers",
			tickerDollars: map[string]float64{"SPXL": 200, "TQQQ": 300, "SSO": 100},
			limit:         2,
			want:          []string{"TQQQ", "SPXL"},
		},
		{
			name:          "limit zero returns empty slice",
			tickerDollars: map[string]float64{"TQQQ": 100},
			limit:         0,
			want:          []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := topTradeSentimentTickers(tt.tickerDollars, tt.limit)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("topTradeSentimentTickers(%v, %d) = %v, want %v", tt.tickerDollars, tt.limit, got, tt.want)
			}
		})
	}
}

func summaryTrade(ticker string, date models.AspNetDate, dollars, dollarsMultiplier float64, darkPool, sweep bool, cumulativeDistribution float64) models.Trade {
	return models.Trade{
		Ticker:                 ticker,
		Date:                   date,
		Dollars:                dollars,
		DollarsMultiplier:      dollarsMultiplier,
		DarkPool:               models.FlexBool(darkPool),
		Sweep:                  models.FlexBool(sweep),
		CumulativeDistribution: cumulativeDistribution,
	}
}

func assertTradeGroupSummaries(t *testing.T, got, want map[string]models.TradeGroupSummary) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("group count = %d, want %d; got %#v", len(got), len(want), got)
	}
	for key, wantSummary := range want {
		gotSummary, ok := got[key]
		if !ok {
			t.Fatalf("group %q missing from %#v", key, got)
		}
		if gotSummary != wantSummary {
			t.Fatalf("group %q = %#v, want %#v", key, gotSummary, wantSummary)
		}
	}
}
