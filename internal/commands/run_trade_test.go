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

func TestTradeRunFunctions(t *testing.T) {
	tests := []struct {
		name string
		cmd  func() *cli.Command
		args []string
		path string
	}{
		{
			name: "list",
			cmd:  newTradeListCommand,
			args: []string{"app", "list", "--start-date", "2025-01-01", "--end-date", "2025-01-31"},
			path: "/Trades/GetTrades",
		},
		{
			name: "clusters",
			cmd:  newTradeClustersCommand,
			args: []string{"app", "clusters", "--start-date", "2025-01-01", "--end-date", "2025-01-31"},
			path: "/TradeClusters/GetTradeClusters",
		},
		{
			name: "cluster-bombs",
			cmd:  newTradeClusterBombsCommand,
			args: []string{"app", "cluster-bombs", "--start-date", "2025-01-01", "--end-date", "2025-01-31"},
			path: "/TradeClusterBombs/GetTradeClusterBombs",
		},
		{
			name: "alerts",
			cmd:  newTradeAlertsCommand,
			args: []string{"app", "alerts", "--date", "2025-01-15"},
			path: "/TradeAlerts/GetTradeAlerts",
		},
		{
			name: "cluster-alerts",
			cmd:  newTradeClusterAlertsCommand,
			args: []string{"app", "cluster-alerts", "--date", "2025-01-15"},
			path: "/TradeClusterAlerts/GetTradeClusterAlerts",
		},
		{
			name: "levels",
			cmd:  newTradeLevelsCommand,
			args: []string{"app", "levels", "--ticker", "AAPL", "--start-date", "2025-01-01", "--end-date", "2025-01-31"},
			path: "/TradeLevels/GetTradeLevels",
		},
		{
			name: "level-touches",
			cmd:  newTradeLevelTouchesCommand,
			args: []string{"app", "level-touches", "--start-date", "2025-01-01", "--end-date", "2025-01-31"},
			path: "/TradeLevelTouches/GetTradeLevelTouches",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.path {
					t.Errorf("expected path %s, got %s", tt.path, r.URL.Path)
				}
				fmt.Fprint(w, dataTablesJSON(`[{}]`))
			}))
			t.Cleanup(server.Close)

			ctx := contextWithTestClient(server.URL)
			captureStdout(t, func() {
				root := &cli.Command{Commands: []*cli.Command{tt.cmd()}}
				if err := root.Run(ctx, tt.args); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			})
		})
	}
}

func TestTradeListPresetIncludesDefaults(t *testing.T) {
	// Regression test for https://github.com/major/volumeleaders-agent/issues/8
	// Presets must include CLI flag defaults for params the API requires
	// (MaxVolume, MaxPrice, session toggles, etc.).
	requiredParams := []string{
		"MaxVolume", "MinVolume",
		"MaxPrice", "MinPrice",
		"MaxDollars", "MinDollars",
		"IncludePremarket", "IncludeRTH", "IncludeAH",
		"IncludeOpening", "IncludeClosing",
		"IncludePhantom", "IncludeOffsetting",
		"SecurityTypeKey", "MarketCap",
		"DarkPools", "Sweeps", "LatePrints",
		"SignaturePrints", "EvenShared",
		"TradeRankSnapshot",
	}

	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{newTradeListCommand()}}
		if err := root.Run(ctx, []string{
			"app", "list",
			"--preset", "Top-100 Rank",
			"--start-date", "2025-04-01",
			"--end-date", "2025-04-25",
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	for _, param := range requiredParams {
		if !strings.Contains(body, param+"=") {
			t.Errorf("preset request missing required param %q", param)
		}
	}
}

func TestTradeListPresetOverridesDefaults(t *testing.T) {
	// The "Top-100 Rank" preset sets TradeRank=100 and MaxDollars=100000000000,
	// overriding CLI defaults of TradeRank=-1 and MaxDollars=30000000000.
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{newTradeListCommand()}}
		if err := root.Run(ctx, []string{
			"app", "list",
			"--preset", "Top-100 Rank",
			"--start-date", "2025-04-01",
			"--end-date", "2025-04-25",
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Preset value should override CLI default.
	if !strings.Contains(body, "TradeRank=100") {
		t.Error("preset TradeRank=100 not found in request body")
	}
	if !strings.Contains(body, "MaxDollars=100000000000") {
		t.Error("preset MaxDollars=100000000000 not found in request body")
	}
}

func TestTradeListExplicitFlagOverridesPreset(t *testing.T) {
	// An explicit CLI flag should override both the default and the preset.
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{newTradeListCommand()}}
		if err := root.Run(ctx, []string{
			"app", "list",
			"--preset", "Top-100 Rank",
			"--start-date", "2025-04-01",
			"--end-date", "2025-04-25",
			"--dark-pools", "1",
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Explicit --dark-pools=1 should override the preset/default.
	if !strings.Contains(body, "DarkPools=1") {
		t.Error("explicit --dark-pools=1 not found in request body")
	}
	// Preset value should still be present.
	if !strings.Contains(body, "TradeRank=100") {
		t.Error("preset TradeRank=100 should still be present")
	}
}

func TestTradeRunFunctionServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	root := &cli.Command{Commands: []*cli.Command{newTradeListCommand()}}
	err := root.Run(ctx, []string{"app", "list", "--start-date", "2025-01-01", "--end-date", "2025-01-31"})
	assertErrContains(t, err, "query trades")
}
