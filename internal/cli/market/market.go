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

// earningsOptions holds flags for the "market earnings" subcommand.
type earningsOptions struct {
	StartDate string              `flag:"start-date" flaggroup:"Dates" flagshort:"s" flagdescr:"Start date YYYY-MM-DD (required unless --days is set)"`
	EndDate   string              `flag:"end-date" flaggroup:"Dates" flagshort:"e" flagdescr:"End date YYYY-MM-DD (required unless --days is set)"`
	Days      int                 `flag:"days" flaggroup:"Dates" flagshort:"d" flagdescr:"Look back this many days from --end-date or today"`
	Format    common.OutputFormat `flag:"format" flaggroup:"Output" flagshort:"f" flagdescr:"Output format: json, csv, or tsv" default:"json"`
	Fields    string              `flag:"fields" flaggroup:"Output" flagdescr:"Comma-separated fields to include (use 'all' for every field)"`
}

// exhaustionOptions holds flags for the "market exhaustion" subcommand.
type exhaustionOptions struct {
	Date string `flag:"date" flaggroup:"Dates" flagshort:"d" flagdescr:"Date YYYY-MM-DD (empty for current day)"`
}

// NewMarketCommand returns the "market" command group with all subcommands.
func NewMarketCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market",
		Short:   "Market-wide data commands",
		GroupID: "market",
		Args:    cobra.NoArgs,
		Long:    "market contains subcommands for querying market-wide data from VolumeLeaders, including earnings calendars and exhaustion signals. None of the subcommands accept positional arguments; all filtering is done via flags.",
	}
	cmd.AddCommand(
		newEarningsCmd(),
		newExhaustionCmd(),
	)
	return cmd
}

// newEarningsCmd returns the "earnings" subcommand.
func newEarningsCmd() *cobra.Command {
	opts := &earningsOptions{}
	cmd := &cobra.Command{
		Use:        "earnings",
		Short:      "Query earnings calendar within a date range",
		Example:    "volumeleaders-agent market earnings --days 5",
		Args:       cobra.NoArgs,
		Long:       "Query the earnings calendar for a date range, showing tickers with earnings dates and associated trade activity counts. Requires --start-date and --end-date (or --days). Outputs compact JSON or CSV/TSV with --format. PREREQUISITES: provide a date range with --days or explicit start and end dates. RECOVERY: if date validation fails, use --days N for the fastest retry or provide both --start-date and --end-date. NEXT STEPS: run trade list for tickers near earnings, then market exhaustion for broader reversal context.",
		SuggestFor: []string{"earning", "earings"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			startDate, endDate, err := common.ResolveDateRange(cmd, 0, true)
			if err != nil {
				return err
			}

			fields, err := common.OutputFields[models.Earnings](opts.Fields, marketEarningsDefaultFields)
			if err != nil {
				return err
			}

			dtOpts := common.NewDataTableOptions(common.DataTableRequestConfig{Start: 0, Length: -1, OrderCol: 0, OrderDir: common.OrderDirectionASC, Fields: fields, Filters: map[string]string{"StartDate": startDate, "EndDate": endDate}})
			return common.RunDataTablesCommand[models.Earnings](cmd, "/Earnings/GetEarnings", datatables.EarningsColumns, dtOpts, opts.Format, "query earnings")
		},
	}
	common.BindOrPanic(cmd, opts, "earnings options")
	return cmd
}

// newExhaustionCmd returns the "exhaustion" subcommand.
func newExhaustionCmd() *cobra.Command {
	opts := &exhaustionOptions{}
	cmd := &cobra.Command{
		Use:        "exhaustion",
		Short:      "Query exhaustion scores for a date",
		Example:    "volumeleaders-agent market exhaustion --date 2025-01-15",
		Args:       cobra.NoArgs,
		Long:       "Query exhaustion scores that indicate overbought or oversold market conditions based on institutional trade clustering patterns. Omit --date to query the current trading day. Outputs compact JSON with rank metrics at different lookback periods.",
		SuggestFor: []string{"exhaust", "exhastion"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			vlClient, err := common.NewCommandClient(ctx)
			if err != nil {
				return err
			}

			payload := map[string]string{"Date": opts.Date}
			var score models.ExhaustionScore
			if err := vlClient.PostJSON(ctx, "/ExecutiveSummary/GetExhaustionScores", payload, &score); err != nil {
				slog.Error("failed to query exhaustion scores", "error", err)
				return fmt.Errorf("query exhaustion scores: %w", err)
			}

			return common.PrintJSON(cmd.OutOrStdout(), ctx, summarizeMarketExhaustion(score))
		},
	}
	common.BindOrPanic(cmd, opts, "exhaustion options")
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
