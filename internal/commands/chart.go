package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"

	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
	cli "github.com/urfave/cli/v3"
)

// --- Option structs ---

type priceDataOptions struct {
	ticker, format                 string
	startDate, endDate             string
	volumeProfile, levels          int
	minVolume, maxVolume           int
	vcd, tradeCount                int
	tradeRank, tradeRankSnapshot   int
	darkPools, sweeps, latePrints  int
	signaturePrints                int
	includePremarket, includeRTH   int
	includeAH, includeOpening      int
	includeClosing, includePhantom int
	includeOffsetting              int
	minPrice, maxPrice             float64
	minDollars, maxDollars         float64
}

type chartLevelsOptions struct {
	ticker, startDate, endDate, format string
	levels                             int
}

// NewChartCommand returns the "chart" command group with all subcommands.
func NewChartCommand() *cli.Command {
	return &cli.Command{
		Name:  "chart",
		Usage: "Chart and price data commands",
		Commands: []*cli.Command{
			newPriceDataCommand(),
			newChartSnapshotCommand(),
			newChartLevelsCommand(),
			newCompanyCommand(),
		},
	}
}

// --- Subcommand factories ---

func newPriceDataCommand() *cli.Command {
	return &cli.Command{
		Name:  "price-data",
		Usage: "Query one-minute chart bars with trade metadata for a ticker",
		UsageText: `volumeleaders-agent chart price-data --ticker AAPL --start-date 2025-01-15 --end-date 2025-01-15
volumeleaders-agent chart price-data --ticker NVDA --start-date 2025-01-01 --end-date 2025-01-31 --dark-pools 1`,
		Flags: slices.Concat(
			dateRangeFlags(),
			volumeRangeFlags(),
			priceRangeFlags(),
			dollarRangeFlags(500000),
			[]cli.Flag{
				&cli.StringFlag{Name: "ticker", Required: true, Usage: "Ticker symbol"},
				&cli.IntFlag{Name: "volume-profile", Value: 0, Usage: "Volume profile flag"},
				&cli.IntFlag{Name: "levels", Value: 5, Usage: "Number of trade levels"},
				&cli.IntFlag{Name: "dark-pools", Value: -1, Usage: "Dark pool filter (-1=all, 0=exclude, 1=only)"},
				&cli.IntFlag{Name: "sweeps", Value: -1, Usage: "Sweep filter (-1=all, 0=exclude, 1=only)"},
				&cli.IntFlag{Name: "late-prints", Value: -1, Usage: "Late print filter (-1=all, 0=exclude, 1=only)"},
				&cli.IntFlag{Name: "signature-prints", Value: -1, Usage: "Signature print filter (-1=all, 0=exclude, 1=only)"},
				&cli.IntFlag{Name: "trade-count", Value: 3, Usage: "Minimum trade count"},
				&cli.IntFlag{Name: "vcd", Value: 0, Usage: "VCD filter"},
				&cli.IntFlag{Name: "trade-rank", Value: -1, Usage: "Trade rank filter"},
				&cli.IntFlag{Name: "trade-rank-snapshot", Value: -1, Usage: "Trade rank snapshot filter"},
				&cli.IntFlag{Name: "include-premarket", Value: 1, Usage: "Include premarket trades (0/1)"},
				&cli.IntFlag{Name: "include-rth", Value: 1, Usage: "Include regular trading hours (0/1)"},
				&cli.IntFlag{Name: "include-ah", Value: 1, Usage: "Include after hours trades (0/1)"},
				&cli.IntFlag{Name: "include-opening", Value: 1, Usage: "Include opening trades (0/1)"},
				&cli.IntFlag{Name: "include-closing", Value: 1, Usage: "Include closing trades (0/1)"},
				&cli.IntFlag{Name: "include-phantom", Value: 1, Usage: "Include phantom prints (0/1)"},
				&cli.IntFlag{Name: "include-offsetting", Value: 1, Usage: "Include offsetting trades (0/1)"},
			},
			outputFormatFlags(),
		),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			opts := priceDataOptions{
				ticker:            cmd.String("ticker"),
				format:            cmd.String("format"),
				startDate:         cmd.String("start-date"),
				endDate:           cmd.String("end-date"),
				volumeProfile:     cmd.Int("volume-profile"),
				levels:            cmd.Int("levels"),
				minVolume:         cmd.Int("min-volume"),
				maxVolume:         cmd.Int("max-volume"),
				minDollars:        cmd.Float("min-dollars"),
				maxDollars:        cmd.Float("max-dollars"),
				darkPools:         cmd.Int("dark-pools"),
				sweeps:            cmd.Int("sweeps"),
				latePrints:        cmd.Int("late-prints"),
				signaturePrints:   cmd.Int("signature-prints"),
				tradeCount:        cmd.Int("trade-count"),
				minPrice:          cmd.Float("min-price"),
				maxPrice:          cmd.Float("max-price"),
				vcd:               cmd.Int("vcd"),
				tradeRank:         cmd.Int("trade-rank"),
				tradeRankSnapshot: cmd.Int("trade-rank-snapshot"),
				includePremarket:  cmd.Int("include-premarket"),
				includeRTH:        cmd.Int("include-rth"),
				includeAH:         cmd.Int("include-ah"),
				includeOpening:    cmd.Int("include-opening"),
				includeClosing:    cmd.Int("include-closing"),
				includePhantom:    cmd.Int("include-phantom"),
				includeOffsetting: cmd.Int("include-offsetting"),
			}
			return runPriceData(ctx, &opts)
		},
	}
}

