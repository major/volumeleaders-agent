package trade

import (
	"fmt"
	"log/slog"
	"maps"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/client"
	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
)

const (
	defaultTradeRequestLength   = 10
	maxTradeRequestLength       = 50
	maxTradeLevelRequestLength  = 50
	tradeListTickerLookbackDays = 365
)

var tradeClusterDefaultFields = []string{
	"Date",
	"Ticker",
	"Price",
	"Dollars",
	"Volume",
	"TradeCount",
	"DollarsMultiplier",
	"CumulativeDistribution",
	"TradeClusterRank",
	"MinFullDateTime",
	"MaxFullDateTime",
}

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

// NewCmd returns the "trade" command group with all subcommands.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "trade", Short: "Trade-related commands", GroupID: "trading"}
	cmd.AddCommand(
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
	)
	return cmd
}

// NewTradeCommand returns the "trade" command group with all subcommands.
func NewTradeCommand() *cobra.Command { return NewCmd() }

func newTradeListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [tickers...]",
		Short: "Query institutional trades",
		Example: `volumeleaders-agent trade list AAPL MSFT
volumeleaders-agent trade list --tickers AAPL,MSFT
volumeleaders-agent trade list --tickers NVDA --dark-pools 1 --min-dollars 1000000
volumeleaders-agent trade list --sector Technology --relative-size 10 --length 50
volumeleaders-agent trade list --preset "Top-100 Rank" --start-date 2025-04-01 --end-date 2025-04-24
volumeleaders-agent trade list --watchlist "Magnificent 7" --start-date 2025-04-01 --end-date 2025-04-24`,
		RunE: runTradeList,
	}
	common.AddOptionalDateRangeFlags(cmd)
	addTradeRangeFlags(cmd, 500000)
	common.AddTickersFlag(cmd)
	addTradeFilterFlags(cmd, 97)
	cmd.Flags().String("sector", "", "Sector/Industry filter")
	cmd.Flags().String("preset", "", "Apply a built-in filter preset (see: trade presets)")
	cmd.Flags().String("watchlist", "", "Apply filters from a saved watchlist by name")
	cmd.Flags().String("fields", "", "Comma-separated trade fields to include in output")
	cmd.Flags().Bool("summary", false, "Return aggregate metrics instead of individual trades")
	cmd.Flags().String("group-by", "ticker", "Summary grouping (requires --summary): ticker, day, or ticker,day")
	common.AddOutputFormatFlags(cmd)
	common.AddPaginationFlags(cmd, defaultTradeRequestLength, 1, "desc")
	return cmd
}

func newTradeSentimentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sentiment",
		Short:   "Summarize leveraged ETF bull/bear flow by day",
		Example: "volumeleaders-agent trade sentiment --start-date 2025-04-21 --end-date 2025-04-25",
		RunE:    runTradeSentiment,
	}
	common.AddDateRangeFlags(cmd)
	addTradeRangeFlags(cmd, 5000000)
	addTradeFilterFlags(cmd, 97)
	common.AddOutputFormatFlags(cmd)
	return cmd
}

func newTradeClustersCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clusters [tickers...]",
		Short:   "Query aggregated trade clusters",
		Example: "volumeleaders-agent trade clusters AAPL --days 7",
		RunE:    runTradeClusters,
	}
	common.AddDateRangeFlags(cmd)
	addTradeRangeFlags(cmd, 10000000)
	common.AddTickersFlag(cmd)
	cmd.Flags().Int("vcd", 0, "VCD filter")
	cmd.Flags().Int("security-type", -1, "Security type key")
	cmd.Flags().Int("relative-size", 5, "Relative size threshold")
	cmd.Flags().Int("trade-cluster-rank", -1, "Trade cluster rank filter")
	cmd.Flags().String("sector", "", "Sector/Industry filter")
	cmd.Flags().String("fields", "", "Comma-separated TradeCluster fields to include in output, or 'all' for every field")
	common.AddOutputFormatFlags(cmd)
	common.AddPaginationFlags(cmd, 1000, 1, "desc")
	return cmd
}

func newTradeClusterBombsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster-bombs [tickers...]",
		Short:   "Query trade cluster bombs",
		Example: "volumeleaders-agent trade cluster-bombs TSLA --days 3",
		RunE:    runTradeClusterBombs,
	}
	common.AddDateRangeFlags(cmd)
	common.AddVolumeRangeFlags(cmd)
	common.AddDollarRangeFlags(cmd, 0)
	common.AddTickersFlag(cmd)
	cmd.Flags().Int("vcd", 0, "VCD filter")
	cmd.Flags().Int("security-type", 0, "Security type key")
	cmd.Flags().Int("relative-size", 0, "Relative size threshold")
	cmd.Flags().Int("trade-cluster-bomb-rank", -1, "Trade cluster bomb rank filter")
	cmd.Flags().String("sector", "", "Sector/Industry filter")
	common.AddOutputFormatFlags(cmd)
	common.AddPaginationFlags(cmd, 100, 1, "desc")
	return cmd
}

func newTradeAlertsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "alerts",
		Short:   "Query trade alerts for a date",
		Example: "volumeleaders-agent trade alerts --date 2025-01-15",
		RunE:    runTradeAlerts,
	}
	cmd.Flags().String("date", "", "Date YYYY-MM-DD")
	_ = cmd.MarkFlagRequired("date")
	common.AddOutputFormatFlags(cmd)
	common.AddPaginationFlags(cmd, 100, 1, "desc")
	return cmd
}

func newTradeClusterAlertsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster-alerts",
		Short:   "Query trade cluster alerts for a date",
		Example: "volumeleaders-agent trade cluster-alerts --date 2025-01-15",
		RunE:    runTradeClusterAlerts,
	}
	cmd.Flags().String("date", "", "Date YYYY-MM-DD")
	_ = cmd.MarkFlagRequired("date")
	common.AddOutputFormatFlags(cmd)
	common.AddPaginationFlags(cmd, 100, 1, "desc")
	return cmd
}

func newTradeLevelsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "levels [ticker]",
		Short:   "Query significant price levels for a ticker",
		Example: "volumeleaders-agent trade levels AAPL",
		RunE:    runTradeLevels,
	}
	common.AddOptionalDateRangeFlags(cmd)
	addTradeRangeFlags(cmd, 500000)
	common.AddTickerFlag(cmd)
	cmd.Flags().Int("vcd", 0, "VCD filter")
	cmd.Flags().Int("relative-size", 0, "Relative size threshold")
	cmd.Flags().Int("trade-level-rank", -1, "Trade level rank filter")
	cmd.Flags().Int("trade-level-count", 10, "Number of price levels to return (1-50)")
	cmd.Flags().String("fields", "", "Comma-separated TradeLevel fields to include in output, or 'all' for every field")
	common.AddOutputFormatFlags(cmd)
	return cmd
}

func newTradeLevelTouchesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "level-touches [ticker]",
		Short:   "Query trade events at notable price levels",
		Example: "volumeleaders-agent trade level-touches AAPL --days 14",
		RunE:    runTradeLevelTouches,
	}
	common.AddDateRangeFlags(cmd)
	addTradeRangeFlags(cmd, 500000)
	common.AddTickerFlag(cmd)
	cmd.Flags().Int("vcd", 0, "VCD filter")
	cmd.Flags().Int("relative-size", 0, "Relative size threshold")
	cmd.Flags().Int("trade-level-rank", 10, "Trade level rank filter")
	cmd.Flags().Int("trade-level-count", maxTradeLevelRequestLength, "Number of price levels to include (1-50)")
	common.AddOutputFormatFlags(cmd)
	common.AddPaginationFlags(cmd, maxTradeLevelRequestLength, 0, "desc")
	return cmd
}

func addTradeRangeFlags(cmd *cobra.Command, minDollarsDefault float64) {
	common.AddVolumeRangeFlags(cmd)
	common.AddPriceRangeFlags(cmd)
	common.AddDollarRangeFlags(cmd, minDollarsDefault)
}

