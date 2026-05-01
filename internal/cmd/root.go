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
	rootCmd.AddCommand(tradesCmd)

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
