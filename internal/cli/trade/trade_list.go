package trade

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"net/url"
	"strings"

	vlgo "github.com/major/volumeleaders-go/volumeleaders"
	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
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

	dtOpts := common.NewDataTableOptions(common.DataTableRequestConfig{Start: opts.Start, Length: -1, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Filters: filters, Fields: fields})
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
	if shouldUseLongTermTradeList(dtOpts, startDate, endDate) {
		dtOpts = longTermTradeListOptions(cmd, dtOpts)
		trades, err := fetchTradeList(cmd, dtOpts)
		if err != nil {
			return err
		}
		if format == common.OutputFormatJSON && len(fields) == 0 {
			return common.PrintJSON(cmd.OutOrStdout(), cmd.Context(), models.NewTradeListRows(trades))
		}
		return common.PrintDataTablesResult(cmd.OutOrStdout(), cmd.Context(), trades, fields, format)
	}
	if format == common.OutputFormatJSON && len(fields) == 0 {
		return runTradeListRows(cmd, dtOpts)
	}
	return common.RunVLDataTablesCommandWithPageSize[vlgo.Trade, models.Trade](cmd, dtOpts, opts.Format, tradeBrowserPageLength, "query trades", fetchTrades, common.MapVLTrade)
}

func shouldUseLongTermTradeList(opts common.DataTableOptions, startDate, endDate string) bool {
	return opts.Filters["Tickers"] != "" && startDate != endDate
}

func longTermTradeListOptions(cmd *cobra.Command, opts common.DataTableOptions) common.DataTableOptions {
	filters := maps.Clone(opts.Filters)
	filters["Sort"] = "Dollars"
	if !cmd.Flags().Changed("vcd") {
		filters["VCD"] = "0"
	}
	if !cmd.Flags().Changed("relative-size") {
		filters["RelativeSize"] = "0"
	}
	deleteDefaultOnlyFilter(cmd, filters, "security-type", "SecurityTypeKey")
	deleteDefaultOnlyFilter(cmd, filters, "even-shared", "EvenShared")
	deleteDefaultOnlyFilter(cmd, filters, "rank-snapshot", "TradeRankSnapshot")
	deleteDefaultOnlyFilter(cmd, filters, "market-cap", "MarketCap")
	opts.Filters = filters
	opts.Start = 0
	opts.Length = tradeListLongTermLength
	opts.OrderCol = 0
	opts.OrderDir = common.OrderDirection("DESC")
	return opts
}

func deleteDefaultOnlyFilter(cmd *cobra.Command, filters map[string]string, flagName, filterName string) {
	if !cmd.Flags().Changed(flagName) {
		delete(filters, filterName)
	}
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
	vlClient, err := common.NewVLClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create VL client: %w", err)
	}
	if opts.Length < 0 {
		vlTrades, err := common.FetchVLPages[vlgo.Trade](ctx, vlClient, opts, tradeBrowserPageLength, "query trades", fetchTrades)
		if err != nil {
			return nil, fmt.Errorf("fetch trades: %w", err)
		}
		return common.MapSlice(vlTrades, common.MapVLTrade), nil
	}
	dtReq := newTradeListVLRequest(opts)
	resp, err := vlClient.GetTrades(ctx, vlgo.TradesRequest{DataTables: dtReq, Filters: common.FiltersToValues(opts.Filters)})
	if err != nil {
		slog.Error("failed to query trades", "error", err)
		return nil, fmt.Errorf("query trades: %w", err)
	}
	return common.MapSlice(resp.Data, common.MapVLTrade), nil
}

// newTradeListVLRequest builds a vlgo DataTables request for trade list queries.
// The long-term chart path intentionally asks for 10 rows even though the
// observed browser HAR requested 5. The project default follows the user's
// preference for a slightly broader top-N result while preserving the
// lightweight chart request shape that avoids backend timeouts.
func newTradeListVLRequest(opts common.DataTableOptions) vlgo.DataTablesRequest {
	if opts.Filters["Sort"] == "Dollars" && opts.Length == tradeListLongTermLength {
		return newDashboardDTRequest(tradeChartColumns, opts.OrderCol, "FullTimeString24", opts.Length)
	}
	return common.NewVLDataTablesRequest(opts)
}

