package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/major/volumeleaders-agent/internal/models"
	cli "github.com/urfave/cli/v3"
)

func TestRunSnapshots(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Trades/GetAllSnapshots" {
			t.Errorf("expected path /Trades/GetAllSnapshots, got %s", r.URL.Path)
		}
		fmt.Fprint(w, `"AAPL:255.30;MSFT:420.50"`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		if err := runSnapshots(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, "AAPL") {
		t.Errorf("expected output to contain AAPL, got: %s", output)
	}
	if !strings.Contains(output, "255.3") {
		t.Errorf("expected output to contain 255.3, got: %s", output)
	}
}

func TestRunSnapshotsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	err := runSnapshots(ctx)
	assertErrContains(t, err, "query snapshots")
}

func TestRunEarnings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Earnings/GetEarnings" {
			t.Errorf("expected path /Earnings/GetEarnings, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[{
			"Ticker":"AAPL",
			"Name":"Apple Inc.",
			"Sector":"Technology",
			"Industry":"Consumer Electronics",
			"EarningsDate":"/Date(1761091200000)/",
			"AfterMarketClose":true,
			"TradeCount":12,
			"TradeClusterCount":3,
			"TradeClusterBombCount":1
		}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		if err := runEarnings(ctx, "2025-01-20", "2025-01-24", marketEarningsDefaultFields, "json"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	rows := decodeJSONRows(t, output)
	if len(rows) != 1 {
		t.Fatalf("expected 1 earnings row, got %d", len(rows))
	}
	assertJSONFields(t, rows[0], marketEarningsDefaultFields, []string{"Name", "Sector", "Industry"})
}

func TestRunEarningsAllFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Earnings/GetEarnings" {
			t.Errorf("expected path /Earnings/GetEarnings, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[{
			"Ticker":"MSFT",
			"Name":"Microsoft Corporation",
			"Sector":"Technology",
			"Industry":"Software",
			"EarningsDate":"/Date(1761091200000)/",
			"AfterMarketClose":false,
			"TradeCount":8,
			"TradeClusterCount":2,
			"TradeClusterBombCount":0
		}]`))
	}))
	t.Cleanup(server.Close)

	fields, err := outputFields[models.Earnings]("all", marketEarningsDefaultFields)
	if err != nil {
		t.Fatalf("unexpected fields error: %v", err)
	}

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		if err := runEarnings(ctx, "2025-01-20", "2025-01-24", fields, "json"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	rows := decodeJSONRows(t, output)
	if len(rows) != 1 {
		t.Fatalf("expected 1 earnings row, got %d", len(rows))
	}
	assertJSONFields(t, rows[0], []string{
		"Ticker",
		"Name",
		"Sector",
		"Industry",
		"EarningsDate",
		"AfterMarketClose",
		"TradeCount",
		"TradeClusterCount",
		"TradeClusterBombCount",
	}, nil)
}

func TestRunEarningsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	err := runEarnings(ctx, "2025-01-20", "2025-01-24", marketEarningsDefaultFields, "json")
	assertErrContains(t, err, "query earnings")
}

func TestRunExhaustion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ExecutiveSummary/GetExhaustionScores" {
			t.Errorf("expected path /ExecutiveSummary/GetExhaustionScores, got %s", r.URL.Path)
		}
		fmt.Fprint(w, `{
			"DateKey":20250115,
			"ExhaustionScoreRank":4,
			"ExhaustionScoreRank30Day":8,
			"ExhaustionScoreRank90Day":13,
			"ExhaustionScoreRank365Day":21
		}`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		if err := runExhaustion(ctx, "2025-01-15"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var score map[string]int
	if err := json.Unmarshal([]byte(output), &score); err != nil {
		t.Fatalf("decode exhaustion output: %v", err)
	}
	for _, field := range []string{"date_key", "rank", "rank_30d", "rank_90d", "rank_365d"} {
		if _, ok := score[field]; !ok {
			t.Errorf("expected compact exhaustion field %q in output %s", field, output)
		}
	}
	for _, field := range []string{"DateKey", "ExhaustionScoreRank", "ExhaustionScoreRank30Day", "ExhaustionScoreRank90Day", "ExhaustionScoreRank365Day"} {
		if _, ok := score[field]; ok {
			t.Errorf("expected raw exhaustion field %q to be omitted from output %s", field, output)
		}
	}
}

func TestRunExhaustionServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	err := runExhaustion(ctx, "2025-01-15")
	assertErrContains(t, err, "query exhaustion scores")
}

func TestMarketSnapshotsCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Trades/GetAllSnapshots" {
			t.Errorf("expected path /Trades/GetAllSnapshots, got %s", r.URL.Path)
		}
		fmt.Fprint(w, `"AAPL:255.30"`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewMarketCommand()}}
		if err := root.Run(ctx, []string{"app", "market", "snapshots"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, "AAPL") {
		t.Errorf("expected output to contain AAPL, got: %s", output)
	}
}

func TestMarketEarningsCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Earnings/GetEarnings" {
			t.Errorf("expected path /Earnings/GetEarnings, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[{
			"Ticker":"AAPL",
			"Name":"Apple Inc.",
			"Sector":"Technology",
			"Industry":"Consumer Electronics",
			"EarningsDate":"/Date(1761091200000)/",
			"AfterMarketClose":true,
			"TradeCount":12,
			"TradeClusterCount":3,
			"TradeClusterBombCount":1
		}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewMarketCommand()}}
		args := []string{
			"app",
			"market",
			"earnings",
			"--start-date",
			"2025-01-20",
			"--end-date",
			"2025-01-24",
			"--fields",
			"all",
		}
		if err := root.Run(ctx, args); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, "Name") {
		t.Errorf("expected --fields all output to contain Name, got: %s", output)
	}
}

func TestMarketExhaustionCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ExecutiveSummary/GetExhaustionScores" {
			t.Errorf("expected path /ExecutiveSummary/GetExhaustionScores, got %s", r.URL.Path)
		}
		fmt.Fprint(w, `{}`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewMarketCommand()}}
		if err := root.Run(ctx, []string{"app", "market", "exhaustion", "--date", "2025-01-15"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func decodeJSONRows(t *testing.T, output string) []map[string]json.RawMessage {
	t.Helper()

	var rows []map[string]json.RawMessage
	if err := json.Unmarshal([]byte(output), &rows); err != nil {
		t.Fatalf("decode JSON rows: %v", err)
	}
	return rows
}

func assertJSONFields(t *testing.T, row map[string]json.RawMessage, included, omitted []string) {
	t.Helper()

	for _, field := range included {
		if _, ok := row[field]; !ok {
			t.Errorf("expected field %q in row %#v", field, row)
		}
	}
	for _, field := range omitted {
		if _, ok := row[field]; ok {
			t.Errorf("expected field %q to be omitted from row %#v", field, row)
		}
	}
}
