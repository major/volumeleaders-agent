package commands

import (
	"context"
	"slices"

	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
	cli "github.com/urfave/cli/v3"
)

type volumeOptions struct {
	date, tickers, orderDir string
	start, length, orderCol int
}

// volumeFlags returns the shared flag set used by all volume subcommands.
func volumeFlags() []cli.Flag {
	return slices.Concat([]cli.Flag{
		&cli.StringFlag{Name: "date", Required: true, Usage: "Date YYYY-MM-DD"},
		&cli.StringFlag{Name: "tickers", Usage: "Comma-separated ticker symbols"},
	}, paginationFlags(100, 1, "asc"))
}

// parseVolumeOptions extracts volumeOptions from the parsed CLI flags.
func parseVolumeOptions(cmd *cli.Command) volumeOptions {
	return volumeOptions{
		date:     cmd.String("date"),
		tickers:  cmd.String("tickers"),
		start:    cmd.Int("start"),
		length:   cmd.Int("length"),
		orderCol: cmd.Int("order-col"),
		orderDir: cmd.String("order-dir"),
	}
}

// NewVolumeCommand returns the "volume" command group with all subcommands.
func NewVolumeCommand() *cli.Command {
	return &cli.Command{
		Name:  "volume",
		Usage: "Volume leaderboard commands",
		Commands: []*cli.Command{
		{
			Name:      "institutional",
			Usage:     "Query institutional volume leaderboard",
			UsageText: "volumeleaders-agent volume institutional --date 2025-01-15 --tickers AAPL,MSFT",
				Flags: volumeFlags(),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runVolume(ctx, parseVolumeOptions(cmd),
						"/InstitutionalVolume/GetInstitutionalVolume",
						datatables.InstitutionalVolumeColumns)
				},
			},
		{
			Name:      "ah-institutional",
			Usage:     "Query after-hours institutional volume leaderboard",
			UsageText: "volumeleaders-agent volume ah-institutional --date 2025-01-15",
				Flags: volumeFlags(),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runVolume(ctx, parseVolumeOptions(cmd),
						"/AHInstitutionalVolume/GetAHInstitutionalVolume",
						datatables.InstitutionalVolumeColumns)
				},
			},
		{
			Name:      "total",
			Usage:     "Query total volume leaderboard",
			UsageText: "volumeleaders-agent volume total --date 2025-01-15 --length 20",
				Flags: volumeFlags(),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runVolume(ctx, parseVolumeOptions(cmd),
						"/TotalVolume/GetTotalVolume",
						datatables.TotalVolumeColumns)
				},
			},
		},
	}
}

// runVolume is the shared handler for all volume subcommands.
func runVolume(ctx context.Context, opts volumeOptions, path string, columns []string) error {
	return runDataTablesCommand[models.Trade](ctx, path, columns,
		dataTableOptions{
			start:    opts.start,
			length:   opts.length,
			orderCol: opts.orderCol,
			orderDir: opts.orderDir,
			filters: map[string]string{
				"Date":    opts.date,
				"Tickers": opts.tickers,
			},
		},
		"query volume data")
}