// fetchTrades wraps vlgo.Client.GetTrades as a VLFetcher.
func fetchTrades(ctx context.Context, c *vlgo.Client, dt *vlgo.DataTablesRequest, filters url.Values) (*vlgo.DataTablesResponse[vlgo.Trade], error) {
	return c.GetTrades(ctx, vlgo.TradesRequest{DataTables: *dt, Filters: filters})
}

type tradeSummaryGroup string

const (
	tradeSummaryGroupTicker    tradeSummaryGroup = "ticker"
	tradeSummaryGroupDay       tradeSummaryGroup = "day"
	tradeSummaryGroupTickerDay tradeSummaryGroup = "ticker,day"
)

func parseTradeSummaryGroup(value tradeSummaryGroup) (tradeSummaryGroup, error) {
	normalized := strings.ToLower(strings.TrimSpace(string(value)))
	normalized = strings.NewReplacer(" ", "", "-", ",").Replace(normalized)
	if normalized == "tickerday" {
		normalized = string(tradeSummaryGroupTickerDay)
	}
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
	closingTrade           int
	cumulativeDistribution float64
	maxDollars             float64
	maxDollarsMultiplier   float64
	minTradeRank           int
	hasTradeRank           bool
	topPrice               float64
	latestTradeTime        string
	topTradeTime           string
	topDarkPool            bool
	topSweep               bool
	topClosingTrade        bool
}

func summarizeTrades(trades []models.Trade, group tradeSummaryGroup, startDate, endDate string) models.TradeSummary {
	summary := models.TradeSummary{DateRange: models.TradeSummaryDateRange{Start: startDate, End: endDate}}
	groups := make(map[string]*tradeGroupAccumulator)
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

func summarizeTradeGroups(groups map[string]*tradeGroupAccumulator) map[string]models.TradeGroupSummary {
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

func addTradeSummaryGroup(groups map[string]*tradeGroupAccumulator, key string, trade *models.Trade) {
	acc := groups[key]
	if acc == nil {
		acc = &tradeGroupAccumulator{}
		groups[key] = acc
	}
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
	if bool(trade.ClosingTrade) {
		acc.closingTrade++
	}
	if trade.Dollars > acc.maxDollars {
		acc.maxDollars = trade.Dollars
		acc.topPrice = trade.Price
		acc.topTradeTime = tradeSummaryTradeTime(trade)
		acc.topDarkPool = bool(trade.DarkPool)
		acc.topSweep = bool(trade.Sweep)
		acc.topClosingTrade = bool(trade.ClosingTrade)
	}
	if trade.DollarsMultiplier > acc.maxDollarsMultiplier {
		acc.maxDollarsMultiplier = trade.DollarsMultiplier
	}
	if trade.TradeRank > 0 && (!acc.hasTradeRank || trade.TradeRank < acc.minTradeRank) {
		acc.minTradeRank = trade.TradeRank
		acc.hasTradeRank = true
	}
	if tradeTime := tradeSummaryTradeTime(trade); tradeTime > acc.latestTradeTime {
		acc.latestTradeTime = tradeTime
	}
}

func (acc *tradeGroupAccumulator) summary() models.TradeGroupSummary {
	if acc.trades == 0 {
		return models.TradeGroupSummary{}
	}
	count := float64(acc.trades)
	return models.TradeGroupSummary{
		Trades: acc.trades, Dollars: acc.dollars,
		AvgDollarsMultiplier:      acc.dollarsMultiplier / count,
		PctDarkPool:               float64(acc.darkPool) / count * 100,
		PctSweep:                  float64(acc.sweep) / count * 100,
		PctClosingTrade:           float64(acc.closingTrade) / count * 100,
		AvgCumulativeDistribution: acc.cumulativeDistribution / count,
		MaxDollars:                acc.maxDollars,
		MaxDollarsMultiplier:      acc.maxDollarsMultiplier,
		MinTradeRank:              acc.minTradeRank,
		TopPrice:                  acc.topPrice,
		LatestTradeTime:           acc.latestTradeTime,
		TopTradeTime:              acc.topTradeTime,
		TopDarkPool:               acc.topDarkPool,
		TopSweep:                  acc.topSweep,
		TopClosingTrade:           acc.topClosingTrade,
	}
}

func tradeSummaryTradeTime(trade *models.Trade) string {
	if trade.FullDateTime != nil && *trade.FullDateTime != "" {
		return *trade.FullDateTime
	}
	if trade.FullTimeString24 != nil && *trade.FullTimeString24 != "" {
		return *trade.FullTimeString24
	}
	return ""
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
