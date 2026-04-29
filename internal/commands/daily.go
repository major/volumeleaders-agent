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
			&cli.IntFlag{Name: "limit", Value: 10, Usage: "Rows considered per daily summary ranking"},
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
		Date:                  date,
		InstitutionalVolume:   summarizeDailyInstitutionalVolume(institutional, limit),
		Clusters:              summarizeDailyClusters(clusters, limit),
		ClusterBombs:          summarizeDailyClusterBombs(clusterBombs, limit),
		LevelTouches:          summarizeDailyLevelTouches(levelTouches, limit),
		LeveragedETFSentiment: summarizeDailyLeveragedETFSentiment(sentimentTrades, date),
		MarketExhaustion:      summarizeDailyMarketExhaustion(exhaustion),
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

func summarizeDailyInstitutionalVolume(trades []models.Trade, limit int) []models.DailyInstitutionalVolume {
	sorted := slices.Clone(trades)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].TotalInstitutionalDollars == sorted[j].TotalInstitutionalDollars {
			return sorted[i].Ticker < sorted[j].Ticker
		}
		return sorted[i].TotalInstitutionalDollars > sorted[j].TotalInstitutionalDollars
	})
	rows := make([]models.DailyInstitutionalVolume, 0, min(limit, len(sorted)))
	for i := range min(limit, len(sorted)) {
		trade := &sorted[i]
		rows = append(rows, models.DailyInstitutionalVolume{
			Ticker:               trade.Ticker,
			Sector:               trade.Sector,
			Price:                trade.Price,
			InstitutionalDollars: trade.TotalInstitutionalDollars,
			Rank:                 trade.TotalInstitutionalDollarsRank,
		})
	}
	return rows
}

func summarizeDailyClusters(clusters []models.TradeCluster, limit int) models.DailyClusterSummary {
	byDollars := slices.Clone(clusters)
	sort.Slice(byDollars, func(i, j int) bool {
		if byDollars[i].Dollars == byDollars[j].Dollars {
			return byDollars[i].Ticker < byDollars[j].Ticker
		}
		return byDollars[i].Dollars > byDollars[j].Dollars
	})

	byMultiplier := slices.Clone(clusters)
	sort.Slice(byMultiplier, func(i, j int) bool {
		if byMultiplier[i].DollarsMultiplier == byMultiplier[j].DollarsMultiplier {
			return byMultiplier[i].Ticker < byMultiplier[j].Ticker
		}
		return byMultiplier[i].DollarsMultiplier > byMultiplier[j].DollarsMultiplier
	})

	return models.DailyClusterSummary{
		Top:             dailyClusterRows(byDollars, byMultiplier, limit),
		RepeatedTickers: summarizeRepeatedClusterTickers(clusters, limit),
	}
}

func dailyClusterRows(byDollars, byMultiplier []models.TradeCluster, limit int) []models.DailyCluster {
	rowsByKey := make(map[string]models.DailyCluster)
	keys := make([]string, 0, min(limit, len(byDollars))+min(limit, len(byMultiplier)))

	addDailyClusterRows(rowsByKey, &keys, byDollars, limit, "dollars")
	addDailyClusterRows(rowsByKey, &keys, byMultiplier, limit, "multiplier")

	rows := make([]models.DailyCluster, 0, len(keys))
	for _, key := range keys {
		rows = append(rows, rowsByKey[key])
	}
	return rows
}

func addDailyClusterRows(rowsByKey map[string]models.DailyCluster, keys *[]string, clusters []models.TradeCluster, limit int, topBy string) {
	for i := range min(limit, len(clusters)) {
		cluster := &clusters[i]
		key := dailyClusterKey(cluster)
		row, ok := rowsByKey[key]
		if !ok {
			row = models.DailyCluster{
				Ticker:                 cluster.Ticker,
				Sector:                 cluster.Sector,
				Dollars:                cluster.Dollars,
				DollarsMultiplier:      cluster.DollarsMultiplier,
				TradeCount:             cluster.TradeCount,
				Rank:                   cluster.TradeClusterRank,
				CumulativeDistribution: cluster.CumulativeDistribution,
			}
			*keys = append(*keys, key)
		}
		row.TopBy = appendUniqueDailyTopBy(row.TopBy, topBy)
		rowsByKey[key] = row
	}
}

func dailyClusterKey(cluster *models.TradeCluster) string {
	return fmt.Sprintf("%s|%.6f|%d|%d", cluster.Ticker, cluster.Dollars, cluster.TradeCount, cluster.TradeClusterRank)
}

func appendUniqueDailyTopBy(topBy []string, value string) []string {
	if slices.Contains(topBy, value) {
		return topBy
	}
	return append(topBy, value)
}

