package trade

import (
	"fmt"
	"log/slog"
	"maps"
	"strings"

	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
)

func runTradeList(cmd *cobra.Command, opts *tradeListOptions) error {
	presetName := opts.Preset
	watchlistName := opts.Watchlist
	tickers := common.MultiTickerValue(cmd)
	fields, err := common.ParseJSONFieldList[models.Trade](opts.Fields)
	if err != nil {
		return fmt.Errorf("parsing fields flag: %w", err)
	}
	format, err := common.ParseOutputFormat(opts.Format)
	if err != nil {
		return err
	}

	lookbackDays := 0
	if tickers != "" {
		lookbackDays = tradeListTickerLookbackDays
	}
	startDate, endDate, err := common.ResolveDateRange(cmd, lookbackDays, false)
	if err != nil {
		return err
	}

	filters := buildTradeFilters(tradesOptionsFromListOptions(opts, tickers, startDate, endDate))
	if presetName != "" || watchlistName != "" {
		if presetName != "" {
			preset, err := findPreset(presetName)
			if err != nil {
				return err
			}
			maps.Copy(filters, preset.filters)
		}
		if watchlistName != "" {
			wlFilters, err := fetchWatchlistFilters(cmd.Context(), watchlistName)
			if err != nil {
				return err
			}
			maps.Copy(filters, wlFilters)
		}
		applyExplicitFlags(cmd, filters)
	}
	if tickers != "" {
		filters["Tickers"] = tickers
	}
	filters["StartDate"] = startDate
	filters["EndDate"] = endDate

	dtOpts := common.DataTableOptions{Start: opts.Start, Length: opts.Length, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Filters: filters, Fields: fields}
	if !opts.Summary && cmd.Flags().Changed("group-by") {
		return fmt.Errorf("--group-by only works with --summary")
	}
	if opts.Summary {
		if len(fields) > 0 {
			return fmt.Errorf("--fields cannot be used with --summary")
		}
		if format != common.OutputFormatJSON {
			return fmt.Errorf("--format cannot be used with --summary")
		}
		return runTradeSummary(cmd, dtOpts, opts.GroupBy, startDate, endDate)
	}
	if format == common.OutputFormatJSON && len(fields) == 0 {
		return runTradeListRows(cmd, dtOpts)
	}
	return common.RunDataTablesCommand[models.Trade](cmd, "/Trades/GetTrades", datatables.TradeColumns, dtOpts, opts.Format, "query trades")
}

func tradesOptionsFromListOptions(opts *tradeListOptions, tickers, startDate, endDate string) *tradesOptions {
	return &tradesOptions{tickers: tickers, startDate: startDate, endDate: endDate, minVolume: opts.MinVolume, maxVolume: opts.MaxVolume, minPrice: opts.MinPrice, maxPrice: opts.MaxPrice, minDollars: opts.MinDollars, maxDollars: opts.MaxDollars, conditions: opts.Conditions, vcd: opts.VCD, securityType: opts.SecurityType, relativeSize: opts.RelativeSize, darkPools: opts.DarkPools.Int(), sweeps: opts.Sweeps.Int(), latePrints: opts.LatePrints.Int(), sigPrints: opts.SigPrints.Int(), evenShared: opts.EvenShared.Int(), tradeRank: opts.TradeRank, rankSnapshot: opts.RankSnapshot, marketCap: opts.MarketCap, premarket: opts.Premarket.Int(), rth: opts.RTH.Int(), ah: opts.AH.Int(), opening: opts.Opening.Int(), closing: opts.Closing.Int(), phantom: opts.Phantom.Int(), offsetting: opts.Offsetting.Int(), sector: opts.Sector}
}

func runTradeListRows(cmd *cobra.Command, opts common.DataTableOptions) error {
	result, err := fetchTradeList(cmd, opts)
	if err != nil {
		return err
	}
	return common.PrintJSON(cmd.OutOrStdout(), cmd.Context(), models.NewTradeListRows(result))
}

func runTradeSummary(cmd *cobra.Command, opts common.DataTableOptions, groupBy tradeSummaryGroup, startDate, endDate string) error {
	group, err := parseTradeSummaryGroup(groupBy)
	if err != nil {
		return err
	}
	trades, err := fetchTradeList(cmd, opts)
	if err != nil {
		return err
	}
	return common.PrintJSON(cmd.OutOrStdout(), cmd.Context(), summarizeTrades(trades, group, startDate, endDate))
}

