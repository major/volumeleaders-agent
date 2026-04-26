package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/major/volumeleaders-agent/internal/models"
	cli "github.com/urfave/cli/v3"
)

func TestTradeRunFunctions(t *testing.T) {
	tests := []struct {
		name string
		cmd  func() *cli.Command
		args []string
		path string
	}{
		{
			name: "list",
			cmd:  newTradeListCommand,
			args: []string{"app", "list", "--start-date", "2025-01-01", "--end-date", "2025-01-31"},
			path: "/Trades/GetTrades",
		},
		{
			name: "sentiment",
			cmd:  newTradeSentimentCommand,
			args: []string{"app", "sentiment", "--start-date", "2025-01-01", "--end-date", "2025-01-31"},
			path: "/Trades/GetTrades",
		},
		{
			name: "clusters",
			cmd:  newTradeClustersCommand,
			args: []string{"app", "clusters", "--start-date", "2025-01-01", "--end-date", "2025-01-31"},
			path: "/TradeClusters/GetTradeClusters",
		},
		{
			name: "cluster-bombs",
			cmd:  newTradeClusterBombsCommand,
			args: []string{"app", "cluster-bombs", "--start-date", "2025-01-01", "--end-date", "2025-01-31"},
			path: "/TradeClusterBombs/GetTradeClusterBombs",
		},
		{
			name: "alerts",
			cmd:  newTradeAlertsCommand,
			args: []string{"app", "alerts", "--date", "2025-01-15"},
			path: "/TradeAlerts/GetTradeAlerts",
		},
		{
			name: "cluster-alerts",
			cmd:  newTradeClusterAlertsCommand,
			args: []string{"app", "cluster-alerts", "--date", "2025-01-15"},
			path: "/TradeClusterAlerts/GetTradeClusterAlerts",
		},
		{
			name: "levels",
			cmd:  newTradeLevelsCommand,
			args: []string{"app", "levels", "--ticker", "AAPL", "--start-date", "2025-01-01", "--end-date", "2025-01-31"},
			path: "/TradeLevels/GetTradeLevels",
		},
		{
			name: "level-touches",
			cmd:  newTradeLevelTouchesCommand,
			args: []string{"app", "level-touches", "--start-date", "2025-01-01", "--end-date", "2025-01-31"},
			path: "/TradeLevelTouches/GetTradeLevelTouches",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.path {
					t.Errorf("expected path %s, got %s", tt.path, r.URL.Path)
				}
				fmt.Fprint(w, dataTablesJSON(`[{}]`))
			}))
			t.Cleanup(server.Close)

			ctx := contextWithTestClient(server.URL)
			captureStdout(t, func() {
				root := &cli.Command{Commands: []*cli.Command{tt.cmd()}}
				if err := root.Run(ctx, tt.args); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			})
		})
	}
}

func TestTradeSentimentAggregatesLeveragedFlow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Trades/GetTrades" {
			t.Errorf("expected path /Trades/GetTrades, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[
			{"Date":"/Date(1745193600000)/","Ticker":"SH","Sector":"X Bear","Dollars":886000000},
			{"Date":"/Date(1745193600000)/","Ticker":"TQQQ","Sector":"X Bull","Dollars":68000000},
			{"Date":"/Date(1745280000000)/","Ticker":"SQQQ","Industry":"X Bear","Dollars":51000000},
			{"Date":"/Date(1745280000000)/","Ticker":"SOXL","Industry":"X Bull","Dollars":102000000},
			{"Date":"/Date(1745280000000)/","Ticker":"AAPL","Sector":"Technology","Dollars":999999999}
		]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	output := captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{newTradeSentimentCommand()}}
		if err := root.Run(ctx, []string{
			"app", "sentiment",
			"--start-date", "2025-04-21",
			"--end-date", "2025-04-22",
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var got models.TradeSentiment
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatalf("unmarshal sentiment output: %v", err)
	}
	if got.DateRange.Start != "2025-04-21" || got.DateRange.End != "2025-04-22" {
		t.Fatalf("unexpected date range: %#v", got.DateRange)
	}
	if len(got.Daily) != 2 {
		t.Fatalf("expected two daily rows, got %d", len(got.Daily))
	}

	monday := got.Daily[0]
	if monday.Date != "2025-04-21" {
		t.Fatalf("expected first day 2025-04-21, got %s", monday.Date)
	}
	if monday.Bear.Trades != 1 || monday.Bear.Dollars != 886000000 {
		t.Fatalf("unexpected Monday bear summary: %#v", monday.Bear)
	}
	if monday.Bull.Trades != 1 || monday.Bull.Dollars != 68000000 {
		t.Fatalf("unexpected Monday bull summary: %#v", monday.Bull)
	}
	if monday.Ratio == nil || *monday.Ratio < 0.076 || *monday.Ratio > 0.077 {
		t.Fatalf("unexpected Monday ratio: %v", monday.Ratio)
	}
	if monday.Signal != models.TradeSentimentExtremeBear {
		t.Fatalf("expected Monday extreme bear signal, got %s", monday.Signal)
	}
	if strings.Join(monday.Bear.TopTickers, ",") != "SH" {
		t.Fatalf("unexpected Monday bear top tickers: %#v", monday.Bear.TopTickers)
	}
	if strings.Join(monday.Bull.TopTickers, ",") != "TQQQ" {
		t.Fatalf("unexpected Monday bull top tickers: %#v", monday.Bull.TopTickers)
	}

	tuesday := got.Daily[1]
	if tuesday.Ratio == nil || *tuesday.Ratio != 2.0 {
		t.Fatalf("unexpected Tuesday ratio: %v", tuesday.Ratio)
	}
	if tuesday.Signal != models.TradeSentimentNeutral {
		t.Fatalf("expected Tuesday neutral signal at 2.0 boundary, got %s", tuesday.Signal)
	}
	if got.Totals.Bear.Trades != 2 || got.Totals.Bull.Trades != 2 {
		t.Fatalf("unexpected totals: %#v", got.Totals)
	}
}