func addTradeFilterFlags(cmd *cobra.Command, vcdDefault int) {
	cmd.Flags().Int("conditions", -1, "Trade conditions filter")
	cmd.Flags().Int("vcd", vcdDefault, "VCD filter")
	cmd.Flags().Int("security-type", -1, "Security type key")
	cmd.Flags().Int("relative-size", 5, "Relative size threshold")
	cmd.Flags().Int("dark-pools", -1, "Dark pool filter")
	cmd.Flags().Int("sweeps", -1, "Sweep filter")
	cmd.Flags().Int("late-prints", -1, "Late print filter")
	cmd.Flags().Int("sig-prints", -1, "Signature print filter")
	cmd.Flags().Int("even-shared", -1, "Even shared filter")
	cmd.Flags().Int("trade-rank", -1, "Trade rank filter")
	cmd.Flags().Int("rank-snapshot", -1, "Trade rank snapshot filter")
	cmd.Flags().Int("market-cap", 0, "Market cap filter")
	cmd.Flags().Int("premarket", 1, "Include premarket")
	cmd.Flags().Int("rth", 1, "Include regular trading hours")
	cmd.Flags().Int("ah", 1, "Include after hours")
	cmd.Flags().Int("opening", 1, "Include opening trades")
	cmd.Flags().Int("closing", 1, "Include closing trades")
	cmd.Flags().Int("phantom", 1, "Include phantom prints")
	cmd.Flags().Int("offsetting", 1, "Include offsetting trades")
}

func runTradeList(cmd *cobra.Command, _ []string) error {
	presetName, _ := cmd.Flags().GetString("preset")
	watchlistName, _ := cmd.Flags().GetString("watchlist")
	tickers := common.MultiTickerValue(cmd)
	fieldsValue, _ := cmd.Flags().GetString("fields")
	fields, err := common.ParseJSONFieldList[models.Trade](fieldsValue)
	if err != nil {
		return fmt.Errorf("parsing fields flag: %w", err)
	}
	formatValue, _ := cmd.Flags().GetString("format")
	format, err := common.ParseOutputFormat(formatValue)
	if err != nil {
		return err
	}
	length, _ := cmd.Flags().GetInt("length")
	if err := validateTradeRequestLength(length); err != nil {
		return err
	}

	lookbackDays := 0
	if tickers != "" {
		lookbackDays = tradeListTickerLookbackDays
	}
	startDate, endDate, err := common.OptionalDateRange(cmd, lookbackDays)
	if err != nil {
		return err
	}

	filters := buildTradeFilters(tradesOptionsFromFlags(cmd, tickers, startDate, endDate))
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

	start, _ := cmd.Flags().GetInt("start")
	orderCol, _ := cmd.Flags().GetInt("order-col")
	orderDir, _ := cmd.Flags().GetString("order-dir")
	opts := common.DataTableOptions{Start: start, Length: length, OrderCol: orderCol, OrderDir: orderDir, Filters: filters, Fields: fields}
	summary, _ := cmd.Flags().GetBool("summary")
	if !summary && cmd.Flags().Changed("group-by") {
		return fmt.Errorf("--group-by only works with --summary")
	}
	if summary {
		if len(fields) > 0 {
			return fmt.Errorf("--fields cannot be used with --summary")
		}
		if format != common.OutputFormatJSON {
			return fmt.Errorf("--format cannot be used with --summary")
		}
		groupBy, _ := cmd.Flags().GetString("group-by")
		return runTradeSummary(cmd, opts, groupBy, startDate, endDate)
	}
	if format == common.OutputFormatJSON && len(fields) == 0 {
		return runTradeListRows(cmd, opts)
	}
	return common.RunDataTablesCommand[models.Trade](cmd, "/Trades/GetTrades", datatables.TradeColumns, opts, formatValue, "query trades")
}

