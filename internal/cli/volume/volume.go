package volume

import (
	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
)

// NewVolumeCommand returns the "volume" command group with all subcommands.
func NewVolumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:      "volume",
		Short:    "Volume leaderboard commands",
		GroupID:  "volume",
		Args:     cobra.NoArgs,
		Long:     "volume contains subcommands for querying volume leaderboard data from VolumeLeaders, showing tickers ranked by institutional trade volume. Filter by date and optional ticker list. All subcommands output compact JSON or CSV/TSV with --format.",
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
	cmd := &cobra.Command{
		Use:       "institutional [tickers...]",
		Short:     "Query institutional volume leaderboard",
		Long:      "Query the regular-hours institutional volume leaderboard, ranking tickers by total institutional trade volume for a given date. Accepts optional ticker positional arguments to filter results; also accepts --tickers flag. Requires --date. Outputs compact JSON or CSV/TSV with --format.",
		Example:   "volumeleaders-agent volume institutional AAPL MSFT --date 2025-01-15",
		Args:      cobra.ArbitraryArgs,
		SuggestFor: []string{"inst", "insitutional"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runVolume(cmd,
				"/InstitutionalVolume/GetInstitutionalVolume",
				datatables.InstitutionalVolumeColumns)
		},
	}
	addVolumeFlags(cmd)
	return cmd
}

// newAHInstitutionalCmd returns the "ah-institutional" subcommand.
func newAHInstitutionalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "ah-institutional [tickers...]",
		Short:     "Query after-hours institutional volume leaderboard",
		Long:      "Query the after-hours institutional volume leaderboard, ranking tickers by total institutional trade volume during after-hours sessions for a given date. Accepts optional ticker positional arguments; also accepts --tickers flag. Requires --date.",
		Example:   "volumeleaders-agent volume ah-institutional --date 2025-01-15",
		Args:      cobra.ArbitraryArgs,
		SuggestFor: []string{"ah", "afterhours"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runVolume(cmd,
				"/AHInstitutionalVolume/GetAHInstitutionalVolume",
				datatables.InstitutionalVolumeColumns)
		},
	}
	addVolumeFlags(cmd)
	return cmd
}

// newTotalCmd returns the "total" subcommand.
func newTotalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "total [tickers...]",
		Short:     "Query total volume leaderboard",
		Long:      "Query the total volume leaderboard combining all session types, ranking tickers by total institutional trade volume for a given date. Accepts optional ticker positional arguments; also accepts --tickers flag. Requires --date.",
		Example:   "volumeleaders-agent volume total XLE --date 2025-01-15 --length 20",
		Args:      cobra.ArbitraryArgs,
		SuggestFor: []string{"totl", "all"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runVolume(cmd,
				"/TotalVolume/GetTotalVolume",
				datatables.TotalVolumeColumns)
		},
	}
	addVolumeFlags(cmd)
	return cmd
}

// addVolumeFlags registers the shared flag set used by all volume subcommands.
func addVolumeFlags(cmd *cobra.Command) {
	cmd.Flags().String("date", "", "Date YYYY-MM-DD")
	_ = cmd.MarkFlagRequired("date")
	common.AddTickersFlag(cmd)
	common.AddOutputFormatFlags(cmd)
	common.AddPaginationFlags(cmd, 100, 1, "asc")
}

// runVolume is the shared handler for all volume subcommands.
func runVolume(cmd *cobra.Command, path string, columns []string) error {
	date, _ := cmd.Flags().GetString("date")
	tickers := common.MultiTickerValue(cmd)
	start, _ := cmd.Flags().GetInt("start")
	length, _ := cmd.Flags().GetInt("length")
	orderCol, _ := cmd.Flags().GetInt("order-col")
	orderDir, _ := cmd.Flags().GetString("order-dir")
	format, _ := cmd.Flags().GetString("format")

	opts := common.DataTableOptions{
		Start:    start,
		Length:   length,
		OrderCol: orderCol,
		OrderDir: orderDir,
		Filters: map[string]string{
			"Date":    date,
			"Tickers": tickers,
		},
	}
	return common.RunDataTablesCommand[models.Trade](cmd, path, columns, opts, format, "query volume data")
}
