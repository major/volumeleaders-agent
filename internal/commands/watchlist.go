package commands

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strconv"

	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
	cli "github.com/urfave/cli/v3"
)

// NewWatchlistCommand returns the "watchlist" command group with all subcommands.
func NewWatchlistCommand() *cli.Command {
	return &cli.Command{
		Name:  "watchlist",
		Usage: "Watch list commands",
		Commands: []*cli.Command{
			{
				Name:      "configs",
				Usage:     "List saved watch list configurations",
				UsageText: "volumeleaders-agent watchlist configs",
				Flags:     outputFormatFlags(),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runWatchlistConfigs(ctx, cmd.String("format"))
				},
			},
			{
				Name:      "tickers",
				Usage:     "Query tickers for a selected watch list",
				UsageText: "volumeleaders-agent watchlist tickers --watchlist-key 1",
				Flags: append([]cli.Flag{
					&cli.IntFlag{Name: "watchlist-key", Value: -1, Usage: "Watch list key (-1 for all)"},
				}, outputFormatFlags()...),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runWatchlistTickers(ctx, cmd.Int("watchlist-key"), cmd.String("format"))
				},
			},
			{
				Name:      "delete",
				Usage:     "Delete a watch list configuration",
				UsageText: "volumeleaders-agent watchlist delete --key 1",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "key", Required: true, Usage: "Watch list key to delete"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runWatchlistDelete(ctx, cmd.Int("key"))
				},
			},
			{
				Name:      "add-ticker",
				Usage:     "Add a ticker to an existing watch list",
				UsageText: "volumeleaders-agent watchlist add-ticker --watchlist-key 1 --ticker NVDA",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "watchlist-key", Required: true, Usage: "Watch list key"},
					&cli.StringFlag{Name: "ticker", Required: true, Usage: "Ticker symbol to add"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runWatchlistAddTicker(ctx, cmd.Int("watchlist-key"), cmd.String("ticker"))
				},
			},
			newWatchlistCreateCommand(),
			newWatchlistEditCommand(),
		},
	}
}

// watchlistConfigFlags returns the shared CLI flags for watchlist create/edit.
func watchlistConfigFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "name", Usage: "Watch list name"},
		&cli.StringFlag{Name: "tickers", Usage: "Comma-separated ticker symbols (max 500)"},
		&cli.IntFlag{Name: "min-volume", Usage: "Minimum volume filter"},
		&cli.IntFlag{Name: "max-volume", Value: 2000000000, Usage: "Maximum volume filter"},
		&cli.FloatFlag{Name: "min-dollars", Usage: "Minimum dollars filter"},
		&cli.FloatFlag{Name: "max-dollars", Value: 30000000000, Usage: "Maximum dollars filter"},
		&cli.FloatFlag{Name: "min-price", Usage: "Minimum price filter"},
		&cli.FloatFlag{Name: "max-price", Value: 100000, Usage: "Maximum price filter"},
		&cli.FloatFlag{Name: "min-vcd", Usage: "Minimum VCD percentile (0-100)"},
		&cli.StringFlag{Name: "sector-industry", Usage: "Sector/industry filter (max 100 chars)"},
		&cli.IntFlag{Name: "security-type", Value: -1, Usage: "Security type (-1=all, 1=stocks, 26=ETFs, 4=REITs)"},
		&cli.IntFlag{Name: "min-relative-size", Usage: "Minimum relative size (0/5/10/25/50/100)"},
		&cli.IntFlag{Name: "max-trade-rank", Value: -1, Usage: "Maximum trade rank (-1=all, 1/3/5/10/25/50/100)"},
		&cli.BoolFlag{Name: "normal-prints", Value: true, Usage: "Include normal prints"},
		&cli.BoolFlag{Name: "signature-prints", Value: true, Usage: "Include signature prints"},
		&cli.BoolFlag{Name: "late-prints", Value: true, Usage: "Include late prints"},
		&cli.BoolFlag{Name: "timely-prints", Value: true, Usage: "Include timely prints"},
		&cli.BoolFlag{Name: "dark-pools", Value: true, Usage: "Include dark pool trades"},
		&cli.BoolFlag{Name: "lit-exchanges", Value: true, Usage: "Include lit exchange trades"},
		&cli.BoolFlag{Name: "sweeps", Value: true, Usage: "Include sweep trades"},
		&cli.BoolFlag{Name: "blocks", Value: true, Usage: "Include block trades"},
		&cli.BoolFlag{Name: "premarket-trades", Value: true, Usage: "Include premarket trades"},
		&cli.BoolFlag{Name: "rth-trades", Value: true, Usage: "Include regular trading hours trades"},
		&cli.BoolFlag{Name: "ah-trades", Value: true, Usage: "Include after-hours trades"},
		&cli.BoolFlag{Name: "opening-trades", Value: true, Usage: "Include opening trades"},
		&cli.BoolFlag{Name: "closing-trades", Value: true, Usage: "Include closing trades"},
		&cli.BoolFlag{Name: "phantom-trades", Value: true, Usage: "Include phantom trades"},
		&cli.BoolFlag{Name: "offsetting-trades", Value: true, Usage: "Include offsetting trades"},
		&cli.IntFlag{Name: "rsi-overbought-daily", Value: -1, Usage: "RSI overbought daily (1=yes, 0=no, -1=ignore)"},
		&cli.IntFlag{Name: "rsi-overbought-hourly", Value: -1, Usage: "RSI overbought hourly (1=yes, 0=no, -1=ignore)"},
		&cli.IntFlag{Name: "rsi-oversold-daily", Value: -1, Usage: "RSI oversold daily (1=yes, 0=no, -1=ignore)"},
		&cli.IntFlag{Name: "rsi-oversold-hourly", Value: -1, Usage: "RSI oversold hourly (1=yes, 0=no, -1=ignore)"},
	}
}

