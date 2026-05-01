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

// NewRootCmd builds a fresh command tree for the VolumeLeaders CLI.
func NewRootCmd() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:              appName,
		Short:            "VolumeLeaders market intelligence CLI",
		Long:             "VolumeLeaders market intelligence CLI for human and LLM workflows.",
		TraverseChildren: true,
	}

	tradesCmd, err := trades.NewCommand()
	if err != nil {
		return nil, fmt.Errorf("create trades command: %w", err)
	}
	top10Cmd, err := trades.NewTop10Command()
	if err != nil {
		return nil, fmt.Errorf("create top10 command: %w", err)
	}
	top100Cmd, err := trades.NewTop100Command()
	if err != nil {
		return nil, fmt.Errorf("create top100 command: %w", err)
	}
	phantomCmd, err := trades.NewPhantomCommand()
	if err != nil {
		return nil, fmt.Errorf("create phantom command: %w", err)
	}
	offsettingCmd, err := trades.NewOffsettingCommand()
	if err != nil {
		return nil, fmt.Errorf("create offsetting command: %w", err)
	}
	bullLeverageCmd, err := trades.NewBullLeverageCommand()
	if err != nil {
		return nil, fmt.Errorf("create bull-leverage command: %w", err)
	}
	bearLeverageCmd, err := trades.NewBearLeverageCommand()
	if err != nil {
		return nil, fmt.Errorf("create bear-leverage command: %w", err)
	}
	biotechCmd, err := trades.NewBiotechCommand()
	if err != nil {
		return nil, fmt.Errorf("create biotech command: %w", err)
	}
	bondsCmd, err := trades.NewBondsCommand()
	if err != nil {
		return nil, fmt.Errorf("create bonds command: %w", err)
	}
	commoditiesCmd, err := trades.NewCommoditiesCommand()
	if err != nil {
		return nil, fmt.Errorf("create commodities command: %w", err)
	}
	communicationsServicesCmd, err := trades.NewCommunicationsServicesCommand()
	if err != nil {
		return nil, fmt.Errorf("create communications-services command: %w", err)
	}
	rootCmd.AddCommand(tradesCmd, top10Cmd, top100Cmd, phantomCmd, offsettingCmd, bullLeverageCmd, bearLeverageCmd, biotechCmd, bondsCmd, commoditiesCmd, communicationsServicesCmd)

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
