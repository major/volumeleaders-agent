package volume

import (
	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
)

// volumeOptions holds flags shared by all volume subcommands.
type volumeOptions struct {
	Date     string
	Tickers  string
	Format   common.OutputFormat
	Start    int
	Length   int
	OrderCol int
	OrderDir common.OrderDirection
}

// NewVolumeCommand returns the "volume" command group with all subcommands.
func NewVolumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "volume",
		Short:   "Volume leaderboard commands",
		GroupID: "volume",
		Args:    cobra.NoArgs,
		Long:    "volume contains subcommands for querying volume leaderboard data from VolumeLeaders, showing tickers ranked by institutional trade volume. Filter by date and optional ticker list. All subcommands output compact JSON or CSV/TSV with --format.",
	}
	cmd.AddCommand(
		newInstitutionalCmd(),
		newAHInstitutionalCmd(),
		newTotalCmd(),
	)
	return cmd
}

// newInstitutionalCmd returns the "institutional" subcommand.
func newInstitutionalCmd() *cobra.Command {
	opts := volumeOptions{Length: 100, OrderCol: 1, OrderDir: "asc", Format: "json"}
	cmd := &cobra.Command{
		Use:        "institutional [tickers...]",
		Short:      "Query institutional volume leaderboard",
		Long:       "Query the regular-hours institutional volume leaderboard, ranking tickers by total institutional trade volume for a given date. Accepts optional ticker positional arguments to filter results; also accepts --tickers flag. Requires --date. Outputs compact JSON or CSV/TSV with --format. PREREQUISITES: choose a trading date in YYYY-MM-DD format. RECOVERY: if --date is missing or invalid, retry with an explicit trading day. NEXT STEPS: run trade dashboard for interesting single tickers first, then use trade list, trade levels, or trade clusters only when a dashboard section needs deeper detail.",
		Example:    "volumeleaders-agent volume institutional AAPL MSFT --date 2025-01-15",
		Args:       cobra.ArbitraryArgs,
		SuggestFor: []string{"inst", "insitutional"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runVolume(cmd, &opts,
				"/InstitutionalVolume/GetInstitutionalVolume",
				datatables.InstitutionalVolumeColumns)
		},
	}
	cmd.Flags().StringVarP(&opts.Date, "date", "d", "", "Date YYYY-MM-DD")
	common.AnnotateFlagGroup(cmd, "date", "Dates")
	common.MarkFlagRequired(cmd, "date")
	cmd.Flags().StringVarP(&opts.Tickers, "tickers", "t", "", "Comma-separated ticker symbols")
	common.AnnotateFlagGroup(cmd, "tickers", "Input")
	cmd.Flags().VarP(&opts.Format, "format", "f", "Output format: json, csv, or tsv")
	common.AnnotateFlagGroup(cmd, "format", "Output")
	common.AnnotateFlagEnum(cmd, "format", []string{"json", "csv", "tsv"})
	cmd.Flags().IntVar(&opts.Start, "start", 0, "DataTables start offset")
	common.AnnotateFlagGroup(cmd, "start", "Pagination")
	cmd.Flags().IntVarP(&opts.Length, "length", "l", 100, "Number of results")
	common.AnnotateFlagGroup(cmd, "length", "Pagination")
	cmd.Flags().IntVar(&opts.OrderCol, "order-col", 1, "Order column index")
	common.AnnotateFlagGroup(cmd, "order-col", "Pagination")
	cmd.Flags().Var(&opts.OrderDir, "order-dir", "Order direction")
	common.AnnotateFlagGroup(cmd, "order-dir", "Pagination")
	common.AnnotateFlagEnum(cmd, "order-dir", []string{"asc", "desc"})
	common.WrapValidation(cmd, &opts)
	return cmd
}

