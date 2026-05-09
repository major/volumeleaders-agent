package report

import (
	"context"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"

	vlgo "github.com/major/volumeleaders-go/volumeleaders"
	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/models"
)

const reportBrowserPageLength = 100

type reportDefinition struct {
	use     string
	name    string
	short   string
	long    string
	example string
	filters map[string]string
}

type reportOptions struct {
	Tickers      string              `flag:"tickers" flaggroup:"Input" flagshort:"t" flagdescr:"Comma-separated ticker symbols; use this for multi-day report lookbacks"`
	StartDate    string              `flag:"start-date" flaggroup:"Dates" flagshort:"s" flagdescr:"Start date YYYY-MM-DD (default: today)"`
	EndDate      string              `flag:"end-date" flaggroup:"Dates" flagshort:"e" flagdescr:"End date YYYY-MM-DD (default: today)"`
	Days         int                 `flag:"days" flaggroup:"Dates" flagshort:"d" flagdescr:"Look back this many days from --end-date or today; broad scans require a single day"`
	Fields       string              `flag:"fields" flaggroup:"Output" flagdescr:"Comma-separated raw Trade fields to include, or omit for compact JSON"`
	Summary      bool                `flag:"summary" flaggroup:"Output" flagdescr:"Return aggregate metrics instead of individual trades"`
	GroupBy      reportSummaryGroup  `flag:"group-by" flaggroup:"Output" flagdescr:"Summary grouping (requires --summary): ticker, day, or ticker,day"`
	LimitGroups  int                 `flag:"limit-groups" flaggroup:"Output" flagdescr:"Limit ordered summary groups; 0 omits the ordered groups slice"`
	SortGroupsBy reportSummarySort   `flag:"sort-groups-by" flaggroup:"Output" flagdescr:"Sort ordered summary groups by: dollars, trades, avg-multiplier, or min-rank"`
	Format       common.OutputFormat `flag:"format" flaggroup:"Output" flagshort:"f" flagdescr:"Output format: json, csv, or tsv"`
}

type reportListOptions struct {
	Format common.OutputFormat `flag:"format" flagdescr:"Output format: json, csv, or tsv"`
}

// ReportInfo describes one curated report command for report list output.
type ReportInfo struct {
	Name        string            `json:"name"`
	Command     string            `json:"command"`
	Description string            `json:"description"`
	Preset      string            `json:"preset"`
	Filters     map[string]string `json:"filters"`
}

type reportSummaryGroup string

type reportSummarySort string

const (
	reportSummaryGroupTicker    reportSummaryGroup = "ticker"
	reportSummaryGroupDay       reportSummaryGroup = "day"
	reportSummaryGroupTickerDay reportSummaryGroup = "ticker,day"

	reportSummarySortDollars       reportSummarySort = "dollars"
	reportSummarySortTrades        reportSummarySort = "trades"
	reportSummarySortAvgMultiplier reportSummarySort = "avg-multiplier"
	reportSummarySortMinRank       reportSummarySort = "min-rank"
)

// NewCmd returns the "report" command group with curated report presets.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "report",
		Short:   "Run curated VolumeLeaders trade reports",
		GroupID: "trading",
		Args:    cobra.NoArgs,
		Long: `Run curated VolumeLeaders trade reports using vetted browser presets. Start here before raw trade list filters: report commands keep the query shape small, documented, and close to the VolumeLeaders site so humans and LLMs do not need to reason about low-level filter parameters.

Reports default to today-only scans and fetch results in the same browser-sized 100-row pages observed from VolumeLeaders. Broad multi-day scans without tickers are rejected to avoid expensive requests and backend timeouts. Add --tickers for multi-day lookbacks, or use trade list --preset only when you need advanced filters that are not exposed by reports.

PREREQUISITES: Browser authentication must be available.

RECOVERY: If a broad multi-day scan is rejected, rerun for one day or add --tickers. If you need a custom filter combination, inspect report list first, then move to trade list --preset rather than hand-building filters.`,
	}
	cmd.AddCommand(newReportListCommand())
	definitions := reportDefinitions()
	for i := range definitions {
		cmd.AddCommand(newReportCommand(&definitions[i]))
	}
	return cmd
}

// NewReportCommand returns the "report" command group with all report subcommands.
func NewReportCommand() *cobra.Command { return NewCmd() }

