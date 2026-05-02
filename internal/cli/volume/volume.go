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
		Use:     "institutional",
		Short:   "Query institutional volume leaderboard",
		Example: "volumeleaders-agent volume institutional AAPL MSFT --date 2025-01-15",
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
		Use:     "ah-institutional",
		Short:   "Query after-hours institutional volume leaderboard",
		Example: "volumeleaders-agent volume ah-institutional --date 2025-01-15",
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
		Use:     "total",
		Short:   "Query total volume leaderboard",
		Example: "volumeleaders-agent volume total XLE --date 2025-01-15 --length 20",
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