func newChartSnapshotCommand() *cli.Command {
	return &cli.Command{
		Name:      "snapshot",
		Usage:     "Query quote snapshot for a ticker and date",
		UsageText: "volumeleaders-agent chart snapshot --ticker AAPL --date-key 2025-01-15",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "ticker", Required: true, Usage: "Ticker symbol"},
			&cli.StringFlag{Name: "date-key", Required: true, Usage: "Date key YYYY-MM-DD or YYYYMMDD"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runChartSnapshot(ctx, cmd.String("ticker"), cmd.String("date-key"))
		},
	}
}

func newChartLevelsCommand() *cli.Command {
	return &cli.Command{
		Name:  "levels",
		Usage: "Query chart-level overlays for a ticker and date range",
		UsageText: `volumeleaders-agent chart levels --ticker AAPL --start-date 2025-01-01 --end-date 2025-01-31
volumeleaders-agent chart levels --ticker TSLA --start-date 2025-01-01 --levels 10`,
		Flags: slices.Concat(
			dateRangeFlags(),
			[]cli.Flag{
				&cli.StringFlag{Name: "ticker", Required: true, Usage: "Ticker symbol"},
				&cli.IntFlag{Name: "levels", Value: 5, Usage: "Number of trade levels"},
			},
			outputFormatFlags(),
		),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			opts := chartLevelsOptions{
				ticker:    cmd.String("ticker"),
				startDate: cmd.String("start-date"),
				endDate:   cmd.String("end-date"),
				levels:    cmd.Int("levels"),
				format:    cmd.String("format"),
			}
			return runChartLevels(ctx, opts)
		},
	}
}

func newCompanyCommand() *cli.Command {
	return &cli.Command{
		Name:      "company",
		Usage:     "Query company metadata for a ticker",
		UsageText: "volumeleaders-agent chart company --ticker AAPL",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "ticker", Required: true, Usage: "Ticker symbol"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runCompany(ctx, cmd.String("ticker"))
		},
	}
}

// --- Action handlers ---

func runPriceData(ctx context.Context, opts *priceDataOptions) error {
	format, err := parseOutputFormat(opts.format)
	if err != nil {
		return err
	}

	vlClient, err := newCommandClient(ctx)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"StartDateKey":      toDateKey(opts.startDate),
		"EndDateKey":        toDateKey(opts.endDate),
		"Ticker":            opts.ticker,
		"VolumeProfile":     opts.volumeProfile,
		"Levels":            opts.levels,
		"MinVolume":         opts.minVolume,
		"MaxVolume":         opts.maxVolume,
		"MinDollars":        opts.minDollars,
		"MaxDollars":        opts.maxDollars,
		"DarkPools":         opts.darkPools,
		"Sweeps":            opts.sweeps,
		"LatePrints":        opts.latePrints,
		"SignaturePrints":   opts.signaturePrints,
		"TradeCount":        opts.tradeCount,
		"MinPrice":          opts.minPrice,
		"MaxPrice":          opts.maxPrice,
		"VCD":               opts.vcd,
		"TradeRank":         opts.tradeRank,
		"TradeRankSnapshot": opts.tradeRankSnapshot,
		"IncludePremarket":  opts.includePremarket,
		"IncludeRTH":        opts.includeRTH,
		"IncludeAH":         opts.includeAH,
		"IncludeOpening":    opts.includeOpening,
		"IncludeClosing":    opts.includeClosing,
		"IncludePhantom":    opts.includePhantom,
		"IncludeOffsetting": opts.includeOffsetting,
	}

	// The API returns a nested array: [[PriceBar, ...], ...]. Extract the first element.
	var nested []json.RawMessage
	if err := vlClient.PostJSON(ctx, "/Chart0/GetAllPriceVolumeTradeData", payload, &nested); err != nil {
		slog.Error("failed to query price data", "error", err)
		return fmt.Errorf("query price data: %w", err)
	}

	var bars []models.PriceBar
	if len(nested) > 0 {
		if err := json.Unmarshal(nested[0], &bars); err != nil {
			slog.Error("failed to decode price bars", "error", err)
			return fmt.Errorf("decode price bars: %w", err)
		}
	}

	return printDataTablesResult(ctx, bars, nil, format)
}

func runChartSnapshot(ctx context.Context, ticker, dateKey string) error {
	vlClient, err := newCommandClient(ctx)
	if err != nil {
		return err
	}

	payload := map[string]string{
		"Ticker":  ticker,
		"DateKey": toDateKey(dateKey),
	}

	var resp models.SnapshotResponse
	if err := vlClient.PostJSON(ctx, "/Chart0/GetSnapshot", payload, &resp); err != nil {
		slog.Error("failed to query chart snapshot", "error", err)
		return fmt.Errorf("query chart snapshot: %w", err)
	}

	return printJSON(ctx, resp.Snapshot)
}

func runChartLevels(ctx context.Context, opts chartLevelsOptions) error {
	return runDataTablesCommand[models.TradeLevel](ctx, "/Chart0/GetTradeLevels", datatables.TradeLevelColumns,
		dataTableOptions{
			start:    0,
			length:   -1,
			orderCol: 0,
			orderDir: "desc",
			filters: map[string]string{
				"StartDate": opts.startDate,
				"EndDate":   opts.endDate,
				"Ticker":    opts.ticker,
				"Levels":    intStr(opts.levels),
			},
		},
		opts.format,
		"query chart levels")
}

func runCompany(ctx context.Context, ticker string) error {
	vlClient, err := newCommandClient(ctx)
	if err != nil {
		return err
	}

	payload := map[string]string{"Ticker": ticker}
	var company models.Company
	if err := vlClient.PostJSON(ctx, "/Chart0/GetCompany", payload, &company); err != nil {
		slog.Error("failed to query company", "error", err)
		return fmt.Errorf("query company: %w", err)
	}

	return printJSON(ctx, company)
}
