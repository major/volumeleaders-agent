package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommandWiresStructCLIFeatures(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantOuts    []string
		wantNotOuts []string
	}{
		{
			name:        "json schema tree is available",
			args:        []string{"--jsonschema=tree"},
			wantOuts:    []string{"watchlists", "save-watchlist", "delete-watchlist", "trades", "trade-levels", "trade-clusters", "top10-clusters", "top100-clusters", "top30-10x-99pct", "top30-10x-99pct-clusters", "top100-dark-pool-20x", "top100-dark-pool-20x-clusters", "top100-leveraged-etfs", "top100-leveraged-etfs-clusters", "top100-dark-pool-sweeps", "top100-dark-pool-sweeps-clusters", "phantom", "offsetting", "overbought", "overbought-clusters", "oversold", "oversold-clusters"},
			wantNotOuts: []string{"bull-leverage", "bull-leverage-clusters", "bear-leverage", "bear-leverage-clusters", "biotech", "biotech-clusters", "bonds", "bonds-clusters", "commodities", "commodities-clusters", "communications-services", "communications-services-clusters"},
		},
		{
			name:        "env vars reference topic is available",
			args:        []string{"env-vars"},
			wantOuts:    []string{"VOLUMELEADERS_AGENT_WATCHLISTS_PRESET_FIELDS", "VOLUMELEADERS_AGENT_SAVE_WATCHLIST_NAME", "VOLUMELEADERS_AGENT_DELETE_WATCHLIST_SEARCH_TEMPLATE_KEY", "VOLUMELEADERS_AGENT_TRADES_DATE", "VOLUMELEADERS_AGENT_TRADE_LEVELS_TICKER", "VOLUMELEADERS_AGENT_TRADE_CLUSTERS_DATE", "VOLUMELEADERS_AGENT_TOP10_CLUSTERS_DATE", "VOLUMELEADERS_AGENT_TOP100_CLUSTERS_DATE", "VOLUMELEADERS_AGENT_TOP30_10X_99PCT_DATE", "VOLUMELEADERS_AGENT_TOP30_10X_99PCT_CLUSTERS_DATE", "VOLUMELEADERS_AGENT_TOP100_DARK_POOL_20X_DATE", "VOLUMELEADERS_AGENT_TOP100_DARK_POOL_20X_CLUSTERS_DATE", "VOLUMELEADERS_AGENT_TOP100_LEVERAGED_ETFS_DATE", "VOLUMELEADERS_AGENT_TOP100_LEVERAGED_ETFS_CLUSTERS_DATE", "VOLUMELEADERS_AGENT_TOP100_DARK_POOL_SWEEPS_DATE", "VOLUMELEADERS_AGENT_TOP100_DARK_POOL_SWEEPS_CLUSTERS_DATE", "VOLUMELEADERS_AGENT_PHANTOM_DATE", "VOLUMELEADERS_AGENT_OFFSETTING_DATE", "VOLUMELEADERS_AGENT_OVERBOUGHT_DATE", "VOLUMELEADERS_AGENT_OVERBOUGHT_CLUSTERS_DATE", "VOLUMELEADERS_AGENT_OVERSOLD_DATE", "VOLUMELEADERS_AGENT_OVERSOLD_CLUSTERS_DATE"},
			wantNotOuts: []string{"VOLUMELEADERS_AGENT_BULL_LEVERAGE_DATE", "VOLUMELEADERS_AGENT_BULL_LEVERAGE_CLUSTERS_DATE", "VOLUMELEADERS_AGENT_BEAR_LEVERAGE_DATE", "VOLUMELEADERS_AGENT_BEAR_LEVERAGE_CLUSTERS_DATE", "VOLUMELEADERS_AGENT_BIOTECH_DATE", "VOLUMELEADERS_AGENT_BIOTECH_CLUSTERS_DATE", "VOLUMELEADERS_AGENT_BONDS_DATE", "VOLUMELEADERS_AGENT_BONDS_CLUSTERS_DATE", "VOLUMELEADERS_AGENT_COMMODITIES_DATE", "VOLUMELEADERS_AGENT_COMMODITIES_CLUSTERS_DATE", "VOLUMELEADERS_AGENT_COMMUNICATIONS_SERVICES_DATE", "VOLUMELEADERS_AGENT_COMMUNICATIONS_SERVICES_CLUSTERS_DATE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd, err := NewRootCmd()
			if err != nil {
				t.Fatalf("NewRootCmd() error = %v", err)
			}

			var output bytes.Buffer
			rootCmd.SetOut(&output)
			rootCmd.SetErr(&output)
			rootCmd.SetArgs(tt.args)

			if err := rootCmd.Execute(); err != nil {
				t.Fatalf("Execute() error = %v\noutput: %s", err, output.String())
			}
			for _, wantOut := range tt.wantOuts {
				if !strings.Contains(output.String(), wantOut) {
					t.Fatalf("expected output to contain %q, got %q", wantOut, output.String())
				}
			}
			for _, wantNotOut := range tt.wantNotOuts {
				if strings.Contains(output.String(), wantNotOut) {
					t.Fatalf("expected output not to contain %q, got %q", wantNotOut, output.String())
				}
			}
		})
	}
}

func TestNewRootCmdWrapsCommandFactoryError(t *testing.T) {
	wantErr := errors.New("factory failed")
	oldFactories := rootCommandFactories
	rootCommandFactories = []commandFactory{
		{
			name: "broken",
			new: func() (*cobra.Command, error) {
				return nil, wantErr
			},
		},
	}
	t.Cleanup(func() { rootCommandFactories = oldFactories })

	_, err := NewRootCmd()
	if err == nil {
		t.Fatal("expected command factory error")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped factory error, got %v", err)
	}
	if !strings.Contains(err.Error(), "create broken command") {
		t.Fatalf("expected command name context, got %v", err)
	}
}
