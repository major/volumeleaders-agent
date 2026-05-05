package trade

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/major/volumeleaders-agent/internal/cli/testutil"
	"github.com/major/volumeleaders-agent/internal/models"
)

// minimalJSONArray returns a JSON array of n empty objects, usable as a
// DataTables data payload for pagination tests that need page-sized responses.
func minimalJSONArray(n int) string {
	if n == 0 {
		return "[]"
	}
	return "[" + strings.Repeat("{},", n-1) + "{}]"
}

func TestNewCmdRegistersNineRunESubcommands(t *testing.T) {
	t.Parallel()

	cmd := NewCmd()
	if len(cmd.Commands()) != 9 {
		t.Fatalf("subcommand count = %d, want 9", len(cmd.Commands()))
	}
	for _, subcmd := range cmd.Commands() {
		if subcmd.RunE == nil {
			t.Fatalf("subcommand %q missing RunE", subcmd.Name())
		}
	}
}

func TestApplyExplicitFlagsPreservesPresetValuesUnlessChanged(t *testing.T) {
	t.Parallel()

	preset, err := findPreset("Top-100 Rank")
	if err != nil {
		t.Fatalf("find preset: %v", err)
	}
	filters := maps.Clone(preset.filters)
	cmd := newTradeListCommand()
	if err := cmd.Flags().Set("dark-pools", "1"); err != nil {
		t.Fatalf("set dark-pools: %v", err)
	}

	applyExplicitFlags(cmd, filters)

	if got := filters["DarkPools"]; got != "1" {
		t.Fatalf("DarkPools = %q, want 1", got)
	}
	if got := filters["TradeRank"]; got != "100" {
		t.Fatalf("TradeRank = %q, want preset value 100", got)
	}
	if got := filters["MaxDollars"]; got != "100000000000" {
		t.Fatalf("MaxDollars = %q, want preset value 100000000000", got)
	}
}

func TestTradeListUnknownPresetReturnsUsefulError(t *testing.T) {
	t.Parallel()

	cmd := newTradeListCommand()
	_, _, err := testutil.ExecuteCommand(t, cmd, t.Context(), "--preset", "Missing")

	testutil.AssertErrContains(t, err, "preset \"Missing\" not found")
}

func TestChangedTrueWhenFlagSetToDefaultValue(t *testing.T) {
	t.Parallel()

	cmd := newTradeSentimentCommand()
	if err := cmd.Flags().Set("vcd", "97"); err != nil {
		t.Fatalf("set vcd: %v", err)
	}
	if !cmd.Flags().Changed("vcd") {
		t.Fatal("Changed(vcd) = false, want true for explicit default value")
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

func TestTradeListRejectsUserSelectedLength(t *testing.T) {
	t.Parallel()

	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, "http://127.0.0.1")
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "list", "--start-date", "2025-04-21", "--end-date", "2025-04-21", "--length", "100")
	testutil.AssertErrContains(t, err, "unknown flag: --length")
}

func TestTradeListFetchesBrowserSizedPages(t *testing.T) {
	t.Parallel()

	var gotLengths []string
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}
		params, _ := url.ParseQuery(string(body))
		gotLengths = append(gotLengths, params.Get("length"))
		requestCount++
		if requestCount == 1 {
			// Full page signals more data available.
			_, _ = w.Write([]byte(testutil.DataTablesJSONPage(minimalJSONArray(tradeBrowserPageLength), tradeBrowserPageLength+50)))
		} else {
			// Short page ends pagination.
			_, _ = w.Write([]byte(testutil.DataTablesJSONPage(minimalJSONArray(50), tradeBrowserPageLength+50)))
		}
	}))
	t.Cleanup(server.Close)

	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, server.URL)
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "list", "--start-date", "2025-04-21", "--end-date", "2025-04-21")
	testutil.AssertErrContains(t, err, "")
	if !slices.Equal(gotLengths, []string{"100", "100"}) {
		t.Fatalf("lengths = %v, want [100 100]", gotLengths)
	}
}

