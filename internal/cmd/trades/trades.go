// Package trades contains commands for VolumeLeaders trade workflows.
package trades

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/leodido/structcli"
	"github.com/major/volumeleaders-agent/internal/auth"
	"github.com/spf13/cobra"
)

const (
	dateLayout           = "2006-01-02"
	defaultTradeLimit    = 100
	defaultTradePageSize = 100
	maxTradeLimit        = 100
	defaultFieldPreset   = "core"
	defaultOutputShape   = "array"
	fullFieldPreset      = "full"
	objectOutputShape    = "objects"
	calendarEventField   = "CalendarEvent"
	auctionTradeField    = "AuctionTrade"
	cancelledTradeField  = "Cancel" + "led"
	openingTradeField    = "OpeningTrade"
	closingTradeField    = "ClosingTrade"
	getTradesPath        = "https://www.volumeleaders.com/Trades/GetTrades"
	getTradeClustersPath = "https://www.volumeleaders.com/TradeClusters/GetTradeClusters"
	getTradeLevelsPath   = "https://www.volumeleaders.com/TradeLevels/GetTradeLevels"
	tradesPage           = "https://www.volumeleaders.com/Trades"
	tradeClustersPage    = "https://www.volumeleaders.com/TradeClusters"
	tradeLevelsPage      = "https://www.volumeleaders.com/TradeLevels"
	tradeLLMFieldGuide   = "LLM field guide: users and LLM callers should focus on VolumeLeaders table fields: time, ticker/count, CP, TP, sector, industry, Sh, $$, RS, PCT, R, and Last. Ticker is the stock ticker symbol, such as TSLA or AMZN. TradeCount is the #T count shown beside the ticker: the number of large trades for that ticker today, so KRE (2) means two large KRE trades today. Raw fields outside that visible table are secondary debugging or correlation context unless this guide says otherwise. DarkPools and Sweeps are request filters. Dark pool trades are done off exchange and reported later; lit exchange trades are done on exchange and reported immediately. Sweeps are orders spread across multiple exchanges to get done quickly; blocks are orders sent to one exchange. For the trades command, false/false means show everything, --dark-pools alone means dark pools of all kinds, --sweeps alone means sweeps from dark pools or lit exchanges, and both flags together mean dark pool sweeps only. In raw output, DarkPool and Sweep describe the classification of each returned row. RelativeSize is a request filter for minimum relative size. Captured browser values are 0, 5, 10, 25, 50, and 100, where 0 means any size and the others mean at least that many times the ticker's average dollar trade size. DollarsMultiplier, shown as RS in the UI, is the returned relative size value: trade dollars divided by average dollars for that ticker. VolumeLeaders highlights trades at or above 25x average size. CumulativeDistribution, shown as PCT in the UI, is the trade's percentile rank relative to other trades for the same ticker. Conditions carries RSI condition filters: OBD means overbought daily, OBH means overbought hourly, OSD means oversold daily, and OSH means oversold hourly. Captured defaults use -1 for no RSI condition filter. IgnoreOBD, IgnoreOBH, IgnoreOSD, and IgnoreOSH mean do not consider that RSI condition; they do not mean exclude matching rows. VCD appears to carry the minimum CumulativeDistribution percentile. Captures use 0 for no percentile filter and 99 for the 99th percentile or above. TradeID, SequenceNumber, and SecurityKey are VolumeLeaders internal identifiers. DateKey and TimeKey are compact internal date/time keys. Treat these five fields as upstream metadata for correlation or debugging, not as trading-decision signals. Date is the trade date. FullDateTime is the full trade timestamp. StartDate and EndDate appear to be upstream query-range echoes or internal metadata rather than separate trade signals. LastComparibleTradeDate is the upstream spelling for the last date VolumeLeaders saw a trade close to this trade's size. Ask and Bid are the ask and bid prices in the bid/ask spread when the trade happened. ClosePrice is CP in the UI: the close price at the end of the day, or the current price if the market is still open. Price is TP in the UI: the trade price when the large trade hit. AverageDailyVolume is a moving-average measure of the stock's normal volume, and PercentDailyVolume compares today's volume with that moving average. Volume is Sh in the UI: how many shares were in the trade. Dollars is $$ in the UI: how big the trade was in dollars, calculated as shares times the trade price. TradeRank is the trade's current rank among all current trades and can change when larger trades arrive. In the UI R column, a dash means the trade is not ranked in the top 100 trades, while a number such as 27 means the trade is currently ranked 27th. TradeRankSnapshot is immutable: it preserves how the trade ranked at the time it appeared. TotalVolume and TotalDollars are internal upstream values; do not treat them as standalone trading-decision signals."
	calendarEventGuide   = "CalendarEvent is a compact derived core field. It contains true upstream calendar markers joined with commas, or null in array output when no marker is true. Source markers are EOM for end of month, EOQ for end of quarter, EOY for end of year, OPEX for a market options expiration date, and VOLEX for a market volatility expiration date such as VIX options expiration. In object output, CalendarEvent is omitted when no marker is true. AuctionTrade is a compact derived core field from upstream OpeningTrade and ClosingTrade 0/1 flags. It is open for opening auction trades, close for market-on-close auction trades, or null in array output when neither flag is true. In object output, AuctionTrade is omitted when neither flag is true."
	expandedPresetGuide  = "Use --preset-fields expanded when an LLM needs every annotated non-internal signal without raw upstream noise. Trade expanded fields add IDs/timestamps, ticker identity, price/size, comparable dates, ranks, print classifications, RSI values, frequency counts, cancellation state, calendar flags, and auction flags. Cluster expanded fields add the date/time range, ticker identity, price/size, cluster count/rank, comparable cluster date, IPO date, distribution, calendar flags, and inside-bar flags. Use --preset-fields full only for debugging raw upstream payloads, because full includes internal, always-zero, and always-null fields."
	tradeLevelGuide      = "Trade levels group one ticker's large-trade history by price level across a date range. Price is the level price. Dollars and Volume aggregate large prints at that level. Trades is the count of prints contributing to that level. RelativeSize is the level's relative size versus the ticker's average dollar trade size. CumulativeDistribution is the level percentile shown as PCT. TradeLevelRank is the all-time level rank when VolumeLeaders marks the level as ranked; captured browser output uses 0 for unranked rows. Dates is the upstream level date range."
	ignoredRSIConditions = "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH"
)

var calendarEventFields = []string{"EOM", "EOQ", "EOY", "OPEX", "VOLEX"}

var (
	extractCookies           = auth.ExtractCookies
	fetchXSRFToken           = auth.FetchXSRFToken
	getTradesHTTPClient      = http.DefaultClient
	getTradesEndpoint        = getTradesPath
	getTradeClustersEndpoint = getTradeClustersPath
	getTradeLevelsEndpoint   = getTradeLevelsPath
	tickerPattern            = regexp.MustCompile(`^[A-Z0-9.-]+$`)
	nullJSON                 = json.RawMessage("null")
)

var tradeFieldPresets = map[string][]string{
	"core": {
		"Ticker",
		"TradeCount",
		"FullTimeString24",
		"ClosePrice",
		"Price",
		"Sector",
		"Industry",
		"Volume",
		"Dollars",
		"DollarsMultiplier",
		"CumulativeDistribution",
		"TradeRank",
		"LastComparibleTradeDate",
		"CalendarEvent",
		"AuctionTrade",
	},
	"expanded": {
		"Date",
		"DateKey",
		"TimeKey",
		"TradeID",
		"Ticker",
		"Sector",
		"Industry",
		"Name",
		"FullDateTime",
		"FullTimeString24",
		"Price",
		"Dollars",
		"Volume",
		"LastComparibleTradeDate",
		"IPODate",
		"OffsettingTradeDate",
		"TradeCount",
		"CumulativeDistribution",
		"TradeRank",
		"TradeRankSnapshot",
		"LatePrint",
		"Sweep",
		"DarkPool",
		"OpeningTrade",
		"ClosingTrade",
		"AuctionTrade",
		"PhantomPrint",
		"InsideBar",
		"DoubleInsideBar",
		"SignaturePrint",
		"NewPosition",
		"RSIHour",
		"RSIDay",
		"TotalRows",
		"FrequencyLast30TD",
		"FrequencyLast90TD",
		"FrequencyLast1CY",
		cancelledTradeField,
		"EOM",
		"EOQ",
		"EOY",
		"OPEX",
		"VOLEX",
		"CalendarEvent",
	},
}

var clusterFieldPresets = map[string][]string{
	"core": {
		"Ticker",
		"MinFullTimeString24",
		"MaxFullTimeString24",
		"Price",
		"Dollars",
		"DollarsMultiplier",
		"Volume",
		"TradeCount",
		"TradeClusterRank",
		"Sector",
		"CalendarEvent",
		"AuctionTrade",
	},
	"expanded": {
		"Date",
		"DateKey",
		"Ticker",
		"Sector",
		"Industry",
		"Name",
		"MinFullDateTime",
		"MaxFullDateTime",
		"MinFullTimeString24",
		"MaxFullTimeString24",
		"Price",
		"Dollars",
		"Volume",
		"TradeCount",
		"LastComparibleTradeClusterDate",
		"IPODate",
		"CumulativeDistribution",
		"TradeClusterRank",
		"EOM",
		"EOQ",
		"EOY",
		"OPEX",
		"VOLEX",
		"CalendarEvent",
		"InsideBar",
		"DoubleInsideBar",
	},
}

var tradeLevelFieldPresets = map[string][]string{
	"core": {
		"Ticker",
		"Price",
		"Dollars",
		"Volume",
		"Trades",
		"RelativeSize",
		"CumulativeDistribution",
		"TradeLevelRank",
		"Dates",
	},
	"expanded": {
		"Ticker",
		"Sector",
		"Industry",
		"Name",
		"Date",
		"MinDate",
		"MaxDate",
		"FullDateTime",
		"FullTimeString24",
		"Dates",
		"Price",
		"Dollars",
		"Volume",
		"Trades",
		"RelativeSize",
		"CumulativeDistribution",
		"TradeLevelRank",
		"TradeLevelTouches",
		"TotalRows",
	},
}

// Options defines the LLM-readable contract for fetching unusual trades.
type Options struct {
	Date         string `flag:"date" flagshort:"d" flagdescr:"Single trading date to query, formatted as YYYY-MM-DD. The disproportionately large trades preset is intentionally limited to one day." flagenv:"true" flagrequired:"true" flaggroup:"Query" validate:"required" mod:"trim"`
	Tickers      string `flag:"tickers" flagdescr:"Optional ticker filter. Use one symbol or a comma-delimited list without spaces, for example AAPL or AAPL,MSFT." flagenv:"true" flaggroup:"Query" mod:"trim"`
	DarkPools    bool   `flag:"dark-pools" flagdescr:"Filter to dark-pool/off-exchange prints only. Dark pool trades are done off exchange and reported later; leave false to include both dark-pool and lit-exchange trades." flagenv:"true" flaggroup:"Filters"`
	Sweeps       bool   `flag:"sweeps" flagdescr:"Filter to sweep orders only. Sweeps are orders spread across multiple exchanges to get done quickly; leave false to include both sweep and block executions." flagenv:"true" flaggroup:"Filters"`
	Limit        int    `flag:"limit" flagdescr:"Maximum trade rows to return. Must be between 1 and 100. Defaults to 100 when omitted." flagenv:"true" flaggroup:"Output"`
	Fields       string `flag:"fields" flagdescr:"Comma-separated trade fields to include. Overrides --preset-fields. Use upstream field names such as Ticker,Dollars,TradeRank." flagenv:"true" flaggroup:"Output" mod:"trim"`
	PresetFields string `flag:"preset-fields" flagdescr:"Field preset to include: core, expanded, or full. Defaults to core for token-efficient output. Expanded includes annotated non-internal signal fields; full returns raw upstream payloads." flagenv:"true" flaggroup:"Output" mod:"trim"`
	Shape        string `flag:"shape" flagdescr:"Trade row shape: array or objects. Array is the default and is most token-efficient." flagenv:"true" flaggroup:"Output" mod:"trim"`
	Pretty       bool   `flag:"pretty" flagdescr:"Pretty-print JSON output. Compact JSON is the default for token-efficient LLM and MCP use." flagenv:"true" flaggroup:"Output"`
}

