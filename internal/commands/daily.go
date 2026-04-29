package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"sort"

	"github.com/major/volumeleaders-agent/internal/client"
	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
	cli "github.com/urfave/cli/v3"
)

const maxDailySummaryPages = 100

// NewDailyCommand returns the "daily" command group.
func NewDailyCommand() *cli.Command {
	return &cli.Command{
		Name:  "daily",
		Usage: "Daily cross-endpoint summaries",
		Commands: []*cli.Command{
			newDailySummaryCommand(),
		},
	}
}

func newDailySummaryCommand() *cli.Command {
	return &cli.Command{
		Name:      "summary",
		Usage:     "Summarize institutional activity for one trading day",
		UsageText: "volumeleaders-agent daily summary --date 2026-04-28 --limit 10",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "date", Required: true, Usage: "Date YYYY-MM-DD"},
			&cli.IntFlag{Name: "limit", Value: 10, Usage: "Maximum rows per summary section"},
		},
		Action: runDailySummary,
	}
}

func runDailySummary(ctx context.Context, cmd *cli.Command) error {
	limit := cmd.Int("limit")
	if limit < 1 {
		return fmt.Errorf("--limit must be greater than 0")
	}

	vlClient, err := newCommandClient(ctx)
	if err != nil {
		return err
	}

	date := cmd.String("date")
	institutional, err := fetchDailyInstitutionalVolume(ctx, vlClient, date)
	if err != nil {
		return err
	}
	clusters, err := fetchDailyClusters(ctx, vlClient, date)
	if err != nil {
		return err
	}
	clusterBombs, err := fetchDailyClusterBombs(ctx, vlClient, date)
	if err != nil {
		return err
	}
	levelTouches, err := fetchDailyLevelTouches(ctx, vlClient, date)
	if err != nil {
		return err
	}
	sentimentTrades, err := fetchDailySentimentTrades(ctx, vlClient, date)
	if err != nil {
		return err
	}
	exhaustion, err := fetchDailyExhaustion(ctx, vlClient, date)
	if err != nil {
		return err
	}

	summary := models.DailySummary{
		Date:                          date,
		TopInstitutionalVolumeTickers: summarizeDailyInstitutionalVolume(institutional, limit),
		TopClustersByDollars:          summarizeDailyClustersByDollars(clusters, limit),
		TopClustersByMultiplier:       summarizeDailyClustersByMultiplier(clusters, limit),
		RepeatedClusterTickers:        summarizeRepeatedClusterTickers(clusters, limit),
		SectorTotals:                  summarizeDailySectorTotals(institutional, clusters, clusterBombs, levelTouches, limit),
		ClusterBombs:                  summarizeDailyClusterBombs(clusterBombs, limit),
		LevelTouches:                  summarizeDailyLevelTouches(levelTouches, limit),
		LeveragedETFSentiment:         summarizeDailyLeveragedETFSentiment(sentimentTrades, date),
		MarketExhaustion:              exhaustion,
	}
	return printJSON(ctx, summary)
}

func fetchDailyInstitutionalVolume(ctx context.Context, vlClient *client.Client, date string) ([]models.Trade, error) {
	return fetchDailyDataTables[models.Trade](ctx, vlClient, "/InstitutionalVolume/GetInstitutionalVolume", datatables.InstitutionalVolumeColumns, dataTableOptions{
		orderCol: 6,
		orderDir: "desc",
		filters: map[string]string{
			"Date":    date,
			"Tickers": "",
		},
	}, "query daily institutional volume")
}

