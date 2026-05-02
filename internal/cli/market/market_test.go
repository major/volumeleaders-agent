package market

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/major/volumeleaders-agent/internal/cli/testutil"
	"github.com/major/volumeleaders-agent/internal/models"
)

func TestSnapshots(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Trades/GetAllSnapshots" {
			t.Errorf("expected path /Trades/GetAllSnapshots, got %s", r.URL.Path)
		}
		fmt.Fprint(w, `"AAPL:255.30;MSFT:420.50"`)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewMarketCommand()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "snapshots")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "AAPL") {
		t.Errorf("expected output to contain AAPL, got: %s", stdout)
	}
	if !strings.Contains(stdout, "255.3") {
		t.Errorf("expected output to contain 255.3, got: %s", stdout)
	}
	if !strings.Contains(stdout, "MSFT") {
		t.Errorf("expected output to contain MSFT, got: %s", stdout)
	}
}

func TestSnapshotsServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewMarketCommand()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "snapshots")
	testutil.AssertErrContains(t, err, "query snapshots")
}

func TestEarnings(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Earnings/GetEarnings" {
			t.Errorf("expected path /Earnings/GetEarnings, got %s", r.URL.Path)
		}
		fmt.Fprint(w, testutil.DataTablesJSON(`[{
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

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewMarketCommand()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "earnings", "--start-date", "2025-01-20", "--end-date", "2025-01-24")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rows := decodeJSONRows(t, stdout)
	if len(rows) != 1 {
		t.Fatalf("expected 1 earnings row, got %d", len(rows))
	}
	assertJSONFields(t, rows[0], marketEarningsDefaultFields, []string{"Name", "Sector", "Industry"})
}

func TestEarningsAllFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, testutil.DataTablesJSON(`[{
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

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewMarketCommand()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "earnings", "--start-date", "2025-01-20", "--end-date", "2025-01-24", "--fields", "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rows := decodeJSONRows(t, stdout)
	if len(rows) != 1 {
		t.Fatalf("expected 1 earnings row, got %d", len(rows))
	}
	allFields := []string{
		"Ticker", "Name", "Sector", "Industry",
		"EarningsDate", "AfterMarketClose",
		"TradeCount", "TradeClusterCount", "TradeClusterBombCount",
	}
	assertJSONFields(t, rows[0], allFields, nil)
}

func TestEarningsMissingDates(t *testing.T) {
	t.Parallel()

	ctx := testutil.ContextWithTestClient(t, "http://unused")
	cmd := NewMarketCommand()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "earnings")
	if err == nil {
		t.Fatal("expected error for missing date flags, got nil")
	}
}

func TestEarningsServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewMarketCommand()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "earnings", "--start-date", "2025-01-20", "--end-date", "2025-01-24")
	testutil.AssertErrContains(t, err, "query earnings")
}

func TestEarningsDaysFlag(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, testutil.DataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewMarketCommand()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "earnings", "--days", "5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout == "" {
		t.Error("expected non-empty stdout")
	}
}

func TestExhaustion(t *testing.T) {
	t.Parallel()

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

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewMarketCommand()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "exhaustion", "--date", "2025-01-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var score map[string]int
	if err := json.Unmarshal([]byte(stdout), &score); err != nil {
		t.Fatalf("decode exhaustion output: %v", err)
	}
	for _, field := range []string{"date_key", "rank", "rank_30d", "rank_90d", "rank_365d"} {
		if _, ok := score[field]; !ok {
			t.Errorf("expected compact field %q in output %s", field, stdout)
		}
	}
	for _, field := range []string{"DateKey", "ExhaustionScoreRank", "ExhaustionScoreRank30Day"} {
		if _, ok := score[field]; ok {
			t.Errorf("expected raw field %q to be omitted from output %s", field, stdout)
		}
	}
}

func TestExhaustionNoDate(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{}`)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewMarketCommand()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "exhaustion")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExhaustionServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewMarketCommand()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "exhaustion", "--date", "2025-01-15")
	testutil.AssertErrContains(t, err, "query exhaustion scores")
}

func TestSummarizeMarketExhaustion(t *testing.T) {
	t.Parallel()

	score := models.ExhaustionScore{
		DateKey:                   20250115,
		ExhaustionScoreRank:       4,
		ExhaustionScoreRank30Day:  8,
		ExhaustionScoreRank90Day:  13,
		ExhaustionScoreRank365Day: 21,
	}
	result := summarizeMarketExhaustion(score)

	if result.DateKey != 20250115 {
		t.Errorf("DateKey = %d, want 20250115", result.DateKey)
	}
	if result.Rank != 4 {
		t.Errorf("Rank = %d, want 4", result.Rank)
	}
	if result.Rank30D != 8 {
		t.Errorf("Rank30D = %d, want 8", result.Rank30D)
	}
	if result.Rank90D != 13 {
		t.Errorf("Rank90D = %d, want 13", result.Rank90D)
	}
	if result.Rank365D != 21 {
		t.Errorf("Rank365D = %d, want 21", result.Rank365D)
	}
}

// --- test helpers ---

// decodeJSONRows unmarshals a JSON array of objects from the command output.
func decodeJSONRows(t *testing.T, output string) []map[string]json.RawMessage {
	t.Helper()

	var rows []map[string]json.RawMessage
	if err := json.Unmarshal([]byte(output), &rows); err != nil {
		t.Fatalf("decode JSON rows: %v", err)
	}
	return rows
}

// assertJSONFields verifies that the row contains included fields and omits excluded fields.
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
