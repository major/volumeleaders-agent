package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/leodido/structcli"
	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/alert"
	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/cli/market"
	"github.com/major/volumeleaders-agent/internal/cli/trade"
	"github.com/major/volumeleaders-agent/internal/cli/volume"
	"github.com/major/volumeleaders-agent/internal/cli/watchlist"
)

// rootOptions holds flags bound to the root command via structcli.Bind.
// The bind pipeline populates these fields before PersistentPreRunE fires.
type rootOptions struct {
	Pretty bool `flag:"pretty" flagdescr:"Pretty-print JSON output with indentation"`
}

// NewRootCmd returns the root cobra command for volumeleaders-agent.
func NewRootCmd(version string) *cobra.Command {
	opts := &rootOptions{}
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
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
			cmd.SetContext(context.WithValue(cmd.Context(), common.PrettyJSONKey, opts.Pretty))
			return nil
		},
	}
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
	if err := structcli.Bind(cmd, opts); err != nil {
		panic(fmt.Sprintf("structcli.Bind root options: %v", err))
	}
	return cmd
}

// SetupCLI configures structcli features (JSON schema, structured flag errors)
// on the root command. Called from main after NewRootCmd; separated because
// WithJSONSchema uses cobra.OnInitialize (process-global) which races in
// parallel tests.
func SetupCLI(cmd *cobra.Command) {
	if err := structcli.Setup(cmd, structcli.WithJSONSchema(), structcli.WithFlagErrors()); err != nil {
		panic(fmt.Sprintf("structcli.Setup: %v", err))
	}
}