// RankedOptions defines the LLM-readable contract for fetching all-time ranked trades.
type RankedOptions struct {
	Date         string `flag:"date" flagshort:"d" flagdescr:"Single trading date to query, formatted as YYYY-MM-DD. Ranked trade presets are intentionally limited to one day." flagenv:"true" flagrequired:"true" flaggroup:"Query" validate:"required" mod:"trim"`
	Tickers      string `flag:"tickers" flagdescr:"Optional ticker filter. Use one symbol or a comma-delimited list without spaces, for example AAPL or AAPL,MSFT." flagenv:"true" flaggroup:"Query" mod:"trim"`
	Limit        int    `flag:"limit" flagdescr:"Maximum trade rows to return. Must be between 1 and 100. Defaults to the command preset when omitted." flagenv:"true" flaggroup:"Output"`
	Fields       string `flag:"fields" flagdescr:"Comma-separated trade fields to include. Overrides --preset-fields. Use upstream field names such as Ticker,Dollars,TradeRank." flagenv:"true" flaggroup:"Output" mod:"trim"`
	PresetFields string `flag:"preset-fields" flagdescr:"Field preset to include: core, expanded, or full. Defaults to core for token-efficient output. Expanded includes annotated non-internal signal fields; full returns raw upstream payloads." flagenv:"true" flaggroup:"Output" mod:"trim"`
	Shape        string `flag:"shape" flagdescr:"Trade row shape: array or objects. Array is the default and is most token-efficient." flagenv:"true" flaggroup:"Output" mod:"trim"`
	Pretty       bool   `flag:"pretty" flagdescr:"Pretty-print JSON output. Compact JSON is the default for token-efficient LLM and MCP use." flagenv:"true" flaggroup:"Output"`
}

// ClusterOptions defines the LLM-readable contract for fetching trade cluster presets.
type ClusterOptions struct {
	Date         string `flag:"date" flagshort:"d" flagdescr:"Single trading date to query, formatted as YYYY-MM-DD. Trade cluster presets are intentionally limited to one day." flagenv:"true" flagrequired:"true" flaggroup:"Query" validate:"required" mod:"trim"`
	Tickers      string `flag:"tickers" flagdescr:"Optional ticker filter. Use one symbol or a comma-delimited list without spaces, for example AAPL or AAPL,MSFT." flagenv:"true" flaggroup:"Query" mod:"trim"`
	Limit        int    `flag:"limit" flagdescr:"Maximum cluster rows to return. Must be between 1 and 100. Defaults to 100 when omitted." flagenv:"true" flaggroup:"Output"`
	Fields       string `flag:"fields" flagdescr:"Comma-separated cluster fields to include. Overrides --preset-fields. Use upstream field names such as Ticker,Dollars,TradeCount,TradeClusterRank." flagenv:"true" flaggroup:"Output" mod:"trim"`
	PresetFields string `flag:"preset-fields" flagdescr:"Field preset to include: core, expanded, or full. Defaults to core for token-efficient output. Expanded includes annotated non-internal signal fields; full returns raw upstream payloads." flagenv:"true" flaggroup:"Output" mod:"trim"`
	Shape        string `flag:"shape" flagdescr:"Cluster row shape: array or objects. Array is the default and is most token-efficient." flagenv:"true" flaggroup:"Output" mod:"trim"`
	Pretty       bool   `flag:"pretty" flagdescr:"Pretty-print JSON output. Compact JSON is the default for token-efficient LLM and MCP use." flagenv:"true" flaggroup:"Output"`
}

// LevelOptions defines the LLM-readable contract for fetching trade price levels.
type LevelOptions struct {
	Ticker          string `flag:"ticker" flagdescr:"Ticker symbol to query. Trade levels are ticker-specific, for example SPY or BAND." flagenv:"true" flagrequired:"true" flaggroup:"Query" validate:"required" mod:"trim"`
	StartDate       string `flag:"start-date" flagdescr:"Start date for level history, formatted as YYYY-MM-DD. Defaults to one year before today when omitted." flagenv:"true" flaggroup:"Query" mod:"trim"`
	EndDate         string `flag:"end-date" flagdescr:"End date for level history, formatted as YYYY-MM-DD. Defaults to today when omitted." flagenv:"true" flaggroup:"Query" mod:"trim"`
	MinDollars      string `flag:"min-dollars" flagdescr:"Minimum aggregate dollars filter sent to VolumeLeaders. Captured browser default is 500000." flagenv:"true" flaggroup:"Filters" mod:"trim"`
	MaxDollars      string `flag:"max-dollars" flagdescr:"Maximum aggregate dollars filter sent to VolumeLeaders. Captured browser default is 30000000000." flagenv:"true" flaggroup:"Filters" mod:"trim"`
	MinVolume       string `flag:"min-volume" flagdescr:"Minimum aggregate share volume filter. Captured browser default is 0." flagenv:"true" flaggroup:"Filters" mod:"trim"`
	MaxVolume       string `flag:"max-volume" flagdescr:"Maximum aggregate share volume filter. Captured browser default is 2000000000." flagenv:"true" flaggroup:"Filters" mod:"trim"`
	MinPrice        string `flag:"min-price" flagdescr:"Minimum level price filter. Captured browser default is 0." flagenv:"true" flaggroup:"Filters" mod:"trim"`
	MaxPrice        string `flag:"max-price" flagdescr:"Maximum level price filter. Captured browser default is 100000." flagenv:"true" flaggroup:"Filters" mod:"trim"`
	VCD             string `flag:"vcd" flagdescr:"Minimum CumulativeDistribution percentile filter. Captured browser default is 0." flagenv:"true" flaggroup:"Filters" mod:"trim"`
	RelativeSize    string `flag:"relative-size" flagdescr:"Minimum relative size filter. Captured values are 0, 3, 5, and 10, where 0 means any size." flagenv:"true" flaggroup:"Filters" mod:"trim"`
	TradeLevelRank  int    `flag:"trade-level-rank" flagdescr:"Rank bucket filter. Captured values are -1 for any, then 1, 3, 5, 10, 25, 50, and 100." flagenv:"true" flaggroup:"Filters"`
	TradeLevelCount int    `flag:"trade-level-count" flagdescr:"Maximum level count requested from VolumeLeaders. Captured values are 5, 10, 20, 50, and -1 for all levels." flagenv:"true" flaggroup:"Filters"`
	Fields          string `flag:"fields" flagdescr:"Comma-separated trade level fields to include. Overrides --preset-fields. Use upstream field names such as Price,Dollars,TradeLevelRank,Dates." flagenv:"true" flaggroup:"Output" mod:"trim"`
	PresetFields    string `flag:"preset-fields" flagdescr:"Field preset to include: core, expanded, or full. Defaults to core for token-efficient output. Full returns raw upstream payloads." flagenv:"true" flaggroup:"Output" mod:"trim"`
	Shape           string `flag:"shape" flagdescr:"Trade level row shape: array or objects. Array is the default and is most token-efficient." flagenv:"true" flaggroup:"Output" mod:"trim"`
	Pretty          bool   `flag:"pretty" flagdescr:"Pretty-print JSON output. Compact JSON is the default for token-efficient LLM and MCP use." flagenv:"true" flaggroup:"Output"`
}

// SignalOptions defines the LLM-readable contract for fetching trade signal presets.
type SignalOptions struct {
	Date         string `flag:"date" flagshort:"d" flagdescr:"Single trading date to query, formatted as YYYY-MM-DD. Trade signal presets are intentionally limited to one day." flagenv:"true" flagrequired:"true" flaggroup:"Query" validate:"required" mod:"trim"`
	Tickers      string `flag:"tickers" flagdescr:"Optional ticker filter. Use one symbol or a comma-delimited list without spaces, for example AAPL or AAPL,MSFT." flagenv:"true" flaggroup:"Query" mod:"trim"`
	Limit        int    `flag:"limit" flagdescr:"Maximum trade rows to return. Must be between 1 and 100. Defaults to 100 when omitted." flagenv:"true" flaggroup:"Output"`
	Fields       string `flag:"fields" flagdescr:"Comma-separated trade fields to include. Overrides --preset-fields. Use upstream field names such as Ticker,Dollars,TradeRank." flagenv:"true" flaggroup:"Output" mod:"trim"`
	PresetFields string `flag:"preset-fields" flagdescr:"Field preset to include: core, expanded, or full. Defaults to core for token-efficient output. Expanded includes annotated non-internal signal fields; full returns raw upstream payloads." flagenv:"true" flaggroup:"Output" mod:"trim"`
	Shape        string `flag:"shape" flagdescr:"Trade row shape: array or objects. Array is the default and is most token-efficient." flagenv:"true" flaggroup:"Output" mod:"trim"`
	Pretty       bool   `flag:"pretty" flagdescr:"Pretty-print JSON output. Compact JSON is the default for token-efficient LLM and MCP use." flagenv:"true" flaggroup:"Output"`
}

// Result is the stable response shape for the unusual trades command.
type Result struct {
	Status          string              `json:"status"`
	Date            string              `json:"date"`
	RecordsTotal    int                 `json:"recordsTotal"`
	RecordsFiltered int                 `json:"recordsFiltered"`
	Fields          []string            `json:"fields,omitempty"`
	Rows            [][]json.RawMessage `json:"rows,omitempty"`
	Trades          []json.RawMessage   `json:"trades,omitempty"`
}

// ClusterResult is the stable response shape for trade cluster presets.
type ClusterResult struct {
	Status          string              `json:"status"`
	Date            string              `json:"date"`
	RankLimit       int                 `json:"rankLimit,omitempty"`
	RecordsTotal    int                 `json:"recordsTotal"`
	RecordsFiltered int                 `json:"recordsFiltered"`
	Fields          []string            `json:"fields,omitempty"`
	Rows            [][]json.RawMessage `json:"rows,omitempty"`
	Clusters        []json.RawMessage   `json:"clusters,omitempty"`
}

// LevelResult is the stable response shape for the trade levels command.
type LevelResult struct {
	Status          string              `json:"status"`
	Ticker          string              `json:"ticker"`
	StartDate       string              `json:"startDate"`
	EndDate         string              `json:"endDate"`
	RecordsTotal    int                 `json:"recordsTotal"`
	RecordsFiltered int                 `json:"recordsFiltered"`
	Fields          []string            `json:"fields,omitempty"`
	Rows            [][]json.RawMessage `json:"rows,omitempty"`
	Levels          []json.RawMessage   `json:"levels,omitempty"`
}

// RankedResult is the stable response shape for all-time ranked trade presets.
type RankedResult struct {
	Status          string              `json:"status"`
	Date            string              `json:"date"`
	RankLimit       int                 `json:"rankLimit"`
	RecordsTotal    int                 `json:"recordsTotal"`
	RecordsFiltered int                 `json:"recordsFiltered"`
	Fields          []string            `json:"fields,omitempty"`
	Rows            [][]json.RawMessage `json:"rows,omitempty"`
	Trades          []json.RawMessage   `json:"trades,omitempty"`
}

type getTradesResponse struct {
	Draw            int               `json:"draw"`
	RecordsTotal    int               `json:"recordsTotal"`
	RecordsFiltered int               `json:"recordsFiltered"`
	Data            []json.RawMessage `json:"data"`
	Error           string            `json:"error"`
}

type tradeColumn struct {
	data      string
	name      string
	orderable string
}

type getTradeClustersRequestOptions struct {
	tradeClusterRank       int
	draw                   int
	start                  int
	length                 int
	minVolume              string
	maxDollars             string
	conditions             string
	vcd                    string
	relativeSize           string
	darkPools              string
	sweeps                 string
	latePrints             string
	signaturePrints        string
	evenShared             string
	securityTypeKey        string
	marketCap              string
	includePremarket       string
	includeRTH             string
	includeAH              string
	includeOpening         string
	includeClosing         string
	includePhantom         string
	includeOffsetting      string
	sectorIndustry         string
	presetSearchTemplateID string
}

type getTradesRequestOptions struct {
	tradeRank              int
	draw                   int
	start                  int
	length                 int
	minVolume              string
	maxDollars             string
	conditions             string
	vcd                    string
	relativeSize           string
	darkPools              string
	sweeps                 string
	latePrints             string
	signaturePrints        string
	evenShared             string
	securityTypeKey        string
	tradeRankSnapshot      string
	marketCap              string
	includePremarket       string
	includeRTH             string
	includeAH              string
	includeOpening         string
	includeClosing         string
	includePhantom         string
	includeOffsetting      string
	sectorIndustry         string
	presetSearchTemplateID string
}

