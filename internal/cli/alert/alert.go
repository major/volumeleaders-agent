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

type alertTickerGroup string

const (
	alertTickerGroupAll      alertTickerGroup = "AllTickers"
	alertTickerGroupSelected alertTickerGroup = "SelectedTickers"
)

// Set implements pflag.Value for alertTickerGroup.
func (v *alertTickerGroup) Set(value string) error {
	switch alertTickerGroup(value) {
	case alertTickerGroupAll, alertTickerGroupSelected:
		*v = alertTickerGroup(value)
		return nil
	default:
		return fmt.Errorf("invalid value %q, expected one of AllTickers, SelectedTickers", value)
	}
}

// String implements pflag.Value for alertTickerGroup.
func (v alertTickerGroup) String() string {
	return string(v)
}

// Type implements pflag.Value for alertTickerGroup.
func (v alertTickerGroup) Type() string {
	return "string"
}

// alertConfigsOptions holds flags for the "alert configs" subcommand.
type alertConfigsOptions struct {
	Format common.OutputFormat
	Fields string
}

// alertDeleteOptions holds flags for the "alert delete" subcommand.
type alertDeleteOptions struct {
	Key int
}

// alertConfigFlags holds the shared flag set for alert create/edit commands.
type alertConfigFlags struct {
	Name                   string
	TickerGroup            alertTickerGroup
	Tickers                string
	TradeRankLTE           int
	TradeVCDGTE            int
	TradeMultGTE           int
	TradeVolumeGTE         int
	TradeDollarsGTE        int
	TradeConditions        string
	DarkPool               bool
	Sweep                  bool
	ClosingTradeRankLTE    int
	ClosingTradeVCDGTE     int
	ClosingTradeMultGTE    int
	ClosingTradeVolumeGTE  int
	ClosingTradeDollarsGTE int
	ClosingTradeConditions string
	ClusterRankLTE         int
	ClusterVCDGTE          int
	ClusterMultGTE         int
	ClusterVolumeGTE       int
	ClusterDollarsGTE      int
	TotalRankLTE           int
	TotalVolumeGTE         int
	TotalDollarsGTE        int
	AHRankLTE              int
	AHVolumeGTE            int
	AHDollarsGTE           int
	OffsettingPrint        bool
	PhantomPrint           bool
}

// alertCreateOptions holds flags for the "alert create" subcommand.
type alertCreateOptions struct {
	alertConfigFlags
}

// alertEditOptions holds flags for the "alert edit" subcommand.
type alertEditOptions struct {
	Key int
	alertConfigFlags
}

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

