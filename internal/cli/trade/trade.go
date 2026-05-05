package trade

import (
	"context"
	"fmt"
	"strconv"

	"github.com/leodido/structcli"
	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
)

const (
	tradeBrowserPageLength      = 100
	tradeListLongTermLength     = 10
	tradeDashboardDefaultCount  = 10
	maxTradeRequestLength       = 50
	maxTradeLevelRequestLength  = 50
	tradeListTickerLookbackDays = 365
)

func init() {
	structcli.RegisterEnum(map[tradeSummaryGroup][]string{
		tradeSummaryGroupTicker:    {"ticker"},
		tradeSummaryGroupDay:       {"day"},
		tradeSummaryGroupTickerDay: {"ticker,day", "ticker, day", "ticker day", "ticker-day"},
	})
}

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

type tradeRangeFilters struct {
	Tickers    string
	StartDate  string
	EndDate    string
	MinVolume  int
	MaxVolume  int
	MinPrice   float64
	MaxPrice   float64
	MinDollars float64
	MaxDollars float64
	VCD        int
	Sector     string
}

func (filters *tradeRangeFilters) baseMap() map[string]string {
	return map[string]string{
		"Tickers":    filters.Tickers,
		"StartDate":  filters.StartDate,
		"EndDate":    filters.EndDate,
		"MinVolume":  common.IntStr(filters.MinVolume),
		"MaxVolume":  common.IntStr(filters.MaxVolume),
		"MinPrice":   common.FormatFloat(filters.MinPrice),
		"MaxPrice":   common.FormatFloat(filters.MaxPrice),
		"MinDollars": common.FormatFloat(filters.MinDollars),
		"MaxDollars": common.FormatFloat(filters.MaxDollars),
		"VCD":        common.IntStr(filters.VCD),
	}
}

func (filters *tradeRangeFilters) clusterMap(securityType, relativeSize, rank int) map[string]string {
	values := filters.baseMap()
	values["SecurityTypeKey"] = common.IntStr(securityType)
	values["RelativeSize"] = common.IntStr(relativeSize)
	values["TradeClusterRank"] = common.IntStr(rank)
	values["SectorIndustry"] = filters.Sector
	return values
}

func (filters *tradeRangeFilters) clusterBombMap(securityType, relativeSize, rank int) map[string]string {
	values := filters.baseMap()
	delete(values, "MinPrice")
	delete(values, "MaxPrice")
	values["SecurityTypeKey"] = common.IntStr(securityType)
	values["RelativeSize"] = common.IntStr(relativeSize)
	values["TradeClusterBombRank"] = common.IntStr(rank)
	values["SectorIndustry"] = filters.Sector
	return values
}

func (filters *tradeRangeFilters) levelTouchMap(relativeSize, rank, count int) map[string]string {
	values := filters.baseMap()
	values["RelativeSize"] = common.IntStr(relativeSize)
	values["TradeLevelRank"] = common.IntStr(rank)
	values["Levels"] = common.IntStr(count)
	return values
}

type tradeDateRangeFlags struct {
	StartDate string `flag:"start-date" flaggroup:"Dates" flagshort:"s" flagdescr:"Start date YYYY-MM-DD (required unless --days is set)"`
	EndDate   string `flag:"end-date" flaggroup:"Dates" flagshort:"e" flagdescr:"End date YYYY-MM-DD (required unless --days is set)"`
	Days      int    `flag:"days" flaggroup:"Dates" flagshort:"d" flagdescr:"Look back this many days from --end-date or today"`
}

type tradeOptionalDateRangeFlags struct {
	StartDate string `flag:"start-date" flaggroup:"Dates" flagshort:"s" flagdescr:"Start date YYYY-MM-DD (default: auto)"`
	EndDate   string `flag:"end-date" flaggroup:"Dates" flagshort:"e" flagdescr:"End date YYYY-MM-DD (default: today)"`
	Days      int    `flag:"days" flaggroup:"Dates" flagshort:"d" flagdescr:"Look back this many days from --end-date or today"`
}

type tradeRangeFlags struct {
	MinVolume  int     `flag:"min-volume" flaggroup:"Ranges" flagdescr:"Minimum volume"`
	MaxVolume  int     `flag:"max-volume" flaggroup:"Ranges" flagdescr:"Maximum volume"`
	MinPrice   float64 `flag:"min-price" flaggroup:"Ranges" flagdescr:"Minimum price"`
	MaxPrice   float64 `flag:"max-price" flaggroup:"Ranges" flagdescr:"Maximum price"`
	MinDollars float64 `flag:"min-dollars" flaggroup:"Ranges" flagdescr:"Minimum dollar value"`
	MaxDollars float64 `flag:"max-dollars" flaggroup:"Ranges" flagdescr:"Maximum dollar value"`
}

