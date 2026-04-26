package commands

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
	cli "github.com/urfave/cli/v3"
)

// Trade filter presets from the VolumeLeaders /Trades page.
//
// Source: the <select id="PresetSearchTemplates"> dropdown rendered
// server-side in the /Trades HTML page. Each <option> value is a URL
// whose query parameters set filters on the DataTables trade request.
//
// To refresh this list:
//  1. Open https://www.volumeleaders.com/Trades in a browser.
//  2. Inspect the <select id="PresetSearchTemplates"> element.
//  3. Extract each <option> value's query parameters.
//  4. Update the presets below, keeping only parameters that differ
//     from the trade list CLI defaults.
//
// The dropdown has three optgroups: "Common", "My Watch Lists"
// (user-created, fetched at runtime via --watchlist), and
// "Disproportionately Large (>=5x Avg Size)".
//
// The list rarely changes. Last updated: 2025-04-24.

type tradePreset struct {
	name    string
	group   string
	filters map[string]string
}

var tradePresets = buildPresets()

// buildPresets constructs the full preset list. Helper closures reduce
// repetition for the two preset groups that share a common filter base.
func buildPresets() []tradePreset {
	// common builds a "Common" preset. The base filters match the website's
	// default trade query: ignore RSI conditions, include phantom and
	// offsetting, cap dollars at 10B, require 10k min volume, drop the
	// relative-size floor to 0, and require 3+ trade count.
	common := func(name string, extra map[string]string) tradePreset {
		f := map[string]string{
			"Conditions":        "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH",
			"IncludeOffsetting": "-1",
			"IncludePhantom":    "-1",
			"MaxDollars":        "10000000000",
			"MinVolume":         "10000",
			"RelativeSize":      "0",
			"TradeCount":        "3",
		}
		for k, v := range extra {
			f[k] = v
		}
		return tradePreset{name: name, group: "Common", filters: f}
	}

	// dpl builds a "Disproportionately Large" preset. Same base as common
	// but WITHOUT RelativeSize=0, so the CLI default of 5 (>=5x avg size)
	// stays in effect.
	dpl := func(name string, extra map[string]string) tradePreset {
		f := map[string]string{
			"Conditions":        "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH",
			"IncludeOffsetting": "-1",
			"IncludePhantom":    "-1",
			"MaxDollars":        "10000000000",
			"MinVolume":         "10000",
			"TradeCount":        "3",
		}
		for k, v := range extra {
			f[k] = v
		}
		return tradePreset{name: name, group: "Disproportionately Large", filters: f}
	}

	return []tradePreset{
		// ---- Common ----
		common("All Trades", nil),
		common("Top-10 Rank", map[string]string{
			"TradeRank": "10",
		}),
		common("Top-100 Rank", map[string]string{
			"MaxDollars": "100000000000",
			"TradeRank":  "100",
		}),
		{name: "Top-100 Rank; Dark Pool Sweeps", group: "Common", filters: map[string]string{
			"Conditions":        "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH",
			"DarkPools":         "1",
			"IncludeAH":         "0",
			"IncludeClosing":    "0",
			"IncludeOffsetting": "-1",
			"IncludeOpening":    "0",
			"IncludePhantom":    "0",
			"MaxDollars":        "100000000000",
			"MinVolume":         "10000",
			"RelativeSize":      "0",
			"SignaturePrints":   "0",
			"Sweeps":            "1",
			"TradeCount":        "3",
			"TradeRank":         "100",
		}},
		common("Top-100 Rank; Leveraged ETFs", map[string]string{
			"MaxDollars":     "1000000000000",
			"SectorIndustry": "X B",
			"TradeRank":      "100",
		}),
		// RSI Overbought: different Conditions, no RelativeSize override
		// (keeps CLI default of 5), adds SignaturePrints=0.
		{name: "Top-100 Rank; RSI OB; >=5x Avg Size", group: "Common", filters: map[string]string{
			"Conditions":        "OBD,OBH",
			"IncludeOffsetting": "-1",
			"IncludePhantom":    "-1",
			"MaxDollars":        "10000000000",
			"MinVolume":         "10000",
			"SignaturePrints":   "0",
			"TradeCount":        "3",
			"TradeRank":         "100",
		}},
		// RSI Oversold: mirror of RSI OB with oversold conditions.
		{name: "Top-100 Rank; RSI OS; >=5x Avg Size", group: "Common", filters: map[string]string{
			"Conditions":        "OSD,OSH",
			"IncludeOffsetting": "-1",
			"IncludePhantom":    "-1",
			"MaxDollars":        "10000000000",
			"MinVolume":         "10000",
			"SignaturePrints":   "0",
			"TradeCount":        "3",
			"TradeRank":         "100",
		}},
		common("Top-100 Rank; >=20x avg size; DP Only", map[string]string{
			"DarkPools":       "1",
			"RelativeSize":    "20",
			"SignaturePrints": "0",
			"TradeRank":       "100",
		}),
		common("Top-30 Rank; >10x avg size; 99th %", map[string]string{
			"RelativeSize":    "10",
			"SignaturePrints": "0",
			"TradeRank":       "30",
			"VCD":             "99.00",
		}),
		// Phantom Trades: dark pool only, exclude all session types except
		// phantom (which stays at CLI default 1).
		{name: "Phantom Trades", group: "Common", filters: map[string]string{
			"Conditions":        "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH",
			"DarkPools":         "1",
			"IncludeAH":         "0",
			"IncludeClosing":    "0",
			"IncludeOffsetting": "0",
			"IncludeOpening":    "0",
			"IncludePremarket":  "0",
			"IncludeRTH":        "0",
			"MaxDollars":        "100000000000",
			"RelativeSize":      "0",
			"SignaturePrints":   "0",
			"TradeCount":        "3",
		}},
		// Offsetting Trades: exclude all session types except offsetting
		// (which stays at CLI default 1).
		{name: "Offsetting Trades", group: "Common", filters: map[string]string{
			"Conditions":       "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH",
			"IncludeAH":        "0",
			"IncludeClosing":   "0",
			"IncludeOpening":   "0",
			"IncludePhantom":   "0",
			"IncludePremarket": "0",
			"IncludeRTH":       "0",
			"MaxDollars":       "100000000000",
			"RelativeSize":     "0",
			"SignaturePrints":  "0",
			"TradeCount":       "3",
		}},

		// ---- Disproportionately Large (>=5x Avg Size) ----
		// Sector-based presets filter by SectorIndustry value.
		dpl("All Disproportionately Large Trades", nil),
		dpl("Bear Leverage", map[string]string{"SectorIndustry": "X Bear", "VCD": "97.00"}),
		dpl("Biotechnology", map[string]string{"SectorIndustry": "Biotech"}),
		dpl("Bonds", map[string]string{"SectorIndustry": "Bonds"}),
		dpl("Bull Leverage", map[string]string{"SectorIndustry": "X Bull", "VCD": "97.00"}),
		dpl("China", map[string]string{"SectorIndustry": "China", "MaxDollars": "100000000000"}),
		dpl("Communication Services", map[string]string{"SectorIndustry": "Comm Services"}),
		dpl("Consumer Discretionary", map[string]string{"SectorIndustry": "Consumer Disc"}),
		dpl("Consumer Staples", map[string]string{"SectorIndustry": "Consumer Staples"}),
		dpl("Crypto", map[string]string{"SectorIndustry": "Crypto", "VCD": "97.00"}),
		dpl("Emerging Markets", map[string]string{"SectorIndustry": "Emerging Markets"}),
		dpl("Energy", map[string]string{"SectorIndustry": "Energy"}),
		dpl("Financials", map[string]string{"SectorIndustry": "Financial"}),
		dpl("Healthcare", map[string]string{"SectorIndustry": "Healthcare"}),
		dpl("Industrials", map[string]string{"SectorIndustry": "Industrials"}),
		dpl("Materials", map[string]string{"SectorIndustry": "Materials"}),
		dpl("Metals and Mining", map[string]string{"SectorIndustry": "Metals and Mining"}),
		dpl("Real Estate", map[string]string{"SectorIndustry": "Real Estate"}),
		dpl("Semiconductors", map[string]string{"SectorIndustry": "Semis"}),
		dpl("Technology", map[string]string{"SectorIndustry": "Technology"}),
		dpl("Utilities", map[string]string{"SectorIndustry": "Utilities"}),

		// Ticker-based presets filter by explicit ticker lists.
		dpl("Commodities", map[string]string{
			"Tickers": "AGQ,BOIL,CORN,COPX,CPER,DBC,DJP,GLD,GLDM,IAU,KOLD,PPLT,SCO,SLV,SOYB,UCO,UGL,UNG,URA,USO,UUP,WEAT,ZSL",
			"VCD":     "97.00",
		}),
		dpl("Electric Vehicles", map[string]string{
			"Tickers": "BLNK,F,GM,LI,NIO,NKLA,TSLA,WKHS,QS,LCID,RIVN,TSLQ,TSLL,TSLS,TSLY,TSDD",
			"VCD":     "97.00",
		}),
		dpl("Megacaps", map[string]string{
			"Tickers": "AAPL,AMZN,META,GOOG,GOOGL,MSFT,NFLX,NVDA,TSLA",
			"VCD":     "97.00",
		}),
		dpl("Meme Stocks", map[string]string{
			"Tickers": "AMC,BB,CLF,GME,NOK,SAVA,SPCE,TLRY,LOGC,CLOV,SOFI,BKKT,PUBM",
			"VCD":     "97.00",
		}),
		dpl("Sector ETFs", map[string]string{
			"Tickers": "DGRO,EEM,GLD,IBB,ITOT,IVE,IVW,IVV,IWM,IWY,MDY,QQQ,RSP,SLV,SMH,SPYD,SPY,SPYV,SPYG,TLT,USO,XBI,XLE,XLK,XLP,XLI,XLF,XLC,XLY,XLV,XLU",
			"VCD":     "97.00",
		}),
		dpl("SPY/QQQ Surrogates", map[string]string{
			"Tickers":      "ACWI,DGRO,FBCG,FBCV,IWL,IWB,IVW,IVV,IWF,IWX,IWV,IWY,MGC,MGK,MGV,MTUM,OEF,PSQ,QLD,QID,QQQE,QQQ,QQEW,RSP,SCHG,SCHK,SCHV,SCHX,SDS,SH,SPYM,SPXS,SPXL,SPYD,SPY,SQQQ,SPYV,SPXU,SPYG,SSO,SUSA,TCHP,TQQQ,UDOW,UPRO,VFVA,VOO,VOOG,VOOV,VUG,VV,VTV,XLK,CGGR,JGRO,SPYU",
			"MaxDollars":   "100000000000",
			"RelativeSize": "0",
		}),
		dpl("Volatility", map[string]string{
			"Tickers": "SVXY,UVXY,VIXY,VXX,SVIX,UVIX",
			"VCD":     "97.00",
		}),
	}
}