func tradesOptionsFromFlags(cmd *cobra.Command, tickers, startDate, endDate string) *tradesOptions {
	return &tradesOptions{
		tickers: tickers, startDate: startDate, endDate: endDate,
		minVolume: getInt(cmd, "min-volume"), maxVolume: getInt(cmd, "max-volume"),
		minPrice: getFloat(cmd, "min-price"), maxPrice: getFloat(cmd, "max-price"),
		minDollars: getFloat(cmd, "min-dollars"), maxDollars: getFloat(cmd, "max-dollars"),
		conditions: getInt(cmd, "conditions"), vcd: getInt(cmd, "vcd"), securityType: getInt(cmd, "security-type"), relativeSize: getInt(cmd, "relative-size"),
		darkPools: getInt(cmd, "dark-pools"), sweeps: getInt(cmd, "sweeps"), latePrints: getInt(cmd, "late-prints"), sigPrints: getInt(cmd, "sig-prints"),
		evenShared: getInt(cmd, "even-shared"), tradeRank: getInt(cmd, "trade-rank"), rankSnapshot: getInt(cmd, "rank-snapshot"), marketCap: getInt(cmd, "market-cap"),
		premarket: getInt(cmd, "premarket"), rth: getInt(cmd, "rth"), ah: getInt(cmd, "ah"), opening: getInt(cmd, "opening"), closing: getInt(cmd, "closing"),
		phantom: getInt(cmd, "phantom"), offsetting: getInt(cmd, "offsetting"), sector: getString(cmd, "sector"),
	}
}

func validateTradeRequestLength(length int) error {
	if length < 1 || length > maxTradeRequestLength {
		return fmt.Errorf("--length must be between 1 and %d for trade retrieval", maxTradeRequestLength)
	}
	return nil
}

func validateTradeLevelCount(count int) error {
	if count < 1 || count > maxTradeLevelRequestLength {
		return fmt.Errorf("--trade-level-count must be between 1 and %d for trade level retrieval", maxTradeLevelRequestLength)
	}
	return nil
}

func validateTradeLevelTouchesLength(length int) error {
	if length < 1 || length > maxTradeLevelRequestLength {
		return fmt.Errorf("--length must be between 1 and %d for trade level touch retrieval", maxTradeLevelRequestLength)
	}
	return nil
}

func runTradeListRows(cmd *cobra.Command, opts common.DataTableOptions) error {
	ctx := cmd.Context()
	vlClient, err := common.NewCommandClient(ctx)
	if err != nil {
		return err
	}
	request := common.NewDataTablesRequest(datatables.TradeColumns, opts)
	var result []models.Trade
	if err := vlClient.PostDataTables(ctx, "/Trades/GetTrades", request.Encode(), &result); err != nil {
		slog.Error("failed to query trades", "error", err)
		return fmt.Errorf("query trades: %w", err)
	}
	return common.PrintJSON(cmd.OutOrStdout(), ctx, models.NewTradeListRows(result))
}

func runTradeSummary(cmd *cobra.Command, opts common.DataTableOptions, groupBy, startDate, endDate string) error {
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

func parseTradeSummaryGroup(value string) (tradeSummaryGroup, error) {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), " ", ""))
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

func runTradeSentiment(cmd *cobra.Command, _ []string) error {
	formatValue, _ := cmd.Flags().GetString("format")
	format, err := common.ParseOutputFormat(formatValue)
	if err != nil {
		return err
	}
	startDate, endDate, err := common.RequiredDateRange(cmd)
	if err != nil {
		return err
	}
	opts := tradesOptionsFromFlags(cmd, "", startDate, endDate)
	opts.sector = "X B"
	filters := buildTradeFilters(opts)
	vlClient, err := common.NewCommandClient(cmd.Context())
	if err != nil {
		return err
	}
	trades, err := fetchTradeSentimentTrades(cmd, vlClient, common.DataTableOptions{Start: 0, Length: maxTradeRequestLength, OrderCol: 1, OrderDir: "desc", Filters: filters})
	if err != nil {
		return err
	}
	sentiment := summarizeTradeSentiment(trades, startDate, endDate)
	if format == common.OutputFormatJSON {
		return common.PrintJSON(cmd.OutOrStdout(), cmd.Context(), sentiment)
	}
	return common.PrintDataTablesResult(cmd.OutOrStdout(), cmd.Context(), flattenTradeSentiment(&sentiment), nil, format)
}