func TestTradeSentimentUsesCombinedLeverageFilter(t *testing.T) {
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		fmt.Fprint(w, dataTablesJSON(`[]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{newTradeSentimentCommand()}}
		if err := root.Run(ctx, []string{
			"app", "sentiment",
			"--start-date", "2025-04-21",
			"--end-date", "2025-04-25",
			"--min-dollars", "7000000",
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	for _, want := range []string{"SectorIndustry=X+B", "VCD=97", "MinDollars=7000000", "length=1000"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected request body to contain %q, got %s", want, body)
		}
	}
}

func TestTradeSentimentRejectsMalformedPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, dataTablesJSON(`{"not":"a trade array"}`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	root := &cli.Command{Commands: []*cli.Command{newTradeSentimentCommand()}}
	err := root.Run(ctx, []string{
		"app", "sentiment",
		"--start-date", "2025-04-21",
		"--end-date", "2025-04-25",
	})
	assertErrContains(t, err, "query trade sentiment: decode response")
}

func TestTradeSentimentTickerFallbackClassifiesCWEBAsBull(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, dataTablesJSON(`[
			{"Date":"/Date(1745193600000)/","Ticker":"CWEB","Sector":"ETF","Industry":"China","Name":"ProShares Ultra China ETF","Dollars":25000000},
			{"Date":"/Date(1745193600000)/","Ticker":"SH","Sector":"ETF","Industry":"Index","Name":"Short S&P 500 ETF","Dollars":5000000}
		]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	output := captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{newTradeSentimentCommand()}}
		if err := root.Run(ctx, []string{
			"app", "sentiment",
			"--start-date", "2025-04-21",
			"--end-date", "2025-04-21",
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var got models.TradeSentiment
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatalf("unmarshal sentiment output: %v", err)
	}
	if got.Totals.Bull.Trades != 1 || got.Totals.Bull.Dollars != 25000000 {
		t.Fatalf("expected CWEB to be classified as bull, got %#v", got.Totals.Bull)
	}
	if strings.Join(got.Totals.Bull.TopTickers, ",") != "CWEB" {
		t.Fatalf("unexpected bull top tickers: %#v", got.Totals.Bull.TopTickers)
	}
}

func TestTradeSentimentNoBearFlowUsesNullRatioAndExtremeBull(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, dataTablesJSON(`[
			{"Date":"/Date(1745193600000)/","Ticker":"TQQQ","Sector":"X Bull","Dollars":75000000}
		]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	output := captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{newTradeSentimentCommand()}}
		if err := root.Run(ctx, []string{
			"app", "sentiment",
			"--start-date", "2025-04-21",
			"--end-date", "2025-04-21",
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var got models.TradeSentiment
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatalf("unmarshal sentiment output: %v", err)
	}
	if got.Totals.Ratio != nil {
		t.Fatalf("expected nil total ratio with no bear flow, got %v", got.Totals.Ratio)
	}
	if got.Totals.Signal != models.TradeSentimentExtremeBull {
		t.Fatalf("expected extreme bull signal, got %s", got.Totals.Signal)
	}
}

func TestTradeListFieldsFiltersOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Trades/GetTrades" {
			t.Errorf("expected path /Trades/GetTrades, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[{"Ticker":"SPY","Dollars":26025000,"Volume":50000}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	output := captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{newTradeListCommand()}}
		if err := root.Run(ctx, []string{
			"app", "list",
			"--start-date", "2025-01-01",
			"--end-date", "2025-01-31",
			"--fields", "Ticker,Dollars",
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var got []map[string]json.RawMessage
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatalf("unmarshal filtered output: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected one row, got %d", len(got))
	}
	if _, ok := got[0]["Ticker"]; !ok {
		t.Fatal("expected Ticker field")
	}
	if _, ok := got[0]["Dollars"]; !ok {
		t.Fatal("expected Dollars field")
	}
	if _, ok := got[0]["Volume"]; ok {
		t.Fatal("did not expect Volume field")
	}
}

func TestTradeListFieldsRejectsInvalidField(t *testing.T) {
	ctx := contextWithTestClient("http://127.0.0.1")
	root := &cli.Command{Commands: []*cli.Command{newTradeListCommand()}}
	err := root.Run(ctx, []string{
		"app", "list",
		"--start-date", "2025-01-01",
		"--end-date", "2025-01-31",
		"--fields", "Ticker,NotAField",
	})
	assertErrContains(t, err, "invalid field \"NotAField\"")
	assertErrContains(t, err, "Ticker")
}

func TestTradeListInvalidFormatWithWatchlistDoesNotQueryAPI(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		fmt.Fprint(w, dataTablesJSON(`[]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	root := &cli.Command{Commands: []*cli.Command{newTradeListCommand()}}
	err := root.Run(ctx, []string{
		"app", "list",
		"--start-date", "2025-01-01",
		"--end-date", "2025-01-31",
		"--watchlist", "Core",
		"--format", "table",
	})
	assertErrContains(t, err, "valid formats: json,csv,tsv")
	if requestCount != 0 {
		t.Fatalf("expected invalid format to fail before watchlist/API query, got %d requests", requestCount)
	}
}

func TestTradeListPresetIncludesDefaults(t *testing.T) {
	// Regression test for https://github.com/major/volumeleaders-agent/issues/8
	// Presets must include CLI flag defaults for params the API requires
	// (MaxVolume, MaxPrice, session toggles, etc.).
	requiredParams := []string{
		"MaxVolume", "MinVolume",
		"MaxPrice", "MinPrice",
		"MaxDollars", "MinDollars",
		"IncludePremarket", "IncludeRTH", "IncludeAH",
		"IncludeOpening", "IncludeClosing",
		"IncludePhantom", "IncludeOffsetting",
		"SecurityTypeKey", "MarketCap",
		"DarkPools", "Sweeps", "LatePrints",
		"SignaturePrints", "EvenShared",
		"TradeRankSnapshot",
	}

	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{newTradeListCommand()}}
		if err := root.Run(ctx, []string{
			"app", "list",
			"--preset", "Top-100 Rank",
			"--start-date", "2025-04-01",
			"--end-date", "2025-04-25",
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	for _, param := range requiredParams {
		if !strings.Contains(body, param+"=") {
			t.Errorf("preset request missing required param %q", param)
		}
	}
}

func TestTradeListPresetOverridesDefaults(t *testing.T) {
	// The "Top-100 Rank" preset sets TradeRank=100 and MaxDollars=100000000000,
	// overriding CLI defaults of TradeRank=-1 and MaxDollars=30000000000.
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{newTradeListCommand()}}
		if err := root.Run(ctx, []string{
			"app", "list",
			"--preset", "Top-100 Rank",
			"--start-date", "2025-04-01",
			"--end-date", "2025-04-25",
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Preset value should override CLI default.
	if !strings.Contains(body, "TradeRank=100") {
		t.Error("preset TradeRank=100 not found in request body")
	}
	if !strings.Contains(body, "MaxDollars=100000000000") {
		t.Error("preset MaxDollars=100000000000 not found in request body")
	}
}

func TestTradeListExplicitFlagOverridesPreset(t *testing.T) {
	// An explicit CLI flag should override both the default and the preset.
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{newTradeListCommand()}}
		if err := root.Run(ctx, []string{
			"app", "list",
			"--preset", "Top-100 Rank",
			"--start-date", "2025-04-01",
			"--end-date", "2025-04-25",
			"--dark-pools", "1",
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Explicit --dark-pools=1 should override the preset/default.
	if !strings.Contains(body, "DarkPools=1") {
		t.Error("explicit --dark-pools=1 not found in request body")
	}
	// Preset value should still be present.
	if !strings.Contains(body, "TradeRank=100") {
		t.Error("preset TradeRank=100 should still be present")
	}
}

func TestTradeRunFunctionServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	root := &cli.Command{Commands: []*cli.Command{newTradeListCommand()}}
	err := root.Run(ctx, []string{"app", "list", "--start-date", "2025-01-01", "--end-date", "2025-01-31"})
	assertErrContains(t, err, "query trades")
}