func fetchDailyClusters(ctx context.Context, vlClient *client.Client, date string) ([]models.TradeCluster, error) {
	return fetchDailyDataTables[models.TradeCluster](ctx, vlClient, "/TradeClusters/GetTradeClusters", datatables.TradeClusterColumns, dataTableOptions{
		orderCol: 9,
		orderDir: "desc",
		filters: map[string]string{
			"Tickers":          "",
			"StartDate":        date,
			"EndDate":          date,
			"MinVolume":        "0",
			"MaxVolume":        "2000000000",
			"MinPrice":         "0",
			"MaxPrice":         "100000",
			"MinDollars":       "10000000",
			"MaxDollars":       "30000000000",
			"VCD":              "0",
			"SecurityTypeKey":  "-1",
			"RelativeSize":     "5",
			"TradeClusterRank": "-1",
			"SectorIndustry":   "",
		},
	}, "query daily trade clusters")
}

func fetchDailyClusterBombs(ctx context.Context, vlClient *client.Client, date string) ([]models.TradeClusterBomb, error) {
	return fetchDailyDataTables[models.TradeClusterBomb](ctx, vlClient, "/TradeClusterBombs/GetTradeClusterBombs", datatables.TradeClusterBombColumns, dataTableOptions{
		orderCol: 7,
		orderDir: "desc",
		filters: map[string]string{
			"Tickers":              "",
			"StartDate":            date,
			"EndDate":              date,
			"MinVolume":            "0",
			"MaxVolume":            "2000000000",
			"MinDollars":           "0",
			"MaxDollars":           "30000000000",
			"VCD":                  "0",
			"SecurityTypeKey":      "0",
			"RelativeSize":         "0",
			"TradeClusterBombRank": "-1",
			"SectorIndustry":       "",
		},
	}, "query daily trade cluster bombs")
}

func fetchDailyLevelTouches(ctx context.Context, vlClient *client.Client, date string) ([]models.TradeLevelTouch, error) {
	return fetchDailyDataTables[models.TradeLevelTouch](ctx, vlClient, "/TradeLevelTouches/GetTradeLevelTouches", datatables.TradeLevelTouchColumns, dataTableOptions{
		orderCol: 8,
		orderDir: "desc",
		filters: map[string]string{
			"Tickers":        "",
			"StartDate":      date,
			"EndDate":        date,
			"MinVolume":      "0",
			"MaxVolume":      "2000000000",
			"MinPrice":       "0",
			"MaxPrice":       "100000",
			"MinDollars":     "500000",
			"MaxDollars":     "30000000000",
			"VCD":            "0",
			"RelativeSize":   "0",
			"TradeLevelRank": "10",
		},
	}, "query daily trade level touches")
}

func fetchDailySentimentTrades(ctx context.Context, vlClient *client.Client, date string) ([]models.Trade, error) {
	opts := &tradesOptions{
		startDate:    date,
		endDate:      date,
		minVolume:    0,
		maxVolume:    2000000000,
		minPrice:     0,
		maxPrice:     100000,
		minDollars:   5000000,
		maxDollars:   30000000000,
		conditions:   -1,
		vcd:          97,
		securityType: -1,
		relativeSize: 5,
		darkPools:    -1,
		sweeps:       -1,
		latePrints:   -1,
		sigPrints:    -1,
		evenShared:   -1,
		tradeRank:    -1,
		rankSnapshot: -1,
		marketCap:    0,
		premarket:    1,
		rth:          1,
		ah:           1,
		opening:      1,
		closing:      1,
		phantom:      1,
		offsetting:   1,
		sector:       "X B",
	}
	return fetchTradeSentimentTrades(ctx, vlClient, dataTableOptions{
		orderCol: 1,
		orderDir: "desc",
		length:   maxTradeRequestLength,
		filters:  buildTradeFilters(opts),
	})
}

func fetchDailyExhaustion(ctx context.Context, vlClient *client.Client, date string) (models.ExhaustionScore, error) {
	var exhaustion models.ExhaustionScore
	if err := vlClient.PostJSON(ctx, "/ExecutiveSummary/GetExhaustionScores", map[string]string{"Date": date}, &exhaustion); err != nil {
		slog.Error("failed to query daily market exhaustion", "error", err)
		return models.ExhaustionScore{}, fmt.Errorf("query daily market exhaustion: %w", err)
	}
	return exhaustion, nil
}

