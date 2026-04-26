package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/major/volumeleaders-agent/internal/client"
	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
	cli "github.com/urfave/cli/v3"
)

// --- Option structs ---

type tradesOptions struct {
	tickers, startDate, endDate, sector            string
	minVolume, maxVolume                           int
	conditions, vcd, securityType, relativeSize    int
	darkPools, sweeps, latePrints, sigPrints       int
	evenShared, tradeRank, rankSnapshot, marketCap int
	premarket, rth, ah, opening, closing           int
	phantom, offsetting                            int
	minPrice, maxPrice, minDollars, maxDollars     float64
}

type tradeLevelOptions struct {
	ticker, startDate, endDate string
	minVolume, maxVolume       int
	vcd, relativeSize          int
	tradeLevelRank             int
	tradeLevelCount            int
	minPrice, maxPrice         float64
	minDollars, maxDollars     float64
}

// NewTradeCommand returns the "trade" command group with all subcommands.
func NewTradeCommand() *cli.Command {
	return &cli.Command{
		Name:  "trade",
		Usage: "Trade-related commands",
		Commands: []*cli.Command{
			newTradeListCommand(),
			newTradeSentimentCommand(),
			newTradePresetsCommand(),
			newTradePresetTickersCommand(),
			newTradeClustersCommand(),
			newTradeClusterBombsCommand(),
			newTradeAlertsCommand(),
			newTradeClusterAlertsCommand(),
			newTradeLevelsCommand(),
			newTradeLevelTouchesCommand(),
		},
	}
}

func newTradeSentimentCommand() *cli.Command {
	return &cli.Command{
		Name:  "sentiment",
		Usage: "Summarize leveraged ETF bull/bear flow by day",
		UsageText: `volumeleaders-agent trade sentiment --start-date 2025-04-21 --end-date 2025-04-25
volumeleaders-agent trade sentiment --start-date 2025-04-21 --end-date 2025-04-25 --min-dollars 5000000`,
		Flags: slices.Concat(
			dateRangeFlags(),
			volumeRangeFlags(),
			priceRangeFlags(),
			dollarRangeFlags(5000000),
			[]cli.Flag{
				&cli.IntFlag{Name: "conditions", Value: -1, Usage: "Trade conditions filter"},
				&cli.IntFlag{Name: "vcd", Value: 97, Usage: "VCD filter"},
				&cli.IntFlag{Name: "security-type", Value: -1, Usage: "Security type key"},
				&cli.IntFlag{Name: "relative-size", Value: 5, Usage: "Relative size threshold"},
				&cli.IntFlag{Name: "dark-pools", Value: -1, Usage: "Dark pool filter"},
				&cli.IntFlag{Name: "sweeps", Value: -1, Usage: "Sweep filter"},
				&cli.IntFlag{Name: "late-prints", Value: -1, Usage: "Late print filter"},
				&cli.IntFlag{Name: "sig-prints", Value: -1, Usage: "Signature print filter"},
				&cli.IntFlag{Name: "even-shared", Value: -1, Usage: "Even shared filter"},
				&cli.IntFlag{Name: "trade-rank", Value: -1, Usage: "Trade rank filter"},
				&cli.IntFlag{Name: "rank-snapshot", Value: -1, Usage: "Trade rank snapshot filter"},
				&cli.IntFlag{Name: "market-cap", Value: 0, Usage: "Market cap filter"},
				&cli.IntFlag{Name: "premarket", Value: 1, Usage: "Include premarket"},
				&cli.IntFlag{Name: "rth", Value: 1, Usage: "Include regular trading hours"},
				&cli.IntFlag{Name: "ah", Value: 1, Usage: "Include after hours"},
				&cli.IntFlag{Name: "opening", Value: 1, Usage: "Include opening trades"},
				&cli.IntFlag{Name: "closing", Value: 1, Usage: "Include closing trades"},
				&cli.IntFlag{Name: "phantom", Value: 1, Usage: "Include phantom prints"},
				&cli.IntFlag{Name: "offsetting", Value: 1, Usage: "Include offsetting trades"},
			},
		),
		Action: runTradeSentiment,
	}
}

// --- Subcommand factories ---

func newTradeListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "Query institutional trades",
		UsageText: `volumeleaders-agent trade list --tickers AAPL,MSFT --start-date 2025-01-01 --end-date 2025-01-31
volumeleaders-agent trade list --tickers NVDA --dark-pools 1 --min-dollars 1000000
volumeleaders-agent trade list --sector Technology --relative-size 10 --length 50
volumeleaders-agent trade list --preset "Top-100 Rank" --start-date 2025-04-01 --end-date 2025-04-24
volumeleaders-agent trade list --watchlist "Magnificent 7" --start-date 2025-04-01 --end-date 2025-04-24`,
		Flags: slices.Concat(
			dateRangeFlags(),
			volumeRangeFlags(),
			priceRangeFlags(),
			dollarRangeFlags(500000),
			[]cli.Flag{
				&cli.StringFlag{Name: "tickers", Aliases: []string{"ticker", "symbol", "symbols"}, Usage: "Comma-separated ticker symbols"},
				&cli.IntFlag{Name: "conditions", Value: -1, Usage: "Trade conditions filter"},
				&cli.IntFlag{Name: "vcd", Value: 0, Usage: "VCD filter"},
				&cli.IntFlag{Name: "security-type", Value: -1, Usage: "Security type key"},
				&cli.IntFlag{Name: "relative-size", Value: 5, Usage: "Relative size threshold"},
				&cli.IntFlag{Name: "dark-pools", Value: -1, Usage: "Dark pool filter"},
				&cli.IntFlag{Name: "sweeps", Value: -1, Usage: "Sweep filter"},
				&cli.IntFlag{Name: "late-prints", Value: -1, Usage: "Late print filter"},
				&cli.IntFlag{Name: "sig-prints", Value: -1, Usage: "Signature print filter"},
				&cli.IntFlag{Name: "even-shared", Value: -1, Usage: "Even shared filter"},
				&cli.IntFlag{Name: "trade-rank", Value: -1, Usage: "Trade rank filter"},
				&cli.IntFlag{Name: "rank-snapshot", Value: -1, Usage: "Trade rank snapshot filter"},
				&cli.IntFlag{Name: "market-cap", Value: 0, Usage: "Market cap filter"},
				&cli.IntFlag{Name: "premarket", Value: 1, Usage: "Include premarket"},
				&cli.IntFlag{Name: "rth", Value: 1, Usage: "Include regular trading hours"},
				&cli.IntFlag{Name: "ah", Value: 1, Usage: "Include after hours"},
				&cli.IntFlag{Name: "opening", Value: 1, Usage: "Include opening trades"},
				&cli.IntFlag{Name: "closing", Value: 1, Usage: "Include closing trades"},
				&cli.IntFlag{Name: "phantom", Value: 1, Usage: "Include phantom prints"},
				&cli.IntFlag{Name: "offsetting", Value: 1, Usage: "Include offsetting trades"},
				&cli.StringFlag{Name: "sector", Usage: "Sector/Industry filter"},
				&cli.StringFlag{Name: "preset", Usage: "Apply a built-in filter preset (see: trade presets)"},
				&cli.StringFlag{Name: "watchlist", Usage: "Apply filters from a saved watchlist by name"},
				&cli.StringFlag{Name: "fields", Usage: "Comma-separated trade fields to include in output"},
			},
			outputFormatFlags(),
			paginationFlags(100, 1, "desc"),
		),
		Action: runTradeList,
	}
}

func newTradeClustersCommand() *cli.Command {
	return &cli.Command{
		Name:  "clusters",
		Usage: "Query aggregated trade clusters",
		UsageText: `volumeleaders-agent trade clusters --tickers AAPL --start-date 2025-01-01 --end-date 2025-01-31
volumeleaders-agent trade clusters --min-dollars 50000000 --vcd 1`,
		Flags: slices.Concat(
			dateRangeFlags(),
			volumeRangeFlags(),
			priceRangeFlags(),
			dollarRangeFlags(10000000),
			[]cli.Flag{
				&cli.StringFlag{Name: "tickers", Aliases: []string{"ticker", "symbol", "symbols"}, Usage: "Comma-separated ticker symbols"},
				&cli.IntFlag{Name: "vcd", Value: 0, Usage: "VCD filter"},
				&cli.IntFlag{Name: "security-type", Value: -1, Usage: "Security type key"},
				&cli.IntFlag{Name: "relative-size", Value: 5, Usage: "Relative size threshold"},
				&cli.IntFlag{Name: "trade-cluster-rank", Value: -1, Usage: "Trade cluster rank filter"},
				&cli.StringFlag{Name: "sector", Usage: "Sector/Industry filter"},
			},
			outputFormatFlags(),
			paginationFlags(1000, 1, "desc"),
		),
		Action: runTradeClusters,
	}
}