func newReportListCommand() *cobra.Command {
	opts := &reportListOptions{Format: common.OutputFormatJSON}
	cmd := &cobra.Command{
		Use:        "list",
		Short:      "List curated report presets",
		Long:       "List curated report commands, their source VolumeLeaders preset names, and their fixed filter configurations. Use these reports before raw trade list filters because they avoid expensive, timeout-prone filter combinations and expose only the safe override surface.",
		Example:    "volumeleaders-agent report list",
		Args:       cobra.NoArgs,
		SuggestFor: []string{"ls", "presets", "reports"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runReportList(cmd, opts)
		},
	}
	common.BindOrPanic(cmd, opts, "report list")
	return cmd
}

func newReportCommand(definition *reportDefinition) *cobra.Command {
	opts := &reportOptions{Format: common.OutputFormatJSON, GroupBy: reportSummaryGroupTicker, SortGroupsBy: reportSummarySortDollars}
	cmd := &cobra.Command{
		Use:     definition.use + " [tickers...]",
		Short:   definition.short,
		Long:    definition.long,
		Example: definition.example,
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runReport(cmd, opts, definition)
		},
	}
	common.BindOrPanic(cmd, opts, definition.use)
	return cmd
}

func runReportList(cmd *cobra.Command, opts *reportListOptions) error {
	format, err := common.ParseOutputFormat(opts.Format)
	if err != nil {
		return err
	}
	reports := reportDefinitions()
	items := make([]ReportInfo, 0, len(reports))
	for _, report := range reports {
		items = append(items, ReportInfo{Name: report.name, Command: "report " + report.use, Description: report.short, Preset: report.name, Filters: maps.Clone(report.filters)})
	}
	return common.PrintDataTablesResult(cmd.OutOrStdout(), cmd.Context(), items, nil, format)
}

func runReport(cmd *cobra.Command, opts *reportOptions, definition *reportDefinition) error {
	fields, err := common.ParseJSONFieldList[models.Trade](opts.Fields)
	if err != nil {
		return fmt.Errorf("parsing fields flag: %w", err)
	}
	format, err := common.ParseOutputFormat(opts.Format)
	if err != nil {
		return err
	}
	tickers := common.MultiTickerValue(cmd)
	startDate, endDate, err := common.ResolveDateRange(cmd, 0, false)
	if err != nil {
		return err
	}
	if tickers == "" && startDate != endDate {
		return fmt.Errorf("broad report scans must use a single day to avoid VolumeLeaders timeouts; add --tickers for multi-day lookbacks or rerun with --start-date equal to --end-date")
	}
	if !opts.Summary && cmd.Flags().Changed("group-by") {
		return fmt.Errorf("--group-by only works with --summary")
	}
	if !opts.Summary && cmd.Flags().Changed("limit-groups") {
		return fmt.Errorf("--limit-groups only works with --summary")
	}
	if !opts.Summary && cmd.Flags().Changed("sort-groups-by") {
		return fmt.Errorf("--sort-groups-by only works with --summary")
	}
	if opts.Summary {
		if len(fields) > 0 {
			return fmt.Errorf("--fields cannot be used with --summary")
		}
		if format != common.OutputFormatJSON {
			return fmt.Errorf("--format cannot be used with --summary")
		}
		if opts.LimitGroups < 0 {
			return fmt.Errorf("--limit-groups must be 0 or greater")
		}
	}
	group := reportSummaryGroup("")
	sortBy := reportSummarySort("")
	if opts.Summary {
		group, err = parseReportSummaryGroup(opts.GroupBy)
		if err != nil {
			return err
		}
		sortBy, err = parseReportSummarySort(opts.SortGroupsBy)
		if err != nil {
			return err
		}
	}

	filters := maps.Clone(definition.filters)
	filters["Tickers"] = tickers
	filters["StartDate"] = startDate
	filters["EndDate"] = endDate
	dtOpts := common.NewDataTableOptions(common.DataTableRequestConfig{Start: 0, Length: -1, OrderCol: 0, OrderDir: common.OrderDirectionDESC, Filters: filters, Fields: fields})
	trades, err := fetchReportTrades(cmd, dtOpts)
	if err != nil {
		return err
	}
	if opts.Summary {
		return common.PrintJSON(cmd.OutOrStdout(), cmd.Context(), summarizeReportTrades(trades, group, startDate, endDate, opts.LimitGroups, sortBy))
	}
	if format == common.OutputFormatJSON && len(fields) == 0 {
		return common.PrintJSON(cmd.OutOrStdout(), cmd.Context(), models.NewTradeListRows(trades))
	}
	return common.PrintDataTablesResult(cmd.OutOrStdout(), cmd.Context(), trades, fields, format)
}