func fetchDailyDataTables[T any](ctx context.Context, vlClient *client.Client, path string, columns []string, opts dataTableOptions, label string) ([]T, error) {
	opts.length = paginationPageSize
	all := make([]T, 0)
	for range maxDailySummaryPages {
		request := newDataTablesRequest(columns, opts)
		resp, err := vlClient.PostDataTablesPage(ctx, path, request.Encode())
		if err != nil {
			slog.Error("failed to "+label, "error", err)
			return nil, fmt.Errorf("%s: %w", label, err)
		}

		var page []T
		if err := json.Unmarshal(resp.Data, &page); err != nil {
			slog.Error("failed to decode "+label, "error", err)
			return nil, fmt.Errorf("%s: decode response: %w", label, err)
		}
		if len(page) == 0 {
			return all, nil
		}

		all = append(all, page...)
		if resp.RecordsFiltered > 0 && len(all) >= resp.RecordsFiltered {
			return all, nil
		}
		if len(page) < paginationPageSize {
			return all, nil
		}
		opts.start += len(page)
	}

	return nil, fmt.Errorf("%s: pagination exceeded %d pages", label, maxDailySummaryPages)
}

func summarizeDailyInstitutionalVolume(trades []models.Trade, limit int) []models.DailyInstitutionalVolumeRow {
	sorted := slices.Clone(trades)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].TotalInstitutionalDollars == sorted[j].TotalInstitutionalDollars {
			return sorted[i].Ticker < sorted[j].Ticker
		}
		return sorted[i].TotalInstitutionalDollars > sorted[j].TotalInstitutionalDollars
	})
	rows := make([]models.DailyInstitutionalVolumeRow, 0, min(limit, len(sorted)))
	for i := range min(limit, len(sorted)) {
		trade := &sorted[i]
		rows = append(rows, models.DailyInstitutionalVolumeRow{
			Ticker:                       trade.Ticker,
			Sector:                       trade.Sector,
			Price:                        trade.Price,
			Volume:                       trade.Volume,
			TotalInstitutionalDollars:    trade.TotalInstitutionalDollars,
			TotalInstitutionalDollarRank: trade.TotalInstitutionalDollarsRank,
		})
	}
	return rows
}

func summarizeDailyClustersByDollars(clusters []models.TradeCluster, limit int) []models.DailyClusterRow {
	sorted := slices.Clone(clusters)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Dollars == sorted[j].Dollars {
			return sorted[i].Ticker < sorted[j].Ticker
		}
		return sorted[i].Dollars > sorted[j].Dollars
	})
	return dailyClusterRows(sorted, limit)
}

func summarizeDailyClustersByMultiplier(clusters []models.TradeCluster, limit int) []models.DailyClusterRow {
	sorted := slices.Clone(clusters)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].DollarsMultiplier == sorted[j].DollarsMultiplier {
			return sorted[i].Ticker < sorted[j].Ticker
		}
		return sorted[i].DollarsMultiplier > sorted[j].DollarsMultiplier
	})
	return dailyClusterRows(sorted, limit)
}

func dailyClusterRows(clusters []models.TradeCluster, limit int) []models.DailyClusterRow {
	rows := make([]models.DailyClusterRow, 0, min(limit, len(clusters)))
	for i := range min(limit, len(clusters)) {
		cluster := &clusters[i]
		rows = append(rows, models.DailyClusterRow{
			Ticker:                 cluster.Ticker,
			Sector:                 cluster.Sector,
			Dollars:                cluster.Dollars,
			DollarsMultiplier:      cluster.DollarsMultiplier,
			Volume:                 cluster.Volume,
			TradeCount:             cluster.TradeCount,
			TradeClusterRank:       cluster.TradeClusterRank,
			CumulativeDistribution: cluster.CumulativeDistribution,
		})
	}
	return rows
}