func flattenTradeSentiment(summary *models.TradeSentiment) []models.TradeSentimentRow {
	rows := make([]models.TradeSentimentRow, 0, len(summary.Daily)+1)
	for i := range summary.Daily {
		day := &summary.Daily[i]
		rows = append(rows, tradeSentimentDayRow(day))
	}
	rows = append(rows, tradeSentimentTotalsRow(&summary.Totals))
	return rows
}

func tradeSentimentDayRow(day *models.TradeSentimentDay) models.TradeSentimentRow {
	return models.TradeSentimentRow{Date: day.Date, BearTrades: day.Bear.Trades, BearDollars: day.Bear.Dollars, BearTopTickers: strings.Join(day.Bear.TopTickers, ";"), BullTrades: day.Bull.Trades, BullDollars: day.Bull.Dollars, BullTopTickers: strings.Join(day.Bull.TopTickers, ";"), Ratio: day.Ratio, Signal: day.Signal}
}

func tradeSentimentTotalsRow(totals *models.TradeSentimentTotals) models.TradeSentimentRow {
	return models.TradeSentimentRow{Date: "total", BearTrades: totals.Bear.Trades, BearDollars: totals.Bear.Dollars, BearTopTickers: strings.Join(totals.Bear.TopTickers, ";"), BullTrades: totals.Bull.Trades, BullDollars: totals.Bull.Dollars, BullTopTickers: strings.Join(totals.Bull.TopTickers, ";"), Ratio: totals.Ratio, Signal: totals.Signal}
}

func runTradeClusters(cmd *cobra.Command, _ []string) error {
	startDate, endDate, err := common.RequiredDateRange(cmd)
	if err != nil {
		return err
	}
	fields, err := common.OutputFields[models.TradeCluster](getString(cmd, "fields"), tradeClusterDefaultFields)
	if err != nil {
		return fmt.Errorf("parsing fields flag: %w", err)
	}
	return common.RunDataTablesCommand[models.TradeCluster](cmd, "/TradeClusters/GetTradeClusters", datatables.TradeClusterColumns, common.DataTableOptions{Start: getInt(cmd, "start"), Length: getInt(cmd, "length"), OrderCol: getInt(cmd, "order-col"), OrderDir: getString(cmd, "order-dir"), Fields: fields, Filters: map[string]string{"Tickers": common.MultiTickerValue(cmd), "StartDate": startDate, "EndDate": endDate, "MinVolume": common.IntStr(getInt(cmd, "min-volume")), "MaxVolume": common.IntStr(getInt(cmd, "max-volume")), "MinPrice": common.FormatFloat(getFloat(cmd, "min-price")), "MaxPrice": common.FormatFloat(getFloat(cmd, "max-price")), "MinDollars": common.FormatFloat(getFloat(cmd, "min-dollars")), "MaxDollars": common.FormatFloat(getFloat(cmd, "max-dollars")), "VCD": common.IntStr(getInt(cmd, "vcd")), "SecurityTypeKey": common.IntStr(getInt(cmd, "security-type")), "RelativeSize": common.IntStr(getInt(cmd, "relative-size")), "TradeClusterRank": common.IntStr(getInt(cmd, "trade-cluster-rank")), "SectorIndustry": getString(cmd, "sector")}}, getString(cmd, "format"), "query trade clusters")
}

