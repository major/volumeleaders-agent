package common

import (
	"strconv"

	"github.com/spf13/cobra"
)

// AddDateRangeFlags registers required date range flags.
func AddDateRangeFlags(cmd *cobra.Command) {
	cmd.Flags().String("start-date", "", "Start date YYYY-MM-DD (required unless --days is set)")
	cmd.Flags().String("end-date", "", "End date YYYY-MM-DD (required unless --days is set)")
	cmd.Flags().Int("days", 0, "Look back this many days from --end-date or today")
}

// AddOptionalDateRangeFlags registers date range flags for commands with defaults.
func AddOptionalDateRangeFlags(cmd *cobra.Command) {
	cmd.Flags().String("start-date", "", "Start date YYYY-MM-DD (default: auto)")
	cmd.Flags().String("end-date", "", "End date YYYY-MM-DD (default: today)")
	cmd.Flags().Int("days", 0, "Look back this many days from --end-date or today")
}

// AddVolumeRangeFlags registers the standard volume range flags.
func AddVolumeRangeFlags(cmd *cobra.Command) {
	cmd.Flags().Int("min-volume", 0, "Minimum volume")
	cmd.Flags().Int("max-volume", 2000000000, "Maximum volume")
}

// AddPriceRangeFlags registers the standard price range flags.
func AddPriceRangeFlags(cmd *cobra.Command) {
	addFloat64Flag(cmd, "min-price", 0, "Minimum price")
	addFloat64Flag(cmd, "max-price", 100000, "Maximum price")
}

// AddDollarRangeFlags registers the standard dollar value range flags.
func AddDollarRangeFlags(cmd *cobra.Command, minDefault float64) {
	addFloat64Flag(cmd, "min-dollars", minDefault, "Minimum dollar value")
	addFloat64Flag(cmd, "max-dollars", 30000000000, "Maximum dollar value")
}

// AddPaginationFlags registers the standard DataTables pagination and ordering flags.
func AddPaginationFlags(cmd *cobra.Command, length, orderCol int, orderDir string) {
	cmd.Flags().Int("start", 0, "DataTables start offset")
	cmd.Flags().Int("length", length, "Number of results")
	cmd.Flags().Int("order-col", orderCol, "Order column index")
	cmd.Flags().String("order-dir", orderDir, "Order direction")
}

// AddOutputFormatFlags registers the standard tabular output format flag.
func AddOutputFormatFlags(cmd *cobra.Command) {
	cmd.Flags().String("format", "json", "Output format: json, csv, or tsv")
}

// AddTickersFlag registers the multi-ticker input flag.
func AddTickersFlag(cmd *cobra.Command) {
	cmd.Flags().String("tickers", "", "Comma-separated ticker symbols")
}

// AddTickerFlag registers the single-ticker input flag.
func AddTickerFlag(cmd *cobra.Command) {
	cmd.Flags().String("ticker", "", "Ticker symbol")
}

func addFloat64Flag(cmd *cobra.Command, name string, value float64, usage string) {
	cmd.Flags().Float64(name, value, usage)
	cmd.Flags().Lookup(name).DefValue = strconv.FormatFloat(value, 'f', -1, 64)
}