// findPreset returns the preset matching name (case-insensitive).
func findPreset(name string) (*tradePreset, error) {
	for i := range tradePresets {
		if strings.EqualFold(tradePresets[i].name, name) {
			return &tradePresets[i], nil
		}
	}

	// Build a list of available names for the error message.
	names := make([]string, len(tradePresets))
	for i, p := range tradePresets {
		names[i] = p.name
	}
	return nil, fmt.Errorf("preset %q not found; available presets: %s", name, strings.Join(names, ", "))
}

// applyExplicitFlags overwrites entries in filters for any trade list CLI
// flag the user explicitly provided. Flags that were not set on the command
// line are left alone, allowing preset/watchlist values to remain.
func applyExplicitFlags(cmd *cli.Command, filters map[string]string) {
	stringFlags := [][2]string{
		{"tickers", "Tickers"},
		{"sector", "SectorIndustry"},
	}
	for _, sf := range stringFlags {
		if cmd.IsSet(sf[0]) {
			filters[sf[1]] = cmd.String(sf[0])
		}
	}

	intFlags := [][2]string{
		{"conditions", "Conditions"},
		{"vcd", "VCD"},
		{"security-type", "SecurityTypeKey"},
		{"relative-size", "RelativeSize"},
		{"dark-pools", "DarkPools"},
		{"sweeps", "Sweeps"},
		{"late-prints", "LatePrints"},
		{"sig-prints", "SignaturePrints"},
		{"even-shared", "EvenShared"},
		{"trade-rank", "TradeRank"},
		{"rank-snapshot", "TradeRankSnapshot"},
		{"market-cap", "MarketCap"},
		{"premarket", "IncludePremarket"},
		{"rth", "IncludeRTH"},
		{"ah", "IncludeAH"},
		{"opening", "IncludeOpening"},
		{"closing", "IncludeClosing"},
		{"phantom", "IncludePhantom"},
		{"offsetting", "IncludeOffsetting"},
		{"min-volume", "MinVolume"},
		{"max-volume", "MaxVolume"},
	}
	for _, inf := range intFlags {
		if cmd.IsSet(inf[0]) {
			filters[inf[1]] = intStr(cmd.Int(inf[0]))
		}
	}

	floatFlags := [][2]string{
		{"min-price", "MinPrice"},
		{"max-price", "MaxPrice"},
		{"min-dollars", "MinDollars"},
		{"max-dollars", "MaxDollars"},
	}
	for _, ff := range floatFlags {
		if cmd.IsSet(ff[0]) {
			filters[ff[1]] = formatFloat(cmd.Float(ff[0]))
		}
	}
}

