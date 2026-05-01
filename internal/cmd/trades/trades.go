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
	getTradesPath        = "https://www.volumeleaders.com/Trades/GetTrades"
	tradesPage           = "https://www.volumeleaders.com/Trades"
)

var (
	extractCookies      = auth.ExtractCookies
	fetchXSRFToken      = auth.FetchXSRFToken
	getTradesHTTPClient = http.DefaultClient
	getTradesEndpoint   = getTradesPath
	tickerPattern       = regexp.MustCompile(`^[A-Z0-9.-]+$`)
	nullJSON            = json.RawMessage("null")
)

var tradeFieldPresets = map[string][]string{
	"core": {
		"Ticker",
		"FullTimeString24",
		"Price",
		"Dollars",
		"DollarsMultiplier",
		"Volume",
		"TradeRank",
		"DarkPool",
		"Sweep",
		"LatePrint",
		"SignaturePrint",
		"Sector",
	},
	"signals": {
		"Ticker",
		"FullTimeString24",
		"Price",
		"Dollars",
		"DollarsMultiplier",
		"Volume",
		"CumulativeDistribution",
		"TradeRank",
		"TradeRankSnapshot",
		"DarkPool",
		"Sweep",
		"LatePrint",
		"SignaturePrint",
		"PhantomPrint",
		"InsideBar",
		"RSIHour",
		"RSIDay",
		"FrequencyLast30TD",
		"FrequencyLast90TD",
		"FrequencyLast1CY",
		"Sector",
		"Industry",
	},
}

// Options defines the LLM-readable contract for fetching unusual trades.
type Options struct {
	Date         string `flag:"date" flagshort:"d" flagdescr:"Single trading date to query, formatted as YYYY-MM-DD. The disproportionately large trades preset is intentionally limited to one day." flagenv:"true" flagrequired:"true" flaggroup:"Query" validate:"required" mod:"trim"`
	Tickers      string `flag:"tickers" flagdescr:"Optional ticker filter. Use one symbol or a comma-delimited list without spaces, for example AAPL or AAPL,MSFT." flagenv:"true" flaggroup:"Query" mod:"trim"`
	Limit        int    `flag:"limit" flagdescr:"Maximum trade rows to return. Must be between 1 and 100. Defaults to 100 when omitted." flagenv:"true" flaggroup:"Output"`
	Fields       string `flag:"fields" flagdescr:"Comma-separated trade fields to include. Overrides --preset-fields. Use upstream field names such as Ticker,Dollars,TradeRank." flagenv:"true" flaggroup:"Output" mod:"trim"`
	PresetFields string `flag:"preset-fields" flagdescr:"Field preset to include: core, signals, or full. Defaults to core for token-efficient output." flagenv:"true" flaggroup:"Output" mod:"trim"`
	Shape        string `flag:"shape" flagdescr:"Trade row shape: array or objects. Array is the default and is most token-efficient." flagenv:"true" flaggroup:"Output" mod:"trim"`
	Pretty       bool   `flag:"pretty" flagdescr:"Pretty-print JSON output. Compact JSON is the default for token-efficient LLM and MCP use." flagenv:"true" flaggroup:"Output"`
}