func fetchReportTrades(cmd *cobra.Command, opts common.DataTableOptions) ([]models.Trade, error) {
	ctx := cmd.Context()
	vlClient, err := common.NewVLClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create VL client: %w", err)
	}
	vlTrades, err := common.FetchVLPages[vlgo.Trade](ctx, vlClient, opts, reportBrowserPageLength, "query report trades", fetchTrades)
	if err != nil {
		return nil, fmt.Errorf("fetch report trades: %w", err)
	}
	return common.MapSlice(vlTrades, common.MapVLTrade), nil
}

// fetchTrades wraps vlgo.Client.GetTrades as a VLFetcher for the report package.
func fetchTrades(ctx context.Context, c *vlgo.Client, dt *vlgo.DataTablesRequest, filters url.Values) (*vlgo.DataTablesResponse[vlgo.Trade], error) {
	return c.GetTrades(ctx, vlgo.TradesRequest{DataTables: *dt, Filters: filters})
}

func parseReportSummaryGroup(value reportSummaryGroup) (reportSummaryGroup, error) {
	normalized := strings.ToLower(strings.TrimSpace(string(value)))
	normalized = strings.NewReplacer(" ", "", "-", ",").Replace(normalized)
	if normalized == "tickerday" {
		normalized = string(reportSummaryGroupTickerDay)
	}
	switch reportSummaryGroup(normalized) {
	case reportSummaryGroupTicker, reportSummaryGroupDay, reportSummaryGroupTickerDay:
		return reportSummaryGroup(normalized), nil
	default:
		return "", fmt.Errorf("invalid group-by %q; valid values: ticker, day, ticker,day", value)
	}
}

func parseReportSummarySort(value reportSummarySort) (reportSummarySort, error) {
	normalized := strings.ToLower(strings.TrimSpace(string(value)))
	normalized = strings.NewReplacer("_", "-", " ", "-").Replace(normalized)
	switch reportSummarySort(normalized) {
	case reportSummarySortDollars, reportSummarySortTrades, reportSummarySortAvgMultiplier, reportSummarySortMinRank:
		return reportSummarySort(normalized), nil
	default:
		return "", fmt.Errorf("invalid sort-groups-by %q; valid values: dollars, trades, avg-multiplier, min-rank", value)
	}
}

func summarizeReportTrades(trades []models.Trade, group reportSummaryGroup, startDate, endDate string, limitGroups int, sortBy reportSummarySort) models.TradeSummary {
	summary := models.TradeSummary{DateRange: models.TradeSummaryDateRange{Start: startDate, End: endDate}}
	groups := make(map[string]*reportGroupAccumulator)
	keyFunc := reportSummaryKeyFunc(group)
	for i := range trades {
		trade := &trades[i]
		summary.TotalTrades++
		summary.TotalDollars += trade.Dollars
		addReportSummaryGroup(groups, keyFunc(trade), trade)
	}
	groupSummaries := summarizeReportGroups(groups)
	switch group {
	case reportSummaryGroupDay:
		summary.ByDay = groupSummaries
	case reportSummaryGroupTickerDay:
		summary.ByTickerDay = groupSummaries
	default:
		summary.ByTicker = groupSummaries
	}
	if limitGroups > 0 {
		summary.Groups = orderedReportGroups(groupSummaries, limitGroups, sortBy)
	}
	return summary
}