func fetchTradeList(cmd *cobra.Command, opts common.DataTableOptions) ([]models.Trade, error) {
	ctx := cmd.Context()
	vlClient, err := common.NewCommandClient(ctx)
	if err != nil {
		return nil, err
	}
	request := common.NewDataTablesRequest(datatables.TradeColumns, opts)
	var result []models.Trade
	if err := vlClient.PostDataTables(ctx, "/Trades/GetTrades", request.Encode(), &result); err != nil {
		slog.Error("failed to query trades", "error", err)
		return nil, fmt.Errorf("query trades: %w", err)
	}
	return result, nil
}

type tradeSummaryGroup string

const (
	tradeSummaryGroupTicker    tradeSummaryGroup = "ticker"
	tradeSummaryGroupDay       tradeSummaryGroup = "day"
	tradeSummaryGroupTickerDay tradeSummaryGroup = "ticker,day"
)

func parseTradeSummaryGroup(value tradeSummaryGroup) (tradeSummaryGroup, error) {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(string(value)), " ", ""))
	switch tradeSummaryGroup(normalized) {
	case tradeSummaryGroupTicker, tradeSummaryGroupDay, tradeSummaryGroupTickerDay:
		return tradeSummaryGroup(normalized), nil
	default:
		return "", fmt.Errorf("invalid group-by %q; valid values: ticker, day, ticker,day", value)
	}
}

type tradeGroupAccumulator struct {
	trades                 int
	dollars                float64
	dollarsMultiplier      float64
	darkPool, sweep        int
	cumulativeDistribution float64
}

func summarizeTrades(trades []models.Trade, group tradeSummaryGroup, startDate, endDate string) models.TradeSummary {
	summary := models.TradeSummary{DateRange: models.TradeSummaryDateRange{Start: startDate, End: endDate}}
	groups := make(map[string]tradeGroupAccumulator)
	keyFunc := tradeSummaryKeyFunc(group)
	for i := range trades {
		trade := &trades[i]
		summary.TotalTrades++
		summary.TotalDollars += trade.Dollars
		addTradeSummaryGroup(groups, keyFunc(trade), trade)
	}
	switch group {
	case tradeSummaryGroupTicker:
		summary.ByTicker = summarizeTradeGroups(groups)
	case tradeSummaryGroupDay:
		summary.ByDay = summarizeTradeGroups(groups)
	case tradeSummaryGroupTickerDay:
		summary.ByTickerDay = summarizeTradeGroups(groups)
	}
	return summary
}

func summarizeTradeGroups(groups map[string]tradeGroupAccumulator) map[string]models.TradeGroupSummary {
	summaries := make(map[string]models.TradeGroupSummary, len(groups))
	for key, acc := range groups {
		summaries[key] = acc.summary()
	}
	return summaries
}

func tradeSummaryKeyFunc(group tradeSummaryGroup) func(*models.Trade) string {
	switch group {
	case tradeSummaryGroupDay:
		return tradeDayKey
	case tradeSummaryGroupTickerDay:
		return tradeTickerDayKey
	default:
		return tradeTickerKey
	}
}

func addTradeSummaryGroup(groups map[string]tradeGroupAccumulator, key string, trade *models.Trade) {
	acc := groups[key]
	acc.trades++
	acc.dollars += trade.Dollars
	acc.dollarsMultiplier += trade.DollarsMultiplier
	acc.cumulativeDistribution += trade.CumulativeDistribution
	if bool(trade.DarkPool) {
		acc.darkPool++
	}
	if bool(trade.Sweep) {
		acc.sweep++
	}
	groups[key] = acc
}

func (acc tradeGroupAccumulator) summary() models.TradeGroupSummary {
	if acc.trades == 0 {
		return models.TradeGroupSummary{}
	}
	count := float64(acc.trades)
	return models.TradeGroupSummary{
		Trades: acc.trades, Dollars: acc.dollars,
		AvgDollarsMultiplier:      acc.dollarsMultiplier / count,
		PctDarkPool:               float64(acc.darkPool) / count * 100,
		PctSweep:                  float64(acc.sweep) / count * 100,
		AvgCumulativeDistribution: acc.cumulativeDistribution / count,
	}
}

func tradeTickerKey(trade *models.Trade) string {
	if trade.Ticker == "" {
		return "unknown"
	}
	return trade.Ticker
}

func tradeDayKey(trade *models.Trade) string {
	if !trade.Date.Valid {
		return "unknown"
	}
	return trade.Date.Format("2006-01-02")
}

func tradeTickerDayKey(trade *models.Trade) string {
	return tradeTickerKey(trade) + "|" + tradeDayKey(trade)
}