func newTradeClusterBombsCommand() *cli.Command {
	return &cli.Command{
		Name:  "cluster-bombs",
		Usage: "Query trade cluster bombs",
		UsageText: `volumeleaders-agent trade cluster-bombs --tickers TSLA --start-date 2025-01-01
volumeleaders-agent trade cluster-bombs --vcd 1 --min-volume 100000`,
		Flags: slices.Concat(
			dateRangeFlags(),
			volumeRangeFlags(),
			dollarRangeFlags(0),
			[]cli.Flag{
				&cli.StringFlag{Name: "tickers", Aliases: []string{"ticker", "symbol", "symbols"}, Usage: "Comma-separated ticker symbols"},
				&cli.IntFlag{Name: "vcd", Value: 0, Usage: "VCD filter"},
				&cli.IntFlag{Name: "security-type", Value: 0, Usage: "Security type key"},
				&cli.IntFlag{Name: "relative-size", Value: 0, Usage: "Relative size threshold"},
				&cli.IntFlag{Name: "trade-cluster-bomb-rank", Value: -1, Usage: "Trade cluster bomb rank filter"},
				&cli.StringFlag{Name: "sector", Usage: "Sector/Industry filter"},
			},
			outputFormatFlags(),
			paginationFlags(100, 1, "desc"),
		),
		Action: runTradeClusterBombs,
	}
}

func newTradeAlertsCommand() *cli.Command {
	return &cli.Command{
		Name:      "alerts",
		Usage:     "Query trade alerts for a date",
		UsageText: "volumeleaders-agent trade alerts --date 2025-01-15",
		Flags: slices.Concat(
			[]cli.Flag{
				&cli.StringFlag{Name: "date", Required: true, Usage: "Date YYYY-MM-DD"},
			},
			outputFormatFlags(),
			paginationFlags(100, 1, "desc"),
		),
		Action: runTradeAlerts,
	}
}

func newTradeClusterAlertsCommand() *cli.Command {
	return &cli.Command{
		Name:      "cluster-alerts",
		Usage:     "Query trade cluster alerts for a date",
		UsageText: "volumeleaders-agent trade cluster-alerts --date 2025-01-15",
		Flags: slices.Concat(
			[]cli.Flag{
				&cli.StringFlag{Name: "date", Required: true, Usage: "Date YYYY-MM-DD"},
			},
			outputFormatFlags(),
			paginationFlags(100, 1, "desc"),
		),
		Action: runTradeClusterAlerts,
	}
}

func newTradeLevelsCommand() *cli.Command {
	return &cli.Command{
		Name:  "levels",
		Usage: "Query significant price levels for a ticker",
		UsageText: `volumeleaders-agent trade levels --ticker AAPL --start-date 2025-01-01 --end-date 2025-01-31
volumeleaders-agent trade levels --ticker MSFT --trade-level-count 20 --min-dollars 1000000`,
		Flags: slices.Concat(
			dateRangeFlags(),
			volumeRangeFlags(),
			priceRangeFlags(),
			dollarRangeFlags(500000),
			[]cli.Flag{
				&cli.StringFlag{Name: "ticker", Aliases: []string{"tickers", "symbol", "symbols"}, Required: true, Usage: "Ticker symbol"},
				&cli.IntFlag{Name: "vcd", Value: 0, Usage: "VCD filter"},
				&cli.IntFlag{Name: "relative-size", Value: 0, Usage: "Relative size threshold"},
				&cli.IntFlag{Name: "trade-level-rank", Value: -1, Usage: "Trade level rank filter"},
				&cli.IntFlag{Name: "trade-level-count", Value: 10, Usage: "Number of price levels to return"},
			},
			outputFormatFlags(),
		),
		Action: runTradeLevels,
	}
}

func newTradeLevelTouchesCommand() *cli.Command {
	return &cli.Command{
		Name:  "level-touches",
		Usage: "Query trade events at notable price levels",
		UsageText: `volumeleaders-agent trade level-touches --tickers AAPL --start-date 2025-01-01
volumeleaders-agent trade level-touches --tickers NVDA,AMD --trade-level-rank 5`,
		Flags: slices.Concat(
			dateRangeFlags(),
			volumeRangeFlags(),
			priceRangeFlags(),
			dollarRangeFlags(500000),
			[]cli.Flag{
				&cli.StringFlag{Name: "tickers", Aliases: []string{"ticker", "symbol", "symbols"}, Usage: "Comma-separated ticker symbols"},
				&cli.IntFlag{Name: "vcd", Value: 0, Usage: "VCD filter"},
				&cli.IntFlag{Name: "relative-size", Value: 0, Usage: "Relative size threshold"},
				&cli.IntFlag{Name: "trade-level-rank", Value: 10, Usage: "Trade level rank filter"},
			},
			outputFormatFlags(),
			paginationFlags(100, 0, "desc"),
		),
		Action: runTradeLevelTouches,
	}
}