// RankedOptions defines the LLM-readable contract for fetching all-time ranked trades.
type RankedOptions struct {
	Date         string `flag:"date" flagshort:"d" flagdescr:"Single trading date to query, formatted as YYYY-MM-DD. Ranked trade presets are intentionally limited to one day." flagenv:"true" flagrequired:"true" flaggroup:"Query" validate:"required" mod:"trim"`
	Tickers      string `flag:"tickers" flagdescr:"Optional ticker filter. Use one symbol or a comma-delimited list without spaces, for example AAPL or AAPL,MSFT." flagenv:"true" flaggroup:"Query" mod:"trim"`
	Limit        int    `flag:"limit" flagdescr:"Maximum trade rows to return. Must be between 1 and 100. Defaults to the command preset when omitted." flagenv:"true" flaggroup:"Output"`
	Fields       string `flag:"fields" flagdescr:"Comma-separated trade fields to include. Overrides --preset-fields. Use upstream field names such as Ticker,Dollars,TradeRank." flagenv:"true" flaggroup:"Output" mod:"trim"`
	PresetFields string `flag:"preset-fields" flagdescr:"Field preset to include: core, signals, or full. Defaults to core for token-efficient output." flagenv:"true" flaggroup:"Output" mod:"trim"`
	Shape        string `flag:"shape" flagdescr:"Trade row shape: array or objects. Array is the default and is most token-efficient." flagenv:"true" flaggroup:"Output" mod:"trim"`
	Pretty       bool   `flag:"pretty" flagdescr:"Pretty-print JSON output. Compact JSON is the default for token-efficient LLM and MCP use." flagenv:"true" flaggroup:"Output"`
}

// SignalOptions defines the LLM-readable contract for fetching trade signal presets.
type SignalOptions struct {
	Date         string `flag:"date" flagshort:"d" flagdescr:"Single trading date to query, formatted as YYYY-MM-DD. Trade signal presets are intentionally limited to one day." flagenv:"true" flagrequired:"true" flaggroup:"Query" validate:"required" mod:"trim"`
	Tickers      string `flag:"tickers" flagdescr:"Optional ticker filter. Use one symbol or a comma-delimited list without spaces, for example AAPL or AAPL,MSFT." flagenv:"true" flaggroup:"Query" mod:"trim"`
	Limit        int    `flag:"limit" flagdescr:"Maximum trade rows to return. Must be between 1 and 100. Defaults to 100 when omitted." flagenv:"true" flaggroup:"Output"`
	Fields       string `flag:"fields" flagdescr:"Comma-separated trade fields to include. Overrides --preset-fields. Use upstream field names such as Ticker,Dollars,TradeRank." flagenv:"true" flaggroup:"Output" mod:"trim"`
	PresetFields string `flag:"preset-fields" flagdescr:"Field preset to include: core, signals, or full. Defaults to core for token-efficient output." flagenv:"true" flaggroup:"Output" mod:"trim"`
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
	includePhantom         string
	includeOffsetting      string
	sectorIndustry         string
	presetSearchTemplateID string
}

type rankedPreset struct {
	use      string
	aliases  []string
	short    string
	long     string
	example  string
	rank     int
	length   int
	presetID string
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

type leveragePreset struct {
	use            string
	aliases        []string
	short          string
	long           string
	example        string
	sectorIndustry string
	presetID       string
}

type sectorPreset struct {
	use            string
	aliases        []string
	short          string
	long           string
	example        string
	sectorIndustry string
	presetID       string
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

// NewCommand builds the large unusual trades command.
func NewCommand() (*cobra.Command, error) {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:     "trades",
		Aliases: []string{"large-trades", "unusual-trades"},
		Short:   "Fetch disproportionately large trades for one date",
		Long:    "Fetch VolumeLeaders disproportionately large trades for a single trading day. This reproduces the default Disproportionately large trades GetTrades request and intentionally does not allow multi-day ranges.",
		Example: "volumeleaders-agent trades --date 2026-04-30\nvolumeleaders-agent trades --date 2026-04-30 --tickers AAPL,MSFT",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), cmd, opts)
		},
	}

	if err := structcli.Bind(cmd, opts); err != nil {
		return nil, fmt.Errorf("bind trades options: %w", err)
	}
	cmd.Flags().StringVar(&opts.Tickers, "ticker", "", "Optional ticker filter. Alias for --tickers.")

	return cmd, nil
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

