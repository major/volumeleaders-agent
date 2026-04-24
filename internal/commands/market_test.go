package commands

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunSnapshots(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Trades/GetAllSnapshots" {
			t.Errorf("expected path /Trades/GetAllSnapshots, got %s", r.URL.Path)
		}
		fmt.Fprint(w, `"AAPL:255.30;MSFT:420.50"`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
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

	ctx := contextWithTestClient(server.URL)
	err := runSnapshots(ctx)
	assertErrContains(t, err, "query snapshots")
}

func TestRunEarnings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Earnings/GetEarnings" {
			t.Errorf("expected path /Earnings/GetEarnings, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		if err := runEarnings(ctx, "2025-01-20", "2025-01-24"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunEarningsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	err := runEarnings(ctx, "2025-01-20", "2025-01-24")
	assertErrContains(t, err, "query earnings")
}

func TestRunExhaustion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ExecutiveSummary/GetExhaustionScores" {
			t.Errorf("expected path /ExecutiveSummary/GetExhaustionScores, got %s", r.URL.Path)
		}
		fmt.Fprint(w, `{}`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		if err := runExhaustion(ctx, "2025-01-15"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunExhaustionServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	err := runExhaustion(ctx, "2025-01-15")
	assertErrContains(t, err, "query exhaustion scores")
}
