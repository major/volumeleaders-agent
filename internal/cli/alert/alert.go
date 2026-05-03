package alert

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/leodido/structcli"
	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
)

// alertConfigsOptions holds flags for the "alert configs" subcommand.
type alertConfigsOptions struct {
	Format common.OutputFormat `flag:"format" flaggroup:"Output" flagshort:"f" default:"json" flagdescr:"Output format: json, csv, or tsv"`
	Fields string              `flag:"fields" flaggroup:"Output" flagdescr:"Comma-separated fields to include (use 'all' for every field)"`
}

// alertDeleteOptions holds flags for the "alert delete" subcommand.
type alertDeleteOptions struct {
	Key int `flag:"key" flaggroup:"Input" flagshort:"k" flagrequired:"true" flagdescr:"Alert config key to delete"`
}

// alertConfigFlags holds the shared flag set for alert create/edit commands.
type alertConfigFlags struct {
	Name                   string `flag:"name" flaggroup:"Basic" flagdescr:"Alert name (max 50 chars)"`
	TickerGroup            string `flag:"ticker-group" flaggroup:"Basic" flagdescr:"Ticker group (AllTickers or SelectedTickers)"`
	Tickers                string `flag:"tickers" flaggroup:"Basic" flagshort:"t" flagdescr:"Comma-separated ticker symbols (max 500, used with SelectedTickers)"`
	TradeRankLTE           int    `flag:"trade-rank-lte" flaggroup:"Trade Filters" flagdescr:"Trade rank <= (0=N/A, 1/3/5/10/25/50/100)"`
	TradeVCDGTE            int    `flag:"trade-vcd-gte" flaggroup:"Trade Filters" flagdescr:"Trade VCD >= (0=N/A, 99/100)"`
	TradeMultGTE           int    `flag:"trade-mult-gte" flaggroup:"Trade Filters" flagdescr:"Trade multiplier >= (0=N/A, 5/10/25/50/100)"`
	TradeVolumeGTE         int    `flag:"trade-volume-gte" flaggroup:"Trade Filters" flagdescr:"Trade volume >= (0=N/A, 1000000/2000000/5000000/10000000)"`
	TradeDollarsGTE        int    `flag:"trade-dollars-gte" flaggroup:"Trade Filters" flagdescr:"Trade dollars >= (0=N/A, 1000000/10000000/...)"`
	TradeConditions        string `flag:"trade-conditions" flaggroup:"Trade Filters" flagdescr:"Trade conditions (0=N/A, OBH/OBD/OSH/OSD combos)"`
	DarkPool               bool   `flag:"dark-pool" flaggroup:"Trade Filters" flagdescr:"Dark pool filter"`
	Sweep                  bool   `flag:"sweep" flaggroup:"Trade Filters" flagdescr:"Sweep filter"`
	ClosingTradeRankLTE    int    `flag:"closing-trade-rank-lte" flaggroup:"Closing Filters" flagdescr:"Closing trade rank <="`
	ClosingTradeVCDGTE     int    `flag:"closing-trade-vcd-gte" flaggroup:"Closing Filters" flagdescr:"Closing trade VCD >= (0/97/98/99/100)"`
	ClosingTradeMultGTE    int    `flag:"closing-trade-mult-gte" flaggroup:"Closing Filters" flagdescr:"Closing trade multiplier >="`
	ClosingTradeVolumeGTE  int    `flag:"closing-trade-volume-gte" flaggroup:"Closing Filters" flagdescr:"Closing trade volume >="`
	ClosingTradeDollarsGTE int    `flag:"closing-trade-dollars-gte" flaggroup:"Closing Filters" flagdescr:"Closing trade dollars >="`
	ClosingTradeConditions string `flag:"closing-trade-conditions" flaggroup:"Closing Filters" flagdescr:"Closing trade conditions"`
	ClusterRankLTE         int    `flag:"cluster-rank-lte" flaggroup:"Cluster Filters" flagdescr:"Trade cluster rank <="`
	ClusterVCDGTE          int    `flag:"cluster-vcd-gte" flaggroup:"Cluster Filters" flagdescr:"Trade cluster VCD >= (0/97/98/99/100)"`
	ClusterMultGTE         int    `flag:"cluster-mult-gte" flaggroup:"Cluster Filters" flagdescr:"Trade cluster multiplier >="`
	ClusterVolumeGTE       int    `flag:"cluster-volume-gte" flaggroup:"Cluster Filters" flagdescr:"Trade cluster volume >="`
	ClusterDollarsGTE      int    `flag:"cluster-dollars-gte" flaggroup:"Cluster Filters" flagdescr:"Trade cluster dollars >="`
	TotalRankLTE           int    `flag:"total-rank-lte" flaggroup:"Total Filters" flagdescr:"Total rank <= (0/1/3/10/25/50/100)"`
	TotalVolumeGTE         int    `flag:"total-volume-gte" flaggroup:"Total Filters" flagdescr:"Total volume >="`
	TotalDollarsGTE        int    `flag:"total-dollars-gte" flaggroup:"Total Filters" flagdescr:"Total dollars >="`
	AHRankLTE              int    `flag:"ah-rank-lte" flaggroup:"After-Hours Filters" flagdescr:"After-hours rank <="`
	AHVolumeGTE            int    `flag:"ah-volume-gte" flaggroup:"After-Hours Filters" flagdescr:"After-hours volume >="`
	AHDollarsGTE           int    `flag:"ah-dollars-gte" flaggroup:"After-Hours Filters" flagdescr:"After-hours dollars >="`
	OffsettingPrint        bool   `flag:"offsetting-print" flaggroup:"Trade Filters" flagdescr:"Offsetting print filter"`
	PhantomPrint           bool   `flag:"phantom-print" flaggroup:"Trade Filters" flagdescr:"Phantom print filter"`
}