type dailyRepeatedClusterAccumulator struct {
	sector        string
	clusters      int
	dollars       float64
	tradeCount    int
	maxMultiplier float64
}

func summarizeRepeatedClusterTickers(clusters []models.TradeCluster, limit int) []models.DailyRepeatedClusterTicker {
	byTicker := make(map[string]dailyRepeatedClusterAccumulator)
	for i := range clusters {
		cluster := &clusters[i]
		acc := byTicker[cluster.Ticker]
		acc.sector = cluster.Sector
		acc.clusters++
		acc.dollars += cluster.Dollars
		acc.tradeCount += cluster.TradeCount
		acc.maxMultiplier = max(acc.maxMultiplier, cluster.DollarsMultiplier)
		byTicker[cluster.Ticker] = acc
	}

	rows := make([]models.DailyRepeatedClusterTicker, 0, len(byTicker))
	for ticker, acc := range byTicker {
		if acc.clusters < 2 {
			continue
		}
		rows = append(rows, models.DailyRepeatedClusterTicker{
			Ticker:        ticker,
			Sector:        acc.sector,
			Clusters:      acc.clusters,
			Dollars:       acc.dollars,
			TradeCount:    acc.tradeCount,
			MaxMultiplier: acc.maxMultiplier,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Clusters == rows[j].Clusters {
			if rows[i].Dollars == rows[j].Dollars {
				return rows[i].Ticker < rows[j].Ticker
			}
			return rows[i].Dollars > rows[j].Dollars
		}
		return rows[i].Clusters > rows[j].Clusters
	})
	return rows[:min(limit, len(rows))]
}

type dailySectorAccumulator struct {
	tickers    map[string]struct{}
	trades     int
	dollars    float64
	volume     int
	tradeCount int
}

func summarizeDailySectorTotals(trades []models.Trade, clusters []models.TradeCluster, bombs []models.TradeClusterBomb, touches []models.TradeLevelTouch, limit int) []models.DailySectorTotal {
	bySector := make(map[string]dailySectorAccumulator)
	for i := range trades {
		trade := &trades[i]
		acc := sectorAccumulator(bySector, trade.Sector)
		acc.tickers[trade.Ticker] = struct{}{}
		acc.trades++
		acc.dollars += trade.TotalInstitutionalDollars
		acc.volume += trade.Volume
		bySector[sectorName(trade.Sector)] = acc
	}
	for i := range clusters {
		cluster := &clusters[i]
		acc := sectorAccumulator(bySector, cluster.Sector)
		acc.tickers[cluster.Ticker] = struct{}{}
		acc.dollars += cluster.Dollars
		acc.volume += cluster.Volume
		acc.tradeCount += cluster.TradeCount
		bySector[sectorName(cluster.Sector)] = acc
	}
	for i := range bombs {
		bomb := &bombs[i]
		acc := sectorAccumulator(bySector, bomb.Sector)
		acc.tickers[bomb.Ticker] = struct{}{}
		acc.dollars += bomb.Dollars
		acc.volume += bomb.Volume
		acc.tradeCount += bomb.TradeCount
		bySector[sectorName(bomb.Sector)] = acc
	}
	for i := range touches {
		touch := &touches[i]
		sector := ""
		if touch.Sector != nil {
			sector = *touch.Sector
		}
		acc := sectorAccumulator(bySector, sector)
		acc.tickers[touch.Ticker] = struct{}{}
		acc.dollars += touch.Dollars
		acc.volume += touch.Volume
		acc.tradeCount += touch.Trades
		bySector[sectorName(sector)] = acc
	}

	rows := make([]models.DailySectorTotal, 0, len(bySector))
	for sector, acc := range bySector {
		rows = append(rows, models.DailySectorTotal{
			Sector:     sector,
			Tickers:    len(acc.tickers),
			Trades:     acc.trades,
			Dollars:    acc.dollars,
			Volume:     acc.volume,
			TradeCount: acc.tradeCount,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Dollars == rows[j].Dollars {
			return rows[i].Sector < rows[j].Sector
		}
		return rows[i].Dollars > rows[j].Dollars
	})
	return rows[:min(limit, len(rows))]
}

func sectorAccumulator(items map[string]dailySectorAccumulator, sector string) dailySectorAccumulator {
	acc := items[sectorName(sector)]
	if acc.tickers == nil {
		acc.tickers = make(map[string]struct{})
	}
	return acc
}

func sectorName(sector string) string {
	if sector == "" {
		return "unknown"
	}
	return sector
}

func summarizeDailyClusterBombs(bombs []models.TradeClusterBomb, limit int) []models.DailyClusterBombRow {
	sorted := slices.Clone(bombs)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Dollars == sorted[j].Dollars {
			return sorted[i].Ticker < sorted[j].Ticker
		}
		return sorted[i].Dollars > sorted[j].Dollars
	})
	rows := make([]models.DailyClusterBombRow, 0, min(limit, len(sorted)))
	for i := range min(limit, len(sorted)) {
		bomb := &sorted[i]
		rows = append(rows, models.DailyClusterBombRow{
			Ticker:                 bomb.Ticker,
			Sector:                 bomb.Sector,
			Dollars:                bomb.Dollars,
			DollarsMultiplier:      bomb.DollarsMultiplier,
			Volume:                 bomb.Volume,
			TradeCount:             bomb.TradeCount,
			TradeClusterBombRank:   bomb.TradeClusterBombRank,
			CumulativeDistribution: bomb.CumulativeDistribution,
		})
	}
	return rows
}