func newWatchlistCreateCommand() *cli.Command {
	flags := watchlistConfigFlags()
	// Name is required for create.
	requireStringFlag(flags, "name")
	return &cli.Command{
		Name:  "create",
		Usage: "Create a new watch list configuration",
		UsageText: `volumeleaders-agent watchlist create --name "Tech stocks" --tickers AAPL,MSFT,GOOGL
volumeleaders-agent watchlist create --name "Large caps" --security-type 1 --min-dollars 10000000`,
		Flags: flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runWatchlistCreateEdit(ctx, cmd, 0)
		},
	}
}

func newWatchlistEditCommand() *cli.Command {
	flags := slices.Concat([]cli.Flag{
		&cli.IntFlag{Name: "key", Required: true, Usage: "Watch list key to edit"},
	}, watchlistConfigFlags())
	return &cli.Command{
		Name:      "edit",
		Usage:     "Edit an existing watch list configuration",
		UsageText: "volumeleaders-agent watchlist edit --key 1 --name \"Updated watchlist\" --tickers AAPL,MSFT",
		Flags:     flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runWatchlistCreateEdit(ctx, cmd, cmd.Int("key"))
		},
	}
}

// buildWatchlistConfigFields maps CLI flag values to form field names.
func buildWatchlistConfigFields(cmd *cli.Command, key int) map[string]string {
	return map[string]string{
		"SearchTemplateKey":           strconv.Itoa(key),
		"Name":                        cmd.String("name"),
		"Tickers":                     cmd.String("tickers"),
		"MinVolume":                   strconv.Itoa(cmd.Int("min-volume")),
		"MaxVolume":                   strconv.Itoa(cmd.Int("max-volume")),
		"MinDollars":                  formatFloat(cmd.Float("min-dollars")),
		"MaxDollars":                  formatFloat(cmd.Float("max-dollars")),
		"MinPrice":                    formatFloat(cmd.Float("min-price")),
		"MaxPrice":                    formatFloat(cmd.Float("max-price")),
		"MinVCD":                      formatFloat(cmd.Float("min-vcd")),
		"SectorIndustry":              cmd.String("sector-industry"),
		"SecurityTypeKey":             strconv.Itoa(cmd.Int("security-type")),
		"MinRelativeSizeSelected":     strconv.Itoa(cmd.Int("min-relative-size")),
		"MaxTradeRankSelected":        strconv.Itoa(cmd.Int("max-trade-rank")),
		"NormalPrintsSelected":        boolString(cmd.Bool("normal-prints")),
		"SignaturePrintsSelected":     boolString(cmd.Bool("signature-prints")),
		"LatePrintsSelected":          boolString(cmd.Bool("late-prints")),
		"TimelyPrintsSelected":        boolString(cmd.Bool("timely-prints")),
		"DarkPoolsSelected":           boolString(cmd.Bool("dark-pools")),
		"LitExchangesSelected":        boolString(cmd.Bool("lit-exchanges")),
		"SweepsSelected":              boolString(cmd.Bool("sweeps")),
		"BlocksSelected":              boolString(cmd.Bool("blocks")),
		"PremarketTradesSelected":     boolString(cmd.Bool("premarket-trades")),
		"RTHTradesSelected":           boolString(cmd.Bool("rth-trades")),
		"AHTradesSelected":            boolString(cmd.Bool("ah-trades")),
		"OpeningTradesSelected":       boolString(cmd.Bool("opening-trades")),
		"ClosingTradesSelected":       boolString(cmd.Bool("closing-trades")),
		"PhantomTradesSelected":       boolString(cmd.Bool("phantom-trades")),
		"OffsettingTradesSelected":    boolString(cmd.Bool("offsetting-trades")),
		"RSIOverboughtDailySelected":  strconv.Itoa(cmd.Int("rsi-overbought-daily")),
		"RSIOverboughtHourlySelected": strconv.Itoa(cmd.Int("rsi-overbought-hourly")),
		"RSIOversoldDailySelected":    strconv.Itoa(cmd.Int("rsi-oversold-daily")),
		"RSIOversoldHourlySelected":   strconv.Itoa(cmd.Int("rsi-oversold-hourly")),
	}
}