// alertCreateOptions holds flags for the "alert create" subcommand.
type alertCreateOptions struct {
	alertConfigFlags
}

// alertEditOptions holds flags for the "alert edit" subcommand.
type alertEditOptions struct {
	Key int `flag:"key" flaggroup:"Input" flagshort:"k" flagrequired:"true" flagdescr:"Alert config key to edit"`
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
	opts := &alertConfigsOptions{}
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
	if err := structcli.Bind(cmd, opts); err != nil {
		panic(fmt.Sprintf("structcli.Bind configs: %v", err))
	}
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
	if err := structcli.Bind(cmd, opts); err != nil {
		panic(fmt.Sprintf("structcli.Bind delete: %v", err))
	}
	return cmd
}

// newCreateCmd returns the "create" subcommand.
func newCreateCmd() *cobra.Command {
	opts := &alertCreateOptions{}
	opts.TickerGroup = "AllTickers"
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
	if err := structcli.Bind(cmd, opts); err != nil {
		panic(fmt.Sprintf("structcli.Bind create: %v", err))
	}
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

// newEditCmd returns the "edit" subcommand.
func newEditCmd() *cobra.Command {
	opts := &alertEditOptions{}
	opts.TickerGroup = "AllTickers"
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
	if err := structcli.Bind(cmd, opts); err != nil {
		panic(fmt.Sprintf("structcli.Bind edit: %v", err))
	}
	return cmd
}

// buildAlertConfigFields maps struct field values to form field names for the
// multipart POST request.
func buildAlertConfigFields(opts *alertConfigFlags, key int) map[string]string {
	// Auto-select SelectedTickers when tickers are specified but ticker-group
	// was left at the default.
	tickerGroup := opts.TickerGroup
	if tickerGroup == "AllTickers" && opts.Tickers != "" {
		tickerGroup = "SelectedTickers"
	}

	return map[string]string{
		"AlertConfigKey":         strconv.Itoa(key),
		"Name":                   opts.Name,
		"TickerGroup":            tickerGroup,
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

	fields := buildAlertConfigFields(opts, key)
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
