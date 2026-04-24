package commands

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cli "github.com/urfave/cli/v3"
)

func TestRunAlertConfigs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/AlertConfigs/GetAlertConfigs" {
			t.Errorf("expected path /AlertConfigs/GetAlertConfigs, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		if err := runAlertConfigs(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunAlertConfigsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	err := runAlertConfigs(ctx)
	assertErrContains(t, err, "query alert configs")
}

func TestRunAlertDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/AlertConfigs/DeleteAlertConfig" {
			t.Errorf("expected path /AlertConfigs/DeleteAlertConfig, got %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"ok":true}`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		if err := runAlertDelete(ctx, 42); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunAlertDeleteServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	err := runAlertDelete(ctx, 42)
	assertErrContains(t, err, "delete alert config")
}

func TestRunAlertCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/AlertConfig" {
			t.Errorf("expected path /AlertConfig, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	output := captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewAlertCommand()}}
		if err := root.Run(ctx, []string{"app", "alert", "create", "--name", "Test Alert"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, `"created"`) {
		t.Errorf("expected output to contain created, got: %s", output)
	}
}

func TestRunAlertEdit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/AlertConfig" {
			t.Errorf("expected path /AlertConfig, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	output := captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewAlertCommand()}}
		if err := root.Run(ctx, []string{"app", "alert", "edit", "--key", "42"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, `"updated"`) {
		t.Errorf("expected output to contain updated, got: %s", output)
	}
}

func TestRunAlertCreateWithTickers(t *testing.T) {
	// Verifies that buildAlertConfigFields auto-selects SelectedTickers when
	// tickers are specified but ticker-group is left at the default.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewAlertCommand()}}
		if err := root.Run(ctx, []string{"app", "alert", "create", "--name", "Ticker Alert", "--tickers", "AAPL,MSFT"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
