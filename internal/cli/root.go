package cli

import (
	"context"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/alert"
	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/cli/market"
	"github.com/major/volumeleaders-agent/internal/cli/trade"
	"github.com/major/volumeleaders-agent/internal/cli/volume"
	"github.com/major/volumeleaders-agent/internal/cli/watchlist"
)

// NewRootCmd returns the root cobra command for volumeleaders-agent.
func NewRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "volumeleaders-agent",
		Short:            "CLI tool for querying VolumeLeaders institutional trade data",
		Long:             "volumeleaders-agent queries institutional trade data from VolumeLeaders, providing access to trades, volume leaderboards, market snapshots, alert configurations, and watchlists. All commands output compact JSON to stdout by default; use --pretty for indented output or --format csv/tsv where supported. Authenticate via browser cookie extraction (see documentation for setup).",
		Version:          version,
		SilenceErrors:    true,
		SilenceUsage:     true,
		TraverseChildren: true,
		Args:             cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			prettyFlag, _ := cmd.Flags().GetBool("pretty")
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
			cmd.SetContext(context.WithValue(cmd.Context(), common.PrettyJSONKey, prettyFlag))
			return nil
		},
	}
	cmd.PersistentFlags().Bool("pretty", false, "Pretty-print JSON output with indentation")
	cmd.AddGroup(
		&cobra.Group{ID: "trading", Title: "Trading Commands:"},
		&cobra.Group{ID: "volume", Title: "Volume Commands:"},
		&cobra.Group{ID: "market", Title: "Market Commands:"},
		&cobra.Group{ID: "alerts", Title: "Alert Commands:"},
		&cobra.Group{ID: "watchlists", Title: "Watchlist Commands:"},
	)
	cmd.AddCommand(
		trade.NewCmd(),
		volume.NewVolumeCommand(),
		market.NewMarketCommand(),
		alert.NewAlertCommand(),
		watchlist.NewCmd(),
	)
	return cmd
}
