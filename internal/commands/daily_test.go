package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/major/volumeleaders-agent/internal/models"
	cli "github.com/urfave/cli/v3"
)

func TestDailySummaryAggregatesInstitutionalActivity(t *testing.T) {
	seen := make(map[string]bool)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen[r.URL.Path] = true
		switch r.URL.Path {
		case "/InstitutionalVolume/GetInstitutionalVolume":
			fmt.Fprint(w, dataTablesJSON(`[
				{"Ticker":"NVDA","Sector":"Technology","Price":900,"Volume":1000,"TotalInstitutionalDollars":500000000,"TotalInstitutionalDollarsRank":1},
				{"Ticker":"SPY","Sector":"ETF","Price":500,"Volume":800,"TotalInstitutionalDollars":250000000,"TotalInstitutionalDollarsRank":2}
			]`))
		case "/TradeClusters/GetTradeClusters":
			fmt.Fprint(w, dataTablesJSON(`[
				{"Ticker":"NVDA","Sector":"Technology","Dollars":300000000,"DollarsMultiplier":6,"Volume":1000,"TradeCount":3,"TradeClusterRank":1,"CumulativeDistribution":0.99},
				{"Ticker":"NVDA","Sector":"Technology","Dollars":150000000,"DollarsMultiplier":8,"Volume":500,"TradeCount":2,"TradeClusterRank":2,"CumulativeDistribution":0.95},
				{"Ticker":"SPY","Sector":"ETF","Dollars":200000000,"DollarsMultiplier":4,"Volume":700,"TradeCount":1,"TradeClusterRank":3,"CumulativeDistribution":0.9}
			]`))
		case "/TradeClusterBombs/GetTradeClusterBombs":
			fmt.Fprint(w, dataTablesJSON(`[
				{"Ticker":"TSLA","Sector":"Consumer Cyclical","Dollars":125000000,"DollarsMultiplier":7,"Volume":400,"TradeCount":2,"TradeClusterBombRank":1,"CumulativeDistribution":0.98}
			]`))
		case "/TradeLevelTouches/GetTradeLevelTouches":
			fmt.Fprint(w, dataTablesJSON(`[
				{"Ticker":"AMD","Sector":"Technology","Price":120,"Dollars":90000000,"RelativeSize":12,"Volume":300,"Trades":2,"TradeLevelRank":1,"TradeLevelTouches":4},
				{"Ticker":"SPY","Sector":"ETF","Price":500,"Dollars":110000000,"RelativeSize":5,"Volume":200,"Trades":1,"TradeLevelRank":2,"TradeLevelTouches":2}
			]`))
		case "/Trades/GetTrades":
			body, _ := io.ReadAll(r.Body)
			params, _ := url.ParseQuery(string(body))
			if got := params.Get("length"); got != "50" {
				t.Fatalf("daily sentiment trade length = %q, want 50", got)
			}
			fmt.Fprint(w, dataTablesJSON(`[
				{"Date":"/Date(1777334400000)/","Ticker":"SH","Sector":"X Bear","Dollars":100000000},
				{"Date":"/Date(1777334400000)/","Ticker":"TQQQ","Sector":"X Bull","Dollars":200000000}
			]`))
		case "/ExecutiveSummary/GetExhaustionScores":
			fmt.Fprint(w, `{"DateKey":20260428,"ExhaustionScoreRank":4,"ExhaustionScoreRank30Day":6,"ExhaustionScoreRank90Day":12,"ExhaustionScoreRank365Day":20}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewDailyCommand()}}
		if err := root.Run(ctx, []string{"app", "daily", "summary", "--date", "2026-04-28", "--limit", "2"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	for _, path := range []string{
		"/InstitutionalVolume/GetInstitutionalVolume",
		"/TradeClusters/GetTradeClusters",
		"/TradeClusterBombs/GetTradeClusterBombs",
		"/TradeLevelTouches/GetTradeLevelTouches",
		"/Trades/GetTrades",
		"/ExecutiveSummary/GetExhaustionScores",
	} {
		if !seen[path] {
			t.Fatalf("expected request to %s", path)
		}
	}

	var got models.DailySummary
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatalf("unmarshal daily summary: %v", err)
	}
	if got.Date != "2026-04-28" {
		t.Fatalf("date = %q, want 2026-04-28", got.Date)
	}
	if got.InstitutionalVolume[0].Ticker != "NVDA" || got.InstitutionalVolume[0].InstitutionalDollars != 500000000 {
		t.Fatalf("expected NVDA top institutional volume, got %#v", got.InstitutionalVolume)
	}
	if len(got.Clusters.Top) != 3 {
		t.Fatalf("expected deduped cluster union with 3 rows, got %#v", got.Clusters.Top)
	}
	if got.Clusters.Top[0].Ticker != "NVDA" || !slices.Contains(got.Clusters.Top[0].TopBy, "dollars") || !slices.Contains(got.Clusters.Top[0].TopBy, "multiplier") {
		t.Fatalf("unexpected top dollar clusters: %#v", got.Clusters.Top)
	}
	if got.Clusters.Top[2].Ticker != "NVDA" || got.Clusters.Top[2].DollarsMultiplier != 8 || got.Clusters.Top[2].TopBy[0] != "multiplier" {
		t.Fatalf("unexpected top multiplier cluster union: %#v", got.Clusters.Top)
	}
	if len(got.Clusters.RepeatedTickers) != 1 || got.Clusters.RepeatedTickers[0].Ticker != "NVDA" || got.Clusters.RepeatedTickers[0].BestRank != 1 {
		t.Fatalf("unexpected repeated clusters: %#v", got.Clusters.RepeatedTickers)
	}
	if len(got.LevelTouches) != 2 || got.LevelTouches[0].Ticker != "AMD" || got.LevelTouches[1].Ticker != "SPY" {
		t.Fatalf("unexpected level touches: %#v", got.LevelTouches)
	}
	if !slices.Contains(got.LevelTouches[0].TopBy, "relative_size") || !slices.Contains(got.LevelTouches[0].TopBy, "dollars") || !slices.Contains(got.LevelTouches[1].TopBy, "relative_size") || !slices.Contains(got.LevelTouches[1].TopBy, "dollars") || got.LevelTouches[1].Touches != 2 {
		t.Fatalf("unexpected level touch top_by metadata: %#v", got.LevelTouches)
	}
	if got.LeveragedETFSentiment.Ratio == nil || *got.LeveragedETFSentiment.Ratio != 2 || got.LeveragedETFSentiment.BullDollars != 200000000 {
		t.Fatalf("unexpected leveraged ETF sentiment: %#v", got.LeveragedETFSentiment)
	}
	if got.MarketExhaustion.Rank != 4 || got.MarketExhaustion.Rank365D != 20 {
		t.Fatalf("unexpected exhaustion score: %#v", got.MarketExhaustion)
	}
}

func TestDailySummaryRejectsInvalidLimit(t *testing.T) {
	ctx := contextWithTestClient(t, "http://127.0.0.1")
	root := &cli.Command{Commands: []*cli.Command{NewDailyCommand()}}
	err := root.Run(ctx, []string{"app", "daily", "summary", "--date", "2026-04-28", "--limit", "0"})
	assertErrContains(t, err, "--limit must be greater than 0")
}

func TestDailySummaryOutputUsesRequestedSectionNames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ExecutiveSummary/GetExhaustionScores" {
			fmt.Fprint(w, `{}`)
			return
		}
		fmt.Fprint(w, dataTablesJSON(`[]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewDailyCommand()}}
		if err := root.Run(ctx, []string{"app", "daily", "summary", "--date", "2026-04-28"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	for _, field := range []string{
		"institutional_volume",
		"clusters",
		"repeated_tickers",
		"cluster_bombs",
		"level_touches",
		"leveraged_etf_sentiment",
		"market_exhaustion",
	} {
		if !strings.Contains(output, field) {
			t.Fatalf("expected output to contain %q, got %s", field, output)
		}
	}
	for _, field := range []string{
		"top_clusters_by_dollars",
		"top_clusters_by_multiplier",
		"sector_totals",
		"by_relative_size",
		"by_dollars",
	} {
		if strings.Contains(output, field) {
			t.Fatalf("expected compact output to omit %q, got %s", field, output)
		}
	}
}