func runTradeClusterBombs(cmd *cobra.Command, _ []string) error {
	startDate, endDate, err := common.RequiredDateRange(cmd)
	if err != nil {
		return err
	}
	return common.RunDataTablesCommand[models.TradeClusterBomb](cmd, "/TradeClusterBombs/GetTradeClusterBombs", datatables.TradeClusterBombColumns, common.DataTableOptions{Start: getInt(cmd, "start"), Length: getInt(cmd, "length"), OrderCol: getInt(cmd, "order-col"), OrderDir: getString(cmd, "order-dir"), Filters: map[string]string{"Tickers": common.MultiTickerValue(cmd), "StartDate": startDate, "EndDate": endDate, "MinVolume": common.IntStr(getInt(cmd, "min-volume")), "MaxVolume": common.IntStr(getInt(cmd, "max-volume")), "MinDollars": common.FormatFloat(getFloat(cmd, "min-dollars")), "MaxDollars": common.FormatFloat(getFloat(cmd, "max-dollars")), "VCD": common.IntStr(getInt(cmd, "vcd")), "SecurityTypeKey": common.IntStr(getInt(cmd, "security-type")), "RelativeSize": common.IntStr(getInt(cmd, "relative-size")), "TradeClusterBombRank": common.IntStr(getInt(cmd, "trade-cluster-bomb-rank")), "SectorIndustry": getString(cmd, "sector")}}, getString(cmd, "format"), "query trade cluster bombs")
}

func runTradeAlerts(cmd *cobra.Command, _ []string) error {
	return common.RunDataTablesCommand[models.TradeAlert](cmd, "/TradeAlerts/GetTradeAlerts", datatables.TradeColumns, common.DataTableOptions{Start: getInt(cmd, "start"), Length: getInt(cmd, "length"), OrderCol: getInt(cmd, "order-col"), OrderDir: getString(cmd, "order-dir"), Filters: map[string]string{"Date": getString(cmd, "date")}}, getString(cmd, "format"), "query trade alerts")
}

func runTradeClusterAlerts(cmd *cobra.Command, _ []string) error {
	return common.RunDataTablesCommand[models.TradeClusterAlert](cmd, "/TradeClusterAlerts/GetTradeClusterAlerts", datatables.TradeClusterColumns, common.DataTableOptions{Start: getInt(cmd, "start"), Length: getInt(cmd, "length"), OrderCol: getInt(cmd, "order-col"), OrderDir: getString(cmd, "order-dir"), Filters: map[string]string{"Date": getString(cmd, "date")}}, getString(cmd, "format"), "query trade cluster alerts")
}

func runTradeLevels(cmd *cobra.Command, _ []string) error {
	formatValue := getString(cmd, "format")
	format, err := common.ParseOutputFormat(formatValue)
	if err != nil {
		return err
	}
	startDate, endDate, err := common.OptionalDateRange(cmd, 365)
	if err != nil {
		return err
	}
	fields, err := common.OutputFields[models.TradeLevel](getString(cmd, "fields"), nil)
	if err != nil {
		return fmt.Errorf("parsing fields flag: %w", err)
	}
	ticker, err := common.SingleTickerValue(cmd)
	if err != nil {
		return err
	}
	opts := tradeLevelOptionsFromFlags(cmd, ticker, startDate, endDate)
	if err := validateTradeLevelCount(opts.tradeLevelCount); err != nil {
		return err
	}
	dataOpts := common.DataTableOptions{Start: 0, Length: opts.tradeLevelCount, OrderCol: 1, OrderDir: "desc", Filters: buildTradeLevelFilters(opts), Fields: fields}
	if format == common.OutputFormatJSON && len(fields) == 0 {
		levels, err := fetchTradeLevels(cmd, dataOpts)
		if err != nil {
			return err
		}
		return common.PrintJSON(cmd.OutOrStdout(), cmd.Context(), models.NewTradeLevelRows(levels))
	}
	return common.RunDataTablesSingleRequestCommand[models.TradeLevel](cmd, "/TradeLevels/GetTradeLevels", datatables.TradeLevelColumns, dataOpts, formatValue, "query trade levels")
}