func TestTradeDashboardFetchesChartOptimizedSections(t *testing.T) {
	t.Parallel()

	requestsByPath := make(map[string]url.Values)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}
		params, _ := url.ParseQuery(string(body))
		requestsByPath[r.URL.Path] = params
		switch r.URL.Path {
		case "/Trades/GetTrades":
			_, _ = w.Write([]byte(testutil.DataTablesJSON(`[{"Ticker":"IGV","Dollars":1000,"Volume":10}]`)))
		case "/TradeClusters/GetTradeClusters":
			_, _ = w.Write([]byte(testutil.DataTablesJSON(`[{"Ticker":"IGV","Dollars":2000,"Volume":20,"TradeCount":2}]`)))
		case "/Chart0/GetTradeLevels":
			_, _ = w.Write([]byte(testutil.DataTablesJSON(`[{"Price":101.5,"Dollars":3000,"Volume":30,"Trades":3}]`)))
		case "/TradeClusterBombs/GetTradeClusterBombs":
			_, _ = w.Write([]byte(testutil.DataTablesJSON(`[{"Ticker":"IGV","Dollars":4000,"Volume":40,"TradeCount":4,"TradeClusterBombRank":1}]`)))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, server.URL)
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "dashboard", "IGV", "--start-date", "2025-05-04", "--end-date", "2026-05-04")
	testutil.AssertErrContains(t, err, "")

	var dashboard models.TradeDashboard
	if err := json.Unmarshal([]byte(stdout), &dashboard); err != nil {
		t.Fatalf("failed to unmarshal dashboard: %v\nstdout=%s", err, stdout)
	}
	if dashboard.Ticker != "IGV" || dashboard.Count != tradeDashboardDefaultCount {
		t.Fatalf("dashboard metadata = (%q, %d), want (IGV, %d)", dashboard.Ticker, dashboard.Count, tradeDashboardDefaultCount)
	}
	if len(dashboard.Trades) != 1 || len(dashboard.Clusters) != 1 || len(dashboard.Levels) != 1 || len(dashboard.ClusterBombs) != 1 {
		t.Fatalf("dashboard section lengths = trades:%d clusters:%d levels:%d bombs:%d, want all 1", len(dashboard.Trades), len(dashboard.Clusters), len(dashboard.Levels), len(dashboard.ClusterBombs))
	}

	for _, path := range []string{"/Trades/GetTrades", "/TradeClusters/GetTradeClusters", "/Chart0/GetTradeLevels", "/TradeClusterBombs/GetTradeClusterBombs"} {
		params, ok := requestsByPath[path]
		if !ok {
			t.Fatalf("missing dashboard request for %s", path)
		}
		if got := params.Get("StartDate"); got != "2025-05-04" {
			t.Fatalf("%s StartDate = %q, want 2025-05-04", path, got)
		}
		if got := params.Get("EndDate"); got != "2026-05-04" {
			t.Fatalf("%s EndDate = %q, want 2026-05-04", path, got)
		}
	}
	if got := requestsByPath["/Trades/GetTrades"].Get("length"); got != "10" {
		t.Fatalf("trades length = %q, want 10", got)
	}
	if got := requestsByPath["/Trades/GetTrades"].Get("VCD"); got != "0" {
		t.Fatalf("trades VCD = %q, want 0", got)
	}
	if got := requestsByPath["/Trades/GetTrades"].Get("RelativeSize"); got != "0" {
		t.Fatalf("trades RelativeSize = %q, want 0", got)
	}
	if got := requestsByPath["/Trades/GetTrades"].Get("Sort"); got != "Dollars" {
		t.Fatalf("trades Sort = %q, want Dollars", got)
	}
	if got := requestsByPath["/Chart0/GetTradeLevels"].Get("Ticker"); got != "IGV" {
		t.Fatalf("levels Ticker = %q, want IGV", got)
	}
	if got := requestsByPath["/Chart0/GetTradeLevels"].Get("Levels"); got != "10" {
		t.Fatalf("levels Levels = %q, want 10", got)
	}
	if got := requestsByPath["/Chart0/GetTradeLevels"].Get("length"); got != "-1" {
		t.Fatalf("levels length = %q, want -1", got)
	}
}

