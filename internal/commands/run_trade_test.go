package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestTradeListSummaryDefaultsToTickerGrouping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Trades/GetTrades" {
			t.Errorf("expected path /Trades/GetTrades, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[
			{"Date":"/Date(1745193600000)/","Ticker":"SPY","Dollars":100,"DollarsMultiplier":2,"CumulativeDistribution":0.8,"DarkPool":1,"Sweep":0},
			{"Date":"/Date(1745193600000)/","Ticker":"SPY","Dollars":300,"DollarsMultiplier":4,"CumulativeDistribution":1,"DarkPool":0,"Sweep":1},
			{"Date":"/Date(1745280000000)/","Ticker":"QQQ","Dollars":200,"DollarsMultiplier":6,"CumulativeDistribution":0.5,"DarkPool":1,"Sweep":1}
		]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	output := captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{newTradeListCommand()}}
		if err := root.Run(ctx, []string{
			"app", "list",
			"--start-date", "2025-04-21",
			"--end-date", "2025-04-22",
			"--summary",
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var got models.TradeSummary
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatalf("unmarshal summary output: %v", err)
	}
	if got.TotalTrades != 3 {
		t.Fatalf("total trades = %d, want 3", got.TotalTrades)
	}
	if got.TotalDollars != 600 {
		t.Fatalf("total dollars = %v, want 600", got.TotalDollars)
	}
	if got.DateRange.Start != "2025-04-21" || got.DateRange.End != "2025-04-22" {
		t.Fatalf("date range = %#v, want 2025-04-21 to 2025-04-22", got.DateRange)
	}
	if got.ByDay != nil || got.ByTickerDay != nil {
		t.Fatalf("unexpected non-ticker groups: byDay=%#v byTickerDay=%#v", got.ByDay, got.ByTickerDay)
	}

	spy := got.ByTicker["SPY"]
	if spy.Trades != 2 || spy.Dollars != 400 {
		t.Fatalf("SPY summary = %#v, want 2 trades and 400 dollars", spy)
	}
	if spy.AvgDollarsMultiplier != 3 {
		t.Fatalf("SPY avg dollars multiplier = %v, want 3", spy.AvgDollarsMultiplier)
	}
	if spy.PctDarkPool != 50 || spy.PctSweep != 50 {
		t.Fatalf("SPY dark/sweep pct = %v/%v, want 50/50", spy.PctDarkPool, spy.PctSweep)
	}
	if spy.AvgCumulativeDistribution != 0.9 {
		t.Fatalf("SPY avg cumulative distribution = %v, want 0.9", spy.AvgCumulativeDistribution)
	}
}

func TestTradeListSummaryGroupsByTickerDay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Trades/GetTrades" {
			t.Errorf("expected path /Trades/GetTrades, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[
			{"Date":"/Date(1745193600000)/","Ticker":"SPY","Dollars":100,"DollarsMultiplier":2,"CumulativeDistribution":0.8,"DarkPool":1,"Sweep":0},
			{"Date":"/Date(1745280000000)/","Ticker":"SPY","Dollars":300,"DollarsMultiplier":4,"CumulativeDistribution":1,"DarkPool":0,"Sweep":1}
		]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	output := captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{newTradeListCommand()}}
		if err := root.Run(ctx, []string{
			"app", "list",
			"--start-date", "2025-04-21",
			"--end-date", "2025-04-22",
			"--summary",
			"--group-by", "ticker,day",
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var got models.TradeSummary
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatalf("unmarshal summary output: %v", err)
	}
	if got.ByTicker != nil || got.ByDay != nil {
		t.Fatalf("unexpected non-cross-tab groups: byTicker=%#v byDay=%#v", got.ByTicker, got.ByDay)
	}
	if got.ByTickerDay["SPY|2025-04-21"].Dollars != 100 {
		t.Fatalf("SPY first day dollars = %v, want 100", got.ByTickerDay["SPY|2025-04-21"].Dollars)
	}
	if got.ByTickerDay["SPY|2025-04-22"].Dollars != 300 {
		t.Fatalf("SPY second day dollars = %v, want 300", got.ByTickerDay["SPY|2025-04-22"].Dollars)
	}
}

