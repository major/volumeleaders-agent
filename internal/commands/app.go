// Package commands defines the CLI command tree and action handlers for
// querying VolumeLeaders institutional trade data.
package commands

import (
	"context"
	"log/slog"
	"os"

	cli "github.com/urfave/cli/v3"
)

// NewApp returns the root CLI command with all subcommand groups registered.
func NewApp(version string) *cli.Command {
	app := &cli.Command{
		Name:    "volumeleaders-agent",
		Version: version,
		Usage:   "CLI tool for querying VolumeLeaders institutional trade data",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "pretty", Usage: "Pretty-print JSON output with indentation"},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			slog.SetDefault(logger)
			return context.WithValue(ctx, prettyJSONKey, cmd.Bool("pretty")), nil
		},
		Commands: []*cli.Command{
			NewTradeCommand(),
			NewVolumeCommand(),
			NewMarketCommand(),
			NewAlertCommand(),
			NewWatchlistCommand(),
		},
	}
	app.Commands = append(app.Commands, SchemaCommand(app, os.Stdout))
	return app
}
