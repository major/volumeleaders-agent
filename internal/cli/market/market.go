package market

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
)

// marketEarningsDefaultFields defines the default field subset for earnings output.
var marketEarningsDefaultFields = []string{
	"Ticker",
	"EarningsDate",
	"AfterMarketClose",
	"TradeCount",
	"TradeClusterCount",
	"TradeClusterBombCount",
}

// NewMarketCommand returns the "market" command group with all subcommands.
func NewMarketCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market",
		Short:   "Market-wide data commands",
		GroupID: "market",
		Args:    cobra.NoArgs,
		Long:    "market contains subcommands for querying market-wide data from VolumeLeaders, including price snapshots, earnings calendars, and exhaustion signals. None of the subcommands accept positional arguments; all filtering is done via flags.",
	}
	cmd.AddCommand(
		newSnapshotsCmd(),
		newEarningsCmd(),
		newExhaustionCmd(),
	)
	return cmd
}

// newSnapshotsCmd returns the "snapshots" subcommand.
func newSnapshotsCmd() *cobra.Command {
	return &cobra.Command{
		Use:        "snapshots",
		Short:      "Get current price snapshots for all symbols",
		Example:    "volumeleaders-agent market snapshots",
		Args:       cobra.NoArgs,
		Long:       "Retrieve current price snapshot data for all symbols tracked by VolumeLeaders, returning the latest available price and volume data. No date filtering is available; always returns the most recent data. Outputs compact JSON by default.",
		SuggestFor: []string{"snapshot", "snaps"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			vlClient, err := common.NewCommandClient(ctx)
			if err != nil {
				return err
			}

			var raw string
			if err := vlClient.PostJSON(ctx, "/Trades/GetAllSnapshots", struct{}{}, &raw); err != nil {
				slog.Error("failed to query snapshots", "error", err)
				return fmt.Errorf("query snapshots: %w", err)
			}

			snapshots := common.ParseSnapshotString(raw)
			return common.PrintJSON(cmd.OutOrStdout(), ctx, snapshots)
		},
	}
}

// newEarningsCmd returns the "earnings" subcommand.
func newEarningsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "earnings",
		Short:      "Query earnings calendar within a date range",
		Example:    "volumeleaders-agent market earnings --days 5",
		Args:       cobra.NoArgs,
		Long:       "Query the earnings calendar for a date range, showing tickers with earnings dates and associated trade activity counts. Requires --start-date and --end-date (or --days). Outputs compact JSON or CSV/TSV with --format.",
		SuggestFor: []string{"earning", "earings"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			startDate, endDate, err := common.RequiredDateRange(cmd)
			if err != nil {
				return err
			}

			fieldsValue, _ := cmd.Flags().GetString("fields")
			fields, err := common.OutputFields[models.Earnings](fieldsValue, marketEarningsDefaultFields)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")

			opts := common.DataTableOptions{
				Start:    0,
				Length:   -1,
				OrderCol: 0,
				OrderDir: "asc",
				Fields:   fields,
				Filters: map[string]string{
					"StartDate": startDate,
					"EndDate":   endDate,
				},
			}
			return common.RunDataTablesCommand[models.Earnings](cmd, "/Earnings/GetEarnings", datatables.EarningsColumns, opts, format, "query earnings")
		},
	}
	common.AddDateRangeFlags(cmd)
	common.AddOutputFormatFlags(cmd)
	cmd.Flags().String("fields", "", "Comma-separated fields to include (use 'all' for every field)")
	return cmd
}

// newExhaustionCmd returns the "exhaustion" subcommand.
func newExhaustionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "exhaustion",
		Short:      "Query exhaustion scores for a date",
		Example:    "volumeleaders-agent market exhaustion --date 2025-01-15",
		Args:       cobra.NoArgs,
		Long:       "Query exhaustion scores that indicate overbought or oversold market conditions based on institutional trade clustering patterns. Omit --date to query the current trading day. Outputs compact JSON with rank metrics at different lookback periods.",
		SuggestFor: []string{"exhaust", "exhastion"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			date, _ := cmd.Flags().GetString("date")

			vlClient, err := common.NewCommandClient(ctx)
			if err != nil {
				return err
			}

			payload := map[string]string{"Date": date}
			var score models.ExhaustionScore
			if err := vlClient.PostJSON(ctx, "/ExecutiveSummary/GetExhaustionScores", payload, &score); err != nil {
				slog.Error("failed to query exhaustion scores", "error", err)
				return fmt.Errorf("query exhaustion scores: %w", err)
			}

			return common.PrintJSON(cmd.OutOrStdout(), ctx, summarizeMarketExhaustion(score))
		},
	}
	cmd.Flags().String("date", "", "Date YYYY-MM-DD (empty for current day)")
	return cmd
}

// summarizeMarketExhaustion maps raw exhaustion score data to the compact CLI projection.
func summarizeMarketExhaustion(score models.ExhaustionScore) models.MarketExhaustion {
	return models.MarketExhaustion{
		DateKey:  score.DateKey,
		Rank:     score.ExhaustionScoreRank,
		Rank30D:  score.ExhaustionScoreRank30Day,
		Rank90D:  score.ExhaustionScoreRank90Day,
		Rank365D: score.ExhaustionScoreRank365Day,
	}
}