// registerAlertConfigFlags registers the shared alert configuration flags on cmd.
// Both create and edit commands embed alertConfigFlags, so this avoids duplicating
// 30 flag registrations across both commands.
func registerAlertConfigFlags(cmd *cobra.Command, opts *alertConfigFlags) {
	f := cmd.Flags()

	// Basic
	f.StringVar(&opts.Name, "name", "", "Alert name (max 50 chars)")
	f.Var(&opts.TickerGroup, "ticker-group", "Ticker group: AllTickers or SelectedTickers")
	f.StringVarP(&opts.Tickers, "tickers", "t", "", "Comma-separated ticker symbols (max 500, used with SelectedTickers)")

	// Trade Filters
	f.IntVar(&opts.TradeRankLTE, "trade-rank-lte", 0, "Trade rank <= (0=N/A, 1/3/5/10/25/50/100)")
	f.IntVar(&opts.TradeVCDGTE, "trade-vcd-gte", 0, "Trade VCD >= (0=N/A, 99/100)")
	f.IntVar(&opts.TradeMultGTE, "trade-mult-gte", 0, "Trade multiplier >= (0=N/A, 5/10/25/50/100)")
	f.IntVar(&opts.TradeVolumeGTE, "trade-volume-gte", 0, "Trade volume >= (0=N/A, 1000000/2000000/5000000/10000000)")
	f.IntVar(&opts.TradeDollarsGTE, "trade-dollars-gte", 0, "Trade dollars >= (0=N/A, 1000000/10000000/...)")
	f.StringVar(&opts.TradeConditions, "trade-conditions", opts.TradeConditions, "Trade conditions (0=N/A, OBH/OBD/OSH/OSD combos)")
	f.BoolVar(&opts.DarkPool, "dark-pool", false, "Dark pool filter")
	f.BoolVar(&opts.Sweep, "sweep", false, "Sweep filter")
	f.BoolVar(&opts.OffsettingPrint, "offsetting-print", false, "Offsetting print filter")
	f.BoolVar(&opts.PhantomPrint, "phantom-print", false, "Phantom print filter")

	// Closing Filters
	f.IntVar(&opts.ClosingTradeRankLTE, "closing-trade-rank-lte", 0, "Closing trade rank <=")
	f.IntVar(&opts.ClosingTradeVCDGTE, "closing-trade-vcd-gte", 0, "Closing trade VCD >= (0/97/98/99/100)")
	f.IntVar(&opts.ClosingTradeMultGTE, "closing-trade-mult-gte", 0, "Closing trade multiplier >=")
	f.IntVar(&opts.ClosingTradeVolumeGTE, "closing-trade-volume-gte", 0, "Closing trade volume >=")
	f.IntVar(&opts.ClosingTradeDollarsGTE, "closing-trade-dollars-gte", 0, "Closing trade dollars >=")
	f.StringVar(&opts.ClosingTradeConditions, "closing-trade-conditions", opts.ClosingTradeConditions, "Closing trade conditions")

	// Cluster Filters
	f.IntVar(&opts.ClusterRankLTE, "cluster-rank-lte", 0, "Trade cluster rank <=")
	f.IntVar(&opts.ClusterVCDGTE, "cluster-vcd-gte", 0, "Trade cluster VCD >= (0/97/98/99/100)")
	f.IntVar(&opts.ClusterMultGTE, "cluster-mult-gte", 0, "Trade cluster multiplier >=")
	f.IntVar(&opts.ClusterVolumeGTE, "cluster-volume-gte", 0, "Trade cluster volume >=")
	f.IntVar(&opts.ClusterDollarsGTE, "cluster-dollars-gte", 0, "Trade cluster dollars >=")

	// Total Filters
	f.IntVar(&opts.TotalRankLTE, "total-rank-lte", 0, "Total rank <= (0/1/3/10/25/50/100)")
	f.IntVar(&opts.TotalVolumeGTE, "total-volume-gte", 0, "Total volume >=")
	f.IntVar(&opts.TotalDollarsGTE, "total-dollars-gte", 0, "Total dollars >=")

	// After-Hours Filters
	f.IntVar(&opts.AHRankLTE, "ah-rank-lte", 0, "After-hours rank <=")
	f.IntVar(&opts.AHVolumeGTE, "ah-volume-gte", 0, "After-hours volume >=")
	f.IntVar(&opts.AHDollarsGTE, "ah-dollars-gte", 0, "After-hours dollars >=")

	// Group annotations
	common.AnnotateFlagGroup(cmd, "name", "Basic")
	common.AnnotateFlagGroup(cmd, "ticker-group", "Basic")
	common.AnnotateFlagGroup(cmd, "tickers", "Basic")

	common.AnnotateFlagGroup(cmd, "trade-rank-lte", "Trade Filters")
	common.AnnotateFlagGroup(cmd, "trade-vcd-gte", "Trade Filters")
	common.AnnotateFlagGroup(cmd, "trade-mult-gte", "Trade Filters")
	common.AnnotateFlagGroup(cmd, "trade-volume-gte", "Trade Filters")
	common.AnnotateFlagGroup(cmd, "trade-dollars-gte", "Trade Filters")
	common.AnnotateFlagGroup(cmd, "trade-conditions", "Trade Filters")
	common.AnnotateFlagGroup(cmd, "dark-pool", "Trade Filters")
	common.AnnotateFlagGroup(cmd, "sweep", "Trade Filters")
	common.AnnotateFlagGroup(cmd, "offsetting-print", "Trade Filters")
	common.AnnotateFlagGroup(cmd, "phantom-print", "Trade Filters")

	common.AnnotateFlagGroup(cmd, "closing-trade-rank-lte", "Closing Filters")
	common.AnnotateFlagGroup(cmd, "closing-trade-vcd-gte", "Closing Filters")
	common.AnnotateFlagGroup(cmd, "closing-trade-mult-gte", "Closing Filters")
	common.AnnotateFlagGroup(cmd, "closing-trade-volume-gte", "Closing Filters")
	common.AnnotateFlagGroup(cmd, "closing-trade-dollars-gte", "Closing Filters")
	common.AnnotateFlagGroup(cmd, "closing-trade-conditions", "Closing Filters")

	common.AnnotateFlagGroup(cmd, "cluster-rank-lte", "Cluster Filters")
	common.AnnotateFlagGroup(cmd, "cluster-vcd-gte", "Cluster Filters")
	common.AnnotateFlagGroup(cmd, "cluster-mult-gte", "Cluster Filters")
	common.AnnotateFlagGroup(cmd, "cluster-volume-gte", "Cluster Filters")
	common.AnnotateFlagGroup(cmd, "cluster-dollars-gte", "Cluster Filters")

	common.AnnotateFlagGroup(cmd, "total-rank-lte", "Total Filters")
	common.AnnotateFlagGroup(cmd, "total-volume-gte", "Total Filters")
	common.AnnotateFlagGroup(cmd, "total-dollars-gte", "Total Filters")

	common.AnnotateFlagGroup(cmd, "ah-rank-lte", "After-Hours Filters")
	common.AnnotateFlagGroup(cmd, "ah-volume-gte", "After-Hours Filters")
	common.AnnotateFlagGroup(cmd, "ah-dollars-gte", "After-Hours Filters")

	// Enum annotation
	common.AnnotateFlagEnum(cmd, "ticker-group", []string{"AllTickers", "SelectedTickers"})
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
	opts := &alertConfigsOptions{
		Format: common.OutputFormatJSON,
	}
	cmd := &cobra.Command{
		Use:        "configs",
		Short:      "List saved alert configurations",
		Long:       "List all saved alert configurations with their keys, names, ticker filters, trade conditions, and notification settings. Outputs compact JSON or CSV/TSV with --format. Use --fields to select specific output fields.",
		Example:    "volumeleaders-agent alert configs",
		Args:       cobra.NoArgs,
		Aliases:    []string{"ls"},
		SuggestFor: []string{"config", "cfg"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			fields, err := common.OutputFields[models.AlertConfig](opts.Fields, alertConfigDefaultFields)
			if err != nil {
				return err
			}

			dtOpts := common.DataTableOptions{
				Start:    0,
				Length:   -1,
				OrderCol: 1,
				OrderDir: "asc",
				Fields:   fields,
			}
			return common.RunDataTablesCommand[models.AlertConfig](cmd,
				"/AlertConfigs/GetAlertConfigs",
				datatables.AlertConfigColumns,
				dtOpts, opts.Format, "query alert configs")
		},
	}

	f := cmd.Flags()
	f.VarP(&opts.Format, "format", "f", "Output format: json, csv, or tsv")
	f.StringVar(&opts.Fields, "fields", "", "Comma-separated fields to include (use 'all' for every field)")

	common.AnnotateFlagGroup(cmd, "format", "Output")
	common.AnnotateFlagGroup(cmd, "fields", "Output")
	common.AnnotateFlagEnum(cmd, "format", []string{"json", "csv", "tsv"})
	common.WrapValidation(cmd, opts)

	return cmd
}

