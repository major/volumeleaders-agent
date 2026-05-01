// Package cmd wires the VolumeLeaders command tree.
package cmd

import (
	"fmt"

	"github.com/leodido/structcli"
	"github.com/leodido/structcli/helptopics"
	"github.com/major/volumeleaders-agent/internal/cmd/trades"
	"github.com/spf13/cobra"
)

const appName = "volumeleaders-agent"

type commandFactory struct {
	name string
	new  func() (*cobra.Command, error)
}

var rootCommandFactories = []commandFactory{
	{name: "trades", new: trades.NewCommand},
	{name: "trade-clusters", new: trades.NewTradeClustersCommand},
	{name: "top10", new: trades.NewTop10Command},
	{name: "top10-clusters", new: trades.NewTop10ClustersCommand},
	{name: "top100", new: trades.NewTop100Command},
	{name: "top100-clusters", new: trades.NewTop100ClustersCommand},
	{name: "phantom", new: trades.NewPhantomCommand},
	{name: "offsetting", new: trades.NewOffsettingCommand},
	{name: "bull-leverage", new: trades.NewBullLeverageCommand},
	{name: "bull-leverage-clusters", new: trades.NewBullLeverageClustersCommand},
	{name: "bear-leverage", new: trades.NewBearLeverageCommand},
	{name: "bear-leverage-clusters", new: trades.NewBearLeverageClustersCommand},
	{name: "biotech", new: trades.NewBiotechCommand},
	{name: "bonds", new: trades.NewBondsCommand},
	{name: "commodities", new: trades.NewCommoditiesCommand},
	{name: "communications-services", new: trades.NewCommunicationsServicesCommand},
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