type tradeVolumeDollarRangeFlags struct {
	MinVolume  int     `flag:"min-volume" flaggroup:"Ranges" flagdescr:"Minimum volume"`
	MaxVolume  int     `flag:"max-volume" flaggroup:"Ranges" flagdescr:"Maximum volume"`
	MinDollars float64 `flag:"min-dollars" flaggroup:"Ranges" flagdescr:"Minimum dollar value"`
	MaxDollars float64 `flag:"max-dollars" flaggroup:"Ranges" flagdescr:"Maximum dollar value"`
}

type tradeFilterFlags struct {
	Conditions   int                   `flag:"conditions" flaggroup:"Filters" flagdescr:"Trade conditions filter"`
	VCD          int                   `flag:"vcd" flaggroup:"Filters" flagdescr:"VCD filter"`
	SecurityType int                   `flag:"security-type" flaggroup:"Filters" flagdescr:"Security type key"`
	RelativeSize int                   `flag:"relative-size" flaggroup:"Filters" flagdescr:"Relative size threshold"`
	DarkPools    common.TriStateFilter `flag:"dark-pools" flaggroup:"Filters" flagdescr:"Dark pool filter (-1=all, 0=exclude, 1=only)"`
	Sweeps       common.TriStateFilter `flag:"sweeps" flaggroup:"Filters" flagdescr:"Sweep filter (-1=all, 0=exclude, 1=only)"`
	LatePrints   common.TriStateFilter `flag:"late-prints" flaggroup:"Filters" flagdescr:"Late print filter (-1=all, 0=exclude, 1=only)"`
	SigPrints    common.TriStateFilter `flag:"sig-prints" flaggroup:"Filters" flagdescr:"Signature print filter (-1=all, 0=exclude, 1=only)"`
	EvenShared   common.TriStateFilter `flag:"even-shared" flaggroup:"Filters" flagdescr:"Even shared filter (-1=all, 0=exclude, 1=only)"`
	TradeRank    int                   `flag:"trade-rank" flaggroup:"Filters" flagdescr:"Trade rank filter"`
	RankSnapshot int                   `flag:"rank-snapshot" flaggroup:"Filters" flagdescr:"Trade rank snapshot filter"`
	MarketCap    int                   `flag:"market-cap" flaggroup:"Filters" flagdescr:"Market cap filter"`
	Premarket    common.TriStateFilter `flag:"premarket" flaggroup:"Sessions" flagdescr:"Premarket session filter (-1=all, 0=exclude, 1=include)"`
	RTH          common.TriStateFilter `flag:"rth" flaggroup:"Sessions" flagdescr:"Regular trading hours filter (-1=all, 0=exclude, 1=include)"`
	AH           common.TriStateFilter `flag:"ah" flaggroup:"Sessions" flagdescr:"After-hours session filter (-1=all, 0=exclude, 1=include)"`
	Opening      common.TriStateFilter `flag:"opening" flaggroup:"Sessions" flagdescr:"Opening trade filter (-1=all, 0=exclude, 1=include)"`
	Closing      common.TriStateFilter `flag:"closing" flaggroup:"Sessions" flagdescr:"Closing trade filter (-1=all, 0=exclude, 1=include)"`
	Phantom      common.TriStateFilter `flag:"phantom" flaggroup:"Sessions" flagdescr:"Phantom print filter (-1=all, 0=exclude, 1=include)"`
	Offsetting   common.TriStateFilter `flag:"offsetting" flaggroup:"Sessions" flagdescr:"Offsetting trade filter (-1=all, 0=exclude, 1=include)"`
}

type tradePaginationFlags struct {
	Start    int                   `flag:"start" flaggroup:"Pagination" flagdescr:"DataTables start offset"`
	Length   int                   `flag:"length" flaggroup:"Pagination" flagshort:"l" flagdescr:"Number of results"`
	OrderCol int                   `flag:"order-col" flaggroup:"Pagination" flagdescr:"Order column index"`
	OrderDir common.OrderDirection `flag:"order-dir" flaggroup:"Pagination" flagdescr:"Order direction"`
}

type tradeFixedPageFlags struct {
	Start    int                   `flag:"start" flaggroup:"Pagination" flagdescr:"DataTables start offset"`
	OrderCol int                   `flag:"order-col" flaggroup:"Pagination" flagdescr:"Order column index"`
	OrderDir common.OrderDirection `flag:"order-dir" flaggroup:"Pagination" flagdescr:"Order direction"`
}

type tradeFormatFlag struct {
	Format common.OutputFormat `flag:"format" flaggroup:"Output" flagshort:"f" flagdescr:"Output format: json, csv, or tsv"`
}