// newDeleteCmd returns the "delete" subcommand.
func newDeleteCmd() *cobra.Command {
	opts := &alertDeleteOptions{}
	cmd := &cobra.Command{
		Use:        "delete",
		Short:      "Delete an alert configuration",
		Long:       "Remove a saved alert configuration by its numeric key. Requires --key with the alert config key (visible in configs output). The deletion is permanent and cannot be undone.",
		Example:    "volumeleaders-agent alert delete --key 42",
		Args:       cobra.NoArgs,
		Aliases:    []string{"rm"},
		SuggestFor: []string{"del", "remove"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			vlClient, err := common.NewCommandClient(ctx)
			if err != nil {
				return err
			}

			payload := map[string]int{"AlertConfigKey": opts.Key}
			var result any
			if err := vlClient.PostJSON(ctx, "/AlertConfigs/DeleteAlertConfig", payload, &result); err != nil {
				slog.Error("failed to delete alert config", "error", err)
				return fmt.Errorf("delete alert config: %w", err)
			}

			return common.PrintJSON(cmd.OutOrStdout(), ctx, result)
		},
	}

	f := cmd.Flags()
	f.IntVarP(&opts.Key, "key", "k", 0, "Alert config key to delete")

	common.AnnotateFlagGroup(cmd, "key", "Input")
	common.MarkFlagRequired(cmd, "key")
	common.WrapValidation(cmd, opts)

	return cmd
}

// newCreateCmd returns the "create" subcommand.
func newCreateCmd() *cobra.Command {
	opts := &alertCreateOptions{}
	opts.TickerGroup = alertTickerGroupAll
	opts.TradeConditions = "0"
	opts.ClosingTradeConditions = "0"
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
			return runAlertCreateEdit(cmd, &opts.alertConfigFlags, 0)
		},
	}

	registerAlertConfigFlags(cmd, &opts.alertConfigFlags)
	common.MarkFlagRequired(cmd, "name")
	common.WrapValidation(cmd, opts)

	return cmd
}