type getTradeLevelsRequestOptions struct {
	draw            int
	start           int
	length          int
	minVolume       string
	maxVolume       string
	minPrice        string
	maxPrice        string
	minDollars      string
	maxDollars      string
	vcd             string
	relativeSize    string
	tradeLevelRank  int
	tradeLevelCount int
}

type commandMetadata struct {
	use     string
	aliases []string
	short   string
	long    string
	example string
}

type standardRunConfig struct {
	cmdName       string
	date          string
	tickers       string
	limit         int
	fields        string
	presetFields  string
	shape         string
	pretty        bool
	fetchResponse func(context.Context, string, string, int) (getTradesResponse, error)
}

type tradeOutput struct {
	fields []string
	rows   [][]json.RawMessage
	trades []json.RawMessage
}

type rankedPreset struct {
	use               string
	aliases           []string
	short             string
	long              string
	example           string
	rank              int
	length            int
	minVolume         string
	maxDollars        string
	conditions        string
	vcd               string
	relativeSize      string
	darkPools         string
	sweeps            string
	latePrints        string
	signaturePrints   string
	evenShared        string
	securityTypeKey   string
	tradeRankSnapshot string
	marketCap         string
	includePremarket  string
	includeRTH        string
	includeAH         string
	includeOpening    string
	includeClosing    string
	includePhantom    string
	includeOffsetting string
	sectorIndustry    string
	presetID          string
}

type signalPreset struct {
	use        string
	aliases    []string
	short      string
	long       string
	example    string
	phantom    string
	offsetting string
	darkPools  string
	presetID   string
}

type conditionPreset struct {
	use        string
	aliases    []string
	short      string
	long       string
	example    string
	conditions string
	presetID   string
}

type clusterPreset struct {
	use               string
	aliases           []string
	short             string
	long              string
	example           string
	tradeClusterRank  int
	length            int
	minVolume         string
	maxDollars        string
	conditions        string
	vcd               string
	relativeSize      string
	darkPools         string
	sweeps            string
	latePrints        string
	signaturePrints   string
	evenShared        string
	securityTypeKey   string
	marketCap         string
	includePremarket  string
	includeRTH        string
	includeAH         string
	includeOpening    string
	includeClosing    string
	includePhantom    string
	includeOffsetting string
	sectorIndustry    string
	presetID          string
}

type tradeRequestConfig struct {
	operation   string
	endpoint    string
	form        url.Values
	setHeaders  func(*http.Request, string)
	afterDecode func(*getTradesResponse)
}

var getTradeClustersColumns = []tradeColumn{
	{data: "MinFullTimeString24", name: "", orderable: "false"},
	{data: "MinFullTimeString24", name: "MinFullTimeString24", orderable: "true"},
	{data: "Ticker", name: "Ticker", orderable: "true"},
	{data: "TradeCount", name: "TradeCount", orderable: "true"},
	{data: "Current", name: "Current", orderable: "false"},
	{data: "Cluster", name: "Cluster", orderable: "false"},
	{data: "Sector", name: "Sector", orderable: "true"},
	{data: "Industry", name: "Industry", orderable: "true"},
	{data: "Volume", name: "Sh", orderable: "true"},
	{data: "Dollars", name: "$$", orderable: "true"},
	{data: "DollarsMultiplier", name: "RS", orderable: "true"},
	{data: "CumulativeDistribution", name: "PCT", orderable: "true"},
	{data: "TradeClusterRank", name: "Rank", orderable: "true"},
	{data: "LastComparibleTradeClusterDate", name: "Last Date", orderable: "true"},
	{data: "LastComparibleTradeClusterDate", name: "Charts", orderable: "false"},
}

var getTradesColumns = []tradeColumn{
	{data: "FullTimeString24", name: "", orderable: "false"},
	{data: "FullTimeString24", name: "FullTimeString24", orderable: "true"},
	{data: "Ticker", name: "Ticker", orderable: "true"},
	{data: "Current", name: "Current", orderable: "false"},
	{data: "Trade", name: "Trade", orderable: "false"},
	{data: "Sector", name: "Sector", orderable: "true"},
	{data: "Industry", name: "Industry", orderable: "true"},
	{data: "Volume", name: "Sh", orderable: "true"},
	{data: "Dollars", name: "$$", orderable: "true"},
	{data: "DollarsMultiplier", name: "RS", orderable: "true"},
	{data: "CumulativeDistribution", name: "PCT", orderable: "true"},
	{data: "TradeRank", name: "Rank", orderable: "true"},
	{data: "LastComparibleTradeDate", name: "Last Traded", orderable: "true"},
	{data: "LastComparibleTradeDate", name: "Charts", orderable: "false"},
}

var getTradeLevelsColumns = []tradeColumn{
	{data: "Price", name: "Price", orderable: "true"},
	{data: "Dollars", name: "$$", orderable: "true"},
	{data: "Volume", name: "Shares", orderable: "true"},
	{data: "Trades", name: "Trades", orderable: "true"},
	{data: "RelativeSize", name: "RS", orderable: "true"},
	{data: "CumulativeDistribution", name: "PCT", orderable: "true"},
	{data: "TradeLevelRank", name: "Level Rank", orderable: "true"},
	{data: "Level Date Range", name: "Level Date Range", orderable: "false"},
}

// NewCommand builds the large unusual trades command.
func NewCommand() (*cobra.Command, error) {
	opts := &Options{}
	return newBoundTradeCommand(&commandMetadata{
		use:     "trades",
		aliases: []string{"large-trades", "unusual-trades"},
		short:   "Fetch disproportionately large trades for one date",
		long:    "Fetch VolumeLeaders disproportionately large trades for a single trading day. This reproduces the default Disproportionately large trades GetTrades request and intentionally does not allow multi-day ranges.",
		example: "volumeleaders-agent trades --date 2026-04-30\nvolumeleaders-agent trades --date 2026-04-30 --tickers AAPL,MSFT",
	}, opts, &opts.Tickers, func(cmd *cobra.Command, _ []string) error {
		return run(cmd.Context(), cmd, opts)
	})
}

func newBoundTradeCommand(meta *commandMetadata, opts any, tickerFlag *string, runE func(*cobra.Command, []string) error) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     meta.use,
		Aliases: meta.aliases,
		Short:   meta.short,
		Long:    tradeCommandLong(meta.long),
		Example: meta.example,
		RunE:    runE,
	}

	if err := structcli.Bind(cmd, opts); err != nil {
		return nil, fmt.Errorf("bind %s options: %w", meta.use, err)
	}
	if tickerFlag != nil {
		cmd.Flags().StringVar(tickerFlag, "ticker", "", "Optional ticker filter. Alias for --tickers.")
	}

	return cmd, nil
}

func tradeCommandLong(base string) string {
	return base + "\n\n" + tradeLLMFieldGuide + " " + calendarEventGuide + " " + expandedPresetGuide + " " + tradeLevelGuide
}

// NewTradeClustersCommand builds the disproportionately large trade clusters command.
func NewTradeClustersCommand() (*cobra.Command, error) {
	return newClusterCommand(&clusterPreset{
		use:     "trade-clusters",
		aliases: []string{"clusters", "large-clusters"},
		short:   "Fetch disproportionately large trade clusters for one date",
		long:    "Fetch VolumeLeaders disproportionately large trade clusters for a single trading day. Trade clusters aggregate many smaller trades close together in time into larger dollar-volume events.",
		example: "volumeleaders-agent trade-clusters --date 2026-04-30\nvolumeleaders-agent trade-clusters --date 2026-04-30 --tickers AAPL,MSFT",
	})
}

// NewTradeLevelsCommand builds the trade price levels command.
func NewTradeLevelsCommand() (*cobra.Command, error) {
	opts := &LevelOptions{}
	return newBoundTradeCommand(&commandMetadata{
		use:     "trade-levels",
		aliases: []string{"levels", "price-levels"},
		short:   "Fetch large-trade price levels for one ticker",
		long:    "Fetch VolumeLeaders trade levels for one ticker across a date range. Trade levels aggregate large prints by price and expose the level's dollars, shares, relative size, percentile, rank, and date range.",
		example: "volumeleaders-agent trade-levels --ticker BAND\nvolumeleaders-agent trade-levels --ticker SPY --start-date 2025-05-01 --end-date 2026-05-01 --trade-level-count 10",
	}, opts, nil, func(cmd *cobra.Command, _ []string) error {
		return runTradeLevels(cmd.Context(), cmd, opts)
	})
}

// NewTop10ClustersCommand builds the top 10 all-time ranked trade clusters command.
func NewTop10ClustersCommand() (*cobra.Command, error) {
	return newClusterCommand(&clusterPreset{
		use:              "top10-clusters",
		aliases:          []string{"top-10-clusters", "cluster-top10", "cluster-top-10"},
		short:            "Fetch trade clusters ranked in a stock's all-time top 10",
		long:             "Fetch VolumeLeaders trade clusters for one day where each cluster ranks in that stock's all-time top 10 clusters.",
		example:          "volumeleaders-agent top10-clusters --date 2026-04-30\nvolumeleaders-agent top10-clusters --date 2026-04-30 --tickers AAPL,MSFT",
		tradeClusterRank: 10,
		length:           10,
		minVolume:        "10000",
		maxDollars:       "100000000000",
		presetID:         "623",
	})
}

// NewTop100ClustersCommand builds the top 100 all-time ranked trade clusters command.
func NewTop100ClustersCommand() (*cobra.Command, error) {
	return newClusterCommand(&clusterPreset{
		use:              "top100-clusters",
		aliases:          []string{"top-100-clusters", "cluster-top100", "cluster-top-100"},
		short:            "Fetch trade clusters ranked in a stock's all-time top 100",
		long:             "Fetch VolumeLeaders trade clusters for one day where each cluster ranks in that stock's all-time top 100 clusters.",
		example:          "volumeleaders-agent top100-clusters --date 2026-04-30\nvolumeleaders-agent top100-clusters --date 2026-04-30 --tickers AAPL,MSFT",
		tradeClusterRank: 100,
		length:           100,
		minVolume:        "10000",
		maxDollars:       "100000000000",
		presetID:         "568",
	})
}

// NewTop10Command builds the top 10 all-time ranked trades command.
func NewTop10Command() (*cobra.Command, error) {
	return newRankedCommand(&rankedPreset{
		use:      "top10",
		aliases:  []string{"top-10", "ranked-top10", "ranked-top-10"},
		short:    "Fetch trades ranked in a stock's all-time top 10",
		long:     "Fetch VolumeLeaders trades for one day where each trade ranks in that stock's all-time top 10 single trades.",
		example:  "volumeleaders-agent top10 --date 2026-04-30\nvolumeleaders-agent top10 --date 2026-04-30 --tickers AAPL,MSFT",
		rank:     10,
		length:   10,
		presetID: "623",
	})
}

// NewTop100Command builds the top 100 all-time ranked trades command.
func NewTop100Command() (*cobra.Command, error) {
	return newRankedCommand(&rankedPreset{
		use:      "top100",
		aliases:  []string{"top-100", "ranked-top100", "ranked-top-100"},
		short:    "Fetch trades ranked in a stock's all-time top 100",
		long:     "Fetch VolumeLeaders trades for one day where each trade ranks in that stock's all-time top 100 single trades.",
		example:  "volumeleaders-agent top100 --date 2026-04-30\nvolumeleaders-agent top100 --date 2026-04-30 --tickers AAPL,MSFT",
		rank:     100,
		length:   100,
		presetID: "568",
	})
}

// NewTop3010x99PctCommand builds the top 30, 10x relative size, 99th percentile trades command.
func NewTop3010x99PctCommand() (*cobra.Command, error) {
	return newRankedCommand(top3010x99PctTradePreset())
}

// NewTop3010x99PctClustersCommand builds the cluster equivalent of the top 30, 10x relative size, 99th percentile command.
func NewTop3010x99PctClustersCommand() (*cobra.Command, error) {
	return newClusterCommand(top3010x99PctClusterPreset())
}

// NewTop100DarkPool20xCommand builds the top 100, 20x relative size, dark-pool trades command.
func NewTop100DarkPool20xCommand() (*cobra.Command, error) {
	return newRankedCommand(top100DarkPool20xTradePreset())
}

// NewTop100DarkPool20xClustersCommand builds the cluster equivalent of the top 100, 20x relative size, dark-pool command.
func NewTop100DarkPool20xClustersCommand() (*cobra.Command, error) {
	return newClusterCommand(top100DarkPool20xClusterPreset())
}