type tradeTickersFlag struct {
	Tickers string `flag:"tickers" flaggroup:"Input" flagshort:"t" flagdescr:"Comma-separated ticker symbols"`
}

type tradeTickerFlag struct {
	Ticker string `flag:"ticker" flaggroup:"Input" flagshort:"t" flagdescr:"Ticker symbol"`
}

type tradeListOptions struct {
	tradeTickersFlag
	tradeOptionalDateRangeFlags
	tradeRangeFlags
	tradeFilterFlags
	Sector    string            `flag:"sector" flaggroup:"Input" flagdescr:"Sector/Industry filter"`
	Preset    string            `flag:"preset" flaggroup:"Input" flagdescr:"Apply a built-in filter preset by name; use report list for curated preset-backed reports"`
	Watchlist string            `flag:"watchlist" flaggroup:"Input" flagdescr:"Apply filters from a saved watchlist by name"`
	Fields    string            `flag:"fields" flaggroup:"Output" flagdescr:"Comma-separated trade fields to include in output"`
	Summary   bool              `flag:"summary" flaggroup:"Output" flagdescr:"Return aggregate metrics instead of individual trades"`
	GroupBy   tradeSummaryGroup `flag:"group-by" flaggroup:"Output" flagdescr:"Summary grouping (requires --summary): ticker, day, or ticker,day"`
	tradeFormatFlag
	tradeFixedPageFlags
}

type tradeSentimentOptions struct {
	tradeDateRangeFlags
	tradeRangeFlags
	tradeFilterFlags
	tradeFormatFlag
}

type tradeClustersOptions struct {
	tradeTickersFlag
	tradeDateRangeFlags
	tradeRangeFlags
	VCD              int    `flag:"vcd" flaggroup:"Filters" flagdescr:"VCD filter"`
	SecurityType     int    `flag:"security-type" flaggroup:"Filters" flagdescr:"Security type key"`
	RelativeSize     int    `flag:"relative-size" flaggroup:"Filters" flagdescr:"Relative size threshold"`
	TradeClusterRank int    `flag:"trade-cluster-rank" flaggroup:"Filters" flagdescr:"Trade cluster rank filter"`
	Sector           string `flag:"sector" flaggroup:"Input" flagdescr:"Sector/Industry filter"`
	Fields           string `flag:"fields" flaggroup:"Output" flagdescr:"Comma-separated TradeCluster fields to include in output, or 'all' for every field"`
	tradeFormatFlag
	tradeFixedPageFlags
}

type tradeClusterBombsOptions struct {
	tradeTickersFlag
	tradeDateRangeFlags
	tradeVolumeDollarRangeFlags
	VCD                  int    `flag:"vcd" flaggroup:"Filters" flagdescr:"VCD filter"`
	SecurityType         int    `flag:"security-type" flaggroup:"Filters" flagdescr:"Security type key"`
	RelativeSize         int    `flag:"relative-size" flaggroup:"Filters" flagdescr:"Relative size threshold"`
	TradeClusterBombRank int    `flag:"trade-cluster-bomb-rank" flaggroup:"Filters" flagdescr:"Trade cluster bomb rank filter"`
	Sector               string `flag:"sector" flaggroup:"Input" flagdescr:"Sector/Industry filter"`
	tradeFormatFlag
	tradeFixedPageFlags
}

type tradeLevelsOptions struct {
	tradeTickerFlag
	tradeOptionalDateRangeFlags
	TradeLevelCount int    `flag:"trade-level-count" flaggroup:"Output" flagdescr:"Number of price levels to return (5, 10, 20, or 50)"`
	Fields          string `flag:"fields" flaggroup:"Output" flagdescr:"Comma-separated TradeLevel fields to include in output, or 'all' for every field"`
	tradeFormatFlag
}

type tradeLevelTouchesOptions struct {
	tradeTickerFlag
	tradeDateRangeFlags
	tradeRangeFlags
	VCD             int `flag:"vcd" flaggroup:"Filters" flagdescr:"VCD filter"`
	RelativeSize    int `flag:"relative-size" flaggroup:"Filters" flagdescr:"Relative size threshold"`
	TradeLevelRank  int `flag:"trade-level-rank" flaggroup:"Filters" flagdescr:"Trade level rank filter"`
	TradeLevelCount int `flag:"trade-level-count" flaggroup:"Output" flagdescr:"Number of price levels to include (5, 10, 20, or 50)"`
	tradeFormatFlag
	tradePaginationFlags
}

type tradeDashboardOptions struct {
	tradeTickerFlag
	tradeOptionalDateRangeFlags
	tradeRangeFlags
	tradeFilterFlags
	Count int `flag:"count" flaggroup:"Output" flagshort:"c" flagdescr:"Rows to return per dashboard section (5, 10, 20, or 50)"`
}