// newEditCmd returns the "edit" subcommand.
func newEditCmd() *cobra.Command {
	opts := &alertEditOptions{}
	opts.TickerGroup = alertTickerGroupAll
	opts.TradeConditions = "0"
	opts.ClosingTradeConditions = "0"
	cmd := &cobra.Command{
		Use:        "edit",
		Short:      "Edit an existing alert configuration",
		Long:       "Modify an existing alert configuration identified by its numeric key. Requires --key with the alert config key. Specify the fields you want to set; unspecified fields are replaced with their default values.",
		Example:    `volumeleaders-agent alert edit --key 42 --name "Updated alert" --trade-rank-lte 3`,
		Args:       cobra.NoArgs,
		SuggestFor: []string{"edt", "modify"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAlertCreateEdit(cmd, &opts.alertConfigFlags, opts.Key)
		},
	}

	f := cmd.Flags()
	f.IntVarP(&opts.Key, "key", "k", 0, "Alert config key to edit")

	common.AnnotateFlagGroup(cmd, "key", "Input")
	common.MarkFlagRequired(cmd, "key")

	registerAlertConfigFlags(cmd, &opts.alertConfigFlags)
	common.WrapValidation(cmd, opts)

	return cmd
}

// buildAlertConfigFields maps struct field values to form field names for the
// multipart POST request.
func buildAlertConfigFields(opts *alertConfigFlags, key int, tickerGroupChanged bool) map[string]string {
	// Auto-select SelectedTickers when tickers are specified but ticker-group
	// was left at the default and not explicitly provided by the caller.
	tickerGroup := opts.TickerGroup
	if !tickerGroupChanged && tickerGroup == alertTickerGroupAll && opts.Tickers != "" {
		tickerGroup = alertTickerGroupSelected
	}

	return map[string]string{
		"AlertConfigKey":         strconv.Itoa(key),
		"Name":                   opts.Name,
		"TickerGroup":            string(tickerGroup),
		"Tickers":                opts.Tickers,
		"TradeRankLTE":           strconv.Itoa(opts.TradeRankLTE),
		"TradeVCDGTE":            strconv.Itoa(opts.TradeVCDGTE),
		"TradeMultGTE":           strconv.Itoa(opts.TradeMultGTE),
		"TradeVolumeGTE":         strconv.Itoa(opts.TradeVolumeGTE),
		"TradeDollarsGTE":        strconv.Itoa(opts.TradeDollarsGTE),
		"TradeConditions":        opts.TradeConditions,
		"DarkPool":               strconv.FormatBool(opts.DarkPool),
		"Sweep":                  strconv.FormatBool(opts.Sweep),
		"ClosingTradeRankLTE":    strconv.Itoa(opts.ClosingTradeRankLTE),
		"ClosingTradeVCDGTE":     strconv.Itoa(opts.ClosingTradeVCDGTE),
		"ClosingTradeMultGTE":    strconv.Itoa(opts.ClosingTradeMultGTE),
		"ClosingTradeVolumeGTE":  strconv.Itoa(opts.ClosingTradeVolumeGTE),
		"ClosingTradeDollarsGTE": strconv.Itoa(opts.ClosingTradeDollarsGTE),
		"ClosingTradeConditions": opts.ClosingTradeConditions,
		"TradeClusterRankLTE":    strconv.Itoa(opts.ClusterRankLTE),
		"TradeClusterVCDGTE":     strconv.Itoa(opts.ClusterVCDGTE),
		"TradeClusterMultGTE":    strconv.Itoa(opts.ClusterMultGTE),
		"TradeClusterVolumeGTE":  strconv.Itoa(opts.ClusterVolumeGTE),
		"TradeClusterDollarsGTE": strconv.Itoa(opts.ClusterDollarsGTE),
		"TotalRankLTE":           strconv.Itoa(opts.TotalRankLTE),
		"TotalVolumeGTE":         strconv.Itoa(opts.TotalVolumeGTE),
		"TotalDollarsGTE":        strconv.Itoa(opts.TotalDollarsGTE),
		"AHRankLTE":              strconv.Itoa(opts.AHRankLTE),
		"AHVolumeGTE":            strconv.Itoa(opts.AHVolumeGTE),
		"AHDollarsGTE":           strconv.Itoa(opts.AHDollarsGTE),
		"OffsettingPrint":        strconv.FormatBool(opts.OffsettingPrint),
		"PhantomPrint":           strconv.FormatBool(opts.PhantomPrint),
	}
}

// runAlertCreateEdit is the shared handler for create and edit subcommands.
func runAlertCreateEdit(cmd *cobra.Command, opts *alertConfigFlags, key int) error {
	ctx := cmd.Context()
	vlClient, err := common.NewCommandClient(ctx)
	if err != nil {
		return err
	}

	fields := buildAlertConfigFields(opts, key, cmd.Flags().Changed("ticker-group"))
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