// --- Action handlers ---

func runTradeList(ctx context.Context, cmd *cli.Command) error {
	presetName := cmd.String("preset")
	watchlistName := cmd.String("watchlist")
	fields, err := parseJSONFieldList[models.Trade](cmd.String("fields"))
	if err != nil {
		return fmt.Errorf("parsing fields flag: %w", err)
	}
	if _, err := parseOutputFormat(cmd.String("format")); err != nil {
		return err
	}

	// Build the full filter map from CLI flags (includes defaults for unset
	// flags). Every key the API requires is present after this call.
	opts := &tradesOptions{
		tickers:      cmd.String("tickers"),
		startDate:    cmd.String("start-date"),
		endDate:      cmd.String("end-date"),
		minVolume:    cmd.Int("min-volume"),
		maxVolume:    cmd.Int("max-volume"),
		minPrice:     cmd.Float("min-price"),
		maxPrice:     cmd.Float("max-price"),
		minDollars:   cmd.Float("min-dollars"),
		maxDollars:   cmd.Float("max-dollars"),
		conditions:   cmd.Int("conditions"),
		vcd:          cmd.Int("vcd"),
		securityType: cmd.Int("security-type"),
		relativeSize: cmd.Int("relative-size"),
		darkPools:    cmd.Int("dark-pools"),
		sweeps:       cmd.Int("sweeps"),
		latePrints:   cmd.Int("late-prints"),
		sigPrints:    cmd.Int("sig-prints"),
		evenShared:   cmd.Int("even-shared"),
		tradeRank:    cmd.Int("trade-rank"),
		rankSnapshot: cmd.Int("rank-snapshot"),
		marketCap:    cmd.Int("market-cap"),
		premarket:    cmd.Int("premarket"),
		rth:          cmd.Int("rth"),
		ah:           cmd.Int("ah"),
		opening:      cmd.Int("opening"),
		closing:      cmd.Int("closing"),
		phantom:      cmd.Int("phantom"),
		offsetting:   cmd.Int("offsetting"),
		sector:       cmd.String("sector"),
	}
	filters := buildTradeFilters(opts)

	if presetName != "" || watchlistName != "" {
		// Preset/watchlist values override the CLI defaults.
		if presetName != "" {
			preset, err := findPreset(presetName)
			if err != nil {
				return err
			}
			maps.Copy(filters, preset.filters)
		}

		if watchlistName != "" {
			wlFilters, err := fetchWatchlistFilters(ctx, watchlistName)
			if err != nil {
				return err
			}
			maps.Copy(filters, wlFilters)
		}

		// User-explicit CLI flags take final precedence.
		applyExplicitFlags(cmd, filters)
	}

	// Dates always come from CLI (required flags).
	filters["StartDate"] = cmd.String("start-date")
	filters["EndDate"] = cmd.String("end-date")

	return runDataTablesCommand[models.Trade](ctx, "/Trades/GetTrades", datatables.TradeColumns,
		dataTableOptions{
			start:    cmd.Int("start"),
			length:   cmd.Int("length"),
			orderCol: cmd.Int("order-col"),
			orderDir: cmd.String("order-dir"),
			filters:  filters,
			fields:   fields,
		},
		cmd.String("format"),
		"query trades")
}

func runTradeSentiment(ctx context.Context, cmd *cli.Command) error {
	opts := &tradesOptions{
		startDate:    cmd.String("start-date"),
		endDate:      cmd.String("end-date"),
		minVolume:    cmd.Int("min-volume"),
		maxVolume:    cmd.Int("max-volume"),
		minPrice:     cmd.Float("min-price"),
		maxPrice:     cmd.Float("max-price"),
		minDollars:   cmd.Float("min-dollars"),
		maxDollars:   cmd.Float("max-dollars"),
		conditions:   cmd.Int("conditions"),
		vcd:          cmd.Int("vcd"),
		securityType: cmd.Int("security-type"),
		relativeSize: cmd.Int("relative-size"),
		darkPools:    cmd.Int("dark-pools"),
		sweeps:       cmd.Int("sweeps"),
		latePrints:   cmd.Int("late-prints"),
		sigPrints:    cmd.Int("sig-prints"),
		evenShared:   cmd.Int("even-shared"),
		tradeRank:    cmd.Int("trade-rank"),
		rankSnapshot: cmd.Int("rank-snapshot"),
		marketCap:    cmd.Int("market-cap"),
		premarket:    cmd.Int("premarket"),
		rth:          cmd.Int("rth"),
		ah:           cmd.Int("ah"),
		opening:      cmd.Int("opening"),
		closing:      cmd.Int("closing"),
		phantom:      cmd.Int("phantom"),
		offsetting:   cmd.Int("offsetting"),
		sector:       "X B",
	}
	filters := buildTradeFilters(opts)
	trades, err := fetchAllTradeSentimentTrades(ctx, dataTableOptions{
		start:    0,
		length:   -1,
		orderCol: 1,
		orderDir: "desc",
		filters:  filters,
	})
	if err != nil {
		return err
	}

	return printJSON(ctx, summarizeTradeSentiment(trades, cmd.String("start-date"), cmd.String("end-date")))
}