// NewTop100LeveragedETFsCommand builds the top 100 leveraged ETF trades command.
func NewTop100LeveragedETFsCommand() (*cobra.Command, error) {
	return newRankedCommand(top100LeveragedETFsTradePreset())
}

// NewTop100LeveragedETFsClustersCommand builds the cluster equivalent of the top 100 leveraged ETF command.
func NewTop100LeveragedETFsClustersCommand() (*cobra.Command, error) {
	return newClusterCommand(top100LeveragedETFsClusterPreset())
}

// NewTop100DarkPoolSweepsCommand builds the top 100 dark-pool sweep trades command.
func NewTop100DarkPoolSweepsCommand() (*cobra.Command, error) {
	return newRankedCommand(top100DarkPoolSweepsTradePreset())
}

// NewTop100DarkPoolSweepsClustersCommand builds the cluster equivalent of the top 100 dark-pool sweep command.
func NewTop100DarkPoolSweepsClustersCommand() (*cobra.Command, error) {
	return newClusterCommand(top100DarkPoolSweepsClusterPreset())
}

// NewPhantomCommand builds the phantom trades command.
func NewPhantomCommand() (*cobra.Command, error) {
	return newSignalCommand(&signalPreset{
		use:        "phantom",
		aliases:    []string{"phantom-trades"},
		short:      "Fetch phantom trades far from the current price",
		long:       "Fetch VolumeLeaders phantom trades for one day. Phantom trades are prints where the trade price is far from the current price and may hint at future price magnets, though they are not guaranteed signals.",
		example:    "volumeleaders-agent phantom --date 2026-04-30\nvolumeleaders-agent phantom --date 2026-04-30 --tickers AAPL,MSFT",
		phantom:    "1",
		offsetting: "0",
		darkPools:  "1",
		presetID:   "857",
	})
}

// NewOffsettingCommand builds the offsetting trades command.
func NewOffsettingCommand() (*cobra.Command, error) {
	return newSignalCommand(&signalPreset{
		use:        "offsetting",
		aliases:    []string{"offsetting-trades"},
		short:      "Fetch trades with nearly matching offsetting prints",
		long:       "Fetch VolumeLeaders offsetting trades for one day. Offsetting trades show prints with nearly matching share sizes across dates, which can hint that a trader entered and later exited a position.",
		example:    "volumeleaders-agent offsetting --date 2026-04-30\nvolumeleaders-agent offsetting --date 2026-04-30 --tickers AAPL,MSFT",
		phantom:    "0",
		offsetting: "1",
		darkPools:  "-1",
		presetID:   "858",
	})
}

// NewOverboughtCommand builds the RSI overbought trades command.
func NewOverboughtCommand() (*cobra.Command, error) {
	return newConditionCommand(&conditionPreset{
		use:        "overbought",
		aliases:    []string{"overbought-trades", "rsi-overbought"},
		short:      "Fetch trades with overbought daily or hourly RSI conditions",
		long:       "Fetch VolumeLeaders trades for one day where the captured RSI condition filter requires daily or hourly overbought matches.",
		example:    "volumeleaders-agent overbought --date 2026-04-30\nvolumeleaders-agent overbought --date 2026-04-30 --tickers AAPL,MSFT",
		conditions: "OBD,OBH,",
		presetID:   "84",
	})
}

// NewOversoldCommand builds the RSI oversold trades command.
func NewOversoldCommand() (*cobra.Command, error) {
	return newConditionCommand(&conditionPreset{
		use:        "oversold",
		aliases:    []string{"oversold-trades", "rsi-oversold"},
		short:      "Fetch trades with oversold daily or hourly RSI conditions",
		long:       "Fetch VolumeLeaders trades for one day where the captured RSI condition filter requires daily or hourly oversold matches.",
		example:    "volumeleaders-agent oversold --date 2026-04-30\nvolumeleaders-agent oversold --date 2026-04-30 --tickers AAPL,MSFT",
		conditions: "OSD,OSH",
		presetID:   "85",
	})
}

// NewOverboughtClustersCommand builds the RSI overbought trade clusters command.
func NewOverboughtClustersCommand() (*cobra.Command, error) {
	return newClusterCommand(overboughtClusterPreset())
}

// NewOversoldClustersCommand builds the RSI oversold trade clusters command.
func NewOversoldClustersCommand() (*cobra.Command, error) {
	return newClusterCommand(oversoldClusterPreset())
}

func top3010x99PctTradePreset() *rankedPreset {
	return &rankedPreset{
		use:               "top30-10x-99pct",
		aliases:           []string{"top30-rank-over-10x-relative-size-in-99th-percentile"},
		short:             "Fetch top 30 trades at least 10x relative size in the 99th percentile",
		long:              "Fetch VolumeLeaders trades for one day where each trade ranks in the stock's all-time top 30, is at least 10x relative size, and is in the 99th CumulativeDistribution percentile.",
		example:           "volumeleaders-agent top30-10x-99pct --date 2026-04-30\nvolumeleaders-agent top30-10x-99pct --date 2026-04-30 --tickers AAPL,MSFT",
		rank:              30,
		length:            100,
		minVolume:         "10000",
		maxDollars:        "10000000000",
		conditions:        ignoredRSIConditions,
		vcd:               "99",
		relativeSize:      "10",
		darkPools:         "-1",
		signaturePrints:   "0",
		includePhantom:    "-1",
		includeOffsetting: "-1",
		presetID:          "212",
	}
}

func top3010x99PctClusterPreset() *clusterPreset {
	return tradePresetToClusterPreset(top3010x99PctTradePreset(), "top30-10x-99pct-clusters", []string{"top30-rank-over-10x-relative-size-in-99th-percentile-clusters"}, "Fetch top 30 trade clusters at least 10x relative size in the 99th percentile", "Fetch VolumeLeaders trade clusters for one day where each cluster ranks in the stock's all-time top 30, is at least 10x relative size, and is in the 99th CumulativeDistribution percentile.", "volumeleaders-agent top30-10x-99pct-clusters --date 2026-04-30\nvolumeleaders-agent top30-10x-99pct-clusters --date 2026-04-30 --tickers AAPL,MSFT")
}

func top100DarkPool20xTradePreset() *rankedPreset {
	return &rankedPreset{
		use:               "top100-dark-pool-20x",
		aliases:           []string{"top100-over-20x-relative-size-dark-pool-trades-only"},
		short:             "Fetch top 100 dark-pool trades at least 20x relative size",
		long:              "Fetch VolumeLeaders dark-pool trades for one day where each trade ranks in the stock's all-time top 100 and is at least 20x relative size. This preset sends DarkPools=1 and does not filter on Sweeps, so it includes dark-pool blocks and dark-pool sweeps.",
		example:           "volumeleaders-agent top100-dark-pool-20x --date 2026-04-30\nvolumeleaders-agent top100-dark-pool-20x --date 2026-04-30 --tickers AAPL,MSFT",
		rank:              100,
		length:            100,
		minVolume:         "10000",
		maxDollars:        "10000000000",
		conditions:        ignoredRSIConditions,
		vcd:               "0",
		relativeSize:      "20",
		darkPools:         "1",
		signaturePrints:   "0",
		includePhantom:    "-1",
		includeOffsetting: "-1",
		presetID:          "183",
	}
}

func top100DarkPool20xClusterPreset() *clusterPreset {
	return tradePresetToClusterPreset(top100DarkPool20xTradePreset(), "top100-dark-pool-20x-clusters", []string{"top100-over-20x-relative-size-dark-pool-trades-only-clusters"}, "Fetch top 100 dark-pool trade clusters at least 20x relative size", "Fetch VolumeLeaders dark-pool trade clusters for one day where each cluster ranks in the stock's all-time top 100 and is at least 20x relative size. This preset sends DarkPools=1 and does not filter on Sweeps, so it includes dark-pool block clusters and dark-pool sweep clusters.", "volumeleaders-agent top100-dark-pool-20x-clusters --date 2026-04-30\nvolumeleaders-agent top100-dark-pool-20x-clusters --date 2026-04-30 --tickers AAPL,MSFT")
}

func top100LeveragedETFsTradePreset() *rankedPreset {
	return &rankedPreset{
		use:               "top100-leveraged-etfs",
		aliases:           []string{"top100-leveraged-etfs-only"},
		short:             "Fetch top 100 leveraged ETF trades",
		long:              "Fetch VolumeLeaders leveraged ETF trades for one day where each trade ranks in the ticker's all-time top 100 trades.",
		example:           "volumeleaders-agent top100-leveraged-etfs --date 2026-04-30\nvolumeleaders-agent top100-leveraged-etfs --date 2026-04-30 --tickers TQQQ,SQQQ",
		rank:              100,
		length:            100,
		minVolume:         "10000",
		maxDollars:        "1000000000000",
		conditions:        ignoredRSIConditions,
		vcd:               "0",
		relativeSize:      "0",
		darkPools:         "-1",
		signaturePrints:   "-1",
		includePhantom:    "-1",
		includeOffsetting: "-1",
		sectorIndustry:    "X B",
		presetID:          "4724",
	}
}

func top100LeveragedETFsClusterPreset() *clusterPreset {
	return tradePresetToClusterPreset(top100LeveragedETFsTradePreset(), "top100-leveraged-etfs-clusters", []string{"top100-leveraged-etfs-only-clusters"}, "Fetch top 100 leveraged ETF trade clusters", "Fetch VolumeLeaders leveraged ETF trade clusters for one day where each cluster ranks in the ticker's all-time top 100 clusters.", "volumeleaders-agent top100-leveraged-etfs-clusters --date 2026-04-30\nvolumeleaders-agent top100-leveraged-etfs-clusters --date 2026-04-30 --tickers TQQQ,SQQQ")
}

func top100DarkPoolSweepsTradePreset() *rankedPreset {
	return &rankedPreset{
		use:               "top100-dark-pool-sweeps",
		short:             "Fetch top 100 dark-pool sweep trades",
		long:              "Fetch VolumeLeaders dark-pool sweep trades for one day where each trade ranks in the stock's all-time top 100. This preset sends DarkPools=1 and Sweeps=1, so it returns only trades that are both dark-pool prints and sweeps. It includes premarket and regular-hours trades, while excluding after-hours, opening, closing, and phantom prints as captured from the browser.",
		example:           "volumeleaders-agent top100-dark-pool-sweeps --date 2026-04-30\nvolumeleaders-agent top100-dark-pool-sweeps --date 2026-04-30 --tickers AAPL,MSFT",
		rank:              100,
		length:            100,
		minVolume:         "10000",
		maxDollars:        "100000000000",
		conditions:        ignoredRSIConditions,
		vcd:               "0",
		relativeSize:      "0",
		darkPools:         "1",
		sweeps:            "1",
		signaturePrints:   "0",
		includePremarket:  "1",
		includeRTH:        "1",
		includeAH:         "0",
		includeOpening:    "0",
		includeClosing:    "0",
		includePhantom:    "0",
		includeOffsetting: "-1",
		presetID:          "2163",
	}
}

func top100DarkPoolSweepsClusterPreset() *clusterPreset {
	return tradePresetToClusterPreset(top100DarkPoolSweepsTradePreset(), "top100-dark-pool-sweeps-clusters", nil, "Fetch top 100 dark-pool sweep trade clusters", "Fetch VolumeLeaders dark-pool sweep trade clusters for one day where each cluster ranks in the stock's all-time top 100. This preset sends DarkPools=1 and Sweeps=1, so it returns only clusters that are both dark-pool prints and sweeps. It includes premarket and regular-hours clusters, while excluding after-hours, opening, closing, and phantom prints as captured from the browser.", "volumeleaders-agent top100-dark-pool-sweeps-clusters --date 2026-04-30\nvolumeleaders-agent top100-dark-pool-sweeps-clusters --date 2026-04-30 --tickers AAPL,MSFT")
}

func overboughtClusterPreset() *clusterPreset {
	return conditionClusterPreset("overbought-clusters", []string{"rsi-overbought-clusters"}, "Fetch trade clusters with overbought daily or hourly RSI conditions", "Fetch VolumeLeaders trade clusters for one day where the captured RSI condition filter requires daily or hourly overbought matches.", "volumeleaders-agent overbought-clusters --date 2026-04-30\nvolumeleaders-agent overbought-clusters --date 2026-04-30 --tickers AAPL,MSFT", "OBD,OBH,", "84")
}

