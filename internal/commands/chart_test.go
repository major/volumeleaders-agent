package commands

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunPriceData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Chart0/GetAllPriceVolumeTradeData" {
			t.Errorf("expected path /Chart0/GetAllPriceVolumeTradeData, got %s", r.URL.Path)
		}
		// API returns nested array: [[PriceBar, ...], ...]
		fmt.Fprint(w, `[[{}]]`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		opts := &priceDataOptions{
			ticker:    "AAPL",
			startDate: "2025-01-15",
			endDate:   "2025-01-15",
		}
		if err := runPriceData(ctx, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunPriceDataEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `[]`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	output := captureStdout(t, func() {
		opts := &priceDataOptions{ticker: "AAPL", startDate: "2025-01-15", endDate: "2025-01-15"}
		if err := runPriceData(ctx, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, "null") {
		t.Errorf("expected null output for empty response, got: %s", output)
	}
}

func TestRunPriceDataServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
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

	ctx := contextWithTestClient(server.URL)
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

	ctx := contextWithTestClient(server.URL)
	err := runChartSnapshot(ctx, "INVALID", "2025-01-15")
	assertErrContains(t, err, "query chart snapshot")
}

func TestRunChartLevels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Chart0/GetTradeLevels" {
			t.Errorf("expected path /Chart0/GetTradeLevels, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		opts := chartLevelsOptions{
			ticker:    "AAPL",
			startDate: "2025-01-01",
			endDate:   "2025-01-31",
			levels:    5,
		}
		if err := runChartLevels(ctx, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunCompany(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Chart0/GetCompany" {
			t.Errorf("expected path /Chart0/GetCompany, got %s", r.URL.Path)
		}
		fmt.Fprint(w, `{}`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		if err := runCompany(ctx, "AAPL"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunCompanyServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	err := runCompany(ctx, "INVALID")
	assertErrContains(t, err, "query company")
}