// NewBullLeverageCommand builds the bullish leveraged ETF trades command.
func NewBullLeverageCommand() (*cobra.Command, error) {
	return newLeverageCommand(&leveragePreset{
		use:            "bull-leverage",
		aliases:        []string{"bullish-leverage", "bull-leverage-etfs", "bullish-leverage-etfs"},
		short:          "Fetch bullish leveraged ETF trades",
		long:           "Fetch VolumeLeaders bullish leveraged ETF trades for one day. This uses the bullish leverage ETF preset and filters the upstream GetTrades request to the X Bull sector group.",
		example:        "volumeleaders-agent bull-leverage --date 2026-04-30\nvolumeleaders-agent bull-leverage --date 2026-04-30 --tickers TQQQ",
		sectorIndustry: "X Bull",
		presetID:       "5",
	})
}

// NewBearLeverageCommand builds the bearish leveraged ETF trades command.
func NewBearLeverageCommand() (*cobra.Command, error) {
	return newLeverageCommand(&leveragePreset{
		use:            "bear-leverage",
		aliases:        []string{"bearish-leverage", "bear-leverage-etfs", "bearish-leverage-etfs"},
		short:          "Fetch bearish leveraged ETF trades",
		long:           "Fetch VolumeLeaders bearish leveraged ETF trades for one day. This uses the bearish leverage ETF preset and filters the upstream GetTrades request to the X Bear sector group.",
		example:        "volumeleaders-agent bear-leverage --date 2026-04-30\nvolumeleaders-agent bear-leverage --date 2026-04-30 --tickers SPXU",
		sectorIndustry: "X Bear",
		presetID:       "6",
	})
}

// NewBiotechCommand builds the biotechnology stock trades command.
func NewBiotechCommand() (*cobra.Command, error) {
	return newSectorCommand(&sectorPreset{
		use:            "biotech",
		aliases:        []string{"biotechnology", "biotechnology-stocks", "biotech-stocks"},
		short:          "Fetch biotechnology stock trades",
		long:           "Fetch VolumeLeaders biotechnology stock trades for one day. This uses the biotechnology preset captured from VolumeLeaders and filters the upstream GetTrades request to the Biotech sector group.",
		example:        "volumeleaders-agent biotech --date 2026-04-30\nvolumeleaders-agent biotech --date 2026-04-30 --tickers IBB,XBI",
		sectorIndustry: "Biotech",
		presetID:       "89",
	})
}

// NewBondsCommand builds the bonds trades command.
func NewBondsCommand() (*cobra.Command, error) {
	return newSectorCommand(&sectorPreset{
		use:            "bonds",
		aliases:        []string{"bond-trades", "bond-etfs"},
		short:          "Fetch bonds trades",
		long:           "Fetch VolumeLeaders bonds trades for one day. This uses the bonds preset captured from VolumeLeaders and filters the upstream GetTrades request to the Bonds sector group.",
		example:        "volumeleaders-agent bonds --date 2026-04-30\nvolumeleaders-agent bonds --date 2026-04-30 --tickers HYG,TLT",
		sectorIndustry: "Bonds",
		presetID:       "90",
	})
}

func newRankedCommand(preset *rankedPreset) (*cobra.Command, error) {
	opts := &RankedOptions{}
	cmd := &cobra.Command{
		Use:     preset.use,
		Aliases: preset.aliases,
		Short:   preset.short,
		Long:    preset.long,
		Example: preset.example,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRanked(cmd.Context(), cmd, opts, preset)
		},
	}

	if err := structcli.Bind(cmd, opts); err != nil {
		return nil, fmt.Errorf("bind %s options: %w", preset.use, err)
	}
	cmd.Flags().StringVar(&opts.Tickers, "ticker", "", "Optional ticker filter. Alias for --tickers.")

	return cmd, nil
}

func newSignalCommand(preset *signalPreset) (*cobra.Command, error) {
	opts := &SignalOptions{}
	cmd := &cobra.Command{
		Use:     preset.use,
		Aliases: preset.aliases,
		Short:   preset.short,
		Long:    preset.long,
		Example: preset.example,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSignal(cmd.Context(), cmd, opts, preset)
		},
	}

	if err := structcli.Bind(cmd, opts); err != nil {
		return nil, fmt.Errorf("bind %s options: %w", preset.use, err)
	}
	cmd.Flags().StringVar(&opts.Tickers, "ticker", "", "Optional ticker filter. Alias for --tickers.")

	return cmd, nil
}