// newAHInstitutionalCmd returns the "ah-institutional" subcommand.
func newAHInstitutionalCmd() *cobra.Command {
	opts := volumeOptions{Length: 100, OrderCol: 1, OrderDir: "asc", Format: "json"}
	cmd := &cobra.Command{
		Use:        "ah-institutional [tickers...]",
		Short:      "Query after-hours institutional volume leaderboard",
		Long:       "Query the after-hours institutional volume leaderboard, ranking tickers by total institutional trade volume during after-hours sessions for a given date. Accepts optional ticker positional arguments; also accepts --tickers flag. Requires --date.",
		Example:    "volumeleaders-agent volume ah-institutional --date 2025-01-15",
		Args:       cobra.ArbitraryArgs,
		SuggestFor: []string{"ah", "afterhours"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runVolume(cmd, &opts,
				"/AHInstitutionalVolume/GetAHInstitutionalVolume",
				datatables.InstitutionalVolumeColumns)
		},
	}
	cmd.Flags().StringVarP(&opts.Date, "date", "d", "", "Date YYYY-MM-DD")
	common.AnnotateFlagGroup(cmd, "date", "Dates")
	common.MarkFlagRequired(cmd, "date")
	cmd.Flags().StringVarP(&opts.Tickers, "tickers", "t", "", "Comma-separated ticker symbols")
	common.AnnotateFlagGroup(cmd, "tickers", "Input")
	cmd.Flags().VarP(&opts.Format, "format", "f", "Output format: json, csv, or tsv")
	common.AnnotateFlagGroup(cmd, "format", "Output")
	common.AnnotateFlagEnum(cmd, "format", []string{"json", "csv", "tsv"})
	cmd.Flags().IntVar(&opts.Start, "start", 0, "DataTables start offset")
	common.AnnotateFlagGroup(cmd, "start", "Pagination")
	cmd.Flags().IntVarP(&opts.Length, "length", "l", 100, "Number of results")
	common.AnnotateFlagGroup(cmd, "length", "Pagination")
	cmd.Flags().IntVar(&opts.OrderCol, "order-col", 1, "Order column index")
	common.AnnotateFlagGroup(cmd, "order-col", "Pagination")
	cmd.Flags().Var(&opts.OrderDir, "order-dir", "Order direction")
	common.AnnotateFlagGroup(cmd, "order-dir", "Pagination")
	common.AnnotateFlagEnum(cmd, "order-dir", []string{"asc", "desc"})
	common.WrapValidation(cmd, &opts)
	return cmd
}

// newTotalCmd returns the "total" subcommand.
func newTotalCmd() *cobra.Command {
	opts := volumeOptions{Length: 100, OrderCol: 1, OrderDir: "asc", Format: "json"}
	cmd := &cobra.Command{
		Use:        "total [tickers...]",
		Short:      "Query total volume leaderboard",
		Long:       "Query the total volume leaderboard combining all session types, ranking tickers by total institutional trade volume for a given date. Accepts optional ticker positional arguments; also accepts --tickers flag. Requires --date.",
		Example:    "volumeleaders-agent volume total XLE --date 2025-01-15 --length 20",
		Args:       cobra.ArbitraryArgs,
		SuggestFor: []string{"totl", "all"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runVolume(cmd, &opts,
				"/TotalVolume/GetTotalVolume",
				datatables.TotalVolumeColumns)
		},
	}
	cmd.Flags().StringVarP(&opts.Date, "date", "d", "", "Date YYYY-MM-DD")
	common.AnnotateFlagGroup(cmd, "date", "Dates")
	common.MarkFlagRequired(cmd, "date")
	cmd.Flags().StringVarP(&opts.Tickers, "tickers", "t", "", "Comma-separated ticker symbols")
	common.AnnotateFlagGroup(cmd, "tickers", "Input")
	cmd.Flags().VarP(&opts.Format, "format", "f", "Output format: json, csv, or tsv")
	common.AnnotateFlagGroup(cmd, "format", "Output")
	common.AnnotateFlagEnum(cmd, "format", []string{"json", "csv", "tsv"})
	cmd.Flags().IntVar(&opts.Start, "start", 0, "DataTables start offset")
	common.AnnotateFlagGroup(cmd, "start", "Pagination")
	cmd.Flags().IntVarP(&opts.Length, "length", "l", 100, "Number of results")
	common.AnnotateFlagGroup(cmd, "length", "Pagination")
	cmd.Flags().IntVar(&opts.OrderCol, "order-col", 1, "Order column index")
	common.AnnotateFlagGroup(cmd, "order-col", "Pagination")
	cmd.Flags().Var(&opts.OrderDir, "order-dir", "Order direction")
	common.AnnotateFlagGroup(cmd, "order-dir", "Pagination")
	common.AnnotateFlagEnum(cmd, "order-dir", []string{"asc", "desc"})
	common.WrapValidation(cmd, &opts)
	return cmd
}

// runVolume is the shared handler for all volume subcommands.
func runVolume(cmd *cobra.Command, opts *volumeOptions, path string, columns []string) error {
	tickers := common.MultiTickerValue(cmd)

	dtOpts := common.NewDataTableOptions(common.DataTableRequestConfig{Start: opts.Start, Length: opts.Length, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Filters: map[string]string{"Date": opts.Date, "Tickers": tickers}})
	return common.RunDataTablesCommand[models.Trade](cmd, path, columns, dtOpts, opts.Format, "query volume data")
}