func runTradeClusters(ctx context.Context, cmd *cli.Command) error {
	return runDataTablesCommand[models.TradeCluster](ctx, "/TradeClusters/GetTradeClusters", datatables.TradeClusterColumns,
		dataTableOptions{
			start: cmd.Int("start"), length: cmd.Int("length"),
			orderCol: cmd.Int("order-col"), orderDir: cmd.String("order-dir"),
			filters: map[string]string{
				"Tickers":          cmd.String("tickers"),
				"StartDate":        cmd.String("start-date"),
				"EndDate":          cmd.String("end-date"),
				"MinVolume":        intStr(cmd.Int("min-volume")),
				"MaxVolume":        intStr(cmd.Int("max-volume")),
				"MinPrice":         formatFloat(cmd.Float("min-price")),
				"MaxPrice":         formatFloat(cmd.Float("max-price")),
				"MinDollars":       formatFloat(cmd.Float("min-dollars")),
				"MaxDollars":       formatFloat(cmd.Float("max-dollars")),
				"VCD":              intStr(cmd.Int("vcd")),
				"SecurityTypeKey":  intStr(cmd.Int("security-type")),
				"RelativeSize":     intStr(cmd.Int("relative-size")),
				"TradeClusterRank": intStr(cmd.Int("trade-cluster-rank")),
				"SectorIndustry":   cmd.String("sector"),
			},
		}, cmd.String("format"), "query trade clusters")
}

func runTradeClusterBombs(ctx context.Context, cmd *cli.Command) error {
	return runDataTablesCommand[models.TradeClusterBomb](ctx, "/TradeClusterBombs/GetTradeClusterBombs", datatables.TradeClusterBombColumns,
		dataTableOptions{
			start: cmd.Int("start"), length: cmd.Int("length"),
			orderCol: cmd.Int("order-col"), orderDir: cmd.String("order-dir"),
			filters: map[string]string{
				"Tickers":              cmd.String("tickers"),
				"StartDate":            cmd.String("start-date"),
				"EndDate":              cmd.String("end-date"),
				"MinVolume":            intStr(cmd.Int("min-volume")),
				"MaxVolume":            intStr(cmd.Int("max-volume")),
				"MinDollars":           formatFloat(cmd.Float("min-dollars")),
				"MaxDollars":           formatFloat(cmd.Float("max-dollars")),
				"VCD":                  intStr(cmd.Int("vcd")),
				"SecurityTypeKey":      intStr(cmd.Int("security-type")),
				"RelativeSize":         intStr(cmd.Int("relative-size")),
				"TradeClusterBombRank": intStr(cmd.Int("trade-cluster-bomb-rank")),
				"SectorIndustry":       cmd.String("sector"),
			},
		}, cmd.String("format"), "query trade cluster bombs")
}

func runTradeAlerts(ctx context.Context, cmd *cli.Command) error {
	return runDataTablesCommand[models.TradeAlert](ctx, "/TradeAlerts/GetTradeAlerts", datatables.TradeColumns,
		dataTableOptions{
			start: cmd.Int("start"), length: cmd.Int("length"),
			orderCol: cmd.Int("order-col"), orderDir: cmd.String("order-dir"),
			filters: map[string]string{"Date": cmd.String("date")},
		}, cmd.String("format"), "query trade alerts")
}

func runTradeClusterAlerts(ctx context.Context, cmd *cli.Command) error {
	return runDataTablesCommand[models.TradeClusterAlert](ctx, "/TradeClusterAlerts/GetTradeClusterAlerts", datatables.TradeClusterColumns,
		dataTableOptions{
			start: cmd.Int("start"), length: cmd.Int("length"),
			orderCol: cmd.Int("order-col"), orderDir: cmd.String("order-dir"),
			filters: map[string]string{"Date": cmd.String("date")},
		}, cmd.String("format"), "query trade cluster alerts")
}