func newLeverageCommand(preset *leveragePreset) (*cobra.Command, error) {
	opts := &SignalOptions{}
	cmd := &cobra.Command{
		Use:     preset.use,
		Aliases: preset.aliases,
		Short:   preset.short,
		Long:    preset.long,
		Example: preset.example,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runLeverage(cmd.Context(), cmd, opts, preset)
		},
	}

	if err := structcli.Bind(cmd, opts); err != nil {
		return nil, fmt.Errorf("bind %s options: %w", preset.use, err)
	}
	cmd.Flags().StringVar(&opts.Tickers, "ticker", "", "Optional ticker filter. Alias for --tickers.")

	return cmd, nil
}

func newSectorCommand(preset *sectorPreset) (*cobra.Command, error) {
	opts := &SignalOptions{}
	cmd := &cobra.Command{
		Use:     preset.use,
		Aliases: preset.aliases,
		Short:   preset.short,
		Long:    preset.long,
		Example: preset.example,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSector(cmd.Context(), cmd, opts, preset)
		},
	}

	if err := structcli.Bind(cmd, opts); err != nil {
		return nil, fmt.Errorf("bind %s options: %w", preset.use, err)
	}
	cmd.Flags().StringVar(&opts.Tickers, "ticker", "", "Optional ticker filter. Alias for --tickers.")

	return cmd, nil
}

func run(ctx context.Context, cmd *cobra.Command, opts *Options) error {
	formattedDate, tickers, err := parseDateAndTickers(ctx, "trades", opts.Date, opts.Tickers)
	if err != nil {
		return err
	}
	limit, err := normalizeLimit("trades", opts.Limit, defaultTradeLimit, cmd.Flags().Changed("limit"))
	if err != nil {
		return err
	}
	fields, shape, err := normalizeOutputOptions(opts.Fields, opts.PresetFields, opts.Shape)
	if err != nil {
		return err
	}

	apiResponse, err := fetchDisproportionatelyLargeTrades(ctx, formattedDate, tickers, limit)
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

	return encodeResult(cmd.OutOrStdout(), "trades", result, opts.Pretty)
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
	formattedDate, tickers, err := parseDateAndTickers(ctx, preset.use, opts.Date, opts.Tickers)
	if err != nil {
		return err
	}
	limit, err := normalizeLimit(preset.use, opts.Limit, defaultTradeLimit, cmd.Flags().Changed("limit"))
	if err != nil {
		return err
	}
	fields, shape, err := normalizeOutputOptions(opts.Fields, opts.PresetFields, opts.Shape)
	if err != nil {
		return err
	}

	apiResponse, err := fetchSignalTrades(ctx, formattedDate, tickers, preset, limit)
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

	return encodeResult(cmd.OutOrStdout(), preset.use, result, opts.Pretty)
}

func runLeverage(ctx context.Context, cmd *cobra.Command, opts *SignalOptions, preset *leveragePreset) error {
	formattedDate, tickers, err := parseDateAndTickers(ctx, preset.use, opts.Date, opts.Tickers)
	if err != nil {
		return err
	}
	limit, err := normalizeLimit(preset.use, opts.Limit, defaultTradeLimit, cmd.Flags().Changed("limit"))
	if err != nil {
		return err
	}
	fields, shape, err := normalizeOutputOptions(opts.Fields, opts.PresetFields, opts.Shape)
	if err != nil {
		return err
	}

	apiResponse, err := fetchLeverageTrades(ctx, formattedDate, tickers, preset, limit)
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

	return encodeResult(cmd.OutOrStdout(), preset.use, result, opts.Pretty)
}