func TestTradeListSummaryGroupsByDay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Trades/GetTrades" {
			t.Errorf("expected path /Trades/GetTrades, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[
			{"Date":"/Date(1745193600000)/","Ticker":"SPY","Dollars":100,"DollarsMultiplier":2,"CumulativeDistribution":0.8,"DarkPool":1,"Sweep":0},
			{"Date":"/Date(1745280000000)/","Ticker":"QQQ","Dollars":300,"DollarsMultiplier":4,"CumulativeDistribution":1,"DarkPool":0,"Sweep":1}
		]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	output := captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{newTradeListCommand()}}
		if err := root.Run(ctx, []string{
			"app", "list",
			"--start-date", "2025-04-21",
			"--end-date", "2025-04-22",
			"--summary",
			"--group-by", "day",
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var got models.TradeSummary
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatalf("unmarshal summary output: %v", err)
	}
	if got.ByTicker != nil || got.ByTickerDay != nil {
		t.Fatalf("unexpected non-day groups: byTicker=%#v byTickerDay=%#v", got.ByTicker, got.ByTickerDay)
	}
	if got.ByDay["2025-04-21"].Dollars != 100 {
		t.Fatalf("first day dollars = %v, want 100", got.ByDay["2025-04-21"].Dollars)
	}
	if got.ByDay["2025-04-22"].Dollars != 300 {
		t.Fatalf("second day dollars = %v, want 300", got.ByDay["2025-04-22"].Dollars)
	}
}

func TestTradeListSummaryAllResultsUsesPagination(t *testing.T) {
	totalRecords := paginationPageSize + 1
	var requestCount int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		body, _ := io.ReadAll(r.Body)
		params, _ := url.ParseQuery(string(body))

		if params.Get("start") == "0" {
			items := make([]string, paginationPageSize)
			for i := range items {
				items[i] = `{"Date":"/Date(1745193600000)/","Ticker":"SPY","Dollars":1}`
			}
			fmt.Fprint(w, dataTablesJSONPage("["+strings.Join(items, ",")+"]", totalRecords))
			return
		}

		fmt.Fprint(w, dataTablesJSONPage(`[{"Date":"/Date(1745193600000)/","Ticker":"SPY","Dollars":2}]`, totalRecords))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	output := captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{newTradeListCommand()}}
		if err := root.Run(ctx, []string{
			"app", "list",
			"--start-date", "2025-04-21",
			"--end-date", "2025-04-21",
			"--summary",
			"--length", "-1",
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var got models.TradeSummary
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatalf("unmarshal summary output: %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("request count = %d, want 2", requestCount)
	}
	if got.TotalTrades != totalRecords {
		t.Fatalf("total trades = %d, want %d", got.TotalTrades, totalRecords)
	}
	wantDollars := float64(paginationPageSize + 2)
	if got.TotalDollars != wantDollars {
		t.Fatalf("total dollars = %v, want %v", got.TotalDollars, wantDollars)
	}
}

func TestTradeListSummaryAllResultsRejectsMalformedPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, dataTablesJSONPage(`{}`, 1))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	root := &cli.Command{Commands: []*cli.Command{newTradeListCommand()}}
	err := root.Run(ctx, []string{
		"app", "list",
		"--start-date", "2025-04-21",
		"--end-date", "2025-04-21",
		"--summary",
		"--length", "-1",
	})
	assertErrContains(t, err, "query trades: decode response")
}

func TestTradeListSummaryRejectsInvalidGroupBy(t *testing.T) {
	ctx := contextWithTestClient("http://127.0.0.1")
	root := &cli.Command{Commands: []*cli.Command{newTradeListCommand()}}
	err := root.Run(ctx, []string{
		"app", "list",
		"--start-date", "2025-01-01",
		"--end-date", "2025-01-31",
		"--summary",
		"--group-by", "sector",
	})
	assertErrContains(t, err, "invalid group-by \"sector\"")
}

func TestTradeListSummaryRejectsFields(t *testing.T) {
	ctx := contextWithTestClient("http://127.0.0.1")
	root := &cli.Command{Commands: []*cli.Command{newTradeListCommand()}}
	err := root.Run(ctx, []string{
		"app", "list",
		"--start-date", "2025-01-01",
		"--end-date", "2025-01-31",
		"--summary",
		"--fields", "Ticker,Dollars",
	})
	assertErrContains(t, err, "--fields cannot be used with --summary")
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