// fetchWatchlistFilters fetches the user's watchlist configs and returns
// trade filter values derived from the config matching name (case-insensitive).
func fetchWatchlistFilters(ctx context.Context, name string) (map[string]string, error) {
	vlClient, err := newCommandClient(ctx)
	if err != nil {
		return nil, err
	}

	request := newDataTablesRequest(datatables.WatchlistConfigColumns, dataTableOptions{
		start: 0, length: -1, orderCol: 1, orderDir: "asc",
	})

	var configs []models.WatchListConfig
	if err := vlClient.PostDataTables(ctx, "/WatchListConfigs/GetWatchLists", request.Encode(), &configs); err != nil {
		slog.Error("failed to fetch watchlists", "error", err)
		return nil, fmt.Errorf("fetch watchlists: %w", err)
	}

	var match *models.WatchListConfig
	for i := range configs {
		if strings.EqualFold(configs[i].Name, name) {
			match = &configs[i]
			break
		}
	}
	if match == nil {
		names := make([]string, len(configs))
		for i := range configs {
			names[i] = configs[i].Name
		}
		return nil, fmt.Errorf("watchlist %q not found; available watchlists: %s", name, strings.Join(names, ", "))
	}

	return watchlistConfigToFilters(match), nil
}

// watchlistConfigToFilters converts a WatchListConfig into trade filter
// key-value pairs. Only fields that are explicitly configured (via the
// *Selected booleans or non-zero values) are included.
func watchlistConfigToFilters(cfg *models.WatchListConfig) map[string]string {
	filters := make(map[string]string)

	if cfg.Tickers != "" {
		filters["Tickers"] = cfg.Tickers
	}
	if cfg.SectorIndustry != nil && *cfg.SectorIndustry != "" {
		filters["SectorIndustry"] = *cfg.SectorIndustry
	}
	if cfg.Conditions != "" {
		filters["Conditions"] = cfg.Conditions
	}
	if cfg.SecurityTypeKey > 0 {
		filters["SecurityTypeKey"] = intStr(cfg.SecurityTypeKey)
	}
	if cfg.MinVCD > 0 {
		filters["VCD"] = formatFloat(cfg.MinVCD)
	}

	// Relative size and trade rank use *Selected to indicate explicit config.
	if cfg.MinRelativeSizeSelected != nil && *cfg.MinRelativeSizeSelected {
		filters["RelativeSize"] = intStr(cfg.MinRelativeSize)
	}
	if cfg.MaxTradeRankSelected != nil && *cfg.MaxTradeRankSelected {
		filters["TradeRank"] = intStr(cfg.MaxTradeRank)
	}

	// Boolean toggle filters - apply only when the "Selected" flag is true,
	// meaning the user explicitly configured this toggle in the watchlist.
	type boolToggle struct {
		selected bool
		value    bool
		key      string
	}
	toggles := []boolToggle{
		{cfg.DarkPoolsSelected, cfg.DarkPools, "DarkPools"},
		{cfg.SweepsSelected, cfg.Sweeps, "Sweeps"},
		{cfg.SignaturePrintsSelected, cfg.SignaturePrints, "SignaturePrints"},
		{cfg.LatePrintsSelected, cfg.LatePrints, "LatePrints"},
		{cfg.PremarketTradesSelected, cfg.PremarketTrades, "IncludePremarket"},
		{cfg.RTHTradesSelected, cfg.RTHTrades, "IncludeRTH"},
		{cfg.AHTradesSelected, cfg.AHTrades, "IncludeAH"},
		{cfg.OpeningTradesSelected, cfg.OpeningTrades, "IncludeOpening"},
		{cfg.ClosingTradesSelected, cfg.ClosingTrades, "IncludeClosing"},
		{cfg.PhantomTradesSelected, cfg.PhantomTrades, "IncludePhantom"},
		{cfg.OffsettingTradesSelected, cfg.OffsettingTrades, "IncludeOffsetting"},
	}
	for _, t := range toggles {
		if t.selected {
			if t.value {
				filters[t.key] = "1"
			} else {
				filters[t.key] = "0"
			}
		}
	}

	// Numeric range filters - include when meaningfully different from zero
	// (or from the very large defaults for max values).
	if cfg.MinVolume > 0 {
		filters["MinVolume"] = intStr(cfg.MinVolume)
	}
	if cfg.MaxVolume > 0 && cfg.MaxVolume < 2000000000 {
		filters["MaxVolume"] = intStr(cfg.MaxVolume)
	}
	if cfg.MinDollars > 0 {
		filters["MinDollars"] = formatFloat(cfg.MinDollars)
	}
	if cfg.MaxDollars > 0 && cfg.MaxDollars < 30000000000 {
		filters["MaxDollars"] = formatFloat(cfg.MaxDollars)
	}
	if cfg.MinPrice > 0 {
		filters["MinPrice"] = formatFloat(cfg.MinPrice)
	}
	if cfg.MaxPrice > 0 && cfg.MaxPrice < 100000 {
		filters["MaxPrice"] = formatFloat(cfg.MaxPrice)
	}

	return filters
}