func oversoldClusterPreset() *clusterPreset {
	return conditionClusterPreset("oversold-clusters", []string{"rsi-oversold-clusters"}, "Fetch trade clusters with oversold daily or hourly RSI conditions", "Fetch VolumeLeaders trade clusters for one day where the captured RSI condition filter requires daily or hourly oversold matches.", "volumeleaders-agent oversold-clusters --date 2026-04-30\nvolumeleaders-agent oversold-clusters --date 2026-04-30 --tickers AAPL,MSFT", "OSD,OSH", "85")
}

func tradePresetToClusterPreset(preset *rankedPreset, use string, aliases []string, short, long, example string) *clusterPreset {
	return &clusterPreset{
		use:               use,
		aliases:           aliases,
		short:             short,
		long:              long,
		example:           example,
		tradeClusterRank:  preset.rank,
		length:            preset.length,
		minVolume:         preset.minVolume,
		maxDollars:        preset.maxDollars,
		conditions:        preset.conditions,
		vcd:               preset.vcd,
		relativeSize:      preset.relativeSize,
		darkPools:         preset.darkPools,
		sweeps:            preset.sweeps,
		latePrints:        preset.latePrints,
		signaturePrints:   preset.signaturePrints,
		evenShared:        preset.evenShared,
		securityTypeKey:   preset.securityTypeKey,
		marketCap:         preset.marketCap,
		includePremarket:  preset.includePremarket,
		includeRTH:        preset.includeRTH,
		includeAH:         preset.includeAH,
		includeOpening:    preset.includeOpening,
		includeClosing:    preset.includeClosing,
		includePhantom:    preset.includePhantom,
		includeOffsetting: preset.includeOffsetting,
		sectorIndustry:    preset.sectorIndustry,
		presetID:          preset.presetID,
	}
}

func conditionClusterPreset(use string, aliases []string, short, long, example, conditions, presetID string) *clusterPreset {
	return &clusterPreset{
		use:               use,
		aliases:           aliases,
		short:             short,
		long:              long,
		example:           example,
		tradeClusterRank:  100,
		length:            defaultTradeLimit,
		minVolume:         "10000",
		maxDollars:        "10000000000",
		conditions:        conditions,
		vcd:               "0",
		relativeSize:      "5",
		darkPools:         "-1",
		signaturePrints:   "0",
		includePhantom:    "-1",
		includeOffsetting: "-1",
		presetID:          presetID,
	}
}

func newRankedCommand(preset *rankedPreset) (*cobra.Command, error) {
	opts := &RankedOptions{}
	return newBoundTradeCommand(&commandMetadata{
		use:     preset.use,
		aliases: preset.aliases,
		short:   preset.short,
		long:    preset.long,
		example: preset.example,
	}, opts, &opts.Tickers, func(cmd *cobra.Command, _ []string) error {
		return runRanked(cmd.Context(), cmd, opts, preset)
	})
}

func newSignalCommand(preset *signalPreset) (*cobra.Command, error) {
	opts := &SignalOptions{}
	return newBoundTradeCommand(&commandMetadata{
		use:     preset.use,
		aliases: preset.aliases,
		short:   preset.short,
		long:    preset.long,
		example: preset.example,
	}, opts, &opts.Tickers, func(cmd *cobra.Command, _ []string) error {
		return runSignal(cmd.Context(), cmd, opts, preset)
	})
}

func newConditionCommand(preset *conditionPreset) (*cobra.Command, error) {
	opts := &SignalOptions{}
	return newBoundTradeCommand(&commandMetadata{
		use:     preset.use,
		aliases: preset.aliases,
		short:   preset.short,
		long:    preset.long,
		example: preset.example,
	}, opts, &opts.Tickers, func(cmd *cobra.Command, _ []string) error {
		return runCondition(cmd.Context(), cmd, opts, preset)
	})
}

func newClusterCommand(preset *clusterPreset) (*cobra.Command, error) {
	opts := &ClusterOptions{}
	return newBoundTradeCommand(&commandMetadata{
		use:     preset.use,
		aliases: preset.aliases,
		short:   preset.short,
		long:    preset.long,
		example: preset.example,
	}, opts, &opts.Tickers, func(cmd *cobra.Command, _ []string) error {
		return runClusterPreset(cmd.Context(), cmd, opts, preset)
	})
}

func run(ctx context.Context, cmd *cobra.Command, opts *Options) error {
	return runStandardTrades(ctx, cmd, &standardRunConfig{
		cmdName:      "trades",
		date:         opts.Date,
		tickers:      opts.Tickers,
		limit:        opts.Limit,
		fields:       opts.Fields,
		presetFields: opts.PresetFields,
		shape:        opts.Shape,
		pretty:       opts.Pretty,
		fetchResponse: func(ctx context.Context, formattedDate, tickers string, limit int) (getTradesResponse, error) {
			return fetchDisproportionatelyLargeTradesWithFilters(ctx, formattedDate, tickers, boolToTradeFilter(opts.DarkPools), boolToTradeFilter(opts.Sweeps), limit)
		},
	})
}

func runTradeLevels(ctx context.Context, cmd *cobra.Command, opts *LevelOptions) error {
	startDate, endDate, err := normalizeDateRange(opts.StartDate, opts.EndDate)
	if err != nil {
		return err
	}
	ticker, err := normalizeRequiredTicker(opts.Ticker)
	if err != nil {
		return err
	}
	options := defaultGetTradeLevelsRequestOptions()
	applyTradeLevelOptions(&options, opts)
	fields, shape, err := normalizeTradeLevelOutputOptions(opts.Fields, opts.PresetFields, opts.Shape)
	if err != nil {
		return err
	}

	apiResponse, err := fetchTradeLevels(ctx, startDate, endDate, ticker, &options)
	if err != nil {
		return err
	}

	result := LevelResult{
		Status:          "ok",
		Ticker:          ticker,
		StartDate:       startDate,
		EndDate:         endDate,
		RecordsTotal:    apiResponse.RecordsTotal,
		RecordsFiltered: apiResponse.RecordsFiltered,
	}
	if err := applyTradeLevelOutput(&result, apiResponse.Data, fields, shape); err != nil {
		return err
	}

	return encodeResult(cmd.OutOrStdout(), "trade-levels", result, opts.Pretty)
}

func boolToTradeFilter(enabled bool) string {
	if enabled {
		return "1"
	}
	return "-1"
}

func runClusterPreset(ctx context.Context, cmd *cobra.Command, opts *ClusterOptions, preset *clusterPreset) error {
	formattedDate, tickers, err := parseDateAndTickers(ctx, preset.use, opts.Date, opts.Tickers)
	if err != nil {
		return err
	}
	limit, err := normalizeLimit(preset.use, opts.Limit, clusterPresetDefaultLimit(preset), cmd.Flags().Changed("limit"))
	if err != nil {
		return err
	}
	fields, shape, err := normalizeClusterOutputOptions(opts.Fields, opts.PresetFields, opts.Shape)
	if err != nil {
		return err
	}

	apiResponse, err := fetchClusterPreset(ctx, formattedDate, tickers, preset, limit)
	if err != nil {
		return err
	}

	result := ClusterResult{
		Status:          "ok",
		Date:            formattedDate,
		RecordsTotal:    apiResponse.RecordsTotal,
		RecordsFiltered: apiResponse.RecordsFiltered,
	}
	if preset.tradeClusterRank > 0 {
		result.RankLimit = preset.tradeClusterRank
	}
	if err := applyClusterOutput(&result, apiResponse.Data, fields, shape); err != nil {
		return err
	}

	return encodeResult(cmd.OutOrStdout(), preset.use, result, opts.Pretty)
}

func runRanked(ctx context.Context, cmd *cobra.Command, opts *RankedOptions, preset *rankedPreset) error {
	formattedDate, tickers, err := parseDateAndTickers(ctx, preset.use, opts.Date, opts.Tickers)
	if err != nil {
		return err
	}
	limit, err := normalizeLimit(preset.use, opts.Limit, preset.length, cmd.Flags().Changed("limit"))
	if err != nil {
		return err
	}
	fields, shape, err := normalizeOutputOptions(opts.Fields, opts.PresetFields, opts.Shape)
	if err != nil {
		return err
	}

	apiResponse, err := fetchRankedTrades(ctx, formattedDate, tickers, preset, limit)
	if err != nil {
		return err
	}

	result := RankedResult{
		Status:          "ok",
		Date:            formattedDate,
		RankLimit:       preset.rank,
		RecordsTotal:    apiResponse.RecordsTotal,
		RecordsFiltered: apiResponse.RecordsFiltered,
	}
	if err := applyRankedTradeOutput(&result, apiResponse.Data, fields, shape); err != nil {
		return err
	}

	return encodeResult(cmd.OutOrStdout(), preset.use, result, opts.Pretty)
}

func runSignal(ctx context.Context, cmd *cobra.Command, opts *SignalOptions, preset *signalPreset) error {
	return runSignalPreset(ctx, cmd, opts, preset.use, func(ctx context.Context, formattedDate, tickers string, limit int) (getTradesResponse, error) {
		return fetchSignalTrades(ctx, formattedDate, tickers, preset, limit)
	})
}

func runCondition(ctx context.Context, cmd *cobra.Command, opts *SignalOptions, preset *conditionPreset) error {
	return runSignalPreset(ctx, cmd, opts, preset.use, func(ctx context.Context, formattedDate, tickers string, limit int) (getTradesResponse, error) {
		return fetchConditionTrades(ctx, formattedDate, tickers, preset, limit)
	})
}

func runSignalPreset(ctx context.Context, cmd *cobra.Command, opts *SignalOptions, cmdName string, fetchResponse func(context.Context, string, string, int) (getTradesResponse, error)) error {
	return runStandardTrades(ctx, cmd, &standardRunConfig{
		cmdName:       cmdName,
		date:          opts.Date,
		tickers:       opts.Tickers,
		limit:         opts.Limit,
		fields:        opts.Fields,
		presetFields:  opts.PresetFields,
		shape:         opts.Shape,
		pretty:        opts.Pretty,
		fetchResponse: fetchResponse,
	})
}

func runStandardTrades(ctx context.Context, cmd *cobra.Command, config *standardRunConfig) error {
	formattedDate, tickers, err := parseDateAndTickers(ctx, config.cmdName, config.date, config.tickers)
	if err != nil {
		return err
	}
	limit, err := normalizeLimit(config.cmdName, config.limit, defaultTradeLimit, cmd.Flags().Changed("limit"))
	if err != nil {
		return err
	}
	fields, shape, err := normalizeOutputOptions(config.fields, config.presetFields, config.shape)
	if err != nil {
		return err
	}

	apiResponse, err := config.fetchResponse(ctx, formattedDate, tickers, limit)
	if err != nil {
		return err
	}

	result := Result{
		Status:          "ok",
		Date:            formattedDate,
		RecordsTotal:    apiResponse.RecordsTotal,
		RecordsFiltered: apiResponse.RecordsFiltered,
	}
	if err := applyTradeOutput(&result, apiResponse.Data, fields, shape); err != nil {
		return err
	}

	return encodeResult(cmd.OutOrStdout(), config.cmdName, result, config.pretty)
}

func parseDateAndTickers(ctx context.Context, cmdName, rawDate, rawTickers string) (formattedDate, normalizedTickers string, err error) {
	select {
	case <-ctx.Done():
		return "", "", fmt.Errorf("run %s command: %w", cmdName, ctx.Err())
	default:
	}

	tradeDate, err := time.Parse(dateLayout, rawDate)
	if err != nil {
		return "", "", fmt.Errorf("invalid date %q: use YYYY-MM-DD: %w", rawDate, err)
	}
	tickers, err := normalizeTickers(rawTickers)
	if err != nil {
		return "", "", err
	}

	return tradeDate.Format(dateLayout), tickers, nil
}

func normalizeDateRange(rawStartDate, rawEndDate string) (startDate, endDate string, err error) {
	endDate = strings.TrimSpace(rawEndDate)
	if endDate == "" {
		endDate = time.Now().Format(dateLayout)
	}
	parsedEndDate, err := time.Parse(dateLayout, endDate)
	if err != nil {
		return "", "", fmt.Errorf("invalid end-date %q: use YYYY-MM-DD: %w", rawEndDate, err)
	}

	startDate = strings.TrimSpace(rawStartDate)
	if startDate == "" {
		startDate = parsedEndDate.AddDate(-1, 0, 0).Format(dateLayout)
	}
	parsedStartDate, err := time.Parse(dateLayout, startDate)
	if err != nil {
		return "", "", fmt.Errorf("invalid start-date %q: use YYYY-MM-DD: %w", rawStartDate, err)
	}
	if parsedStartDate.After(parsedEndDate) {
		return "", "", fmt.Errorf("invalid date range: start-date %s is after end-date %s", startDate, endDate)
	}

	return parsedStartDate.Format(dateLayout), parsedEndDate.Format(dateLayout), nil
}

