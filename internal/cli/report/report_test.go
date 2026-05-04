package report

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"
	"time"

	"github.com/major/volumeleaders-agent/internal/cli/testutil"
	"github.com/major/volumeleaders-agent/internal/models"
)

func TestReportDefinitionsExposeApprovedCommands(t *testing.T) {
	t.Parallel()

	definitions := reportDefinitions()
	uses := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		uses = append(uses, definition.use)
	}

	want := []string{"top-100-rank", "top-10-rank", "dark-pool-sweeps", "disproportionately-large"}
	if !slices.Equal(uses, want) {
		t.Fatalf("report commands = %v, want %v", uses, want)
	}
}

func TestReportFiltersMatchObservedBrowserPresets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		filters   map[string]string
		mustMatch map[string]string
		mustOmit  []string
	}{
		{
			name:    "top 100 ranked",
			filters: topRankFilters("100"),
			mustMatch: map[string]string{
				"Conditions":        "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH",
				"MaxDollars":        "100000000000",
				"MinVolume":         "10000",
				"RelativeSize":      "0",
				"TradeCount":        "3",
				"TradeRank":         "100",
				"TradeRankSnapshot": "-1",
				"VCD":               "0",
			},
		},
		{
			name:    "top 10 ranked",
			filters: topRankFilters("10"),
			mustMatch: map[string]string{
				"MaxDollars": "100000000000",
				"TradeCount": "3",
				"TradeRank":  "10",
				"VCD":        "0",
			},
		},
		{
			name:    "dark pool sweeps",
			filters: darkPoolSweepFilters(),
			mustMatch: map[string]string{
				"DarkPools":        "1",
				"Sweeps":           "1",
				"SignaturePrints":  "0",
				"IncludeAH":        "0",
				"IncludeClosing":   "0",
				"IncludeOpening":   "0",
				"IncludePhantom":   "0",
				"IncludePremarket": "1",
				"IncludeRTH":       "1",
				"TradeRank":        "100",
			},
		},
		{
			name:    "disproportionately large",
			filters: disproportionatelyLargeFilters(),
			mustMatch: map[string]string{
				"Conditions":        "-1",
				"IncludeAH":         "1",
				"IncludeClosing":    "1",
				"IncludeOffsetting": "1",
				"IncludeOpening":    "1",
				"IncludePhantom":    "1",
				"IncludePremarket":  "1",
				"IncludeRTH":        "1",
				"MaxDollars":        "30000000000",
				"MinVolume":         "0",
				"RelativeSize":      "5",
				"TradeRank":         "-1",
				"VCD":               "0",
			},
			mustOmit: []string{"TradeCount"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for key, want := range tt.mustMatch {
				if got := tt.filters[key]; got != want {
					t.Fatalf("filter %s = %q, want %q in %v", key, got, want, tt.filters)
				}
			}
			for _, key := range tt.mustOmit {
				if _, ok := tt.filters[key]; ok {
					t.Fatalf("filter %s should be omitted from %v", key, tt.filters)
				}
			}
		})
	}
}