// newTradePresetTickersCommand returns the "trade preset-tickers" subcommand
// that extracts ticker symbols from a named preset.
func newTradePresetTickersCommand() *cli.Command {
	return &cli.Command{
		Name:      "preset-tickers",
		Usage:     "Extract ticker symbols from a preset",
		UsageText: "volumeleaders-agent trade preset-tickers --preset NAME",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "preset",
				Usage:    "Preset name (case-insensitive)",
				Required: true,
			},
		},
		Action: runTradePresetTickers,
	}
}

func runTradePresetTickers(ctx context.Context, cmd *cli.Command) error {
	p, err := findPreset(cmd.String("preset"))
	if err != nil {
		return err
	}

	info := models.PresetTickersInfo{
		Preset: p.name,
		Group:  p.group,
	}

	// Explicit ticker presets take precedence over sector filters if both are set.
	switch {
	case p.filters["Tickers"] != "":
		info.Type = "tickers"
		info.Tickers = splitTickers(p.filters["Tickers"])
	case p.filters["SectorIndustry"] != "":
		info.Type = "sector-filter"
		info.SectorIndustry = p.filters["SectorIndustry"]
	default:
		info.Type = "unfiltered"
	}

	return printJSON(ctx, info)
}

// splitTickers parses a comma-separated ticker list defensively so preset
// typos do not leak whitespace, empty symbols, or duplicate symbols to output.
func splitTickers(tickers string) []string {
	parts := strings.Split(tickers, ",")
	result := make([]string, 0, len(parts))
	seen := make(map[string]bool, len(parts))

	for _, part := range parts {
		ticker := strings.TrimSpace(part)
		if ticker == "" || seen[ticker] {
			continue
		}

		seen[ticker] = true
		result = append(result, ticker)
	}

	return result
}

// newTradePresetsCommand returns the "trade presets" subcommand that lists
// all available built-in filter presets.
func newTradePresetsCommand() *cli.Command {
	return &cli.Command{
		Name:      "presets",
		Usage:     "List available trade filter presets",
		UsageText: "volumeleaders-agent trade presets",
		Flags:     outputFormatFlags(),
		Action:    runTradePresets,
	}
}

func runTradePresets(ctx context.Context, cmd *cli.Command) error {
	format, err := parseOutputFormat(cmd.String("format"))
	if err != nil {
		return err
	}

	presets := make([]models.PresetInfo, len(tradePresets))
	for i, p := range tradePresets {
		presets[i] = models.PresetInfo{
			Name:    p.name,
			Group:   p.group,
			Filters: p.filters,
		}
	}
	return printDataTablesResult(ctx, presets, nil, format)
}
