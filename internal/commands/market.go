package commands

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
	cli "github.com/urfave/cli/v3"
)

var marketEarningsDefaultFields = []string{
	"Ticker",
	"EarningsDate",
	"AfterMarketClose",
	"TradeCount",
	"TradeClusterCount",
	"TradeClusterBombCount",
}

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
		UsageText: "volumeleaders-agent market earnings --days 5",
		Flags: slices.Concat(dateRangeFlags(), outputFormatFlags(), []cli.Flag{
			&cli.StringFlag{
				Name:  "fields",
				Usage: "Comma-separated fields to include (use 'all' for every field)",
			},
		}),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			startDate, endDate, err := requiredDateRange(cmd)
			if err != nil {
				return err
			}
			fields, err := outputFields[models.Earnings](cmd.String("fields"), marketEarningsDefaultFields)
			if err != nil {
				return err
			}
			return runEarnings(ctx, startDate, endDate, fields, cmd.String("format"))
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

func runEarnings(ctx context.Context, startDate, endDate string, fields []string, format string) error {
	return runDataTablesCommand[models.Earnings](ctx, "/Earnings/GetEarnings", datatables.EarningsColumns,
		dataTableOptions{
			start:    0,
			length:   -1,
			orderCol: 0,
			orderDir: "asc",
			fields:   fields,
			filters: map[string]string{
				"StartDate": startDate,
				"EndDate":   endDate,
			},
		},
		format,
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

	return printJSON(ctx, summarizeMarketExhaustion(score))
}

func summarizeMarketExhaustion(score models.ExhaustionScore) models.MarketExhaustion {
	return models.MarketExhaustion{
		DateKey:  score.DateKey,
		Rank:     score.ExhaustionScoreRank,
		Rank30D:  score.ExhaustionScoreRank30Day,
		Rank90D:  score.ExhaustionScoreRank90Day,
		Rank365D: score.ExhaustionScoreRank365Day,
	}
}