func runTradeLevels(ctx context.Context, cmd *cli.Command) error {
	opts := &tradeLevelOptions{
		ticker:          cmd.String("ticker"),
		startDate:       cmd.String("start-date"),
		endDate:         cmd.String("end-date"),
		minVolume:       cmd.Int("min-volume"),
		maxVolume:       cmd.Int("max-volume"),
		minPrice:        cmd.Float("min-price"),
		maxPrice:        cmd.Float("max-price"),
		minDollars:      cmd.Float("min-dollars"),
		maxDollars:      cmd.Float("max-dollars"),
		vcd:             cmd.Int("vcd"),
		relativeSize:    cmd.Int("relative-size"),
		tradeLevelRank:  cmd.Int("trade-level-rank"),
		tradeLevelCount: cmd.Int("trade-level-count"),
	}
	return runDataTablesCommand[models.TradeLevel](ctx, "/TradeLevels/GetTradeLevels", datatables.TradeLevelColumns,
		dataTableOptions{
			start:    0,
			length:   -1,
			orderCol: 1,
			orderDir: "desc",
			filters:  buildTradeLevelFilters(opts),
		},
		cmd.String("format"),
		"query trade levels")
}

func runTradeLevelTouches(ctx context.Context, cmd *cli.Command) error {
	return runDataTablesCommand[models.TradeLevelTouch](ctx, "/TradeLevelTouches/GetTradeLevelTouches", datatables.TradeLevelTouchColumns,
		dataTableOptions{
			start: cmd.Int("start"), length: cmd.Int("length"),
			orderCol: cmd.Int("order-col"), orderDir: cmd.String("order-dir"),
			filters: map[string]string{
				"Tickers":        cmd.String("tickers"),
				"StartDate":      cmd.String("start-date"),
				"EndDate":        cmd.String("end-date"),
				"MinVolume":      intStr(cmd.Int("min-volume")),
				"MaxVolume":      intStr(cmd.Int("max-volume")),
				"MinPrice":       formatFloat(cmd.Float("min-price")),
				"MaxPrice":       formatFloat(cmd.Float("max-price")),
				"MinDollars":     formatFloat(cmd.Float("min-dollars")),
				"MaxDollars":     formatFloat(cmd.Float("max-dollars")),
				"VCD":            intStr(cmd.Int("vcd")),
				"RelativeSize":   intStr(cmd.Int("relative-size")),
				"TradeLevelRank": intStr(cmd.Int("trade-level-rank")),
			},
		}, cmd.String("format"), "query trade level touches")
}

func fetchAllTradeSentimentTrades(ctx context.Context, opts dataTableOptions) ([]models.Trade, error) {
	vlClient, err := newCommandClient(ctx)
	if err != nil {
		return nil, err
	}
	return fetchAllTradeSentimentPages(ctx, vlClient, opts)
}

func fetchAllTradeSentimentPages(ctx context.Context, vlClient *client.Client, opts dataTableOptions) ([]models.Trade, error) {
	opts.length = paginationPageSize
	all := make([]models.Trade, 0)

	for {
		request := newDataTablesRequest(datatables.TradeColumns, opts)
		resp, err := vlClient.PostDataTablesPage(ctx, "/Trades/GetTrades", request.Encode())
		if err != nil {
			slog.Error("failed to query trade sentiment", "error", err)
			return nil, fmt.Errorf("query trade sentiment: %w", err)
		}

		var page []models.Trade
		if err := json.Unmarshal(resp.Data, &page); err != nil {
			slog.Error("failed to decode trade sentiment response", "error", err)
			return nil, fmt.Errorf("query trade sentiment: decode response: %w", err)
		}
		if len(page) == 0 {
			break
		}

		all = append(all, page...)
		if resp.RecordsFiltered > 0 && len(all) >= resp.RecordsFiltered {
			break
		}
		if len(page) < paginationPageSize {
			break
		}

		opts.start += len(page)
	}

	return all, nil
}

type tradeSentimentAccumulator struct {
	trades        int
	dollars       float64
	tickerDollars map[string]float64
}

type tradeSentimentDayAccumulator struct {
	bear tradeSentimentAccumulator
	bull tradeSentimentAccumulator
}