func TestTradeLevelsUsesChartOptimizedEndpoint(t *testing.T) {
	t.Parallel()

	var got url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Chart0/GetTradeLevels" {
			t.Errorf("path = %q, want /Chart0/GetTradeLevels", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}
		got, _ = url.ParseQuery(string(body))
		_, _ = w.Write([]byte(testutil.DataTablesJSON(`[
			{"Price":113.7,"Dollars":1130011877.91,"Volume":9936998,"Trades":27,"RelativeSize":52.37,"CumulativeDistribution":0.9995,"TradeLevelRank":2},
			{"Price":117.8,"Dollars":735913002.62,"Volume":6249716,"Trades":20,"RelativeSize":34.1,"CumulativeDistribution":0.9991,"TradeLevelRank":3}
		]`)))
	}))
	t.Cleanup(server.Close)

	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, server.URL)
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "levels", "GNRC", "--start-date", "2025-05-05", "--end-date", "2026-05-05")
	testutil.AssertErrContains(t, err, "")

	var levels []models.TradeLevelRow
	if err := json.Unmarshal([]byte(stdout), &levels); err != nil {
		t.Fatalf("failed to unmarshal levels: %v\nstdout=%s", err, stdout)
	}
	if len(levels) != 2 {
		t.Fatalf("len(levels) = %d, want 2", len(levels))
	}
	if got.Get("Ticker") != "GNRC" {
		t.Fatalf("Ticker = %q, want GNRC", got.Get("Ticker"))
	}
	if got.Get("Levels") != "10" {
		t.Fatalf("Levels = %q, want 10", got.Get("Levels"))
	}
	if got.Get("length") != "-1" {
		t.Fatalf("length = %q, want -1", got.Get("length"))
	}
	if got.Get("order[0][name]") != "Price" {
		t.Fatalf("order name = %q, want Price", got.Get("order[0][name]"))
	}
	if got.Get("MinDollars") != "" {
		t.Fatalf("MinDollars = %q, want omitted", got.Get("MinDollars"))
	}
}

func TestTradeLevelsSupportsFullFieldSelection(t *testing.T) {
	t.Parallel()

	server := tradeLevelsServer(t)
	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, server.URL)
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "levels", "GNRC", "--start-date", "2025-05-05", "--end-date", "2026-05-05", "--fields", "all")
	testutil.AssertErrContains(t, err, "")

	var levels []map[string]any
	if err := json.Unmarshal([]byte(stdout), &levels); err != nil {
		t.Fatalf("failed to unmarshal full levels: %v\nstdout=%s", err, stdout)
	}
	if len(levels) != 1 {
		t.Fatalf("len(levels) = %d, want 1", len(levels))
	}
	for _, field := range []string{"Ticker", "Name", "Dates"} {
		if _, ok := levels[0][field]; !ok {
			t.Fatalf("full level output missing field %q: %#v", field, levels[0])
		}
	}
}

func TestTradeLevelsSupportsCSVOutput(t *testing.T) {
	t.Parallel()

	server := tradeLevelsServer(t)
	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, server.URL)
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "levels", "GNRC", "--start-date", "2025-05-05", "--end-date", "2026-05-05", "--format", "csv")
	testutil.AssertErrContains(t, err, "")

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 2 {
		t.Fatalf("CSV line count = %d, want 2\nstdout=%s", len(lines), stdout)
	}
	wantHeader := "Ticker,Name,Price,Dollars,Volume,Trades,RelativeSize,CumulativeDistribution,TradeLevelRank,MinDate,MaxDate,Dates"
	if lines[0] != wantHeader {
		t.Fatalf("CSV header = %q, want %q", lines[0], wantHeader)
	}
	if !strings.Contains(lines[1], "GNRC") || !strings.Contains(lines[1], "113.7") {
		t.Fatalf("CSV row missing expected level values: %q", lines[1])
	}
}

