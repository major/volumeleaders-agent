package alert

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
)

// alertConfigDefaultFields defines the default field subset for configs output.
var alertConfigDefaultFields = []string{
	"AlertConfigKey",
	"Name",
	"Tickers",
	"TradeConditions",
	"ClosingTradeConditions",
	"DarkPool",
	"Sweep",
	"OffsettingPrint",
	"PhantomPrint",
}

// NewAlertCommand returns the "alert" command group with all subcommands.
func NewAlertCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "alert",
		Short:   "Alert configuration commands",
		Long:    "alert contains subcommands for managing saved alert configurations that notify on institutional trade activity matching your criteria. Use create to define new alerts, edit to update existing ones, delete to remove them, and configs to list all saved configurations.",
		GroupID: "alerts",
		Args:    cobra.NoArgs,
	}
	cmd.AddCommand(
		newConfigsCmd(),
		newDeleteCmd(),
		newCreateCmd(),
		newEditCmd(),
	)
	return cmd
}

// newConfigsCmd returns the "configs" subcommand.
func newConfigsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "configs",
		Short:      "List saved alert configurations",
		Long:       "List all saved alert configurations with their keys, names, ticker filters, trade conditions, and notification settings. Outputs compact JSON or CSV/TSV with --format. Use --fields to select specific output fields.",
		Example:    "volumeleaders-agent alert configs",
		Args:       cobra.NoArgs,
		Aliases:    []string{"ls"},
		SuggestFor: []string{"config", "cfg"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			fieldsValue, _ := cmd.Flags().GetString("fields")
			fields, err := common.OutputFields[models.AlertConfig](fieldsValue, alertConfigDefaultFields)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")

			opts := common.DataTableOptions{
				Start:    0,
				Length:   -1,
				OrderCol: 1,
				OrderDir: "asc",
				Fields:   fields,
			}
			return common.RunDataTablesCommand[models.AlertConfig](cmd,
				"/AlertConfigs/GetAlertConfigs",
				datatables.AlertConfigColumns,
				opts, format, "query alert configs")
		},
	}
	common.AddOutputFormatFlags(cmd)
	cmd.Flags().String("fields", "", "Comma-separated fields to include (use 'all' for every field)")
	return cmd
}

// newDeleteCmd returns the "delete" subcommand.
func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "delete",
		Short:      "Delete an alert configuration",
		Long:       "Remove a saved alert configuration by its numeric key. Requires --key with the alert config key (visible in configs output). The deletion is permanent and cannot be undone.",
		Example:    "volumeleaders-agent alert delete --key 42",
		Args:       cobra.NoArgs,
		Aliases:    []string{"rm"},
		SuggestFor: []string{"del", "remove"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			key, _ := cmd.Flags().GetInt("key")
			ctx := cmd.Context()

			vlClient, err := common.NewCommandClient(ctx)
			if err != nil {
				return err
			}

			payload := map[string]int{"AlertConfigKey": key}
			var result any
			if err := vlClient.PostJSON(ctx, "/AlertConfigs/DeleteAlertConfig", payload, &result); err != nil {
				slog.Error("failed to delete alert config", "error", err)
				return fmt.Errorf("delete alert config: %w", err)
			}

			return common.PrintJSON(cmd.OutOrStdout(), ctx, result)
		},
	}
	cmd.Flags().Int("key", 0, "Alert config key to delete")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}

// newCreateCmd returns the "create" subcommand.
func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new alert configuration",
		Long:  "Create a new alert configuration with a name and filter settings for institutional trade activity. Requires --name. Specify filters such as trade rank, dollar thresholds, dark pool and sweep conditions, and ticker scope. Returns a success response with the new configuration key.",
		Example: `volumeleaders-agent alert create --name "Big trades" --tickers AAPL,MSFT --trade-rank-lte 5
volumeleaders-agent alert create --name "Dark pool sweeps" --sweep --dark-pool --trade-volume-gte 1000000`,
		Args:       cobra.NoArgs,
		Aliases:    []string{"new"},
		SuggestFor: []string{"crate", "creat"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAlertCreateEdit(cmd, 0)
		},
	}
	addAlertConfigFlags(cmd)
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

// newEditCmd returns the "edit" subcommand.
func newEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "edit",
		Short:      "Edit an existing alert configuration",
		Long:       "Modify an existing alert configuration identified by its numeric key. Requires --key with the alert config key. Specify only the fields you want to change; unspecified fields retain their current values.",
		Example:    `volumeleaders-agent alert edit --key 42 --name "Updated alert" --trade-rank-lte 3`,
		Args:       cobra.NoArgs,
		SuggestFor: []string{"edt", "modify"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			key, _ := cmd.Flags().GetInt("key")
			return runAlertCreateEdit(cmd, key)
		},
	}
	cmd.Flags().Int("key", 0, "Alert config key to edit")
	_ = cmd.MarkFlagRequired("key")
	addAlertConfigFlags(cmd)
	return cmd
}

