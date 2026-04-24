package commands

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
	cli "github.com/urfave/cli/v3"
)

// NewAlertCommand returns the "alert" command group with all subcommands.
func NewAlertCommand() *cli.Command {
	return &cli.Command{
		Name:  "alert",
		Usage: "Alert configuration commands",
		Commands: []*cli.Command{
		{
			Name:      "configs",
			Usage:     "List saved alert configurations",
			UsageText: "volumeleaders-agent alert configs",
				Action: func(ctx context.Context, _ *cli.Command) error {
					return runAlertConfigs(ctx)
				},
			},
		{
			Name:      "delete",
			Usage:     "Delete an alert configuration",
			UsageText: "volumeleaders-agent alert delete --key 42",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "key", Required: true, Usage: "Alert config key to delete"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runAlertDelete(ctx, cmd.Int("key"))
				},
			},
			newAlertCreateCommand(),
			newAlertEditCommand(),
		},
	}
}

// alertConfigFlags returns the shared CLI flags for alert create/edit.
func alertConfigFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "name", Usage: "Alert name (max 50 chars)"},
		&cli.StringFlag{Name: "ticker-group", Value: "AllTickers", Usage: "Ticker group (AllTickers or SelectedTickers)"},
		&cli.StringFlag{Name: "tickers", Usage: "Comma-separated ticker symbols (max 500, used with SelectedTickers)"},
		// Trade thresholds
		&cli.IntFlag{Name: "trade-rank-lte", Usage: "Trade rank <= (0=N/A, 1/3/5/10/25/50/100)"},
		&cli.IntFlag{Name: "trade-vcd-gte", Usage: "Trade VCD >= (0=N/A, 99/100)"},
		&cli.IntFlag{Name: "trade-mult-gte", Usage: "Trade multiplier >= (0=N/A, 5/10/25/50/100)"},
		&cli.IntFlag{Name: "trade-volume-gte", Usage: "Trade volume >= (0=N/A, 1000000/2000000/5000000/10000000)"},
		&cli.IntFlag{Name: "trade-dollars-gte", Usage: "Trade dollars >= (0=N/A, 1000000/10000000/...)"},
		&cli.StringFlag{Name: "trade-conditions", Value: "0", Usage: "Trade conditions (0=N/A, OBH/OBD/OSH/OSD combos)"},
		&cli.BoolFlag{Name: "dark-pool", Usage: "Dark pool filter"},
		&cli.BoolFlag{Name: "sweep", Usage: "Sweep filter"},
		// Closing trade thresholds
		&cli.IntFlag{Name: "closing-trade-rank-lte", Usage: "Closing trade rank <="},
		&cli.IntFlag{Name: "closing-trade-vcd-gte", Usage: "Closing trade VCD >= (0/97/98/99/100)"},
		&cli.IntFlag{Name: "closing-trade-mult-gte", Usage: "Closing trade multiplier >="},
		&cli.IntFlag{Name: "closing-trade-volume-gte", Usage: "Closing trade volume >="},
		&cli.IntFlag{Name: "closing-trade-dollars-gte", Usage: "Closing trade dollars >="},
		&cli.StringFlag{Name: "closing-trade-conditions", Value: "0", Usage: "Closing trade conditions"},
		// Trade cluster thresholds
		&cli.IntFlag{Name: "cluster-rank-lte", Usage: "Trade cluster rank <="},
		&cli.IntFlag{Name: "cluster-vcd-gte", Usage: "Trade cluster VCD >= (0/97/98/99/100)"},
		&cli.IntFlag{Name: "cluster-mult-gte", Usage: "Trade cluster multiplier >="},
		&cli.IntFlag{Name: "cluster-volume-gte", Usage: "Trade cluster volume >="},
		&cli.IntFlag{Name: "cluster-dollars-gte", Usage: "Trade cluster dollars >="},
		// Cumulative institutional totals
		&cli.IntFlag{Name: "total-rank-lte", Usage: "Total rank <= (0/1/3/10/25/50/100)"},
		&cli.IntFlag{Name: "total-volume-gte", Usage: "Total volume >="},
		&cli.IntFlag{Name: "total-dollars-gte", Usage: "Total dollars >="},
		// After-hours cumulative
		&cli.IntFlag{Name: "ah-rank-lte", Usage: "After-hours rank <="},
		&cli.IntFlag{Name: "ah-volume-gte", Usage: "After-hours volume >="},
		&cli.IntFlag{Name: "ah-dollars-gte", Usage: "After-hours dollars >="},
		// Special conditions
		&cli.BoolFlag{Name: "offsetting-print", Usage: "Offsetting print filter"},
		&cli.BoolFlag{Name: "phantom-print", Usage: "Phantom print filter"},
	}
}

func newAlertCreateCommand() *cli.Command {
	flags := alertConfigFlags()
	requireStringFlag(flags, "name")
	return &cli.Command{
		Name:  "create",
		Usage: "Create a new alert configuration",
		UsageText: `volumeleaders-agent alert create --name "Big trades" --tickers AAPL,MSFT --trade-rank-lte 5
volumeleaders-agent alert create --name "Dark pool sweeps" --sweep --dark-pool --trade-volume-gte 1000000`,
		Flags: flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runAlertCreateEdit(ctx, cmd, 0)
		},
	}
}

