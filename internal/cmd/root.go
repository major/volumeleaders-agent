// Package cmd wires the VolumeLeaders command tree.
package cmd

import (
	"fmt"

	"github.com/leodido/structcli"
	"github.com/leodido/structcli/helptopics"
	"github.com/major/volumeleaders-agent/internal/cmd/trades"
	"github.com/major/volumeleaders-agent/internal/cmd/watchlists"
	"github.com/spf13/cobra"
)

const appName = "volumeleaders-agent"

type commandFactory struct {
	name string
	new  func() (*cobra.Command, error)
}

var rootCommandFactories = []commandFactory{
	// Watchlists are account-level saved filters, so keep them before dated trade
	// commands that callers may choose after inspecting the configured criteria.
	{name: "watchlists", new: watchlists.NewCommand},
	{name: "save-watchlist", new: watchlists.NewSaveCommand},
	{name: "delete-watchlist", new: watchlists.NewDeleteCommand},

	// Keep individual trade commands together so the help, JSON schema, and MCP
	// discovery output read in the same order a user would browse trade filters.
	{name: "trades", new: trades.NewCommand},
	{name: "top10", new: trades.NewTop10Command},
	{name: "top100", new: trades.NewTop100Command},
	{name: "top30-10x-99pct", new: trades.NewTop3010x99PctCommand},
	{name: "top100-dark-pool-20x", new: trades.NewTop100DarkPool20xCommand},
	{name: "top100-leveraged-etfs", new: trades.NewTop100LeveragedETFsCommand},
	{name: "top100-dark-pool-sweeps", new: trades.NewTop100DarkPoolSweepsCommand},
	{name: "phantom", new: trades.NewPhantomCommand},
	{name: "offsetting", new: trades.NewOffsettingCommand},
	{name: "overbought", new: trades.NewOverboughtCommand},
	{name: "oversold", new: trades.NewOversoldCommand},

	// Keep cluster equivalents in their own block. Most mirror the trade filters
	// above, but they call the TradeClusters endpoint and return cluster rows.
	{name: "trade-clusters", new: trades.NewTradeClustersCommand},
	{name: "top10-clusters", new: trades.NewTop10ClustersCommand},
	{name: "top100-clusters", new: trades.NewTop100ClustersCommand},
	{name: "top30-10x-99pct-clusters", new: trades.NewTop3010x99PctClustersCommand},
	{name: "top100-dark-pool-20x-clusters", new: trades.NewTop100DarkPool20xClustersCommand},
	{name: "top100-leveraged-etfs-clusters", new: trades.NewTop100LeveragedETFsClustersCommand},
	{name: "top100-dark-pool-sweeps-clusters", new: trades.NewTop100DarkPoolSweepsClustersCommand},
	{name: "overbought-clusters", new: trades.NewOverboughtClustersCommand},
	{name: "oversold-clusters", new: trades.NewOversoldClustersCommand},
}

// NewRootCmd builds a fresh command tree for the VolumeLeaders CLI.
func NewRootCmd() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:              appName,
		Short:            "VolumeLeaders market intelligence CLI",
		Long:             "VolumeLeaders market intelligence CLI for human and LLM workflows.",
		TraverseChildren: true,
	}

	for _, factory := range rootCommandFactories {
		cmd, err := factory.new()
		if err != nil {
			return nil, fmt.Errorf("create %s command: %w", factory.name, err)
		}
		rootCmd.AddCommand(cmd)
	}

	if err := structcli.Setup(rootCmd,
		structcli.WithAppName(appName),
		structcli.WithFlagErrors(),
		structcli.WithHelpTopics(helptopics.Options{ReferenceSection: true}),
		structcli.WithJSONSchema(),
		structcli.WithMCP(),
	); err != nil {
		return nil, fmt.Errorf("setup structcli: %w", err)
	}

	return rootCmd, nil
}