func summarizeDailyLevelTouches(touches []models.TradeLevelTouch, limit int) models.DailyLevelTouchSummary {
	byRelativeSize := slices.Clone(touches)
	sort.Slice(byRelativeSize, func(i, j int) bool {
		if byRelativeSize[i].RelativeSize == byRelativeSize[j].RelativeSize {
			return byRelativeSize[i].Ticker < byRelativeSize[j].Ticker
		}
		return byRelativeSize[i].RelativeSize > byRelativeSize[j].RelativeSize
	})

	byDollars := slices.Clone(touches)
	sort.Slice(byDollars, func(i, j int) bool {
		if byDollars[i].Dollars == byDollars[j].Dollars {
			return byDollars[i].Ticker < byDollars[j].Ticker
		}
		return byDollars[i].Dollars > byDollars[j].Dollars
	})

	return models.DailyLevelTouchSummary{
		ByRelativeSize: dailyLevelTouchRows(byRelativeSize, limit),
		ByDollars:      dailyLevelTouchRows(byDollars, limit),
	}
}

func dailyLevelTouchRows(touches []models.TradeLevelTouch, limit int) []models.DailyLevelTouchRow {
	rows := make([]models.DailyLevelTouchRow, 0, min(limit, len(touches)))
	for i := range min(limit, len(touches)) {
		touch := &touches[i]
		sector := ""
		if touch.Sector != nil {
			sector = *touch.Sector
		}
		rows = append(rows, models.DailyLevelTouchRow{
			Ticker:            touch.Ticker,
			Sector:            sector,
			Price:             touch.Price,
			Dollars:           touch.Dollars,
			RelativeSize:      touch.RelativeSize,
			Volume:            touch.Volume,
			Trades:            touch.Trades,
			TradeLevelRank:    touch.TradeLevelRank,
			TradeLevelTouches: touch.TradeLevelTouches,
		})
	}
	return rows
}

func summarizeDailyLeveragedETFSentiment(trades []models.Trade, date string) models.DailyLeveragedETFSentiment {
	sentiment := summarizeTradeSentiment(trades, date, date)
	return models.DailyLeveragedETFSentiment{
		Bear:   sentiment.Totals.Bear,
		Bull:   sentiment.Totals.Bull,
		Ratio:  sentiment.Totals.Ratio,
		Signal: sentiment.Totals.Signal,
	}
}
