package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/major/volumeleaders-agent/internal/models"
	cli "github.com/urfave/cli/v3"
)

func TestRunPriceData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Chart0/GetAllPriceVolumeTradeData" {
			t.Errorf("expected path /Chart0/GetAllPriceVolumeTradeData, got %s", r.URL.Path)
		}
		// API returns nested array: [[PriceBar, ...], ...]
		fmt.Fprint(w, `[[{"DateKey":20250428,"TimeKey":945,"SecurityKey":123,"Ticker":"AAPL","FullDateTime":"2025-04-28T09:45:00","OpenPrice":100,"HighPrice":101,"LowPrice":99,"ClosePrice":100.5,"Volume":100000,"Dollars":10050000,"Trades":4,"FrequencyLast30TD":9}]]`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		opts := &priceDataOptions{
			ticker:    "AAPL",
			startDate: "2025-01-15",
			endDate:   "2025-01-15",
		}
		if err := runPriceData(ctx, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal([]byte(output), &rows); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	for _, field := range []string{"DateKey", "FullDateTime", "OpenPrice", "Volume", "Dollars", "Trades"} {
		if _, ok := rows[0][field]; !ok {
			t.Errorf("expected compact default field %q in output", field)
		}
	}
	for _, field := range []string{"SecurityKey", "Ticker", "FrequencyLast30TD"} {
		if _, ok := rows[0][field]; ok {
			t.Errorf("did not expect noisy field %q in compact default output", field)
		}
	}
}

func TestRunPriceDataEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `[]`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		opts := &priceDataOptions{ticker: "AAPL", startDate: "2025-01-15", endDate: "2025-01-15"}
		if err := runPriceData(ctx, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if strings.TrimSpace(output) != "[]" {
		t.Errorf("expected empty array output for empty response, got: %s", output)
	}
}

func TestRunPriceDataFieldsSelectsRequestedOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `[[{"DateKey":20250428,"SecurityKey":123,"Ticker":"AAPL","Volume":100000}]]`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		opts := &priceDataOptions{
			ticker:    "AAPL",
			startDate: "2025-01-15",
			endDate:   "2025-01-15",
			fields:    []string{"Ticker", "SecurityKey"},
		}
		if err := runPriceData(ctx, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var rows []map[string]json.RawMessage
	if err := json.Unmarshal([]byte(output), &rows); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if len(rows[0]) != 2 {
		t.Fatalf("expected 2 selected fields, got %#v", rows[0])
	}
	for _, field := range []string{"Ticker", "SecurityKey"} {
		if _, ok := rows[0][field]; !ok {
			t.Errorf("expected selected field %q in output", field)
		}
	}
	if _, ok := rows[0]["Volume"]; ok {
		t.Errorf("did not expect unselected field Volume in output")
	}
}

func TestChartPriceDataFieldsRejectsInvalidField(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		fmt.Fprint(w, `[[{}]]`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	root := &cli.Command{Commands: []*cli.Command{NewChartCommand()}}
	err := root.Run(ctx, []string{
		"app", "chart", "price-data", "--ticker", "AAPL", "--start-date", "2025-01-15", "--end-date", "2025-01-15",
		"--fields", "Ticker,NotAField",
	})
	assertErrContains(t, err, "invalid field \"NotAField\"")
	if calls != 0 {
		t.Errorf("expected invalid fields to fail before API call, got %d calls", calls)
	}
}

func TestRunPriceDataServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	opts := &priceDataOptions{ticker: "AAPL", startDate: "2025-01-15", endDate: "2025-01-15"}
	err := runPriceData(ctx, opts)
	assertErrContains(t, err, "query price data")
}

func TestRunChartSnapshot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Chart0/GetSnapshot" {
			t.Errorf("expected path /Chart0/GetSnapshot, got %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"Snapshot":{}}`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	captureStdout(t, func() {
		if err := runChartSnapshot(ctx, "AAPL", "2025-01-15"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunChartSnapshotServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	err := runChartSnapshot(ctx, "INVALID", "2025-01-15")
	assertErrContains(t, err, "query chart snapshot")
}

func TestRunChartLevels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Chart0/GetTradeLevels" {
			t.Errorf("expected path /Chart0/GetTradeLevels, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[{"Ticker":"AAPL","Name":"Apple Inc.","Price":195.5,"Dollars":25000000,"Volume":125000,"Trades":5,"RelativeSize":8.5,"CumulativeDistribution":0.94,"TradeLevelRank":2,"Dates":"2025-01-01 - 2025-01-31"}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		opts := chartLevelsOptions{
			ticker:    "AAPL",
			startDate: "2025-01-01",
			endDate:   "2025-01-31",
			levels:    5,
		}
		if err := runChartLevels(ctx, &opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal([]byte(output), &rows); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	for _, field := range []string{"Price", "Dollars", "Volume", "Dates"} {
		if _, ok := rows[0][field]; !ok {
			t.Errorf("expected compact default field %q in output", field)
		}
	}
	for _, field := range []string{"Ticker", "Name", "MinDate", "MaxDate"} {
		if _, ok := rows[0][field]; ok {
			t.Errorf("did not expect noisy field %q in compact default output", field)
		}
	}
}

func TestRunCompany(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Chart0/GetCompany" {
			t.Errorf("expected path /Chart0/GetCompany, got %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"SecurityKey":123,"Name":"Apple Inc.","Ticker":"AAPL","Sector":"Technology","Industry":"Consumer Electronics","Description":"long company description","AverageBlockSizeDollars":2500000,"AverageBlockSizeDollars30Days":3000000,"AverageDailyVolume":65000000,"CurrentPrice":195.5,"MarketCap":3000000000000,"OptionsEnabled":true,"News":"long news payload"}`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		if err := runCompany(ctx, "AAPL", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	var row map[string]json.RawMessage
	if err := json.Unmarshal([]byte(output), &row); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	for _, field := range []string{"Name", "Ticker", "Sector", "Industry", "MarketCap", "CurrentPrice"} {
		if _, ok := row[field]; !ok {
			t.Errorf("expected compact default field %q in output", field)
		}
	}
	for _, field := range []string{"SecurityKey", "Description", "AverageBlockSizeDollars30Days", "News"} {
		if _, ok := row[field]; ok {
			t.Errorf("did not expect noisy field %q in compact default output", field)
		}
	}
}

func TestRunCompanyServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	err := runCompany(ctx, "INVALID", nil)
	assertErrContains(t, err, "query company")
}

func TestRunCompanyFieldsAllIncludesFullModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"SecurityKey":123,"Name":"Apple Inc.","Ticker":"AAPL","Description":"long company description"}`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		fields, err := outputFields[models.Company]("all", chartCompanyDefaultFields)
		if err != nil {
			t.Fatalf("unexpected field parse error: %v", err)
		}
		if err := runCompany(ctx, "AAPL", fields); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var row map[string]json.RawMessage
	if err := json.Unmarshal([]byte(output), &row); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	for _, field := range []string{"SecurityKey", "Description", "AverageBlockSizeDollars30Days"} {
		if _, ok := row[field]; !ok {
			t.Errorf("expected all-fields output to include %q", field)
		}
	}
}

func TestRunPriceDataDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `[["not a price bar object"]]`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	err := runPriceData(ctx, &priceDataOptions{ticker: "AAPL", startDate: "2025-01-15", endDate: "2025-01-15"})
	assertErrContains(t, err, "decode price bars")
}

func TestChartPriceDataCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `[[{}]]`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewChartCommand()}}
		if err := root.Run(ctx, []string{"app", "chart", "price-data", "--ticker", "AAPL", "--start-date", "2025-01-15", "--end-date", "2025-01-15"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestChartSnapshotCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"Snapshot":{}}`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewChartCommand()}}
		if err := root.Run(ctx, []string{"app", "chart", "snapshot", "--ticker", "AAPL", "--date-key", "2025-01-15"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestChartLevelsCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewChartCommand()}}
		if err := root.Run(ctx, []string{"app", "chart", "levels", "--ticker", "AAPL", "--start-date", "2025-01-01", "--end-date", "2025-01-31"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestChartCompanyCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{}`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewChartCommand()}}
		if err := root.Run(ctx, []string{"app", "chart", "company", "--ticker", "AAPL"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestChartPriceDataPositionalTickerAndDays(t *testing.T) {
	var gotTicker, gotStart, gotEnd string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		gotTicker, _ = payload["Ticker"].(string)
		gotStart, _ = payload["StartDateKey"].(string)
		gotEnd, _ = payload["EndDateKey"].(string)
		fmt.Fprint(w, `[[{}]]`)
	}))
	t.Cleanup(server.Close)

	frozen := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	origTimeNow := timeNow
	timeNow = func() time.Time { return frozen }
	t.Cleanup(func() { timeNow = origTimeNow })

	ctx := contextWithTestClient(t, server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewChartCommand()}}
		if err := root.Run(ctx, []string{"app", "chart", "price-data", "AAPL", "--days", "2"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if gotTicker != "AAPL" {
		t.Errorf("Ticker = %q, want AAPL", gotTicker)
	}
	if gotStart != "20250613" {
		t.Errorf("StartDateKey = %q, want 20250613", gotStart)
	}
	if gotEnd != "20250615" {
		t.Errorf("EndDateKey = %q, want 20250615", gotEnd)
	}
}
