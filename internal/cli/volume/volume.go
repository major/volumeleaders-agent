package volume

import (
	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
)

// volumeOptions holds flags shared by all volume subcommands.
type volumeOptions struct {
	Date     string                `flag:"date" flaggroup:"Dates" flagshort:"d" flagdescr:"Date YYYY-MM-DD" flagrequired:"true"`
	Tickers  string                `flag:"tickers" flaggroup:"Input" flagshort:"t" flagdescr:"Comma-separated ticker symbols"`
	Format   common.OutputFormat   `flag:"format" flaggroup:"Output" flagshort:"f" default:"json" flagdescr:"Output format: json, csv, or tsv"`
	Start    int                   `flag:"start" flaggroup:"Pagination" flagdescr:"DataTables start offset"`
	Length   int                   `flag:"length" flaggroup:"Pagination" flagshort:"l" flagdescr:"Number of results"`
	OrderCol int                   `flag:"order-col" flaggroup:"Pagination" flagdescr:"Order column index"`
	OrderDir common.OrderDirection `flag:"order-dir" flaggroup:"Pagination" flagdescr:"Order direction"`
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
	opts := volumeOptions{Length: 100, OrderCol: 1, OrderDir: "asc"}
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
	common.BindOrPanic(cmd, &opts, "institutional")
	return cmd
}

// newAHInstitutionalCmd returns the "ah-institutional" subcommand.
func newAHInstitutionalCmd() *cobra.Command {
	opts := volumeOptions{Length: 100, OrderCol: 1, OrderDir: "asc"}
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
	common.BindOrPanic(cmd, &opts, "ah-institutional")
	return cmd
}

// newTotalCmd returns the "total" subcommand.
func newTotalCmd() *cobra.Command {
	opts := volumeOptions{Length: 100, OrderCol: 1, OrderDir: "asc"}
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
	common.BindOrPanic(cmd, &opts, "total")
	return cmd
}

// runVolume is the shared handler for all volume subcommands.
func runVolume(cmd *cobra.Command, opts *volumeOptions, path string, columns []string) error {
	tickers := common.MultiTickerValue(cmd)

	dtOpts := common.DataTableOptions{
		Start:    opts.Start,
		Length:   opts.Length,
		OrderCol: opts.OrderCol,
		OrderDir: opts.OrderDir,
		Filters: map[string]string{
			"Date":    opts.Date,
			"Tickers": tickers,
		},
	}
	return common.RunDataTablesCommand[models.Trade](cmd, path, columns, dtOpts, opts.Format, "query volume data")
}