func TestTradeLevelsRejectsStaleFilterFlags(t *testing.T) {
	t.Parallel()

	for _, flag := range []string{"--vcd", "--relative-size", "--trade-level-rank", "--min-dollars"} {
		t.Run(flag, func(t *testing.T) {
			t.Parallel()

			cmd := NewCmd()
			ctx := testutil.ContextWithTestClient(t, "http://127.0.0.1")
			_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "levels", "GNRC", flag, "1")
			testutil.AssertErrContains(t, err, "unknown flag: "+flag)
		})
	}
}

func tradeLevelsServer(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Chart0/GetTradeLevels" {
			t.Errorf("path = %q, want /Chart0/GetTradeLevels", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(testutil.DataTablesJSON(`[
			{"Ticker":"GNRC","Name":"Generac Holdings Inc.","Price":113.7,"Dollars":1130011877.91,"Volume":9936998,"Trades":27,"RelativeSize":52.37,"CumulativeDistribution":0.9995,"TradeLevelRank":2,"Dates":"2025-05-05"}
		]`)))
	}))
	t.Cleanup(server.Close)
	return server
}

func TestTradeListSelectedFieldsFetchBrowserSizedPages(t *testing.T) {
	t.Parallel()

	var gotLengths []string
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}
		params, _ := url.ParseQuery(string(body))
		gotLengths = append(gotLengths, params.Get("length"))
		requestCount++
		if requestCount == 1 {
			_, _ = w.Write([]byte(testutil.DataTablesJSONPage(minimalJSONArray(tradeBrowserPageLength), tradeBrowserPageLength+50)))
		} else {
			_, _ = w.Write([]byte(testutil.DataTablesJSONPage(minimalJSONArray(50), tradeBrowserPageLength+50)))
		}
	}))
	t.Cleanup(server.Close)

	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, server.URL)
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "list", "--start-date", "2025-04-21", "--end-date", "2025-04-21", "--fields", "Ticker")
	testutil.AssertErrContains(t, err, "")
	if !slices.Equal(gotLengths, []string{"100", "100"}) {
		t.Fatalf("lengths = %v, want [100 100]", gotLengths)
	}
}

func TestTradeListMultiDayWithoutTickersKeepsFullPagination(t *testing.T) {
	t.Parallel()

	var gotLengths []string
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}
		params, _ := url.ParseQuery(string(body))
		gotLengths = append(gotLengths, params.Get("length"))
		requestCount++
		if requestCount == 1 {
			_, _ = w.Write([]byte(testutil.DataTablesJSONPage(minimalJSONArray(tradeBrowserPageLength), tradeBrowserPageLength+50)))
		} else {
			_, _ = w.Write([]byte(testutil.DataTablesJSONPage(minimalJSONArray(50), tradeBrowserPageLength+50)))
		}
	}))
	t.Cleanup(server.Close)

	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, server.URL)
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "list", "--start-date", "2025-05-04", "--end-date", "2026-05-04")
	testutil.AssertErrContains(t, err, "")
	if !slices.Equal(gotLengths, []string{"100", "100"}) {
		t.Fatalf("lengths = %v, want [100 100]", gotLengths)
	}
}