func (opts *tradeLevelsOptions) Validate(_ context.Context) []error {
	if err := validateTradeLevelCount(opts.TradeLevelCount); err != nil {
		return []error{err}
	}
	return nil
}

func (opts *tradeLevelTouchesOptions) Validate(_ context.Context) []error {
	if err := validateRange(opts.Length, maxTradeLevelRequestLength, "length", "trade level touch retrieval"); err != nil {
		return []error{err}
	}
	if err := validateTradeLevelCount(opts.TradeLevelCount); err != nil {
		return []error{err}
	}
	if opts.TradeLevelRank < 5 {
		return []error{fmt.Errorf("--trade-level-rank must be 5 or higher for trade level touch retrieval")}
	}
	return nil
}

func (opts *tradeDashboardOptions) Validate(_ context.Context) []error {
	if err := validateTradeLevelCount(opts.Count); err != nil {
		return []error{fmt.Errorf("--count must be one of 5, 10, 20, or 50 for ticker dashboard retrieval")}
	}
	return nil
}

type tradeAlertsOptions struct {
	Date string `flag:"date" flaggroup:"Dates" flagshort:"d" flagdescr:"Date YYYY-MM-DD"`
	tradeFormatFlag
	tradePaginationFlags
}

type tradeClusterAlertsOptions struct {
	Date string `flag:"date" flaggroup:"Dates" flagshort:"d" flagdescr:"Date YYYY-MM-DD"`
	tradeFormatFlag
	tradePaginationFlags
}