type reportGroupAccumulator struct {
	trades                 int
	dollars                float64
	dollarsMultiplier      float64
	darkPool               int
	sweep                  int
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

func summarizeReportGroups(groups map[string]*reportGroupAccumulator) map[string]models.TradeGroupSummary {
	summaries := make(map[string]models.TradeGroupSummary, len(groups))
	for key, acc := range groups {
		summaries[key] = acc.summary()
	}
	return summaries
}

func orderedReportGroups(groups map[string]models.TradeGroupSummary, limit int, sortBy reportSummarySort) []models.TradeSummaryGroup {
	ordered := make([]models.TradeSummaryGroup, 0, len(groups))
	for key := range groups {
		summary := groups[key]
		ordered = append(ordered, models.NewTradeSummaryGroup(key, &summary))
	}
	slices.SortFunc(ordered, func(a, b models.TradeSummaryGroup) int {
		if cmp := compareReportSummaryGroup(&a, &b, sortBy); cmp != 0 {
			return cmp
		}
		return strings.Compare(a.Key, b.Key)
	})
	if limit < len(ordered) {
		return ordered[:limit]
	}
	return ordered
}

func compareReportSummaryGroup(a, b *models.TradeSummaryGroup, sortBy reportSummarySort) int {
	switch sortBy {
	case reportSummarySortTrades:
		return compareIntDesc(a.Trades, b.Trades)
	case reportSummarySortAvgMultiplier:
		return compareFloatDesc(a.AvgDollarsMultiplier, b.AvgDollarsMultiplier)
	case reportSummarySortMinRank:
		return compareRankAsc(a.MinTradeRank, b.MinTradeRank)
	default:
		return compareFloatDesc(a.Dollars, b.Dollars)
	}
}

func reportSummaryKeyFunc(group reportSummaryGroup) func(*models.Trade) string {
	switch group {
	case reportSummaryGroupDay:
		return reportDayKey
	case reportSummaryGroupTickerDay:
		return reportTickerDayKey
	default:
		return reportTickerKey
	}
}

func addReportSummaryGroup(groups map[string]*reportGroupAccumulator, key string, trade *models.Trade) {
	acc := groups[key]
	if acc == nil {
		acc = &reportGroupAccumulator{}
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

func (acc *reportGroupAccumulator) summary() models.TradeGroupSummary {
	if acc.trades == 0 {
		return models.TradeGroupSummary{}
	}
	count := float64(acc.trades)
	return models.TradeGroupSummary{
		Trades:                    acc.trades,
		Dollars:                   acc.dollars,
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

func compareFloatDesc(a, b float64) int {
	if a > b {
		return -1
	}
	if a < b {
		return 1
	}
	return 0
}

func compareIntDesc(a, b int) int {
	if a > b {
		return -1
	}
	if a < b {
		return 1
	}
	return 0
}

func compareRankAsc(a, b int) int {
	if a == 0 && b == 0 {
		return 0
	}
	if a == 0 {
		return 1
	}
	if b == 0 {
		return -1
	}
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func reportTickerKey(trade *models.Trade) string {
	if trade.Ticker == "" {
		return "unknown"
	}
	return trade.Ticker
}

func reportDayKey(trade *models.Trade) string {
	if !trade.Date.Valid {
		return "unknown"
	}
	return trade.Date.Format("2006-01-02")
}

func reportTickerDayKey(trade *models.Trade) string {
	return reportTickerKey(trade) + "|" + reportDayKey(trade)
}

func reportDefinitions() []reportDefinition {
	return []reportDefinition{
		{use: "top-100-rank", name: "Top-100 Rank", short: "Run the top 100 ranked trades report", long: reportLong("Top 100 Ranked Trades", "Returns the site-vetted top 100 ranked institutional trades preset. Use this before manual TradeRank filters because it preserves the browser preset shape and avoids oversized custom queries."), example: "volumeleaders-agent report top-100-rank\nvolumeleaders-agent report top-100-rank --tickers NVDA,MSFT --days 5", filters: topRankFilters("100")},
		{use: "top-10-rank", name: "Top-10 Rank", short: "Run the top 10 ranked trades report", long: reportLong("Top 10 Ranked Trades", "Returns the strongest ranked institutional prints using the site-vetted top 10 preset. Use this when the user asks for the highest-conviction trades without needing a broader top 100 scan."), example: "volumeleaders-agent report top-10-rank\nvolumeleaders-agent report top-10-rank --tickers SPY,QQQ --days 3", filters: topRankFilters("10")},
		{use: "dark-pool-sweeps", name: "Top-100 Rank; Dark Pool Sweeps", short: "Run the dark pool sweeps report", long: reportLong("Dark Pool Sweeps", "Returns the site-vetted dark pool sweep preset: top 100 ranked dark pool sweeps during premarket and regular trading hours, excluding after-hours, opening, closing, phantom, and signature prints."), example: "volumeleaders-agent report dark-pool-sweeps\nvolumeleaders-agent report dark-pool-sweeps --tickers AAPL,TSLA --days 5", filters: darkPoolSweepFilters()},
		{use: "disproportionately-large", name: "All Disproportionately Large Trades", short: "Run the disproportionately large trades report", long: reportLong("Disproportionately Large", "Returns the site-vetted 5x relative size scan. Use this when the user asks for unusually large prints, disproportionate activity, or trades that are at least five times normal block size."), example: "volumeleaders-agent report disproportionately-large\nvolumeleaders-agent report disproportionately-large --tickers XLE,XLK --days 5", filters: disproportionatelyLargeFilters()},
		{use: "leveraged-etfs", name: "Top-100 Rank; Leveraged ETFs", short: "Run the leveraged ETF ranked trades report", long: reportLong("Leveraged ETFs", "Returns the site-vetted top 100 ranked leveraged ETF preset. Use this for broad ranked activity in leveraged and inverse ETF products without hand-building sector filters."), example: "volumeleaders-agent report leveraged-etfs\nvolumeleaders-agent report leveraged-etfs --tickers TQQQ,SQQQ --days 5", filters: leveragedETFFilters()},
		{use: "rsi-overbought", name: "Top-100 Rank; RSI OB; >=5x Avg Size", short: "Run the RSI overbought 5x ranked report", long: reportLong("RSI Overbought", "Returns the site-vetted top 100 ranked RSI overbought preset with trades at least five times average size. Use this when looking for high-rank prints in overbought names."), example: "volumeleaders-agent report rsi-overbought\nvolumeleaders-agent report rsi-overbought --tickers NVDA,AMD --days 5", filters: rsiOverboughtFilters()},
		{use: "rsi-oversold", name: "Top-100 Rank; RSI OS; >=5x Avg Size", short: "Run the RSI oversold 5x ranked report", long: reportLong("RSI Oversold", "Returns the site-vetted top 100 ranked RSI oversold preset with trades at least five times average size. Use this when looking for high-rank prints in oversold names."), example: "volumeleaders-agent report rsi-oversold\nvolumeleaders-agent report rsi-oversold --tickers IWM,QQQ --days 5", filters: rsiOversoldFilters()},
		{use: "dark-pool-20x", name: "Top-100 Rank; >=20x avg size; DP Only", short: "Run the 20x dark-pool-only ranked report", long: reportLong("20x Dark Pool Only", "Returns the site-vetted top 100 ranked dark-pool-only preset for trades at least twenty times average size. Use this for unusually large dark-pool prints without adding raw dark-pool filters."), example: "volumeleaders-agent report dark-pool-20x\nvolumeleaders-agent report dark-pool-20x --tickers SPY,QQQ --days 5", filters: darkPool20xFilters()},
		{use: "top-30-rank-10x-99th", name: "Top-30 Rank; >10x avg size; 99th %", short: "Run the top 30 10x 99th percentile report", long: reportLong("Top 30 Rank, 10x Average Size, 99th Percentile", "Returns the site-vetted top 30 ranked preset for trades above ten times average size and in the 99th cumulative distribution percentile. Use this when the user asks for the strongest extreme-size prints."), example: "volumeleaders-agent report top-30-rank-10x-99th\nvolumeleaders-agent report top-30-rank-10x-99th --tickers XLK,XLF --days 5", filters: top30Rank10x99thFilters()},
		{use: "phantom-trades", name: "Phantom Trades", short: "Run the phantom trades report", long: reportLong("Phantom Trades", "Returns the site-vetted phantom trades preset, excluding normal trading sessions and offsetting trades. Use this when the user specifically asks for phantom print activity."), example: "volumeleaders-agent report phantom-trades\nvolumeleaders-agent report phantom-trades --tickers AAPL,MSFT --days 5", filters: phantomTradeFilters()},
		{use: "offsetting-trades", name: "Offsetting Trades", short: "Run the offsetting trades report", long: reportLong("Offsetting Trades", "Returns the site-vetted offsetting trades preset, excluding normal trading sessions and phantom trades. Use this when the user specifically asks for offsetting trade activity."), example: "volumeleaders-agent report offsetting-trades\nvolumeleaders-agent report offsetting-trades --tickers SPY,QQQ --days 5", filters: offsettingTradeFilters()},
	}
}

func reportLong(title, description string) string {
	return fmt.Sprintf(`Run the %s report with fixed VolumeLeaders browser-preset filters.

%s

Reports are the recommended entry point for users and LLMs. They expose only safe overrides: tickers, dates, fields, summary grouping, limited ordered summary groups, and output format. Do not hand-build low-level filters unless this curated report cannot answer the question; use trade list --preset as the advanced escape hatch.

Defaults to today only. Multi-day broad scans without tickers are rejected to avoid expensive requests and backend timeouts. Results are fetched in browser-sized 100-row pages and ordered by time descending.

RECOVERY: If the report is too broad, add --tickers or query one day at a time. If a custom filter is truly required, run report list to inspect the vetted filters, then use trade list --preset rather than assembling raw filter flags.`, title, description)
}

func topRankFilters(rank string) map[string]string {
	return map[string]string{"Conditions": "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH", "DarkPools": "-1", "EvenShared": "-1", "IncludeAH": "1", "IncludeClosing": "1", "IncludeOffsetting": "-1", "IncludeOpening": "1", "IncludePhantom": "-1", "IncludePremarket": "1", "IncludeRTH": "1", "LatePrints": "-1", "MarketCap": "0", "MaxDollars": "100000000000", "MaxPrice": "100000", "MaxVolume": "2000000000", "MinDollars": "500000", "MinPrice": "0", "MinVolume": "10000", "RelativeSize": "0", "SecurityTypeKey": "-1", "SignaturePrints": "-1", "Sweeps": "-1", "TradeCount": "3", "TradeRank": rank, "TradeRankSnapshot": "-1", "VCD": "0"}
}

func darkPoolSweepFilters() map[string]string {
	return map[string]string{"Conditions": "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH", "DarkPools": "1", "EvenShared": "-1", "IncludeAH": "0", "IncludeClosing": "0", "IncludeOffsetting": "-1", "IncludeOpening": "0", "IncludePhantom": "0", "IncludePremarket": "1", "IncludeRTH": "1", "LatePrints": "-1", "MarketCap": "0", "MaxDollars": "100000000000", "MaxPrice": "100000", "MaxVolume": "2000000000", "MinDollars": "500000", "MinPrice": "0", "MinVolume": "10000", "RelativeSize": "0", "SecurityTypeKey": "-1", "SignaturePrints": "0", "Sweeps": "1", "TradeCount": "3", "TradeRank": "100", "TradeRankSnapshot": "-1", "VCD": "0"}
}

func disproportionatelyLargeFilters() map[string]string {
	return map[string]string{"Conditions": "-1", "DarkPools": "-1", "EvenShared": "-1", "IncludeAH": "1", "IncludeClosing": "1", "IncludeOffsetting": "1", "IncludeOpening": "1", "IncludePhantom": "1", "IncludePremarket": "1", "IncludeRTH": "1", "LatePrints": "-1", "MarketCap": "0", "MaxDollars": "30000000000", "MaxPrice": "100000", "MaxVolume": "2000000000", "MinDollars": "500000", "MinPrice": "0", "MinVolume": "0", "RelativeSize": "5", "SecurityTypeKey": "-1", "SignaturePrints": "-1", "Sweeps": "-1", "TradeRank": "-1", "TradeRankSnapshot": "-1", "VCD": "0"}
}

func leveragedETFFilters() map[string]string {
	return map[string]string{"Conditions": "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH", "DarkPools": "-1", "EvenShared": "-1", "IncludeAH": "1", "IncludeClosing": "1", "IncludeOffsetting": "-1", "IncludeOpening": "1", "IncludePhantom": "-1", "IncludePremarket": "1", "IncludeRTH": "1", "LatePrints": "-1", "MarketCap": "0", "MaxDollars": "1000000000000", "MaxPrice": "100000", "MaxVolume": "2000000000", "MinDollars": "500000", "MinPrice": "0", "MinVolume": "10000", "RelativeSize": "0", "SectorIndustry": "X B", "SecurityTypeKey": "-1", "SignaturePrints": "-1", "Sweeps": "-1", "TradeCount": "3", "TradeRank": "100", "TradeRankSnapshot": "-1", "VCD": "0"}
}

func rsiOverboughtFilters() map[string]string {
	return map[string]string{"Conditions": "OBD,OBH", "DarkPools": "-1", "EvenShared": "-1", "IncludeAH": "1", "IncludeClosing": "1", "IncludeOffsetting": "-1", "IncludeOpening": "1", "IncludePhantom": "-1", "IncludePremarket": "1", "IncludeRTH": "1", "LatePrints": "-1", "MarketCap": "0", "MaxDollars": "10000000000", "MaxPrice": "100000", "MaxVolume": "2000000000", "MinDollars": "500000", "MinPrice": "0", "MinVolume": "10000", "RelativeSize": "5", "SecurityTypeKey": "-1", "SignaturePrints": "0", "Sweeps": "-1", "TradeCount": "3", "TradeRank": "100", "TradeRankSnapshot": "-1", "VCD": "0"}
}

func rsiOversoldFilters() map[string]string {
	return map[string]string{"Conditions": "OSD,OSH", "DarkPools": "-1", "EvenShared": "-1", "IncludeAH": "1", "IncludeClosing": "1", "IncludeOffsetting": "-1", "IncludeOpening": "1", "IncludePhantom": "-1", "IncludePremarket": "1", "IncludeRTH": "1", "LatePrints": "-1", "MarketCap": "0", "MaxDollars": "10000000000", "MaxPrice": "100000", "MaxVolume": "2000000000", "MinDollars": "500000", "MinPrice": "0", "MinVolume": "10000", "RelativeSize": "5", "SecurityTypeKey": "-1", "SignaturePrints": "0", "Sweeps": "-1", "TradeCount": "3", "TradeRank": "100", "TradeRankSnapshot": "-1", "VCD": "0"}
}

func darkPool20xFilters() map[string]string {
	return map[string]string{"Conditions": "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH", "DarkPools": "1", "EvenShared": "-1", "IncludeAH": "1", "IncludeClosing": "1", "IncludeOffsetting": "-1", "IncludeOpening": "1", "IncludePhantom": "-1", "IncludePremarket": "1", "IncludeRTH": "1", "LatePrints": "-1", "MarketCap": "0", "MaxDollars": "10000000000", "MaxPrice": "100000", "MaxVolume": "2000000000", "MinDollars": "500000", "MinPrice": "0", "MinVolume": "10000", "RelativeSize": "20", "SecurityTypeKey": "-1", "SignaturePrints": "0", "Sweeps": "-1", "TradeCount": "3", "TradeRank": "100", "TradeRankSnapshot": "-1", "VCD": "0"}
}

func top30Rank10x99thFilters() map[string]string {
	return map[string]string{"Conditions": "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH", "DarkPools": "-1", "EvenShared": "-1", "IncludeAH": "1", "IncludeClosing": "1", "IncludeOffsetting": "-1", "IncludeOpening": "1", "IncludePhantom": "-1", "IncludePremarket": "1", "IncludeRTH": "1", "LatePrints": "-1", "MarketCap": "0", "MaxDollars": "10000000000", "MaxPrice": "100000", "MaxVolume": "2000000000", "MinDollars": "500000", "MinPrice": "0", "MinVolume": "10000", "RelativeSize": "10", "SecurityTypeKey": "-1", "SignaturePrints": "0", "Sweeps": "-1", "TradeCount": "3", "TradeRank": "30", "TradeRankSnapshot": "-1", "VCD": "99"}
}

func phantomTradeFilters() map[string]string {
	return map[string]string{"Conditions": "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH", "DarkPools": "1", "EvenShared": "-1", "IncludeAH": "0", "IncludeClosing": "0", "IncludeOffsetting": "0", "IncludeOpening": "0", "IncludePhantom": "1", "IncludePremarket": "0", "IncludeRTH": "0", "LatePrints": "-1", "MarketCap": "0", "MaxDollars": "100000000000", "MaxPrice": "100000", "MaxVolume": "2000000000", "MinDollars": "500000", "MinPrice": "0", "MinVolume": "0", "RelativeSize": "0", "SecurityTypeKey": "-1", "SignaturePrints": "0", "Sweeps": "-1", "TradeCount": "3", "TradeRank": "-1", "TradeRankSnapshot": "-1", "VCD": "0"}
}

func offsettingTradeFilters() map[string]string {
	return map[string]string{"Conditions": "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH", "DarkPools": "-1", "EvenShared": "-1", "IncludeAH": "0", "IncludeClosing": "0", "IncludeOffsetting": "1", "IncludeOpening": "0", "IncludePhantom": "0", "IncludePremarket": "0", "IncludeRTH": "0", "LatePrints": "-1", "MarketCap": "0", "MaxDollars": "100000000000", "MaxPrice": "100000", "MaxVolume": "2000000000", "MinDollars": "500000", "MinPrice": "0", "MinVolume": "0", "RelativeSize": "0", "SecurityTypeKey": "-1", "SignaturePrints": "0", "Sweeps": "-1", "TradeCount": "3", "TradeRank": "-1", "TradeRankSnapshot": "-1", "VCD": "0"}
}
