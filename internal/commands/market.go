package commands

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
	cli "github.com/urfave/cli/v3"
)

// NewMarketCommand returns the "market" command group with all subcommands.
func NewMarketCommand() *cli.Command {
	return &cli.Command{
		Name:  "market",
		Usage: "Market-wide data commands",
		Commands: []*cli.Command{
			newSnapshotsCommand(),
			newEarningsCommand(),
			newExhaustionCommand(),
		},
	}
}

// --- Subcommand factories ---

func newSnapshotsCommand() *cli.Command {
	return &cli.Command{
		Name:      "snapshots",
		Usage:     "Get current price snapshots for all symbols",
		UsageText: "volumeleaders-agent market snapshots",
		Action: func(ctx context.Context, _ *cli.Command) error {
			return runSnapshots(ctx)
		},
	}
}

func newEarningsCommand() *cli.Command {
	return &cli.Command{
		Name:      "earnings",
		Usage:     "Query earnings calendar within a date range",
		UsageText: "volumeleaders-agent market earnings --start-date 2025-01-20 --end-date 2025-01-24",
		Flags: dateRangeFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runEarnings(ctx, cmd.String("start-date"), cmd.String("end-date"))
		},
	}
}

func newExhaustionCommand() *cli.Command {
	return &cli.Command{
		Name:      "exhaustion",
		Usage:     "Query exhaustion scores for a date",
		UsageText: "volumeleaders-agent market exhaustion --date 2025-01-15",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "date", Usage: "Date YYYY-MM-DD (empty for current day)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runExhaustion(ctx, cmd.String("date"))
		},
	}
}

// --- Action handlers ---

func runSnapshots(ctx context.Context) error {
	vlClient, err := newCommandClient(ctx)
	if err != nil {
		return err
	}

	var raw string
	if err := vlClient.PostJSON(ctx, "/Trades/GetAllSnapshots", struct{}{}, &raw); err != nil {
		slog.Error("failed to query snapshots", "error", err)
		return fmt.Errorf("query snapshots: %w", err)
	}

	snapshots := parseSnapshotString(raw)
	return printJSON(ctx, snapshots)
}

func runEarnings(ctx context.Context, startDate, endDate string) error {
	return runDataTablesCommand[models.Earnings](ctx, "/Earnings/GetEarnings", datatables.EarningsColumns,
		dataTableOptions{
			start:    0,
			length:   -1,
			orderCol: 0,
			orderDir: "asc",
			filters: map[string]string{
				"StartDate": startDate,
				"EndDate":   endDate,
			},
		},
		"query earnings")
}

func runExhaustion(ctx context.Context, date string) error {
	vlClient, err := newCommandClient(ctx)
	if err != nil {
		return err
	}

	payload := map[string]string{"Date": date}
	var score models.ExhaustionScore
	if err := vlClient.PostJSON(ctx, "/ExecutiveSummary/GetExhaustionScores", payload, &score); err != nil {
		slog.Error("failed to query exhaustion scores", "error", err)
		return fmt.Errorf("query exhaustion scores: %w", err)
	}

	return printJSON(ctx, score)
}