func summarizeTradeSentiment(trades []models.Trade, startDate, endDate string) models.TradeSentiment {
	byDay := make(map[string]*tradeSentimentDayAccumulator)
	var totals tradeSentimentDayAccumulator

	for i := range trades {
		trade := &trades[i]
		if !trade.Date.Valid {
			continue
		}

		side := classifyTradeSentimentSide(trade)
		if side == "" {
			continue
		}

		day := trade.Date.Format("2006-01-02")
		acc, ok := byDay[day]
		if !ok {
			acc = &tradeSentimentDayAccumulator{}
			byDay[day] = acc
		}

		acc.add(side, trade)
		totals.add(side, trade)
	}

	days := make([]string, 0, len(byDay))
	for day := range byDay {
		days = append(days, day)
	}
	sort.Strings(days)

	daily := make([]models.TradeSentimentDay, 0, len(days))
	for _, day := range days {
		daily = append(daily, byDay[day].summary(day))
	}

	return models.TradeSentiment{
		DateRange: models.TradeSentimentDateRange{Start: startDate, End: endDate},
		Daily:     daily,
		Totals:    totals.summaryTotals(),
	}
}

func (a *tradeSentimentDayAccumulator) add(side string, trade *models.Trade) {
	switch side {
	case "bear":
		a.bear.add(trade)
	case "bull":
		a.bull.add(trade)
	}
}

func (a *tradeSentimentAccumulator) add(trade *models.Trade) {
	if a.tickerDollars == nil {
		a.tickerDollars = make(map[string]float64)
	}
	a.trades++
	a.dollars += trade.Dollars
	a.tickerDollars[trade.Ticker] += trade.Dollars
}

func (a tradeSentimentDayAccumulator) summary(day string) models.TradeSentimentDay {
	ratio := tradeSentimentRatio(a.bull.dollars, a.bear.dollars)
	return models.TradeSentimentDay{
		Date:   day,
		Bear:   a.bear.summary(),
		Bull:   a.bull.summary(),
		Ratio:  ratio,
		Signal: tradeSentimentSignal(ratio, a.bull.dollars, a.bear.dollars),
	}
}

func (a tradeSentimentDayAccumulator) summaryTotals() models.TradeSentimentTotals {
	ratio := tradeSentimentRatio(a.bull.dollars, a.bear.dollars)
	return models.TradeSentimentTotals{
		Bear:   a.bear.summary(),
		Bull:   a.bull.summary(),
		Ratio:  ratio,
		Signal: tradeSentimentSignal(ratio, a.bull.dollars, a.bear.dollars),
	}
}

func (a tradeSentimentAccumulator) summary() models.TradeSentimentSide {
	return models.TradeSentimentSide{
		Trades:     a.trades,
		Dollars:    a.dollars,
		TopTickers: topTradeSentimentTickers(a.tickerDollars, 3),
	}
}

func tradeSentimentRatio(bullDollars, bearDollars float64) *float64 {
	if bearDollars == 0 {
		return nil
	}
	ratio := bullDollars / bearDollars
	return &ratio
}

func tradeSentimentSignal(ratio *float64, bullDollars, bearDollars float64) models.TradeSentimentSignal {
	if ratio == nil {
		switch {
		case bullDollars > 0:
			return models.TradeSentimentExtremeBull
		case bearDollars > 0:
			return models.TradeSentimentExtremeBear
		default:
			return models.TradeSentimentNeutral
		}
	}

	switch {
	case *ratio < 0.2:
		return models.TradeSentimentExtremeBear
	case *ratio < 0.5:
		return models.TradeSentimentModerateBear
	case *ratio <= 2.0:
		return models.TradeSentimentNeutral
	case *ratio <= 5.0:
		return models.TradeSentimentModerateBull
	default:
		return models.TradeSentimentExtremeBull
	}
}

type tradeSentimentTickerTotal struct {
	ticker  string
	dollars float64
}

func topTradeSentimentTickers(tickerDollars map[string]float64, limit int) []string {
	if len(tickerDollars) == 0 {
		return []string{}
	}

	totals := make([]tradeSentimentTickerTotal, 0, len(tickerDollars))
	for ticker, dollars := range tickerDollars {
		totals = append(totals, tradeSentimentTickerTotal{ticker: ticker, dollars: dollars})
	}
	sort.Slice(totals, func(i, j int) bool {
		if totals[i].dollars == totals[j].dollars {
			return totals[i].ticker < totals[j].ticker
		}
		return totals[i].dollars > totals[j].dollars
	})

	if len(totals) < limit {
		limit = len(totals)
	}
	tickers := make([]string, 0, limit)
	for _, total := range totals[:limit] {
		tickers = append(tickers, total.ticker)
	}
	return tickers
}

