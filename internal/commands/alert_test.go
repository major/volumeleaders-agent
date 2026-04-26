package commands

import (
	"fmt"
	"io"
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

	ctx := contextWithTestClient(t, server.URL)
	captureStdout(t, func() {
		if err := runAlertConfigs(ctx, "json"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunAlertConfigsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	err := runAlertConfigs(ctx, "json")
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

	ctx := contextWithTestClient(t, server.URL)
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

	ctx := contextWithTestClient(t, server.URL)
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

	ctx := contextWithTestClient(t, server.URL)
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

	ctx := contextWithTestClient(t, server.URL)
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
	var capturedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewAlertCommand()}}
		if err := root.Run(ctx, []string{"app", "alert", "create", "--name", "Ticker Alert", "--tickers", "AAPL,MSFT"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(string(capturedBody), "SelectedTickers") {
		t.Errorf("expected request body to reference SelectedTickers, got: %s", capturedBody)
	}
}

func TestAlertConfigsCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewAlertCommand()}}
		if err := root.Run(ctx, []string{"app", "alert", "configs"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAlertDeleteCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"ok":true}`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewAlertCommand()}}
		if err := root.Run(ctx, []string{"app", "alert", "delete", "--key", "42"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunAlertCreateEditServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)

	t.Run("create", func(t *testing.T) {
		root := &cli.Command{Commands: []*cli.Command{NewAlertCommand()}}
		err := root.Run(ctx, []string{"app", "alert", "create", "--name", "Test"})
		assertErrContains(t, err, "save alert config")
	})

	t.Run("edit", func(t *testing.T) {
		root := &cli.Command{Commands: []*cli.Command{NewAlertCommand()}}
		err := root.Run(ctx, []string{"app", "alert", "edit", "--key", "42"})
		assertErrContains(t, err, "save alert config")
	})
}