// addAlertConfigFlags registers the shared flags for alert create/edit commands.
func addAlertConfigFlags(cmd *cobra.Command) {
	cmd.Flags().String("name", "", "Alert name (max 50 chars)")
	cmd.Flags().String("ticker-group", "AllTickers", "Ticker group (AllTickers or SelectedTickers)")
	cmd.Flags().String("tickers", "", "Comma-separated ticker symbols (max 500, used with SelectedTickers)")
	// Trade thresholds
	cmd.Flags().Int("trade-rank-lte", 0, "Trade rank <= (0=N/A, 1/3/5/10/25/50/100)")
	cmd.Flags().Int("trade-vcd-gte", 0, "Trade VCD >= (0=N/A, 99/100)")
	cmd.Flags().Int("trade-mult-gte", 0, "Trade multiplier >= (0=N/A, 5/10/25/50/100)")
	cmd.Flags().Int("trade-volume-gte", 0, "Trade volume >= (0=N/A, 1000000/2000000/5000000/10000000)")
	cmd.Flags().Int("trade-dollars-gte", 0, "Trade dollars >= (0=N/A, 1000000/10000000/...)")
	cmd.Flags().String("trade-conditions", "0", "Trade conditions (0=N/A, OBH/OBD/OSH/OSD combos)")
	cmd.Flags().Bool("dark-pool", false, "Dark pool filter")
	cmd.Flags().Bool("sweep", false, "Sweep filter")
	// Closing trade thresholds
	cmd.Flags().Int("closing-trade-rank-lte", 0, "Closing trade rank <=")
	cmd.Flags().Int("closing-trade-vcd-gte", 0, "Closing trade VCD >= (0/97/98/99/100)")
	cmd.Flags().Int("closing-trade-mult-gte", 0, "Closing trade multiplier >=")
	cmd.Flags().Int("closing-trade-volume-gte", 0, "Closing trade volume >=")
	cmd.Flags().Int("closing-trade-dollars-gte", 0, "Closing trade dollars >=")
	cmd.Flags().String("closing-trade-conditions", "0", "Closing trade conditions")
	// Trade cluster thresholds
	cmd.Flags().Int("cluster-rank-lte", 0, "Trade cluster rank <=")
	cmd.Flags().Int("cluster-vcd-gte", 0, "Trade cluster VCD >= (0/97/98/99/100)")
	cmd.Flags().Int("cluster-mult-gte", 0, "Trade cluster multiplier >=")
	cmd.Flags().Int("cluster-volume-gte", 0, "Trade cluster volume >=")
	cmd.Flags().Int("cluster-dollars-gte", 0, "Trade cluster dollars >=")
	// Cumulative institutional totals
	cmd.Flags().Int("total-rank-lte", 0, "Total rank <= (0/1/3/10/25/50/100)")
	cmd.Flags().Int("total-volume-gte", 0, "Total volume >=")
	cmd.Flags().Int("total-dollars-gte", 0, "Total dollars >=")
	// After-hours cumulative
	cmd.Flags().Int("ah-rank-lte", 0, "After-hours rank <=")
	cmd.Flags().Int("ah-volume-gte", 0, "After-hours volume >=")
	cmd.Flags().Int("ah-dollars-gte", 0, "After-hours dollars >=")
	// Special conditions
	cmd.Flags().Bool("offsetting-print", false, "Offsetting print filter")
	cmd.Flags().Bool("phantom-print", false, "Phantom print filter")
}