func classifyTradeSentimentSide(trade *models.Trade) string {
	fields := []string{trade.Sector, trade.Name}
	if trade.Industry != nil {
		fields = append(fields, *trade.Industry)
	}
	for _, field := range fields {
		field = strings.ToLower(field)
		switch {
		case strings.Contains(field, "bear"):
			return "bear"
		case strings.Contains(field, "bull"):
			return "bull"
		}
	}

	return leveragedETFDirection(trade.Ticker)
}

func leveragedETFDirection(ticker string) string {
	switch strings.ToUpper(strings.TrimSpace(ticker)) {
	case "AAPD", "AMDD", "BERZ", "BITI", "BNKD", "BZQ", "DUST", "EDZ", "ERY", "FAZ", "HIBS", "KOLD", "LABD", "MEXZ", "MYY", "NVDD", "QID", "REK", "REW", "RXD", "SARK", "SCO", "SDD", "SDOW", "SDS", "SEF", "SH", "SMDD", "SOXS", "SPDN", "SPXU", "SPXS", "SQQQ", "SRS", "SSG", "SVIX", "TSDD", "TSLQ", "TSLS", "TZA", "UVIX", "WEBS", "YANG", "YCS", "ZSL":
		return "bear"
	case "AAPU", "AMDL", "BITU", "BOIL", "BRZU", "CURE", "CWEB", "DFEN", "DIG", "DPST", "DRN", "EDC", "ERX", "FAS", "FNGU", "GUSH", "HIBL", "LABU", "MIDU", "NAIL", "NVDL", "QLD", "ROM", "SOXL", "SPXL", "SSO", "TECL", "TMF", "TNA", "TQQQ", "TSLL", "TURB", "UDOW", "UMDD", "UPRO", "URTY", "USD", "UWM", "WEBL", "YINN":
		return "bull"
	default:
		return ""
	}
}

// --- Filter builders ---

func buildTradeFilters(opts *tradesOptions) map[string]string {
	return map[string]string{
		"Tickers":           opts.tickers,
		"StartDate":         opts.startDate,
		"EndDate":           opts.endDate,
		"MinVolume":         intStr(opts.minVolume),
		"MaxVolume":         intStr(opts.maxVolume),
		"MinPrice":          formatFloat(opts.minPrice),
		"MaxPrice":          formatFloat(opts.maxPrice),
		"MinDollars":        formatFloat(opts.minDollars),
		"MaxDollars":        formatFloat(opts.maxDollars),
		"Conditions":        intStr(opts.conditions),
		"VCD":               intStr(opts.vcd),
		"SecurityTypeKey":   intStr(opts.securityType),
		"RelativeSize":      intStr(opts.relativeSize),
		"DarkPools":         intStr(opts.darkPools),
		"Sweeps":            intStr(opts.sweeps),
		"LatePrints":        intStr(opts.latePrints),
		"SignaturePrints":   intStr(opts.sigPrints),
		"EvenShared":        intStr(opts.evenShared),
		"TradeRank":         intStr(opts.tradeRank),
		"TradeRankSnapshot": intStr(opts.rankSnapshot),
		"MarketCap":         intStr(opts.marketCap),
		"IncludePremarket":  intStr(opts.premarket),
		"IncludeRTH":        intStr(opts.rth),
		"IncludeAH":         intStr(opts.ah),
		"IncludeOpening":    intStr(opts.opening),
		"IncludeClosing":    intStr(opts.closing),
		"IncludePhantom":    intStr(opts.phantom),
		"IncludeOffsetting": intStr(opts.offsetting),
		"SectorIndustry":    opts.sector,
	}
}

func buildTradeLevelFilters(opts *tradeLevelOptions) map[string]string {
	return map[string]string{
		"Ticker":          opts.ticker,
		"MinVolume":       intStr(opts.minVolume),
		"MaxVolume":       intStr(opts.maxVolume),
		"MinPrice":        formatFloat(opts.minPrice),
		"MaxPrice":        formatFloat(opts.maxPrice),
		"MinDollars":      formatFloat(opts.minDollars),
		"MaxDollars":      formatFloat(opts.maxDollars),
		"VCD":             intStr(opts.vcd),
		"RelativeSize":    intStr(opts.relativeSize),
		"MinDate":         opts.startDate,
		"MaxDate":         opts.endDate,
		"TradeLevelRank":  intStr(opts.tradeLevelRank),
		"TradeLevelCount": intStr(opts.tradeLevelCount),
	}
}