func TestTradeListMultiDayPresetWithoutTickersKeepsFullPagination(t *testing.T) {
	t.Parallel()

	var gotLengths []string
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}
		params, _ := url.ParseQuery(string(body))
		gotLengths = append(gotLengths, params.Get("length"))
		requestCount++
		if requestCount == 1 {
			_, _ = w.Write([]byte(testutil.DataTablesJSONPage(minimalJSONArray(tradeBrowserPageLength), tradeBrowserPageLength+50)))
		} else {
			_, _ = w.Write([]byte(testutil.DataTablesJSONPage(minimalJSONArray(50), tradeBrowserPageLength+50)))
		}
	}))
	t.Cleanup(server.Close)

	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, server.URL)
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "list", "--preset", "Top-10 Rank", "--start-date", "2025-05-04", "--end-date", "2026-05-04")
	testutil.AssertErrContains(t, err, "")
	if !slices.Equal(gotLengths, []string{"100", "100"}) {
		t.Fatalf("lengths = %v, want [100 100]", gotLengths)
	}
}

func TestTradeListLongPeriodTickerUsesChartTopTenRequest(t *testing.T) {
	t.Parallel()

	var got url.Values
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}
		got, _ = url.ParseQuery(string(body))
		_, _ = w.Write([]byte(testutil.DataTablesJSONPage(`[{"Ticker":"AAPL","Dollars":1000000}]`, 4433)))
	}))
	t.Cleanup(server.Close)

	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, server.URL)
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "list", "AAPL", "--start-date", "2025-05-04", "--end-date", "2026-05-04")
	testutil.AssertErrContains(t, err, "")
	if requestCount != 1 {
		t.Fatalf("request count = %d, want 1", requestCount)
	}

	checks := map[string]string{
		"start":                     "0",
		"length":                    "10",
		"order[0][column]":          "0",
		"order[0][dir]":             "DESC",
		"order[0][name]":            "FullTimeString24",
		"search[value]":             "",
		"search[regex]":             "false",
		"columns[0][data]":          "FullTimeString24",
		"columns[0][name]":          "FullTimeString24",
		"columns[0][orderable]":     "false",
		"columns[1][data]":          "Volume",
		"columns[1][name]":          "Sh",
		"columns[3][data]":          "Dollars",
		"columns[3][name]":          "$$",
		"columns[6][data]":          "LastComparibleTradeDate",
		"columns[6][name]":          "Last Comp",
		"columns[6][search][value]": "",
		"columns[6][search][regex]": "false",
		"Tickers":                   "AAPL",
		"StartDate":                 "2025-05-04",
		"EndDate":                   "2026-05-04",
		"VCD":                       "0",
		"RelativeSize":              "0",
		"Sort":                      "Dollars",
	}
	for key, expected := range checks {
		if value := got.Get(key); value != expected {
			t.Errorf("%s = %q, want %q", key, value, expected)
		}
	}

	for _, key := range []string{"SecurityTypeKey", "EvenShared", "TradeRankSnapshot", "MarketCap"} {
		if _, ok := got[key]; ok {
			t.Errorf("%s should be omitted from long-period chart request", key)
		}
	}
}

func TestTradeListLongPeriodExplicitFiltersOverrideChartDefaults(t *testing.T) {
	t.Parallel()

	var got url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}
		got, _ = url.ParseQuery(string(body))
		_, _ = w.Write([]byte(testutil.DataTablesJSONPage(`[]`, 500)))
	}))
	t.Cleanup(server.Close)

	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, server.URL)
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "list", "AAPL", "--start-date", "2025-05-04", "--end-date", "2026-05-04", "--vcd", "97", "--relative-size", "5", "--security-type", "-1")
	testutil.AssertErrContains(t, err, "")

	checks := map[string]string{
		"length":          "10",
		"VCD":             "97",
		"RelativeSize":    "5",
		"SecurityTypeKey": "-1",
		"Sort":            "Dollars",
	}
	for key, expected := range checks {
		if value := got.Get(key); value != expected {
			t.Errorf("%s = %q, want %q", key, value, expected)
		}
	}
}