func tradeLevelOptionsFromFlags(cmd *cobra.Command, ticker, startDate, endDate string) *tradeLevelOptions {
	return &tradeLevelOptions{ticker: ticker, startDate: startDate, endDate: endDate, minVolume: getInt(cmd, "min-volume"), maxVolume: getInt(cmd, "max-volume"), minPrice: getFloat(cmd, "min-price"), maxPrice: getFloat(cmd, "max-price"), minDollars: getFloat(cmd, "min-dollars"), maxDollars: getFloat(cmd, "max-dollars"), vcd: getInt(cmd, "vcd"), relativeSize: getInt(cmd, "relative-size"), tradeLevelRank: getInt(cmd, "trade-level-rank"), tradeLevelCount: getInt(cmd, "trade-level-count")}
}

func fetchTradeLevels(cmd *cobra.Command, opts common.DataTableOptions) ([]models.TradeLevel, error) {
	ctx := cmd.Context()
	vlClient, err := common.NewCommandClient(ctx)
	if err != nil {
		return nil, err
	}
	request := common.NewDataTablesRequest(datatables.TradeLevelColumns, opts)
	var result []models.TradeLevel
	if err := vlClient.PostDataTables(ctx, "/TradeLevels/GetTradeLevels", request.Encode(), &result); err != nil {
		slog.Error("failed to query trade levels", "error", err)
		return nil, fmt.Errorf("query trade levels: %w", err)
	}
	return result, nil
}

func runTradeLevelTouches(cmd *cobra.Command, _ []string) error {
	startDate, endDate, err := common.RequiredDateRange(cmd)
	if err != nil {
		return err
	}
	if err := validateTradeLevelTouchesLength(getInt(cmd, "length")); err != nil {
		return err
	}
	if err := validateTradeLevelCount(getInt(cmd, "trade-level-count")); err != nil {
		return err
	}
	ticker, err := common.SingleTickerValue(cmd)
	if err != nil {
		return err
	}
	return common.RunDataTablesCommand[models.TradeLevelTouch](cmd, "/TradeLevelTouches/GetTradeLevelTouches", datatables.TradeLevelTouchColumns, common.DataTableOptions{Start: getInt(cmd, "start"), Length: getInt(cmd, "length"), OrderCol: getInt(cmd, "order-col"), OrderDir: getString(cmd, "order-dir"), Filters: map[string]string{"Tickers": ticker, "StartDate": startDate, "EndDate": endDate, "MinVolume": common.IntStr(getInt(cmd, "min-volume")), "MaxVolume": common.IntStr(getInt(cmd, "max-volume")), "MinPrice": common.FormatFloat(getFloat(cmd, "min-price")), "MaxPrice": common.FormatFloat(getFloat(cmd, "max-price")), "MinDollars": common.FormatFloat(getFloat(cmd, "min-dollars")), "MaxDollars": common.FormatFloat(getFloat(cmd, "max-dollars")), "VCD": common.IntStr(getInt(cmd, "vcd")), "RelativeSize": common.IntStr(getInt(cmd, "relative-size")), "TradeLevelRank": common.IntStr(getInt(cmd, "trade-level-rank")), "Levels": common.IntStr(getInt(cmd, "trade-level-count"))}}, getString(cmd, "format"), "query trade level touches")
}

func fetchTradeSentimentTrades(cmd *cobra.Command, vlClient *client.Client, opts common.DataTableOptions) ([]models.Trade, error) {
	request := common.NewDataTablesRequest(datatables.TradeColumns, opts)
	var result []models.Trade
	if err := vlClient.PostDataTables(cmd.Context(), "/Trades/GetTrades", request.Encode(), &result); err != nil {
		slog.Error("failed to query trade sentiment", "error", err)
		return nil, fmt.Errorf("query trade sentiment: %w", err)
	}
	return result, nil
}

type tradeSentimentAccumulator struct {
	trades        int
	dollars       float64
	tickerDollars map[string]float64
}

type tradeSentimentDayAccumulator struct{ bear, bull tradeSentimentAccumulator }

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
	return models.TradeSentiment{DateRange: models.TradeSentimentDateRange{Start: startDate, End: endDate}, Daily: daily, Totals: totals.summaryTotals()}
}