func TestParseReportSummaryGroupAcceptsAliases(t *testing.T) {
	t.Parallel()

	tests := map[reportSummaryGroup]reportSummaryGroup{
		"ticker":      reportSummaryGroupTicker,
		"day":         reportSummaryGroupDay,
		"ticker,day":  reportSummaryGroupTickerDay,
		"ticker, day": reportSummaryGroupTickerDay,
		"ticker day":  reportSummaryGroupTickerDay,
		"ticker-day":  reportSummaryGroupTickerDay,
	}
	for input, want := range tests {
		got, err := parseReportSummaryGroup(input)
		if err != nil {
			t.Fatalf("parseReportSummaryGroup(%q) returned error: %v", input, err)
		}
		if got != want {
			t.Fatalf("parseReportSummaryGroup(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestParseReportSummaryGroupRejectsInvalidValue(t *testing.T) {
	t.Parallel()

	_, err := parseReportSummaryGroup("sector")
	testutil.AssertErrContains(t, err, "invalid group-by")
}

func TestSummarizeReportTradesGroupsByTickerAndDay(t *testing.T) {
	t.Parallel()

	trades := []models.Trade{
		tradeFixture("AAPL", "2026-05-01", 100, 2, true, false, 40),
		tradeFixture("AAPL", "2026-05-01", 300, 4, false, true, 60),
		tradeFixture("MSFT", "2026-05-02", 600, 6, true, true, 90),
		{Dollars: 50, DollarsMultiplier: 1},
	}

	summary := summarizeReportTrades(trades, reportSummaryGroupTickerDay, "2026-05-01", "2026-05-02")
	if summary.TotalTrades != 4 || summary.TotalDollars != 1050 {
		t.Fatalf("summary totals = trades %d dollars %.2f", summary.TotalTrades, summary.TotalDollars)
	}
	if summary.DateRange.Start != "2026-05-01" || summary.DateRange.End != "2026-05-02" {
		t.Fatalf("date range = %#v", summary.DateRange)
	}

	aapl := summary.ByTickerDay["AAPL|2026-05-01"]
	if aapl.Trades != 2 || aapl.Dollars != 400 || aapl.AvgDollarsMultiplier != 3 || aapl.PctDarkPool != 50 || aapl.PctSweep != 50 || aapl.AvgCumulativeDistribution != 50 {
		t.Fatalf("AAPL group summary = %#v", aapl)
	}
	unknown := summary.ByTickerDay["unknown|unknown"]
	if unknown.Trades != 1 || unknown.Dollars != 50 {
		t.Fatalf("unknown group summary = %#v", unknown)
	}
}

func TestRunReportRejectsBroadMultiDayWithoutTickers(t *testing.T) {
	t.Parallel()

	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, "http://127.0.0.1")
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "top-100-rank", "--start-date", "2026-05-01", "--end-date", "2026-05-02")
	testutil.AssertErrContains(t, err, "broad report scans must use a single day")
}

func TestRunReportRejectsSummaryFieldAndFormatConflicts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "group by without summary", args: []string{"top-100-rank", "--group-by", "day"}, want: "--group-by only works with --summary"},
		{name: "fields with summary", args: []string{"top-100-rank", "--summary", "--fields", "Ticker"}, want: "--fields cannot be used with --summary"},
		{name: "csv with summary", args: []string{"top-100-rank", "--summary", "--format", "csv"}, want: "--format cannot be used with --summary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCmd()
			ctx := testutil.ContextWithTestClient(t, "http://127.0.0.1")
			_, _, err := testutil.ExecuteCommand(t, cmd, ctx, tt.args...)
			testutil.AssertErrContains(t, err, tt.want)
		})
	}
}

func TestRunReportSubmitsBrowserSizedPresetRequest(t *testing.T) {
	t.Parallel()

	var got url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Trades/GetTrades" {
			t.Errorf("path = %q, want /Trades/GetTrades", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		got, _ = url.ParseQuery(string(body))
		_, _ = w.Write([]byte(testutil.DataTablesJSON(`[{"Ticker":"AAPL","Dollars":1000,"TradeRank":10}]`)))
	}))
	t.Cleanup(server.Close)

	cmd := NewCmd()
	ctx := testutil.ContextWithTestClient(t, server.URL)
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "top-10-rank", "AAPL", "--days", "1")
	testutil.AssertErrContains(t, err, "")
	if stdout == "" {
		t.Fatal("expected JSON stdout")
	}

	want := map[string]string{
		"length":            "100",
		"order[0][column]":  "0",
		"order[0][dir]":     "desc",
		"Tickers":           "AAPL",
		"TradeRank":         "10",
		"TradeCount":        "3",
		"MaxDollars":        "100000000000",
		"Conditions":        "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH",
		"TradeRankSnapshot": "-1",
	}
	for key, value := range want {
		if got.Get(key) != value {
			t.Fatalf("request %s = %q, want %q in %v", key, got.Get(key), value, got)
		}
	}
}

func tradeFixture(ticker, day string, dollars, multiplier float64, darkPool, sweep bool, cumulativeDistribution float64) models.Trade {
	parsed, err := time.ParseInLocation("2006-01-02", day, time.UTC)
	if err != nil {
		panic(err)
	}
	return models.Trade{Ticker: ticker, Date: models.AspNetDate{Time: parsed, Valid: true}, Dollars: dollars, DollarsMultiplier: multiplier, DarkPool: models.FlexBool(darkPool), Sweep: models.FlexBool(sweep), CumulativeDistribution: cumulativeDistribution}
}