func runSector(ctx context.Context, cmd *cobra.Command, opts *SignalOptions, preset *sectorPreset) error {
	formattedDate, tickers, err := parseDateAndTickers(ctx, preset.use, opts.Date, opts.Tickers)
	if err != nil {
		return err
	}
	limit, err := normalizeLimit(preset.use, opts.Limit, defaultTradeLimit, cmd.Flags().Changed("limit"))
	if err != nil {
		return err
	}
	fields, shape, err := normalizeOutputOptions(opts.Fields, opts.PresetFields, opts.Shape)
	if err != nil {
		return err
	}

	apiResponse, err := fetchSectorTrades(ctx, formattedDate, tickers, preset, limit)
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

	return encodeResult(cmd.OutOrStdout(), preset.use, result, opts.Pretty)
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
	shape = strings.ToLower(strings.TrimSpace(rawShape))
	if shape == "" {
		shape = defaultOutputShape
	}
	if shape != defaultOutputShape && shape != objectOutputShape {
		return nil, "", fmt.Errorf("invalid shape %q: use array or objects", rawShape)
	}

	fields, err = normalizeFields(rawFields, rawPreset)
	if err != nil {
		return nil, "", err
	}

	return fields, shape, nil
}

func normalizeFields(rawFields, rawPreset string) ([]string, error) {
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
	fields, ok := tradeFieldPresets[preset]
	if !ok {
		return nil, fmt.Errorf("invalid preset-fields %q: use core, signals, or full", rawPreset)
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
	if trades == nil {
		trades = []json.RawMessage{}
	}
	if fields == nil {
		result.Trades = trades
		return nil
	}

	result.Fields = fields
	if shape == objectOutputShape {
		projected, err := projectTradeObjects(trades, fields)
		if err != nil {
			return err
		}
		result.Trades = projected
		return nil
	}

	rows, err := projectTradeRows(trades, fields)
	if err != nil {
		return err
	}
	result.Rows = rows
	return nil
}

func applyRankedTradeOutput(result *RankedResult, trades []json.RawMessage, fields []string, shape string) error {
	if trades == nil {
		trades = []json.RawMessage{}
	}
	if fields == nil {
		result.Trades = trades
		return nil
	}

	result.Fields = fields
	if shape == objectOutputShape {
		projected, err := projectTradeObjects(trades, fields)
		if err != nil {
			return err
		}
		result.Trades = projected
		return nil
	}

	rows, err := projectTradeRows(trades, fields)
	if err != nil {
		return err
	}
	result.Rows = rows
	return nil
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
			if value, ok := object[field]; ok {
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
			value, ok := object[field]
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

func fetchDisproportionatelyLargeTrades(ctx context.Context, tradeDate, tickers string, limit int) (getTradesResponse, error) {
	options := defaultGetTradesRequestOptions()
	return fetchTradesPages(ctx, tradeDate, tickers, &options, limit)
}

func fetchRankedTrades(ctx context.Context, tradeDate, tickers string, preset *rankedPreset, limit int) (getTradesResponse, error) {
	options := rankedGetTradesRequestOptions(preset)
	return fetchTradesPages(ctx, tradeDate, tickers, &options, limit)
}

func fetchSignalTrades(ctx context.Context, tradeDate, tickers string, preset *signalPreset, limit int) (getTradesResponse, error) {
	options := signalGetTradesRequestOptions(preset)
	return fetchTradesPages(ctx, tradeDate, tickers, &options, limit)
}

func fetchLeverageTrades(ctx context.Context, tradeDate, tickers string, preset *leveragePreset, limit int) (getTradesResponse, error) {
	options := leverageGetTradesRequestOptions(preset)
	return fetchTradesPages(ctx, tradeDate, tickers, &options, limit)
}

func fetchSectorTrades(ctx context.Context, tradeDate, tickers string, preset *sectorPreset, limit int) (getTradesResponse, error) {
	options := sectorGetTradesRequestOptions(preset)
	return fetchTradesPages(ctx, tradeDate, tickers, &options, limit)
}

func fetchTradesPages(ctx context.Context, tradeDate, tickers string, options *getTradesRequestOptions, limit int) (getTradesResponse, error) {
	if limit < 1 {
		return getTradesResponse{}, fmt.Errorf("fetch GetTrades pages: limit must be 1 or greater")
	}
	if limit > maxTradeLimit {
		return getTradesResponse{}, fmt.Errorf("fetch GetTrades pages: limit must be %d or less", maxTradeLimit)
	}

	var merged getTradesResponse
	merged.Data = []json.RawMessage{}
	for page := 0; len(merged.Data) < limit; page++ {
		select {
		case <-ctx.Done():
			return getTradesResponse{}, fmt.Errorf("fetch GetTrades page %d: %w", page+1, ctx.Err())
		default:
		}

		remaining := limit - len(merged.Data)
		pageLength := min(defaultTradePageSize, remaining)

		pageOptions := *options
		pageOptions.draw = page + 1
		pageOptions.start = page * defaultTradePageSize
		pageOptions.length = pageLength

		apiResponse, err := fetchTrades(ctx, tradeDate, tickers, &pageOptions)
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
		if apiResponse.RecordsFiltered > 0 && pageOptions.start+len(apiResponse.Data) >= apiResponse.RecordsFiltered {
			break
		}
	}
	if len(merged.Data) > limit {
		merged.Data = merged.Data[:limit]
	}

	return merged, nil
}

func fetchTrades(ctx context.Context, tradeDate, tickers string, options *getTradesRequestOptions) (getTradesResponse, error) {
	cookies, err := extractCookies(ctx)
	if err != nil {
		return getTradesResponse{}, fmt.Errorf("extract VolumeLeaders browser cookies: %w", err)
	}

	token, err := fetchXSRFToken(ctx, getTradesHTTPClient, cookies)
	if err != nil {
		return getTradesResponse{}, fmt.Errorf("fetch VolumeLeaders XSRF token: %w", err)
	}

	form := getTradesForm(tradeDate, tickers, options)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, getTradesEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return getTradesResponse{}, fmt.Errorf("create GetTrades request: %w", err)
	}
	setGetTradesHeaders(req, token, tradeDate, tickers, options)
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	resp, err := getTradesHTTPClient.Do(req)
	if err != nil {
		return getTradesResponse{}, fmt.Errorf("post GetTrades request: %w", err)
	}
	defer resp.Body.Close()

	if sessionExpiredResponse(resp) {
		return getTradesResponse{}, sessionExpiredCommandError()
	}
	if resp.StatusCode != http.StatusOK {
		return getTradesResponse{}, fmt.Errorf("GetTrades request returned status %d", resp.StatusCode)
	}

	bodyReader, closeReader, err := responseBodyReader(resp)
	if err != nil {
		return getTradesResponse{}, err
	}
	defer closeReader()

	var apiResponse getTradesResponse
	if err := json.NewDecoder(bodyReader).Decode(&apiResponse); err != nil {
		return getTradesResponse{}, fmt.Errorf("decode GetTrades response: %w", err)
	}
	if apiResponse.Error != "" {
		return getTradesResponse{}, fmt.Errorf("GetTrades response error: %s", apiResponse.Error)
	}
	if apiResponse.Data == nil {
		apiResponse.Data = []json.RawMessage{}
	}

	return apiResponse, nil
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
		length:                 preset.length,
		minVolume:              "10000",
		maxDollars:             "100000000000",
		conditions:             "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH",
		vcd:                    "0",
		relativeSize:           "0",
		darkPools:              "-1",
		includePhantom:         "-1",
		includeOffsetting:      "-1",
		sectorIndustry:         "",
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
		includePhantom:         preset.phantom,
		includeOffsetting:      preset.offsetting,
		sectorIndustry:         "",
		presetSearchTemplateID: preset.presetID,
	}
}

func leverageGetTradesRequestOptions(preset *leveragePreset) getTradesRequestOptions {
	return getTradesRequestOptions{
		tradeRank:              -1,
		draw:                   1,
		start:                  0,
		length:                 defaultTradeLimit,
		minVolume:              "10000",
		maxDollars:             "10000000000",
		conditions:             "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH",
		vcd:                    "97",
		relativeSize:           "5",
		darkPools:              "-1",
		includePhantom:         "-1",
		includeOffsetting:      "-1",
		sectorIndustry:         preset.sectorIndustry,
		presetSearchTemplateID: preset.presetID,
	}
}

func sectorGetTradesRequestOptions(preset *sectorPreset) getTradesRequestOptions {
	return getTradesRequestOptions{
		tradeRank:              -1,
		draw:                   1,
		start:                  0,
		length:                 defaultTradeLimit,
		minVolume:              "10000",
		maxDollars:             "10000000000",
		conditions:             "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH",
		vcd:                    "0",
		relativeSize:           "5",
		darkPools:              "-1",
		includePhantom:         "-1",
		includeOffsetting:      "-1",
		sectorIndustry:         preset.sectorIndustry,
		presetSearchTemplateID: preset.presetID,
	}
}

func setGetTradesHeaders(req *http.Request, token, tradeDate, tickers string, options *getTradesRequestOptions) {
	req.Header.Set("User-Agent", auth.UserAgent)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-XSRF-Token", token)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", "https://www.volumeleaders.com")
	req.Header.Set("Referer", tradesReferer(tradeDate, tickers, options))
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
}