func normalizeRequiredTicker(rawTicker string) (string, error) {
	ticker := strings.TrimSpace(rawTicker)
	if ticker == "" {
		return "", fmt.Errorf("ticker is required")
	}
	if strings.Contains(ticker, ",") {
		return "", fmt.Errorf("invalid ticker %q: trade-levels accepts exactly one ticker", rawTicker)
	}
	return normalizeTickers(ticker)
}

func applyTradeLevelOptions(options *getTradeLevelsRequestOptions, opts *LevelOptions) {
	setStringIfPresent(&options.minVolume, opts.MinVolume)
	setStringIfPresent(&options.maxVolume, opts.MaxVolume)
	setStringIfPresent(&options.minPrice, opts.MinPrice)
	setStringIfPresent(&options.maxPrice, opts.MaxPrice)
	setStringIfPresent(&options.minDollars, opts.MinDollars)
	setStringIfPresent(&options.maxDollars, opts.MaxDollars)
	setStringIfPresent(&options.vcd, opts.VCD)
	setStringIfPresent(&options.relativeSize, opts.RelativeSize)
	if opts.TradeLevelRank != 0 {
		options.tradeLevelRank = opts.TradeLevelRank
	}
	if opts.TradeLevelCount != 0 {
		options.tradeLevelCount = opts.TradeLevelCount
		options.length = opts.TradeLevelCount
	}
}

func setStringIfPresent(target *string, value string) {
	if strings.TrimSpace(value) != "" {
		*target = strings.TrimSpace(value)
	}
}

func normalizeLimit(cmdName string, rawLimit, defaultLimit int, limitChanged bool) (int, error) {
	if !limitChanged {
		return defaultLimit, nil
	}
	if rawLimit < 1 {
		return 0, fmt.Errorf("invalid %s limit %d: use a value of 1 or greater", cmdName, rawLimit)
	}
	if rawLimit > maxTradeLimit {
		return 0, fmt.Errorf("invalid %s limit %d: use a value of %d or less", cmdName, rawLimit, maxTradeLimit)
	}

	return rawLimit, nil
}

func normalizeOutputOptions(rawFields, rawPreset, rawShape string) (fields []string, shape string, err error) {
	return normalizeOutputOptionsWithPresets(rawFields, rawPreset, rawShape, tradeFieldPresets)
}

func normalizeClusterOutputOptions(rawFields, rawPreset, rawShape string) (fields []string, shape string, err error) {
	return normalizeOutputOptionsWithPresets(rawFields, rawPreset, rawShape, clusterFieldPresets)
}

func normalizeTradeLevelOutputOptions(rawFields, rawPreset, rawShape string) (fields []string, shape string, err error) {
	return normalizeOutputOptionsWithPresets(rawFields, rawPreset, rawShape, tradeLevelFieldPresets)
}

func normalizeOutputOptionsWithPresets(rawFields, rawPreset, rawShape string, presets map[string][]string) (fields []string, shape string, err error) {
	shape = strings.ToLower(strings.TrimSpace(rawShape))
	if shape == "" {
		shape = defaultOutputShape
	}
	if shape != defaultOutputShape && shape != objectOutputShape {
		return nil, "", fmt.Errorf("invalid shape %q: use array or objects", rawShape)
	}

	fields, err = normalizeFields(rawFields, rawPreset, presets)
	if err != nil {
		return nil, "", err
	}

	return fields, shape, nil
}

func normalizeFields(rawFields, rawPreset string, presets map[string][]string) ([]string, error) {
	if strings.TrimSpace(rawFields) != "" {
		return parseFields(rawFields)
	}

	preset := strings.ToLower(strings.TrimSpace(rawPreset))
	if preset == "" {
		preset = defaultFieldPreset
	}
	if preset == fullFieldPreset {
		return nil, nil
	}
	fields, ok := presets[preset]
	if !ok {
		return nil, fmt.Errorf("invalid preset-fields %q: use core, expanded, or full", rawPreset)
	}

	return append([]string(nil), fields...), nil
}

func parseFields(rawFields string) ([]string, error) {
	parts := strings.Split(rawFields, ",")
	fields := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		field := strings.TrimSpace(part)
		if field == "" {
			return nil, fmt.Errorf("invalid fields %q: comma-delimited list contains an empty field", rawFields)
		}
		if strings.ContainsAny(field, " \t\n\r") {
			return nil, fmt.Errorf("invalid field %q: field names cannot contain whitespace", field)
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		fields = append(fields, field)
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("invalid fields %q: provide at least one field", rawFields)
	}

	return fields, nil
}

func applyTradeOutput(result *Result, trades []json.RawMessage, fields []string, shape string) error {
	output, err := projectTradeOutput(trades, fields, shape)
	if err != nil {
		return err
	}
	result.Fields = output.fields
	result.Rows = output.rows
	result.Trades = output.trades
	return nil
}

func applyClusterOutput(result *ClusterResult, clusters []json.RawMessage, fields []string, shape string) error {
	output, err := projectTradeOutput(clusters, fields, shape)
	if err != nil {
		return err
	}
	result.Fields = output.fields
	result.Rows = output.rows
	result.Clusters = output.trades
	return nil
}

func applyTradeLevelOutput(result *LevelResult, levels []json.RawMessage, fields []string, shape string) error {
	output, err := projectTradeOutput(levels, fields, shape)
	if err != nil {
		return err
	}
	result.Fields = output.fields
	result.Rows = output.rows
	result.Levels = output.trades
	return nil
}

func applyRankedTradeOutput(result *RankedResult, trades []json.RawMessage, fields []string, shape string) error {
	output, err := projectTradeOutput(trades, fields, shape)
	if err != nil {
		return err
	}
	result.Fields = output.fields
	result.Rows = output.rows
	result.Trades = output.trades
	return nil
}

func projectTradeOutput(trades []json.RawMessage, fields []string, shape string) (tradeOutput, error) {
	if trades == nil {
		trades = []json.RawMessage{}
	}
	if fields == nil {
		return tradeOutput{trades: trades}, nil
	}

	if shape == objectOutputShape {
		projected, err := projectTradeObjects(trades, fields)
		if err != nil {
			return tradeOutput{}, err
		}
		return tradeOutput{fields: fields, trades: projected}, nil
	}

	rows, err := projectTradeRows(trades, fields)
	if err != nil {
		return tradeOutput{}, err
	}
	return tradeOutput{fields: fields, rows: rows}, nil
}

func projectTradeObjects(trades []json.RawMessage, fields []string) ([]json.RawMessage, error) {
	projected := make([]json.RawMessage, 0, len(trades))
	for _, trade := range trades {
		object, err := decodeTradeObject(trade)
		if err != nil {
			return nil, err
		}
		row := make(map[string]json.RawMessage, len(fields))
		for _, field := range fields {
			if value, ok := projectedFieldValue(object, field); ok {
				row[field] = value
			}
		}
		encoded, err := json.Marshal(row)
		if err != nil {
			return nil, fmt.Errorf("encode projected trade row: %w", err)
		}
		projected = append(projected, encoded)
	}

	return projected, nil
}

