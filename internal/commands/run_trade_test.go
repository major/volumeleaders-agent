package commands

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