func TestTradeListSummaryKeepsFullPaginationForLongPeriodTicker(t *testing.T) {
	t.Parallel()

	var gotLengths []string
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}
		params, _ := url.ParseQuery(string(body))
		gotLengths = append(gotLengths, params.Get("length"))
		requestCount++
		if requestCount == 1 {
			_, _ = w.Write([]byte(testutil.DataTablesJSONPage(minimalJSONArray(tradeBrowserPageLength), tradeBrowserPageLength+50)))
		} else {
			_, _ = w.Write([]byte(testutil.DataTablesJSONPage(minimalJSONArray(50), tradeBrowserPageLength+50)))
		}
	}))
	t.Cleanup(server.Close)

	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, server.URL)
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "list", "AAPL", "--start-date", "2025-05-04", "--end-date", "2026-05-04", "--summary")
	testutil.AssertErrContains(t, err, "")
	if !slices.Equal(gotLengths, []string{"100", "100"}) {
		t.Fatalf("lengths = %v, want [100 100]", gotLengths)
	}
}

func TestTradeClusterCommandsRejectUserSelectedLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{name: "clusters", args: []string{"clusters", "AAPL", "--days", "7", "--length", "10"}},
		{name: "cluster bombs", args: []string{"cluster-bombs", "AAPL", "--days", "3", "--length", "10"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCmd()
			ctx := testutil.ContextWithTestClient(t, "http://127.0.0.1")
			_, _, err := testutil.ExecuteCommand(t, cmd, ctx, tt.args...)
			testutil.AssertErrContains(t, err, "unknown flag: --length")
		})
	}
}

func TestTradeClusterCommandsFetchBrowserSizedPages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		path string
	}{
		{name: "clusters", args: []string{"clusters", "AAPL", "--days", "1"}, path: "/TradeClusters/GetTradeClusters"},
		{name: "cluster-bombs", args: []string{"cluster-bombs", "AAPL", "--days", "1"}, path: "/TradeClusterBombs/GetTradeClusterBombs"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotLengths []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("failed to read request body: %v", err)
				}
				params, _ := url.ParseQuery(string(body))
				gotLengths = append(gotLengths, params.Get("length"))
				_, _ = w.Write([]byte(testutil.DataTablesJSONPage("[]", 0)))
			}))
			t.Cleanup(server.Close)

			cmd := NewCmd()
			ctx := testutil.ContextWithTestClient(t, server.URL)
			_, _, err := testutil.ExecuteCommand(t, cmd, ctx, tt.args...)
			testutil.AssertErrContains(t, err, "")
			if len(gotLengths) == 0 || gotLengths[0] != fmt.Sprintf("%d", tradeBrowserPageLength) {
				t.Fatalf("first request length = %v, want %d", gotLengths, tradeBrowserPageLength)
			}
		})
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
			want: "--trade-level-count must be one of 5, 10, 20, or 50 for trade level retrieval",
		},
		{
			name: "trade level rank below site minimum",
			args: []string{"AAPL", "--start-date", "2025-04-21", "--end-date", "2025-04-21", "--trade-level-rank", "4"},
			want: "--trade-level-rank must be 5 or higher for trade level touch retrieval",
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

func TestTradeLevelsRestrictsTradeLevelCountToSiteValues(t *testing.T) {
	t.Parallel()

	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, "http://127.0.0.1")
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "levels", "AAPL", "--trade-level-count", "6")
	testutil.AssertErrContains(t, err, "--trade-level-count must be one of 5, 10, 20, or 50 for trade level retrieval")
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

func TestTradeLevelTouchesDefaultsToRankFive(t *testing.T) {
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
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "level-touches", "AAPL", "--start-date", "2025-04-21", "--end-date", "2025-04-21")
	testutil.AssertErrContains(t, err, "")
	if got.Get("TradeLevelRank") != "5" {
		t.Fatalf("TradeLevelRank = %q, want 5", got.Get("TradeLevelRank"))
	}
}

func TestNewTradeListCommandCanBeConstructedWithoutServer(t *testing.T) {
	t.Parallel()

	cmd := newTradeListCommand()
	if cmd == nil {
		t.Fatal("expected trade list command")
	}
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