func projectTradeRows(trades []json.RawMessage, fields []string) ([][]json.RawMessage, error) {
	rows := make([][]json.RawMessage, 0, len(trades))
	for _, trade := range trades {
		object, err := decodeTradeObject(trade)
		if err != nil {
			return nil, err
		}
		row := make([]json.RawMessage, 0, len(fields))
		for _, field := range fields {
			value, ok := projectedFieldValue(object, field)
			if !ok {
				value = nullJSON
			}
			row = append(row, value)
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func decodeTradeObject(trade json.RawMessage) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(trade, &object); err != nil {
		return nil, fmt.Errorf("decode trade row: %w", err)
	}

	return object, nil
}

func projectedFieldValue(object map[string]json.RawMessage, field string) (json.RawMessage, bool) {
	if field == calendarEventField {
		return calendarEventValue(object)
	}
	if field == auctionTradeField {
		return auctionTradeValue(object)
	}

	value, ok := object[field]
	return value, ok
}

func auctionTradeValue(object map[string]json.RawMessage) (json.RawMessage, bool) {
	if rawValue, ok := object[openingTradeField]; ok {
		value, valid := rawJSONFlag(rawValue)
		if valid && value {
			return json.RawMessage(`"open"`), true
		}
	}
	if rawValue, ok := object[closingTradeField]; ok {
		value, valid := rawJSONFlag(rawValue)
		if valid && value {
			return json.RawMessage(`"close"`), true
		}
	}
	return nil, false
}

func calendarEventValue(object map[string]json.RawMessage) (json.RawMessage, bool) {
	events := make([]string, 0, len(calendarEventFields))
	for _, field := range calendarEventFields {
		if rawValue, ok := object[field]; ok {
			value, valid := rawJSONFlag(rawValue)
			if valid && value {
				events = append(events, field)
			}
		}
	}
	if len(events) == 0 {
		return nil, false
	}

	encoded, err := json.Marshal(strings.Join(events, ","))
	if err != nil {
		return nil, false
	}
	return encoded, true
}

func rawJSONFlag(rawValue json.RawMessage) (value, ok bool) {
	var boolValue bool
	if err := json.Unmarshal(rawValue, &boolValue); err == nil {
		return boolValue, true
	}

	var numericValue int
	if err := json.Unmarshal(rawValue, &numericValue); err != nil {
		return false, false
	}
	return numericValue != 0, true
}

func encodeResult(w io.Writer, cmdName string, result any, pretty bool) error {
	encoder := json.NewEncoder(w)
	if pretty {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("encode %s response: %w", cmdName, err)
	}

	return nil
}

func fetchDisproportionatelyLargeTradesWithFilters(ctx context.Context, tradeDate, tickers, darkPools, sweeps string, limit int) (getTradesResponse, error) {
	options := defaultGetTradesRequestOptions()
	options.darkPools = darkPools
	options.sweeps = sweeps
	return fetchTradesPages(ctx, tradeDate, tickers, &options, limit)
}

func fetchClusterPreset(ctx context.Context, tradeDate, tickers string, preset *clusterPreset, limit int) (getTradesResponse, error) {
	options := clusterPresetRequestOptions(preset)
	return fetchTradeClustersPages(ctx, tradeDate, tickers, &options, limit)
}

func fetchRankedTrades(ctx context.Context, tradeDate, tickers string, preset *rankedPreset, limit int) (getTradesResponse, error) {
	options := rankedGetTradesRequestOptions(preset)
	return fetchTradesPages(ctx, tradeDate, tickers, &options, limit)
}

func fetchSignalTrades(ctx context.Context, tradeDate, tickers string, preset *signalPreset, limit int) (getTradesResponse, error) {
	options := signalGetTradesRequestOptions(preset)
	return fetchTradesPages(ctx, tradeDate, tickers, &options, limit)
}

func fetchConditionTrades(ctx context.Context, tradeDate, tickers string, preset *conditionPreset, limit int) (getTradesResponse, error) {
	options := conditionGetTradesRequestOptions(preset)
	return fetchTradesPages(ctx, tradeDate, tickers, &options, limit)
}

func fetchTradeClustersPages(ctx context.Context, tradeDate, tickers string, options *getTradeClustersRequestOptions, limit int) (getTradesResponse, error) {
	return fetchTradesResponsePages(ctx, "GetTradeClusters", limit, func(page, pageLength int) (getTradesResponse, error) {
		pageOptions := *options
		pageOptions.draw = page + 1
		pageOptions.start = page * defaultTradePageSize
		pageOptions.length = pageLength

		return fetchTradeClustersPage(ctx, tradeDate, tickers, &pageOptions)
	})
}

func fetchTradeLevels(ctx context.Context, startDate, endDate, ticker string, options *getTradeLevelsRequestOptions) (getTradesResponse, error) {
	return fetchTradeResponse(ctx, tradeRequestConfig{
		operation: "GetTradeLevels",
		endpoint:  getTradeLevelsEndpoint,
		form:      getTradeLevelsForm(startDate, endDate, ticker, options),
		setHeaders: func(req *http.Request, token string) {
			setGetTradeLevelsHeaders(req, token, startDate, endDate, ticker, options)
		},
		afterDecode: inferTradeLevelTotals,
	})
}

func fetchTradeClustersPage(ctx context.Context, tradeDate, tickers string, options *getTradeClustersRequestOptions) (getTradesResponse, error) {
	return fetchTradeResponse(ctx, tradeRequestConfig{
		operation: "GetTradeClusters",
		endpoint:  getTradeClustersEndpoint,
		form:      getTradeClustersForm(tradeDate, tickers, options),
		setHeaders: func(req *http.Request, token string) {
			setGetTradeClustersHeaders(req, token, tradeDate, tickers, options)
		},
		afterDecode: inferTradeClusterTotals,
	})
}

func fetchTradesPages(ctx context.Context, tradeDate, tickers string, options *getTradesRequestOptions, limit int) (getTradesResponse, error) {
	return fetchTradesResponsePages(ctx, "GetTrades", limit, func(page, pageLength int) (getTradesResponse, error) {
		pageOptions := *options
		pageOptions.draw = page + 1
		pageOptions.start = page * defaultTradePageSize
		pageOptions.length = pageLength

		return fetchTrades(ctx, tradeDate, tickers, &pageOptions)
	})
}

func fetchTradesResponsePages(ctx context.Context, operation string, limit int, fetchPage func(page, pageLength int) (getTradesResponse, error)) (getTradesResponse, error) {
	if limit < 1 {
		return getTradesResponse{}, fmt.Errorf("fetch %s pages: limit must be 1 or greater", operation)
	}
	if limit > maxTradeLimit {
		return getTradesResponse{}, fmt.Errorf("fetch %s pages: limit must be %d or less", operation, maxTradeLimit)
	}

	var merged getTradesResponse
	merged.Data = []json.RawMessage{}
	for page := 0; len(merged.Data) < limit; page++ {
		select {
		case <-ctx.Done():
			return getTradesResponse{}, fmt.Errorf("fetch %s page %d: %w", operation, page+1, ctx.Err())
		default:
		}

		remaining := limit - len(merged.Data)
		pageLength := min(defaultTradePageSize, remaining)

		apiResponse, err := fetchPage(page, pageLength)
		if err != nil {
			return getTradesResponse{}, err
		}
		if page == 0 {
			merged.Draw = apiResponse.Draw
			merged.RecordsTotal = apiResponse.RecordsTotal
			merged.RecordsFiltered = apiResponse.RecordsFiltered
		}

		merged.Data = append(merged.Data, apiResponse.Data...)
		if len(apiResponse.Data) < pageLength {
			break
		}
		if apiResponse.RecordsFiltered > 0 && page*defaultTradePageSize+len(apiResponse.Data) >= apiResponse.RecordsFiltered {
			break
		}
	}
	if len(merged.Data) > limit {
		merged.Data = merged.Data[:limit]
	}

	return merged, nil
}

func fetchTrades(ctx context.Context, tradeDate, tickers string, options *getTradesRequestOptions) (getTradesResponse, error) {
	return fetchTradeResponse(ctx, tradeRequestConfig{
		operation: "GetTrades",
		endpoint:  getTradesEndpoint,
		form:      getTradesForm(tradeDate, tickers, options),
		setHeaders: func(req *http.Request, token string) {
			setGetTradesHeaders(req, token, tradeDate, tickers, options)
		},
	})
}

func fetchTradeResponse(ctx context.Context, config tradeRequestConfig) (getTradesResponse, error) {
	cookies, err := extractCookies(ctx)
	if err != nil {
		return getTradesResponse{}, fmt.Errorf("extract VolumeLeaders browser cookies: %w", err)
	}

	token, err := fetchXSRFToken(ctx, getTradesHTTPClient, cookies)
	if err != nil {
		return getTradesResponse{}, fmt.Errorf("fetch VolumeLeaders XSRF token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.endpoint, strings.NewReader(config.form.Encode()))
	if err != nil {
		return getTradesResponse{}, fmt.Errorf("create %s request: %w", config.operation, err)
	}
	config.setHeaders(req, token)
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	resp, err := getTradesHTTPClient.Do(req)
	if err != nil {
		return getTradesResponse{}, fmt.Errorf("post %s request: %w", config.operation, err)
	}
	defer resp.Body.Close()

	if sessionExpiredResponse(resp) {
		return getTradesResponse{}, sessionExpiredCommandError()
	}
	if resp.StatusCode != http.StatusOK {
		return getTradesResponse{}, fmt.Errorf("%s request returned status %d", config.operation, resp.StatusCode)
	}

	bodyReader, closeReader, err := responseBodyReader(resp, config.operation)
	if err != nil {
		return getTradesResponse{}, err
	}
	defer closeReader()

	var apiResponse getTradesResponse
	if err := json.NewDecoder(bodyReader).Decode(&apiResponse); err != nil {
		return getTradesResponse{}, fmt.Errorf("decode %s response: %w", config.operation, err)
	}
	if apiResponse.Error != "" {
		return getTradesResponse{}, fmt.Errorf("%s response error: %s", config.operation, apiResponse.Error)
	}
	if apiResponse.Data == nil {
		apiResponse.Data = []json.RawMessage{}
	}
	if config.afterDecode != nil {
		config.afterDecode(&apiResponse)
	}

	return apiResponse, nil
}

func defaultGetTradeClustersRequestOptions() getTradeClustersRequestOptions {
	return clusterPresetRequestOptions(defaultClusterPreset())
}

func defaultGetTradeLevelsRequestOptions() getTradeLevelsRequestOptions {
	return getTradeLevelsRequestOptions{
		draw:            1,
		start:           0,
		length:          -1,
		minVolume:       "0",
		maxVolume:       "2000000000",
		minPrice:        "0",
		maxPrice:        "100000",
		minDollars:      "500000",
		maxDollars:      "30000000000",
		vcd:             "0",
		relativeSize:    "0",
		tradeLevelRank:  -1,
		tradeLevelCount: 10,
	}
}

func defaultClusterPreset() *clusterPreset {
	return &clusterPreset{
		tradeClusterRank:  -1,
		minVolume:         "0",
		maxDollars:        "30000000000",
		conditions:        "-1",
		vcd:               "0",
		relativeSize:      "0",
		darkPools:         "-1",
		sweeps:            "-1",
		latePrints:        "-1",
		signaturePrints:   "-1",
		evenShared:        "-1",
		securityTypeKey:   "-1",
		marketCap:         "0",
		includePremarket:  "1",
		includeRTH:        "1",
		includeAH:         "1",
		includeOpening:    "1",
		includeClosing:    "1",
		includePhantom:    "1",
		includeOffsetting: "1",
		sectorIndustry:    "",
		presetID:          "87",
	}
}

func clusterPresetRequestOptions(preset *clusterPreset) getTradeClustersRequestOptions {
	return getTradeClustersRequestOptions{
		tradeClusterRank:       clusterPresetRank(preset),
		draw:                   1,
		start:                  0,
		length:                 clusterPresetDefaultLimit(preset),
		minVolume:              clusterPresetString(preset.minVolume, "0"),
		maxDollars:             clusterPresetString(preset.maxDollars, "30000000000"),
		conditions:             clusterPresetString(preset.conditions, "-1"),
		vcd:                    clusterPresetString(preset.vcd, "0"),
		relativeSize:           clusterPresetString(preset.relativeSize, "0"),
		darkPools:              clusterPresetString(preset.darkPools, "-1"),
		sweeps:                 clusterPresetString(preset.sweeps, "-1"),
		latePrints:             clusterPresetString(preset.latePrints, "-1"),
		signaturePrints:        clusterPresetString(preset.signaturePrints, "-1"),
		evenShared:             clusterPresetString(preset.evenShared, "-1"),
		securityTypeKey:        clusterPresetString(preset.securityTypeKey, "-1"),
		marketCap:              clusterPresetString(preset.marketCap, "0"),
		includePremarket:       clusterPresetString(preset.includePremarket, "1"),
		includeRTH:             clusterPresetString(preset.includeRTH, "1"),
		includeAH:              clusterPresetString(preset.includeAH, "1"),
		includeOpening:         clusterPresetString(preset.includeOpening, "1"),
		includeClosing:         clusterPresetString(preset.includeClosing, "1"),
		includePhantom:         clusterPresetString(preset.includePhantom, "1"),
		includeOffsetting:      clusterPresetString(preset.includeOffsetting, "1"),
		sectorIndustry:         preset.sectorIndustry,
		presetSearchTemplateID: clusterPresetString(preset.presetID, "87"),
	}
}

func clusterPresetDefaultLimit(preset *clusterPreset) int {
	if preset.length > 0 {
		return preset.length
	}
	return defaultTradeLimit
}

func clusterPresetRank(preset *clusterPreset) int {
	if preset.tradeClusterRank > 0 {
		return preset.tradeClusterRank
	}
	return -1
}

func clusterPresetString(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func defaultGetTradesRequestOptions() getTradesRequestOptions {
	return getTradesRequestOptions{
		tradeRank:              -1,
		draw:                   1,
		start:                  0,
		length:                 defaultTradeLimit,
		minVolume:              "0",
		maxDollars:             "30000000000",
		conditions:             "-1",
		vcd:                    "0",
		relativeSize:           "5",
		darkPools:              "-1",
		sweeps:                 "-1",
		latePrints:             "-1",
		signaturePrints:        "-1",
		evenShared:             "-1",
		securityTypeKey:        "-1",
		tradeRankSnapshot:      "-1",
		marketCap:              "0",
		includePremarket:       "1",
		includeRTH:             "1",
		includeAH:              "1",
		includeOpening:         "1",
		includeClosing:         "1",
		includePhantom:         "1",
		includeOffsetting:      "1",
		sectorIndustry:         "",
		presetSearchTemplateID: "87",
	}
}

func rankedGetTradesRequestOptions(preset *rankedPreset) getTradesRequestOptions {
	return getTradesRequestOptions{
		tradeRank:              preset.rank,
		draw:                   1,
		start:                  0,
		length:                 clusterPresetDefaultLimit(&clusterPreset{length: preset.length}),
		minVolume:              clusterPresetString(preset.minVolume, "10000"),
		maxDollars:             clusterPresetString(preset.maxDollars, "100000000000"),
		conditions:             clusterPresetString(preset.conditions, ignoredRSIConditions),
		vcd:                    clusterPresetString(preset.vcd, "0"),
		relativeSize:           clusterPresetString(preset.relativeSize, "0"),
		darkPools:              clusterPresetString(preset.darkPools, "-1"),
		sweeps:                 clusterPresetString(preset.sweeps, "-1"),
		latePrints:             clusterPresetString(preset.latePrints, "-1"),
		signaturePrints:        clusterPresetString(preset.signaturePrints, "-1"),
		evenShared:             clusterPresetString(preset.evenShared, "-1"),
		securityTypeKey:        clusterPresetString(preset.securityTypeKey, "-1"),
		tradeRankSnapshot:      clusterPresetString(preset.tradeRankSnapshot, "-1"),
		marketCap:              clusterPresetString(preset.marketCap, "0"),
		includePremarket:       clusterPresetString(preset.includePremarket, "1"),
		includeRTH:             clusterPresetString(preset.includeRTH, "1"),
		includeAH:              clusterPresetString(preset.includeAH, "1"),
		includeOpening:         clusterPresetString(preset.includeOpening, "1"),
		includeClosing:         clusterPresetString(preset.includeClosing, "1"),
		includePhantom:         clusterPresetString(preset.includePhantom, "-1"),
		includeOffsetting:      clusterPresetString(preset.includeOffsetting, "-1"),
		sectorIndustry:         preset.sectorIndustry,
		presetSearchTemplateID: preset.presetID,
	}
}

func signalGetTradesRequestOptions(preset *signalPreset) getTradesRequestOptions {
	return getTradesRequestOptions{
		tradeRank:              -1,
		draw:                   1,
		start:                  0,
		length:                 defaultTradeLimit,
		minVolume:              "0",
		maxDollars:             "100000000000",
		conditions:             "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH",
		vcd:                    "0",
		relativeSize:           "0",
		darkPools:              preset.darkPools,
		sweeps:                 "-1",
		latePrints:             "-1",
		signaturePrints:        "-1",
		evenShared:             "-1",
		securityTypeKey:        "-1",
		tradeRankSnapshot:      "-1",
		marketCap:              "0",
		includePremarket:       "1",
		includeRTH:             "1",
		includeAH:              "1",
		includeOpening:         "1",
		includeClosing:         "1",
		includePhantom:         preset.phantom,
		includeOffsetting:      preset.offsetting,
		sectorIndustry:         "",
		presetSearchTemplateID: preset.presetID,
	}
}

func conditionGetTradesRequestOptions(preset *conditionPreset) getTradesRequestOptions {
	return getTradesRequestOptions{
		tradeRank:              100,
		draw:                   1,
		start:                  0,
		length:                 defaultTradeLimit,
		minVolume:              "10000",
		maxDollars:             "10000000000",
		conditions:             preset.conditions,
		vcd:                    "0",
		relativeSize:           "5",
		darkPools:              "-1",
		sweeps:                 "-1",
		latePrints:             "-1",
		signaturePrints:        "0",
		evenShared:             "-1",
		securityTypeKey:        "-1",
		tradeRankSnapshot:      "-1",
		marketCap:              "0",
		includePremarket:       "1",
		includeRTH:             "1",
		includeAH:              "1",
		includeOpening:         "1",
		includeClosing:         "1",
		includePhantom:         "-1",
		includeOffsetting:      "-1",
		sectorIndustry:         "",
		presetSearchTemplateID: preset.presetID,
	}
}

func setGetTradeClustersHeaders(req *http.Request, token, tradeDate, tickers string, options *getTradeClustersRequestOptions) {
	setCommonXHRHeaders(req, token)
	req.Header.Set("Referer", tradeClustersReferer(tradeDate, tickers, options))
}

func setGetTradesHeaders(req *http.Request, token, tradeDate, tickers string, options *getTradesRequestOptions) {
	setCommonXHRHeaders(req, token)
	req.Header.Set("Referer", tradesReferer(tradeDate, tickers, options))
}

func setGetTradeLevelsHeaders(req *http.Request, token, startDate, endDate, ticker string, options *getTradeLevelsRequestOptions) {
	setCommonXHRHeaders(req, token)
	req.Header.Set("Referer", tradeLevelsReferer(startDate, endDate, ticker, options))
}

func setCommonXHRHeaders(req *http.Request, token string) {
	req.Header.Set("User-Agent", auth.UserAgent)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-XSRF-Token", token)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", "https://www.volumeleaders.com")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
}

func tradeClustersReferer(tradeDate, tickers string, options *getTradeClustersRequestOptions) string {
	query := url.Values{}
	setTradeClustersQueryParams(query, tradeDate, tickers, options)
	query.Set("PresetSearchTemplateID", options.presetSearchTemplateID)
	query.Set("ViewMode", "Automatic")
	return tradeClustersPage + "?" + query.Encode()
}

func tradesReferer(tradeDate, tickers string, options *getTradesRequestOptions) string {
	query := url.Values{}
	setSharedQueryParams(query, tradeDate, tickers, options)
	query.Set("PresetSearchTemplateID", options.presetSearchTemplateID)
	query.Set("ViewMode", "Automatic")
	return tradesPage + "?" + query.Encode()
}

func tradeLevelsReferer(startDate, endDate, ticker string, options *getTradeLevelsRequestOptions) string {
	query := url.Values{}
	setTradeLevelsQueryParams(query, startDate, endDate, ticker, options)
	return tradeLevelsPage + "?" + query.Encode()
}

func setTradeClustersQueryParams(values url.Values, tradeDate, tickers string, options *getTradeClustersRequestOptions) {
	values.Set("Tickers", tickers)
	values.Set("StartDate", tradeDate)
	values.Set("EndDate", tradeDate)
	values.Set("MinVolume", options.minVolume)
	values.Set("MaxVolume", "2000000000")
	values.Set("Conditions", options.conditions)
	values.Set("VCD", options.vcd)
	values.Set("DarkPools", options.darkPools)
	values.Set("Sweeps", options.sweeps)
	values.Set("LatePrints", options.latePrints)
	values.Set("SignaturePrints", options.signaturePrints)
	values.Set("EvenShared", options.evenShared)
	values.Set("SecurityTypeKey", options.securityTypeKey)
	values.Set("RelativeSize", options.relativeSize)
	values.Set("MinPrice", "0")
	values.Set("MaxPrice", "100000")
	values.Set("MinDollars", "500000")
	values.Set("MaxDollars", options.maxDollars)
	values.Set("TradeClusterRank", fmt.Sprintf("%d", options.tradeClusterRank))
	values.Set("MarketCap", options.marketCap)
	values.Set("IncludePremarket", options.includePremarket)
	values.Set("IncludeRTH", options.includeRTH)
	values.Set("IncludeAH", options.includeAH)
	values.Set("IncludeOpening", options.includeOpening)
	values.Set("IncludeClosing", options.includeClosing)
	values.Set("IncludePhantom", options.includePhantom)
	values.Set("IncludeOffsetting", options.includeOffsetting)
	values.Set("SectorIndustry", options.sectorIndustry)
}

func setSharedQueryParams(values url.Values, tradeDate, tickers string, options *getTradesRequestOptions) {
	values.Set("Tickers", tickers)
	values.Set("StartDate", tradeDate)
	values.Set("EndDate", tradeDate)
	values.Set("MinVolume", options.minVolume)
	values.Set("MaxVolume", "2000000000")
	values.Set("Conditions", options.conditions)
	values.Set("VCD", options.vcd)
	values.Set("RelativeSize", options.relativeSize)
	values.Set("DarkPools", options.darkPools)
	values.Set("Sweeps", options.sweeps)
	values.Set("LatePrints", options.latePrints)
	values.Set("SignaturePrints", options.signaturePrints)
	values.Set("EvenShared", options.evenShared)
	values.Set("SecurityTypeKey", options.securityTypeKey)
	values.Set("MinPrice", "0")
	values.Set("MaxPrice", "100000")
	values.Set("MinDollars", "500000")
	values.Set("MaxDollars", options.maxDollars)
	values.Set("TradeRank", fmt.Sprintf("%d", options.tradeRank))
	values.Set("TradeRankSnapshot", options.tradeRankSnapshot)
	values.Set("MarketCap", options.marketCap)
	values.Set("IncludePremarket", options.includePremarket)
	values.Set("IncludeRTH", options.includeRTH)
	values.Set("IncludeAH", options.includeAH)
	values.Set("IncludeOpening", options.includeOpening)
	values.Set("IncludeClosing", options.includeClosing)
	values.Set("IncludePhantom", options.includePhantom)
	values.Set("IncludeOffsetting", options.includeOffsetting)
	values.Set("SectorIndustry", options.sectorIndustry)
}

func setTradeLevelsQueryParams(values url.Values, startDate, endDate, ticker string, options *getTradeLevelsRequestOptions) {
	values.Set("Ticker", ticker)
	values.Set("MinVolume", options.minVolume)
	values.Set("MaxVolume", options.maxVolume)
	values.Set("VCD", options.vcd)
	values.Set("RelativeSize", options.relativeSize)
	values.Set("MinPrice", options.minPrice)
	values.Set("MaxPrice", options.maxPrice)
	values.Set("MinDollars", options.minDollars)
	values.Set("MaxDollars", options.maxDollars)
	values.Set("StartDate", startDate)
	values.Set("EndDate", endDate)
	values.Set("MinDate", startDate)
	values.Set("MaxDate", endDate)
	values.Set("TradeLevelRank", strconv.Itoa(options.tradeLevelRank))
	values.Set("TradeLevelCount", strconv.Itoa(options.tradeLevelCount))
}

func getTradeClustersForm(tradeDate, tickers string, options *getTradeClustersRequestOptions) url.Values {
	form := url.Values{}
	setDataTableFormFields(form, getTradeClustersColumns, "MinFullTimeString24", options.draw, options.start, options.length)
	setTradeClustersQueryParams(form, tradeDate, tickers, options)
	return form
}

func getTradesForm(tradeDate, tickers string, options *getTradesRequestOptions) url.Values {
	form := url.Values{}
	setDataTableFormFields(form, getTradesColumns, "FullTimeString24", options.draw, options.start, options.length)
	setSharedQueryParams(form, tradeDate, tickers, options)
	return form
}

func getTradeLevelsForm(startDate, endDate, ticker string, options *getTradeLevelsRequestOptions) url.Values {
	form := url.Values{}
	setDataTableFormFields(form, getTradeLevelsColumns, "$$", options.draw, options.start, options.length)
	setTradeLevelsQueryParams(form, startDate, endDate, ticker, options)
	return form
}

func setDataTableFormFields(form url.Values, columns []tradeColumn, orderName string, draw, start, length int) {
	if draw == 0 {
		draw = 1
	}
	form.Set("draw", fmt.Sprintf("%d", draw))
	for i, column := range columns {
		prefix := fmt.Sprintf("columns[%d]", i)
		form.Set(prefix+"[data]", column.data)
		form.Set(prefix+"[name]", column.name)
		form.Set(prefix+"[searchable]", "true")
		form.Set(prefix+"[orderable]", column.orderable)
		form.Set(prefix+"[search][value]", "")
		form.Set(prefix+"[search][regex]", "false")
	}
	form.Set("order[0][column]", "1")
	form.Set("order[0][dir]", "DESC")
	form.Set("order[0][name]", orderName)
	form.Set("start", fmt.Sprintf("%d", start))
	form.Set("length", fmt.Sprintf("%d", length))
	form.Set("search[value]", "")
	form.Set("search[regex]", "false")
}

func normalizeTickers(rawTickers string) (string, error) {
	if strings.TrimSpace(rawTickers) == "" {
		return "", nil
	}
	if strings.ContainsAny(rawTickers, " \t\n\r") {
		return "", fmt.Errorf("invalid tickers %q: use a comma-delimited list without spaces", rawTickers)
	}

	parts := strings.Split(rawTickers, ",")
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		ticker := strings.ToUpper(part)
		if ticker == "" {
			return "", fmt.Errorf("invalid tickers %q: comma-delimited list contains an empty ticker", rawTickers)
		}
		if !tickerPattern.MatchString(ticker) {
			return "", fmt.Errorf("invalid ticker %q: use letters, numbers, dots, or hyphens", ticker)
		}
		normalized = append(normalized, ticker)
	}

	return strings.Join(normalized, ","), nil
}

func sessionExpiredResponse(resp *http.Response) bool {
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return true
	}
	if resp.Request == nil || resp.Request.URL == nil {
		return false
	}
	return strings.EqualFold(resp.Request.URL.EscapedPath(), "/Login")
}

func sessionExpiredCommandError() error {
	return fmt.Errorf("%s: %w", auth.SessionExpiredMessage, auth.ErrSessionExpired)
}

func inferTradeClusterTotals(apiResponse *getTradesResponse) {
	if len(apiResponse.Data) == 0 || (apiResponse.RecordsTotal != 0 && apiResponse.RecordsFiltered != 0) {
		return
	}
	var row struct {
		TotalRows int `json:"TotalRows"`
	}
	if err := json.Unmarshal(apiResponse.Data[0], &row); err != nil || row.TotalRows == 0 {
		return
	}
	if apiResponse.RecordsTotal == 0 {
		apiResponse.RecordsTotal = row.TotalRows
	}
	if apiResponse.RecordsFiltered == 0 {
		apiResponse.RecordsFiltered = row.TotalRows
	}
}

func inferTradeLevelTotals(apiResponse *getTradesResponse) {
	if len(apiResponse.Data) == 0 || (apiResponse.RecordsTotal != 0 && apiResponse.RecordsFiltered != 0) {
		return
	}
	var row struct {
		TotalRows int `json:"TotalRows"`
	}
	if err := json.Unmarshal(apiResponse.Data[0], &row); err != nil || row.TotalRows == 0 {
		return
	}
	if apiResponse.RecordsTotal == 0 {
		apiResponse.RecordsTotal = row.TotalRows
	}
	if apiResponse.RecordsFiltered == 0 {
		apiResponse.RecordsFiltered = row.TotalRows
	}
}

func responseBodyReader(resp *http.Response, operation string) (io.Reader, func(), error) {
	if resp.Header.Get("Content-Encoding") != "gzip" {
		return resp.Body, func() {}, nil
	}
	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, func() {}, fmt.Errorf("decompress %s response: %w", operation, err)
	}
	return gr, func() { _ = gr.Close() }, nil
}