// NewCmd returns the "trade" command group with all subcommands.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "trade",
		Short:   "Trade-related commands",
		GroupID: "trading",
		Args:    cobra.NoArgs,
		Long:    "trade contains subcommands for querying institutional trade data, trade clusters, sentiment metrics, price levels, and alert activity from VolumeLeaders. Use subcommands to filter by ticker, date range, dollar amounts, trade conditions, and more. All subcommands output compact JSON to stdout by default.",
	}
	cmd.AddCommand(
		newTradeListCommand(),
		newTradeDashboardCommand(),
		newTradeSentimentCommand(),
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

func newTradeDashboardCommand() *cobra.Command {
	opts := &tradeDashboardOptions{}
	presetTradeFilterDefaults(&opts.tradeFilterFlags, 0)
	presetTradeRangeDefaults(&opts.tradeRangeFlags, 500000)
	opts.RelativeSize = 0
	opts.Count = tradeDashboardDefaultCount
	cmd := &cobra.Command{
		Use:   "dashboard [ticker]",
		Short: "Query a ticker institutional dashboard",
		Long: `Query a fast ticker dashboard with the same chart-optimized institutional context VolumeLeaders shows in the browser. The dashboard fetches the largest trades, trade clusters, trade levels, and cluster bombs for one ticker in a single JSON object.

Defaults to a 365-day lookback, 10 rows per section, --vcd 0, --relative-size 0, and the same broad trade/session filters used by the browser chart page. Use this command as the first stop for any single-ticker investigation, including institutional levels, largest trades, clustered activity, or sudden bursts, then drill into trade list, trade clusters, trade levels, or trade cluster-bombs only when a section needs deeper pagination, CSV/TSV output, or explicit field selection.

PREREQUISITES: Provide exactly one ticker as a positional argument or with --ticker. Browser authentication must be available.

RECOVERY: If ticker validation fails, use one ticker only. If --count is rejected, use 5, 10, 20, or 50. If date flags conflict, use either --days or --start-date with --end-date.`,
		Example:    "volumeleaders-agent trade dashboard IGV",
		Args:       cobra.ArbitraryArgs,
		SuggestFor: []string{"dash", "overview", "chart"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTradeDashboard(cmd, opts)
		},
	}
	common.BindOrPanic(cmd, opts, "dashboard")
	setTradeRangeFlagDefValues(cmd, opts.MinDollars)
	return cmd
}

func newTradeListCommand() *cobra.Command {
	opts := &tradeListOptions{}
	presetTradeFilterDefaults(&opts.tradeFilterFlags, 97)
	presetTradeRangeDefaults(&opts.tradeRangeFlags, 500000)
	presetTradeFixedPageDefaults(&opts.tradeFixedPageFlags, 1)
	opts.Format = "json"
	opts.GroupBy = "ticker"
	cmd := &cobra.Command{
		Use:   "list [tickers...]",
		Short: "Query institutional trades",
		Long: `Query individual institutional trades from VolumeLeaders, filterable by ticker, date range, dollar amounts, volume, trade conditions, session type, and trade rank. Supports built-in filter presets (--preset) and watchlist-based filtering (--watchlist). Start with report list for curated preset-backed reports; use trade list when custom raw trade filters are needed. Outputs compact JSON or CSV/TSV with --format; use --summary for aggregate metrics grouped by ticker or day.

Date defaults: 365-day lookback when tickers are provided, today-only without tickers. Preset and watchlist filters do not supply dates. Filter precedence is preset baseline, then watchlist merge, then explicit CLI flags override both.

Default JSON is compact and omits repetitive/internal fields. Use --fields FIELD1,FIELD2, CSV/TSV, or --fields all where supported when raw API fields are needed. --summary returns aggregate JSON with valid --group-by values of ticker, day, or ticker,day; do not combine summary mode with --fields or non-JSON formats.

KEY METRICS

Field                      Meaning
-------------------------  ---------------------------------------------------------------
CumulativeDistribution     Volume percentile, 0 to 1, higher means more accumulation
DollarsMultiplier          Trade dollars relative to average block size
TradeRank                  VL significance rank now, lower is stronger
TradeRankSnapshot          VL significance rank at print time, lower is stronger
TradeClusterRank           Rank for cluster significance, lower is stronger
TradeClusterBombRank       Rank for burst significance, lower is stronger
TradeLevelRank             Rank for level significance, lower is stronger
RelativeSize               Trade size vs normal activity
PercentDailyVolume         Trade volume as percent of average daily volume
VCD                        Volume Confirmation Distribution score
FrequencyLast30TD          Similar-magnitude trade frequency over last 30 trading days
FrequencyLast90TD          Similar-magnitude trade frequency over last 90 trading days
FrequencyLast1CY           Similar-magnitude trade frequency over last calendar year
RSIHour                    Hourly RSI at time of trade
RSIDay                     Daily RSI at time of trade
DarkPool                   Boolean: trade printed on a dark pool
Sweep                      Boolean: trade was a sweep order
LatePrint                  Boolean: trade was a late print
SignaturePrint             Boolean: trade matched a signature print pattern
PhantomPrint               Boolean: trade was a phantom print
InsideBar                  Boolean: bar was an inside bar

Shared trade filters include volume, price, dollars, conditions, VCD, relative size, security type, market cap, trade rank, dark pools, sweeps, late prints, signature prints, even-share prints, and session/event toggles.

PREREQUISITES: Browser authentication. For reproducible scans, pass explicit dates or --days plus tickers, preset, watchlist, or sector filters.

RECOVERY: Multi-day lookups whose effective filters include tickers return the top 10 long-period trades with the same lightweight chart query shape VolumeLeaders uses in the browser. Single-day scans, all-market scans, sector-only presets, and --summary still fetch all matching rows in browser-sized 100-row pages. If --summary rejects --fields or --format, rerun summary as JSON without --fields. If date flags conflict, use either --days or --start-date with --end-date.

NEXT STEPS: Use trade dashboard first for any single-ticker investigation, then trade levels for level-only support/resistance output, trade clusters when prints concentrate near a price, or trade sentiment for leveraged ETF bull/bear context.`,
		Example: `volumeleaders-agent trade list AAPL MSFT
volumeleaders-agent trade list --tickers AAPL,MSFT
volumeleaders-agent trade list --tickers NVDA --dark-pools 1 --min-dollars 1000000
volumeleaders-agent trade list --sector Technology --relative-size 10
volumeleaders-agent trade list --preset "Top-100 Rank" --start-date 2025-04-01 --end-date 2025-04-24
volumeleaders-agent trade list --watchlist "Magnificent 7" --start-date 2025-04-01 --end-date 2025-04-24`,
		Args:       cobra.ArbitraryArgs,
		Aliases:    []string{"ls"},
		SuggestFor: []string{"lst", "lis"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTradeList(cmd, opts)
		},
	}
	common.BindOrPanic(cmd, opts, "list")
	setTradeRangeFlagDefValues(cmd, opts.MinDollars)
	return cmd
}

func newTradeSentimentCommand() *cobra.Command {
	opts := &tradeSentimentOptions{}
	presetTradeFilterDefaults(&opts.tradeFilterFlags, 97)
	presetTradeRangeDefaults(&opts.tradeRangeFlags, 5000000)
	opts.Format = "json"
	cmd := &cobra.Command{
		Use:   "sentiment",
		Short: "Summarize leveraged ETF bull/bear flow by day",
		Long: `Summarize leveraged ETF bull and bear flow by trading day, showing aggregate institutional dollar volume on the bull and bear side. Requires --start-date and --end-date (or --days). Outputs one record per day with bull and bear totals.

This command always queries the combined leveraged ETF sector filter SectorIndustry=X B, classifies bull and bear ETFs locally, and cannot be constrained by ticker or sector flags. Non-standard defaults include --min-dollars 5000000 and --vcd 97; shared --relative-size 5 still applies.

Ratio is bull dollars divided by bear dollars and is null when bear flow is zero. Treat the output as leveraged ETF proxy flow, not signed buy/sell flow for the broader market.`,
		Example:    "volumeleaders-agent trade sentiment --start-date 2025-04-21 --end-date 2025-04-25",
		Args:       cobra.NoArgs,
		SuggestFor: []string{"sentment", "sentimnt"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTradeSentiment(cmd, opts)
		},
	}
	common.BindOrPanic(cmd, opts, "sentiment")
	setTradeRangeFlagDefValues(cmd, opts.MinDollars)
	return cmd
}

func newTradeClustersCommand() *cobra.Command {
	opts := &tradeClustersOptions{}
	presetTradeRangeDefaults(&opts.tradeRangeFlags, 10000000)
	presetTradeFixedPageDefaults(&opts.tradeFixedPageFlags, 1)
	opts.SecurityType = -1
	opts.RelativeSize = 5
	opts.TradeClusterRank = -1
	opts.Format = "json"
	cmd := &cobra.Command{
		Use:   "clusters [tickers...]",
		Short: "Query aggregated trade clusters",
		Long: `Query aggregated trade clusters, which group multiple trades in a short window into a single cluster record. Filterable by ticker, date range, dollar amounts, sector, and trade cluster rank. Outputs compact JSON or CSV/TSV with --format.


Results are fetched in browser-sized 100-row pages to match VolumeLeaders' frontend behavior. Use clusters when the question is about price-level concentration, not single prints. This command uses larger default dollar thresholds than ordinary trade list. Use trade cluster-bombs instead when looking for sudden aggressive bursts tightly grouped in time and price.`,
		Example:    "volumeleaders-agent trade clusters AAPL --days 7",
		Args:       cobra.ArbitraryArgs,
		SuggestFor: []string{"cluster", "clsters"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTradeClusters(cmd, opts)
		},
	}
	common.BindOrPanic(cmd, opts, "clusters")
	setTradeRangeFlagDefValues(cmd, opts.MinDollars)
	return cmd
}

func newTradeClusterBombsCommand() *cobra.Command {
	opts := &tradeClusterBombsOptions{}
	presetTradeVolumeDollarRangeDefaults(&opts.tradeVolumeDollarRangeFlags, 0)
	presetTradeFixedPageDefaults(&opts.tradeFixedPageFlags, 1)
	opts.TradeClusterBombRank = -1
	opts.Format = "json"
	cmd := &cobra.Command{
		Use:   "cluster-bombs [tickers...]",
		Short: "Query trade cluster bombs",
		Long: `Query trade cluster bombs, which are extreme-magnitude trade clusters that exceed normal institutional activity thresholds. Filterable by ticker, date range, dollar amounts, sector, and cluster bomb rank. Outputs compact JSON by default.

Results are fetched in browser-sized 100-row pages to match VolumeLeaders' frontend behavior. Cluster bombs find sudden aggressive bursts tightly grouped in time and price, with different defaults and rank fields than trade clusters. Use this command when looking for extreme concentration events, not general price-level clustering.`,
		Example:    "volumeleaders-agent trade cluster-bombs TSLA --days 3",
		Args:       cobra.ArbitraryArgs,
		SuggestFor: []string{"clusterbombs", "cluster-bomb"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTradeClusterBombs(cmd, opts)
		},
	}
	common.BindOrPanic(cmd, opts, "cluster-bombs")
	setTradeVolumeDollarFlagDefValues(cmd, opts.MinDollars)
	return cmd
}

func newTradeAlertsCommand() *cobra.Command {
	opts := &tradeAlertsOptions{}
	presetTradePaginationDefaults(&opts.tradePaginationFlags, 100, 1)
	opts.Format = "json"
	cmd := &cobra.Command{
		Use:   "alerts",
		Short: "Query trade alerts for a date",
		Long: `Query trade alerts fired on a specific date based on saved alert configurations. Requires --date. Returns alert records matching your configured filters. Outputs compact JSON or CSV/TSV with --format.

Alert configs trigger when trades match thresholds. Threshold names follow the pattern CategoryMetricLTE or CategoryMetricGTE where LTE is maximum rank and GTE is minimum value. Use alert configs to see your configured thresholds.`,
		Example:    "volumeleaders-agent trade alerts --date 2025-01-15",
		Args:       cobra.NoArgs,
		SuggestFor: []string{"alert", "alrts"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTradeAlerts(cmd, opts)
		},
	}
	common.BindOrPanic(cmd, opts, "alerts")
	_ = cmd.MarkFlagRequired("date")
	return cmd
}

func newTradeClusterAlertsCommand() *cobra.Command {
	opts := &tradeClusterAlertsOptions{}
	presetTradePaginationDefaults(&opts.tradePaginationFlags, 100, 1)
	opts.Format = "json"
	cmd := &cobra.Command{
		Use:   "cluster-alerts",
		Short: "Query trade cluster alerts for a date",
		Long: `Query trade cluster alerts fired on a specific date based on saved alert configurations that target cluster activity. Requires --date. Returns cluster alert records matching your configured filters.

Cluster alert rows use the full cluster-shaped model rather than the compact default from trade clusters. Use trade alerts for individual trade alert rows and this command for cluster-level alert rows.`,
		Example:    "volumeleaders-agent trade cluster-alerts --date 2025-01-15",
		Args:       cobra.NoArgs,
		SuggestFor: []string{"clusteralerts", "cluster-alert"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTradeClusterAlerts(cmd, opts)
		},
	}
	common.BindOrPanic(cmd, opts, "cluster-alerts")
	_ = cmd.MarkFlagRequired("date")
	return cmd
}

func newTradeLevelsCommand() *cobra.Command {
	opts := &tradeLevelsOptions{}
	opts.TradeLevelCount = 10
	opts.Format = "json"
	cmd := &cobra.Command{
		Use:   "levels [ticker]",
		Short: "Query significant price levels for a ticker",
		Long: `Query significant price levels for a ticker, showing historical support and resistance zones identified by institutional trade clustering. Accepts a ticker as positional argument or via --ticker flag. Outputs compact JSON by default.

Defaults to a 365-day lookback when dates are omitted and shares the chart-optimized VolumeLeaders level request used by trade dashboard. This command intentionally exposes a reduced CLI surface: ticker, dates, --trade-level-count, --fields, and --format. For any single-ticker investigation, run trade dashboard TICKER first because it returns trades, clusters, levels, and cluster bombs together; use trade levels only when you need level-only output, CSV/TSV, or explicit field selection. Only --trade-level-count values of 5, 10, 20, or 50 are accepted. Default JSON is compact and omits repetitive ticker metadata and the verbose Dates list; use --fields all or CSV/TSV when raw fields are needed.

PREREQUISITES: Provide exactly one ticker as a positional argument or with --ticker.

RECOVERY: If ticker validation fails, use one ticker only. If --trade-level-count is rejected, use 5, 10, 20, or 50.

NEXT STEPS: Use trade dashboard as the first single-ticker overview, or use trade level-touches with the same ticker and date range to find trades that revisited these levels.`,
		Example:    "volumeleaders-agent trade levels AAPL",
		Args:       cobra.ArbitraryArgs,
		SuggestFor: []string{"level", "lvels"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTradeLevels(cmd, opts)
		},
	}
	common.BindOrPanic(cmd, opts, "levels")
	return cmd
}

func newTradeLevelTouchesCommand() *cobra.Command {
	opts := &tradeLevelTouchesOptions{}
	presetTradeRangeDefaults(&opts.tradeRangeFlags, 500000)
	presetTradePaginationDefaults(&opts.tradePaginationFlags, maxTradeLevelRequestLength, 0)
	opts.TradeLevelRank = 5
	opts.TradeLevelCount = maxTradeLevelRequestLength
	opts.Format = "json"
	cmd := &cobra.Command{
		Use:   "level-touches [ticker]",
		Short: "Query trade events at notable price levels",
		Long: `Query institutional trade events that occurred at notable price levels for a ticker, showing how the market interacted with key support and resistance zones. Accepts a ticker as positional argument or via --ticker flag. Requires --start-date and --end-date (or --days).

Defaults to --trade-level-rank 5 and --length 50, rejects --length -1, --length 0, and values above 50, and only allows --trade-level-count values of 5, 10, 20, or 50. Use trade levels first to identify significant price zones, then use this command to find events where price revisited those levels.

PREREQUISITES: Provide exactly one ticker and a date range with --start-date and --end-date or --days.

RECOVERY: If --length is rejected, use 1 to 50. If --trade-level-count is rejected, use 5, 10, 20, or 50. If --trade-level-rank is rejected, use 5 or higher. If dates are missing, add --days N for a quick retry.

NEXT STEPS: Compare touched levels with fresh trade list output to see whether recent institutional prints confirm or reject the level.`,
		Example:    "volumeleaders-agent trade level-touches AAPL --days 14",
		Args:       cobra.ArbitraryArgs,
		SuggestFor: []string{"leveltouches", "level-touch"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runTradeLevelTouches(cmd, opts)
		},
	}
	common.BindOrPanic(cmd, opts, "level-touches")
	setTradeRangeFlagDefValues(cmd, opts.MinDollars)
	return cmd
}

func presetTradeRangeDefaults(opts *tradeRangeFlags, minDollarsDefault float64) {
	opts.MaxVolume = 2000000000
	opts.MaxPrice = 100000
	opts.MinDollars = minDollarsDefault
	opts.MaxDollars = 30000000000
}

func presetTradeVolumeDollarRangeDefaults(opts *tradeVolumeDollarRangeFlags, minDollarsDefault float64) {
	opts.MaxVolume = 2000000000
	opts.MinDollars = minDollarsDefault
	opts.MaxDollars = 30000000000
}

func presetTradeFilterDefaults(opts *tradeFilterFlags, vcdDefault int) {
	opts.Conditions = -1
	opts.VCD = vcdDefault
	opts.SecurityType = -1
	opts.RelativeSize = 5
	opts.DarkPools = common.TriStateAll
	opts.Sweeps = common.TriStateAll
	opts.LatePrints = common.TriStateAll
	opts.SigPrints = common.TriStateAll
	opts.EvenShared = common.TriStateAll
	opts.TradeRank = -1
	opts.RankSnapshot = -1
	opts.Premarket = common.TriStateOnly
	opts.RTH = common.TriStateOnly
	opts.AH = common.TriStateOnly
	opts.Opening = common.TriStateOnly
	opts.Closing = common.TriStateOnly
	opts.Phantom = common.TriStateOnly
	opts.Offsetting = common.TriStateOnly
}

func presetTradePaginationDefaults(opts *tradePaginationFlags, length, orderCol int) {
	opts.Length = length
	opts.OrderCol = orderCol
	opts.OrderDir = "desc"
}

func presetTradeFixedPageDefaults(opts *tradeFixedPageFlags, orderCol int) {
	opts.OrderCol = orderCol
	opts.OrderDir = "desc"
}

func validateTradeLevelCount(count int) error {
	switch count {
	case 5, 10, 20, maxTradeLevelRequestLength:
		return nil
	default:
		return fmt.Errorf("--trade-level-count must be one of 5, 10, 20, or 50 for trade level retrieval")
	}
}

func setTradeRangeFlagDefValues(cmd *cobra.Command, minDollarsDefault float64) {
	setFloatFlagDefValue(cmd, "min-price", 0)
	setFloatFlagDefValue(cmd, "max-price", 100000)
	setTradeVolumeDollarFlagDefValues(cmd, minDollarsDefault)
}

func setTradeVolumeDollarFlagDefValues(cmd *cobra.Command, minDollarsDefault float64) {
	setFloatFlagDefValue(cmd, "min-dollars", minDollarsDefault)
	setFloatFlagDefValue(cmd, "max-dollars", 30000000000)
}

func setFloatFlagDefValue(cmd *cobra.Command, name string, value float64) {
	if flag := cmd.Flags().Lookup(name); flag != nil {
		flag.DefValue = strconv.FormatFloat(value, 'f', -1, 64)
	}
}

func buildTradeFilters(opts *tradesOptions) map[string]string {
	return map[string]string{"Tickers": opts.tickers, "StartDate": opts.startDate, "EndDate": opts.endDate, "MinVolume": common.IntStr(opts.minVolume), "MaxVolume": common.IntStr(opts.maxVolume), "MinPrice": common.FormatFloat(opts.minPrice), "MaxPrice": common.FormatFloat(opts.maxPrice), "MinDollars": common.FormatFloat(opts.minDollars), "MaxDollars": common.FormatFloat(opts.maxDollars), "Conditions": common.IntStr(opts.conditions), "VCD": common.IntStr(opts.vcd), "SecurityTypeKey": common.IntStr(opts.securityType), "RelativeSize": common.IntStr(opts.relativeSize), "DarkPools": common.IntStr(opts.darkPools), "Sweeps": common.IntStr(opts.sweeps), "LatePrints": common.IntStr(opts.latePrints), "SignaturePrints": common.IntStr(opts.sigPrints), "EvenShared": common.IntStr(opts.evenShared), "TradeRank": common.IntStr(opts.tradeRank), "TradeRankSnapshot": common.IntStr(opts.rankSnapshot), "MarketCap": common.IntStr(opts.marketCap), "IncludePremarket": common.IntStr(opts.premarket), "IncludeRTH": common.IntStr(opts.rth), "IncludeAH": common.IntStr(opts.ah), "IncludeOpening": common.IntStr(opts.opening), "IncludeClosing": common.IntStr(opts.closing), "IncludePhantom": common.IntStr(opts.phantom), "IncludeOffsetting": common.IntStr(opts.offsetting), "SectorIndustry": opts.sector}
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
