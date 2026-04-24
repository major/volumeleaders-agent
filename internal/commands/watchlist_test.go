package commands

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cli "github.com/urfave/cli/v3"
)

func TestRunWatchlistConfigs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/WatchListConfigs/GetWatchLists" {
			t.Errorf("expected path /WatchListConfigs/GetWatchLists, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		if err := runWatchlistConfigs(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunWatchlistConfigsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	err := runWatchlistConfigs(ctx)
	assertErrContains(t, err, "query watchlist configs")
}

func TestRunWatchlistTickers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/WatchLists/GetWatchListTickers" {
			t.Errorf("expected path /WatchLists/GetWatchListTickers, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		if err := runWatchlistTickers(ctx, 1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunWatchlistDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/WatchListConfigs/DeleteWatchList" {
			t.Errorf("expected path /WatchListConfigs/DeleteWatchList, got %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"ok":true}`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		if err := runWatchlistDelete(ctx, 1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunWatchlistDeleteServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	err := runWatchlistDelete(ctx, 1)
	assertErrContains(t, err, "delete watchlist")
}

func TestRunWatchlistAddTicker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Chart0/UpdateWatchList" {
			t.Errorf("expected path /Chart0/UpdateWatchList, got %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"ok":true}`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		if err := runWatchlistAddTicker(ctx, 1, "NVDA"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunWatchlistAddTickerServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	err := runWatchlistAddTicker(ctx, 1, "INVALID")
	assertErrContains(t, err, "add ticker to watchlist")
}

func TestRunWatchlistCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/WatchListConfig" {
			t.Errorf("expected path /WatchListConfig, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	output := captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewWatchlistCommand()}}
		if err := root.Run(ctx, []string{"app", "watchlist", "create", "--name", "Test List"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, `"created"`) {
		t.Errorf("expected output to contain created, got: %s", output)
	}
}

func TestRunWatchlistEdit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/WatchListConfig" {
			t.Errorf("expected path /WatchListConfig, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	output := captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewWatchlistCommand()}}
		if err := root.Run(ctx, []string{"app", "watchlist", "edit", "--key", "1"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, `"updated"`) {
		t.Errorf("expected output to contain updated, got: %s", output)
	}
}

func TestWatchlistConfigsCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/WatchListConfigs/GetWatchLists" {
			t.Errorf("expected path /WatchListConfigs/GetWatchLists, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewWatchlistCommand()}}
		if err := root.Run(ctx, []string{"app", "watchlist", "configs"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestWatchlistTickersCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/WatchLists/GetWatchListTickers" {
			t.Errorf("expected path /WatchLists/GetWatchListTickers, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewWatchlistCommand()}}
		if err := root.Run(ctx, []string{"app", "watchlist", "tickers", "--watchlist-key", "1"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestWatchlistDeleteCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/WatchListConfigs/DeleteWatchList" {
			t.Errorf("expected path /WatchListConfigs/DeleteWatchList, got %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"ok":true}`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewWatchlistCommand()}}
		if err := root.Run(ctx, []string{"app", "watchlist", "delete", "--key", "1"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestWatchlistAddTickerCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Chart0/UpdateWatchList" {
			t.Errorf("expected path /Chart0/UpdateWatchList, got %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"ok":true}`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewWatchlistCommand()}}
		if err := root.Run(ctx, []string{"app", "watchlist", "add-ticker", "--watchlist-key", "1", "--ticker", "NVDA"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunWatchlistTickersServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	err := runWatchlistTickers(ctx, 1)
	assertErrContains(t, err, "query watchlist tickers")
}

func TestRunWatchlistCreateEditServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)

	t.Run("create", func(t *testing.T) {
		root := &cli.Command{Commands: []*cli.Command{NewWatchlistCommand()}}
		err := root.Run(ctx, []string{"app", "watchlist", "create", "--name", "Test"})
		assertErrContains(t, err, "save watchlist config")
	})

	t.Run("edit", func(t *testing.T) {
		root := &cli.Command{Commands: []*cli.Command{NewWatchlistCommand()}}
		err := root.Run(ctx, []string{"app", "watchlist", "edit", "--key", "1"})
		assertErrContains(t, err, "save watchlist config")
	})
}
