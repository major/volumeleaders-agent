package trade

import (
	"context"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/major/volumeleaders-agent/internal/cli/testutil"
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

	cmd := newTradeListCommand()
	ctx := testutil.ContextWithTestClient(t, "http://127.0.0.1")
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "--start-date", "2025-04-21", "--end-date", "2025-04-21", "--length", "100")
	testutil.AssertErrContains(t, err, "--length must be between 1 and 50 for trade retrieval")
}

func TestTradeListAcceptsMaximumLength(t *testing.T) {
	t.Parallel()

	var gotLength string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		params, _ := url.ParseQuery(string(body))
		gotLength = params.Get("length")
		_, _ = w.Write([]byte(testutil.DataTablesJSON(`[]`)))
	}))
	t.Cleanup(server.Close)

	cmd := newTradeListCommand()
	ctx := testutil.ContextWithTestClient(t, server.URL)
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "--start-date", "2025-04-21", "--end-date", "2025-04-21", "--length", "50")
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
			cmd := newTradeLevelTouchesCommand()
			ctx := testutil.ContextWithTestClient(t, "http://127.0.0.1")
			_, _, err := testutil.ExecuteCommand(t, cmd, ctx, tt.args...)
			testutil.AssertErrContains(t, err, tt.want)
		})
	}
}

func TestTradeLevelTouchesAcceptsMaximumTradeLevelCount(t *testing.T) {
	t.Parallel()

	var got url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		got, _ = url.ParseQuery(string(body))
		_, _ = w.Write([]byte(testutil.DataTablesJSON(`[]`)))
	}))
	t.Cleanup(server.Close)

	cmd := newTradeLevelTouchesCommand()
	ctx := testutil.ContextWithTestClient(t, server.URL)
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "AAPL", "--start-date", "2025-04-21", "--end-date", "2025-04-21", "--trade-level-count", "50")
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