// --- Action handlers ---

func runWatchlistConfigs(ctx context.Context, formatValue string) error {
	format, err := parseOutputFormat(formatValue)
	if err != nil {
		return err
	}

	vlClient, err := newCommandClient(ctx)
	if err != nil {
		return err
	}

	request := newDataTablesRequest(datatables.WatchlistConfigColumns, dataTableOptions{start: 0, length: -1, orderCol: 1, orderDir: "asc"})
	var configs []models.WatchListConfig
	if err := vlClient.PostDataTables(ctx, "/WatchListConfigs/GetWatchLists", request.Encode(), &configs); err != nil {
		slog.Error("failed to query watchlist configs", "error", err)
		return fmt.Errorf("query watchlist configs: %w", err)
	}

	return printDataTablesResult(ctx, configs, nil, format)
}

func runWatchlistTickers(ctx context.Context, watchlistKey int, formatValue string) error {
	format, err := parseOutputFormat(formatValue)
	if err != nil {
		return err
	}

	vlClient, err := newCommandClient(ctx)
	if err != nil {
		return err
	}

	request := newDataTablesRequest(datatables.WatchlistTickerColumns, dataTableOptions{
		start:    0,
		length:   -1,
		orderCol: 0,
		orderDir: "asc",
		filters: map[string]string{
			"WatchListKey": strconv.Itoa(watchlistKey),
		},
	})
	var tickers []models.WatchListTicker
	if err := vlClient.PostDataTables(ctx, "/WatchLists/GetWatchListTickers", request.Encode(), &tickers); err != nil {
		slog.Error("failed to query watchlist tickers", "error", err)
		return fmt.Errorf("query watchlist tickers: %w", err)
	}

	return printDataTablesResult(ctx, tickers, nil, format)
}

func runWatchlistDelete(ctx context.Context, key int) error {
	vlClient, err := newCommandClient(ctx)
	if err != nil {
		return err
	}

	payload := map[string]int{"WatchListKey": key}
	var result any
	if err := vlClient.PostJSON(ctx, "/WatchListConfigs/DeleteWatchList", payload, &result); err != nil {
		slog.Error("failed to delete watchlist", "error", err)
		return fmt.Errorf("delete watchlist: %w", err)
	}

	return printJSON(ctx, result)
}

func runWatchlistAddTicker(ctx context.Context, watchlistKey int, ticker string) error {
	vlClient, err := newCommandClient(ctx)
	if err != nil {
		return err
	}

	values := url.Values{
		"WatchListKey": {strconv.Itoa(watchlistKey)},
		"Ticker":       {ticker},
	}
	var result any
	if err := vlClient.PostForm(ctx, "/Chart0/UpdateWatchList", values, &result); err != nil {
		slog.Error("failed to add ticker to watchlist", "error", err)
		return fmt.Errorf("add ticker to watchlist: %w", err)
	}

	return printJSON(ctx, result)
}

func runWatchlistCreateEdit(ctx context.Context, cmd *cli.Command, key int) error {
	vlClient, err := newCommandClient(ctx)
	if err != nil {
		return err
	}

	fields := buildWatchlistConfigFields(cmd, key)
	if err := vlClient.PostMultipart(ctx, "/WatchListConfig", fields); err != nil {
		slog.Error("failed to save watchlist config", "error", err)
		return fmt.Errorf("save watchlist config: %w", err)
	}

	action := "created"
	if key != 0 {
		action = "updated"
	}
	return printJSON(ctx, map[string]any{"success": true, "action": action, "key": key})
}