type dailyRepeatedClusterAccumulator struct {
	sector        string
	clusters      int
	dollars       float64
	tradeCount    int
	maxMultiplier float64
	bestRank      int
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
		if acc.bestRank == 0 || cluster.TradeClusterRank < acc.bestRank {
			acc.bestRank = cluster.TradeClusterRank
		}
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
			BestRank:      acc.bestRank,
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

func summarizeDailyClusterBombs(bombs []models.TradeClusterBomb, limit int) []models.DailyClusterBomb {
	sorted := slices.Clone(bombs)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Dollars == sorted[j].Dollars {
			return sorted[i].Ticker < sorted[j].Ticker
		}
		return sorted[i].Dollars > sorted[j].Dollars
	})
	rows := make([]models.DailyClusterBomb, 0, min(limit, len(sorted)))
	for i := range min(limit, len(sorted)) {
		bomb := &sorted[i]
		rows = append(rows, models.DailyClusterBomb{
			Ticker:                 bomb.Ticker,
			Sector:                 bomb.Sector,
			Dollars:                bomb.Dollars,
			DollarsMultiplier:      bomb.DollarsMultiplier,
			TradeCount:             bomb.TradeCount,
			Rank:                   bomb.TradeClusterBombRank,
			CumulativeDistribution: bomb.CumulativeDistribution,
		})
	}
	return rows
}

func summarizeDailyLevelTouches(touches []models.TradeLevelTouch, limit int) []models.DailyLevelTouch {
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

	return dailyLevelTouchRows(byRelativeSize, byDollars, limit)
}

func dailyLevelTouchRows(byRelativeSize, byDollars []models.TradeLevelTouch, limit int) []models.DailyLevelTouch {
	rowsByKey := make(map[string]models.DailyLevelTouch)
	keys := make([]string, 0, min(limit, len(byRelativeSize))+min(limit, len(byDollars)))

	addDailyLevelTouchRows(rowsByKey, &keys, byRelativeSize, limit, "relative_size")
	addDailyLevelTouchRows(rowsByKey, &keys, byDollars, limit, "dollars")

	rows := make([]models.DailyLevelTouch, 0, len(keys))
	for _, key := range keys {
		rows = append(rows, rowsByKey[key])
	}
	return rows
}

func addDailyLevelTouchRows(rowsByKey map[string]models.DailyLevelTouch, keys *[]string, touches []models.TradeLevelTouch, limit int, topBy string) {
	for i := range min(limit, len(touches)) {
		touch := &touches[i]
		key := dailyLevelTouchKey(touch)
		row, ok := rowsByKey[key]
		if !ok {
			row = dailyLevelTouchRow(touch)
			*keys = append(*keys, key)
		}
		row.TopBy = appendUniqueDailyTopBy(row.TopBy, topBy)
		rowsByKey[key] = row
	}
}

func dailyLevelTouchRow(touch *models.TradeLevelTouch) models.DailyLevelTouch {
	sector := ""
	if touch.Sector != nil {
		sector = *touch.Sector
	}
	return models.DailyLevelTouch{
		Ticker:                 touch.Ticker,
		Sector:                 sector,
		Price:                  touch.Price,
		Dollars:                touch.Dollars,
		RelativeSize:           touch.RelativeSize,
		Trades:                 touch.Trades,
		Rank:                   touch.TradeLevelRank,
		Touches:                touch.TradeLevelTouches,
		CumulativeDistribution: touch.CumulativeDistribution,
	}
}

func dailyLevelTouchKey(touch *models.TradeLevelTouch) string {
	return fmt.Sprintf("%s|%.6f|%.6f|%d", touch.Ticker, touch.Price, touch.Dollars, touch.TradeLevelRank)
}

func summarizeDailyLeveragedETFSentiment(trades []models.Trade, date string) models.DailyLeveragedETFSentiment {
	sentiment := summarizeTradeSentiment(trades, date, date)
	return models.DailyLeveragedETFSentiment{
		Signal:      sentiment.Totals.Signal,
		Ratio:       sentiment.Totals.Ratio,
		BullDollars: sentiment.Totals.Bull.Dollars,
		BearDollars: sentiment.Totals.Bear.Dollars,
		BullTrades:  sentiment.Totals.Bull.Trades,
		BearTrades:  sentiment.Totals.Bear.Trades,
		BullTickers: sentiment.Totals.Bull.TopTickers,
		BearTickers: sentiment.Totals.Bear.TopTickers,
	}
}

func summarizeDailyMarketExhaustion(exhaustion models.ExhaustionScore) models.DailyMarketExhaustion {
	return models.DailyMarketExhaustion{
		Rank:     exhaustion.ExhaustionScoreRank,
		Rank30D:  exhaustion.ExhaustionScoreRank30Day,
		Rank90D:  exhaustion.ExhaustionScoreRank90Day,
		Rank365D: exhaustion.ExhaustionScoreRank365Day,
	}
}