func tradesReferer(tradeDate, tickers string, options *getTradesRequestOptions) string {
	query := url.Values{}
	setSharedQueryParams(query, tradeDate, tickers, options)
	query.Set("PresetSearchTemplateID", options.presetSearchTemplateID)
	query.Set("ViewMode", "Automatic")
	return tradesPage + "?" + query.Encode()
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
	values.Set("Sweeps", "-1")
	values.Set("LatePrints", "-1")
	values.Set("SignaturePrints", "-1")
	values.Set("EvenShared", "-1")
	values.Set("SecurityTypeKey", "-1")
	values.Set("MinPrice", "0")
	values.Set("MaxPrice", "100000")
	values.Set("MinDollars", "500000")
	values.Set("MaxDollars", options.maxDollars)
	values.Set("TradeRank", fmt.Sprintf("%d", options.tradeRank))
	values.Set("TradeRankSnapshot", "-1")
	values.Set("MarketCap", "0")
	values.Set("IncludePremarket", "1")
	values.Set("IncludeRTH", "1")
	values.Set("IncludeAH", "1")
	values.Set("IncludeOpening", "1")
	values.Set("IncludeClosing", "1")
	values.Set("IncludePhantom", options.includePhantom)
	values.Set("IncludeOffsetting", options.includeOffsetting)
	values.Set("SectorIndustry", options.sectorIndustry)
}

func getTradesForm(tradeDate, tickers string, options *getTradesRequestOptions) url.Values {
	form := url.Values{}
	draw := options.draw
	if draw == 0 {
		draw = 1
	}
	form.Set("draw", fmt.Sprintf("%d", draw))
	for i, column := range getTradesColumns {
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
	form.Set("order[0][name]", "FullTimeString24")
	form.Set("start", fmt.Sprintf("%d", options.start))
	form.Set("length", fmt.Sprintf("%d", options.length))
	form.Set("search[value]", "")
	form.Set("search[regex]", "false")
	setSharedQueryParams(form, tradeDate, tickers, options)
	return form
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

func responseBodyReader(resp *http.Response) (io.Reader, func(), error) {
	if resp.Header.Get("Content-Encoding") != "gzip" {
		return resp.Body, func() {}, nil
	}
	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, func() {}, fmt.Errorf("decompress GetTrades response: %w", err)
	}
	return gr, func() { _ = gr.Close() }, nil
}