// buildAlertConfigFields maps CLI flag values to form field names for the
// multipart POST request.
func buildAlertConfigFields(cmd *cobra.Command, key int) map[string]string {
	// Auto-select SelectedTickers when tickers are specified but ticker-group
	// was left at the default.
	tickerGroup, _ := cmd.Flags().GetString("ticker-group")
	tickers, _ := cmd.Flags().GetString("tickers")
	if tickerGroup == "AllTickers" && tickers != "" {
		tickerGroup = "SelectedTickers"
	}

	name, _ := cmd.Flags().GetString("name")
	tradeRankLTE, _ := cmd.Flags().GetInt("trade-rank-lte")
	tradeVCDGTE, _ := cmd.Flags().GetInt("trade-vcd-gte")
	tradeMultGTE, _ := cmd.Flags().GetInt("trade-mult-gte")
	tradeVolumeGTE, _ := cmd.Flags().GetInt("trade-volume-gte")
	tradeDollarsGTE, _ := cmd.Flags().GetInt("trade-dollars-gte")
	tradeConditions, _ := cmd.Flags().GetString("trade-conditions")
	darkPool, _ := cmd.Flags().GetBool("dark-pool")
	sweep, _ := cmd.Flags().GetBool("sweep")
	closingTradeRankLTE, _ := cmd.Flags().GetInt("closing-trade-rank-lte")
	closingTradeVCDGTE, _ := cmd.Flags().GetInt("closing-trade-vcd-gte")
	closingTradeMultGTE, _ := cmd.Flags().GetInt("closing-trade-mult-gte")
	closingTradeVolumeGTE, _ := cmd.Flags().GetInt("closing-trade-volume-gte")
	closingTradeDollarsGTE, _ := cmd.Flags().GetInt("closing-trade-dollars-gte")
	closingTradeConditions, _ := cmd.Flags().GetString("closing-trade-conditions")
	clusterRankLTE, _ := cmd.Flags().GetInt("cluster-rank-lte")
	clusterVCDGTE, _ := cmd.Flags().GetInt("cluster-vcd-gte")
	clusterMultGTE, _ := cmd.Flags().GetInt("cluster-mult-gte")
	clusterVolumeGTE, _ := cmd.Flags().GetInt("cluster-volume-gte")
	clusterDollarsGTE, _ := cmd.Flags().GetInt("cluster-dollars-gte")
	totalRankLTE, _ := cmd.Flags().GetInt("total-rank-lte")
	totalVolumeGTE, _ := cmd.Flags().GetInt("total-volume-gte")
	totalDollarsGTE, _ := cmd.Flags().GetInt("total-dollars-gte")
	ahRankLTE, _ := cmd.Flags().GetInt("ah-rank-lte")
	ahVolumeGTE, _ := cmd.Flags().GetInt("ah-volume-gte")
	ahDollarsGTE, _ := cmd.Flags().GetInt("ah-dollars-gte")
	offsettingPrint, _ := cmd.Flags().GetBool("offsetting-print")
	phantomPrint, _ := cmd.Flags().GetBool("phantom-print")

	return map[string]string{
		"AlertConfigKey":         strconv.Itoa(key),
		"Name":                   name,
		"TickerGroup":            tickerGroup,
		"Tickers":                tickers,
		"TradeRankLTE":           strconv.Itoa(tradeRankLTE),
		"TradeVCDGTE":            strconv.Itoa(tradeVCDGTE),
		"TradeMultGTE":           strconv.Itoa(tradeMultGTE),
		"TradeVolumeGTE":         strconv.Itoa(tradeVolumeGTE),
		"TradeDollarsGTE":        strconv.Itoa(tradeDollarsGTE),
		"TradeConditions":        tradeConditions,
		"DarkPool":               strconv.FormatBool(darkPool),
		"Sweep":                  strconv.FormatBool(sweep),
		"ClosingTradeRankLTE":    strconv.Itoa(closingTradeRankLTE),
		"ClosingTradeVCDGTE":     strconv.Itoa(closingTradeVCDGTE),
		"ClosingTradeMultGTE":    strconv.Itoa(closingTradeMultGTE),
		"ClosingTradeVolumeGTE":  strconv.Itoa(closingTradeVolumeGTE),
		"ClosingTradeDollarsGTE": strconv.Itoa(closingTradeDollarsGTE),
		"ClosingTradeConditions": closingTradeConditions,
		"TradeClusterRankLTE":    strconv.Itoa(clusterRankLTE),
		"TradeClusterVCDGTE":     strconv.Itoa(clusterVCDGTE),
		"TradeClusterMultGTE":    strconv.Itoa(clusterMultGTE),
		"TradeClusterVolumeGTE":  strconv.Itoa(clusterVolumeGTE),
		"TradeClusterDollarsGTE": strconv.Itoa(clusterDollarsGTE),
		"TotalRankLTE":           strconv.Itoa(totalRankLTE),
		"TotalVolumeGTE":         strconv.Itoa(totalVolumeGTE),
		"TotalDollarsGTE":        strconv.Itoa(totalDollarsGTE),
		"AHRankLTE":              strconv.Itoa(ahRankLTE),
		"AHVolumeGTE":            strconv.Itoa(ahVolumeGTE),
		"AHDollarsGTE":           strconv.Itoa(ahDollarsGTE),
		"OffsettingPrint":        strconv.FormatBool(offsettingPrint),
		"PhantomPrint":           strconv.FormatBool(phantomPrint),
	}
}

// runAlertCreateEdit is the shared handler for create and edit subcommands.
func runAlertCreateEdit(cmd *cobra.Command, key int) error {
	ctx := cmd.Context()
	vlClient, err := common.NewCommandClient(ctx)
	if err != nil {
		return err
	}

	fields := buildAlertConfigFields(cmd, key)
	if err := vlClient.PostMultipart(ctx, "/AlertConfig", fields); err != nil {
		slog.Error("failed to save alert config", "error", err)
		return fmt.Errorf("save alert config: %w", err)
	}

	action := "created"
	if key != 0 {
		action = "updated"
	}
	return common.PrintJSON(cmd.OutOrStdout(), ctx, map[string]any{"success": true, "action": action, "key": key})
}