func (a *tradeSentimentDayAccumulator) add(side string, trade *models.Trade) {
	if side == "bear" {
		a.bear.add(trade)
	} else if side == "bull" {
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
	return models.TradeSentimentDay{Date: day, Bear: a.bear.summary(), Bull: a.bull.summary(), Ratio: ratio, Signal: tradeSentimentSignal(ratio, a.bull.dollars, a.bear.dollars)}
}

func (a tradeSentimentDayAccumulator) summaryTotals() models.TradeSentimentTotals {
	ratio := tradeSentimentRatio(a.bull.dollars, a.bear.dollars)
	return models.TradeSentimentTotals{Bear: a.bear.summary(), Bull: a.bull.summary(), Ratio: ratio, Signal: tradeSentimentSignal(ratio, a.bull.dollars, a.bear.dollars)}
}

func (a tradeSentimentAccumulator) summary() models.TradeSentimentSide {
	return models.TradeSentimentSide{Trades: a.trades, Dollars: a.dollars, TopTickers: topTradeSentimentTickers(a.tickerDollars, 3)}
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

func buildTradeFilters(opts *tradesOptions) map[string]string {
	return map[string]string{"Tickers": opts.tickers, "StartDate": opts.startDate, "EndDate": opts.endDate, "MinVolume": common.IntStr(opts.minVolume), "MaxVolume": common.IntStr(opts.maxVolume), "MinPrice": common.FormatFloat(opts.minPrice), "MaxPrice": common.FormatFloat(opts.maxPrice), "MinDollars": common.FormatFloat(opts.minDollars), "MaxDollars": common.FormatFloat(opts.maxDollars), "Conditions": common.IntStr(opts.conditions), "VCD": common.IntStr(opts.vcd), "SecurityTypeKey": common.IntStr(opts.securityType), "RelativeSize": common.IntStr(opts.relativeSize), "DarkPools": common.IntStr(opts.darkPools), "Sweeps": common.IntStr(opts.sweeps), "LatePrints": common.IntStr(opts.latePrints), "SignaturePrints": common.IntStr(opts.sigPrints), "EvenShared": common.IntStr(opts.evenShared), "TradeRank": common.IntStr(opts.tradeRank), "TradeRankSnapshot": common.IntStr(opts.rankSnapshot), "MarketCap": common.IntStr(opts.marketCap), "IncludePremarket": common.IntStr(opts.premarket), "IncludeRTH": common.IntStr(opts.rth), "IncludeAH": common.IntStr(opts.ah), "IncludeOpening": common.IntStr(opts.opening), "IncludeClosing": common.IntStr(opts.closing), "IncludePhantom": common.IntStr(opts.phantom), "IncludeOffsetting": common.IntStr(opts.offsetting), "SectorIndustry": opts.sector}
}

func buildTradeLevelFilters(opts *tradeLevelOptions) map[string]string {
	return map[string]string{"Ticker": opts.ticker, "MinVolume": common.IntStr(opts.minVolume), "MaxVolume": common.IntStr(opts.maxVolume), "MinPrice": common.FormatFloat(opts.minPrice), "MaxPrice": common.FormatFloat(opts.maxPrice), "MinDollars": common.FormatFloat(opts.minDollars), "MaxDollars": common.FormatFloat(opts.maxDollars), "VCD": common.IntStr(opts.vcd), "RelativeSize": common.IntStr(opts.relativeSize), "StartDate": opts.startDate, "EndDate": opts.endDate, "TradeLevelRank": common.IntStr(opts.tradeLevelRank), "Levels": common.IntStr(opts.tradeLevelCount)}
}

func getString(cmd *cobra.Command, name string) string {
	value, _ := cmd.Flags().GetString(name)
	return value
}
func getInt(cmd *cobra.Command, name string) int { value, _ := cmd.Flags().GetInt(name); return value }
func getFloat(cmd *cobra.Command, name string) float64 {
	value, _ := cmd.Flags().GetFloat64(name)
	return value
}