func newAlertEditCommand() *cli.Command {
	flags := append([]cli.Flag{
		&cli.IntFlag{Name: "key", Required: true, Usage: "Alert config key to edit"},
	}, alertConfigFlags()...)
	return &cli.Command{
		Name:      "edit",
		Usage:     "Edit an existing alert configuration",
		UsageText: "volumeleaders-agent alert edit --key 42 --name \"Updated alert\" --trade-rank-lte 3",
		Flags: flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runAlertCreateEdit(ctx, cmd, cmd.Int("key"))
		},
	}
}

// buildAlertConfigFields maps CLI flag values to form field names.
func buildAlertConfigFields(cmd *cli.Command, key int) map[string]string {
	// Auto-select SelectedTickers when tickers are specified but ticker-group
	// was left at the default.
	tickerGroup := cmd.String("ticker-group")
	if tickerGroup == "AllTickers" && cmd.String("tickers") != "" {
		tickerGroup = "SelectedTickers"
	}

	return map[string]string{
		"AlertConfigKey":         strconv.Itoa(key),
		"Name":                   cmd.String("name"),
		"TickerGroup":            tickerGroup,
		"Tickers":                cmd.String("tickers"),
		"TradeRankLTE":           strconv.Itoa(cmd.Int("trade-rank-lte")),
		"TradeVCDGTE":            strconv.Itoa(cmd.Int("trade-vcd-gte")),
		"TradeMultGTE":           strconv.Itoa(cmd.Int("trade-mult-gte")),
		"TradeVolumeGTE":         strconv.Itoa(cmd.Int("trade-volume-gte")),
		"TradeDollarsGTE":        strconv.Itoa(cmd.Int("trade-dollars-gte")),
		"TradeConditions":        cmd.String("trade-conditions"),
		"DarkPool":               boolString(cmd.Bool("dark-pool")),
		"Sweep":                  boolString(cmd.Bool("sweep")),
		"ClosingTradeRankLTE":    strconv.Itoa(cmd.Int("closing-trade-rank-lte")),
		"ClosingTradeVCDGTE":     strconv.Itoa(cmd.Int("closing-trade-vcd-gte")),
		"ClosingTradeMultGTE":    strconv.Itoa(cmd.Int("closing-trade-mult-gte")),
		"ClosingTradeVolumeGTE":  strconv.Itoa(cmd.Int("closing-trade-volume-gte")),
		"ClosingTradeDollarsGTE": strconv.Itoa(cmd.Int("closing-trade-dollars-gte")),
		"ClosingTradeConditions": cmd.String("closing-trade-conditions"),
		"TradeClusterRankLTE":    strconv.Itoa(cmd.Int("cluster-rank-lte")),
		"TradeClusterVCDGTE":     strconv.Itoa(cmd.Int("cluster-vcd-gte")),
		"TradeClusterMultGTE":    strconv.Itoa(cmd.Int("cluster-mult-gte")),
		"TradeClusterVolumeGTE":  strconv.Itoa(cmd.Int("cluster-volume-gte")),
		"TradeClusterDollarsGTE": strconv.Itoa(cmd.Int("cluster-dollars-gte")),
		"TotalRankLTE":           strconv.Itoa(cmd.Int("total-rank-lte")),
		"TotalVolumeGTE":         strconv.Itoa(cmd.Int("total-volume-gte")),
		"TotalDollarsGTE":        strconv.Itoa(cmd.Int("total-dollars-gte")),
		"AHRankLTE":              strconv.Itoa(cmd.Int("ah-rank-lte")),
		"AHVolumeGTE":            strconv.Itoa(cmd.Int("ah-volume-gte")),
		"AHDollarsGTE":           strconv.Itoa(cmd.Int("ah-dollars-gte")),
		"OffsettingPrint":        boolString(cmd.Bool("offsetting-print")),
		"PhantomPrint":           boolString(cmd.Bool("phantom-print")),
	}
}

// --- Action handlers ---

func runAlertConfigs(ctx context.Context) error {
	vlClient, err := newCommandClient(ctx)
	if err != nil {
		return err
	}

	request := newDataTablesRequest(datatables.AlertConfigColumns, dataTableOptions{start: 0, length: -1, orderCol: 1, orderDir: "asc"})
	var configs []models.AlertConfig
	if err := vlClient.PostDataTables(ctx, "/AlertConfigs/GetAlertConfigs", request.Encode(), &configs); err != nil {
		slog.Error("failed to query alert configs", "error", err)
		return fmt.Errorf("query alert configs: %w", err)
	}

	return printJSON(ctx, configs)
}

func runAlertDelete(ctx context.Context, key int) error {
	vlClient, err := newCommandClient(ctx)
	if err != nil {
		return err
	}

	payload := map[string]int{"AlertConfigKey": key}
	var result any
	if err := vlClient.PostJSON(ctx, "/AlertConfigs/DeleteAlertConfig", payload, &result); err != nil {
		slog.Error("failed to delete alert config", "error", err)
		return fmt.Errorf("delete alert config: %w", err)
	}

	return printJSON(ctx, result)
}

func runAlertCreateEdit(ctx context.Context, cmd *cli.Command, key int) error {
	vlClient, err := newCommandClient(ctx)
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
	return printJSON(ctx, map[string]any{"success": true, "action": action, "key": key})
}
