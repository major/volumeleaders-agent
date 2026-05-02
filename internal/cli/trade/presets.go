package trade

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"strings"

	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
)

type tradePreset struct {
	name    string
	group   string
	filters map[string]string
}

var tradePresets = buildPresets()

func buildPresets() []tradePreset {
	commonPreset := func(name string, extra map[string]string) tradePreset {
		f := map[string]string{"Conditions": "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH", "IncludeOffsetting": "-1", "IncludePhantom": "-1", "MaxDollars": "10000000000", "MinVolume": "10000", "RelativeSize": "0", "TradeCount": "3"}
		maps.Copy(f, extra)
		return tradePreset{name: name, group: "Common", filters: f}
	}
	large := func(name string, extra map[string]string) tradePreset {
		f := map[string]string{"Conditions": "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH", "IncludeOffsetting": "-1", "IncludePhantom": "-1", "MaxDollars": "10000000000", "MinVolume": "10000", "TradeCount": "3"}
		maps.Copy(f, extra)
		return tradePreset{name: name, group: "Disproportionately Large", filters: f}
	}

	return []tradePreset{
		commonPreset("All Trades", nil),
		commonPreset("Top-10 Rank", map[string]string{"TradeRank": "10"}),
		commonPreset("Top-100 Rank", map[string]string{"MaxDollars": "100000000000", "TradeRank": "100"}),
		{name: "Top-100 Rank; Dark Pool Sweeps", group: "Common", filters: map[string]string{"Conditions": "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH", "DarkPools": "1", "IncludeAH": "0", "IncludeClosing": "0", "IncludeOffsetting": "-1", "IncludeOpening": "0", "IncludePhantom": "0", "MaxDollars": "100000000000", "MinVolume": "10000", "RelativeSize": "0", "SignaturePrints": "0", "Sweeps": "1", "TradeCount": "3", "TradeRank": "100"}},
		commonPreset("Top-100 Rank; Leveraged ETFs", map[string]string{"MaxDollars": "1000000000000", "SectorIndustry": "X B", "TradeRank": "100"}),
		{name: "Top-100 Rank; RSI OB; >=5x Avg Size", group: "Common", filters: map[string]string{"Conditions": "OBD,OBH", "IncludeOffsetting": "-1", "IncludePhantom": "-1", "MaxDollars": "10000000000", "MinVolume": "10000", "SignaturePrints": "0", "TradeCount": "3", "TradeRank": "100"}},
		{name: "Top-100 Rank; RSI OS; >=5x Avg Size", group: "Common", filters: map[string]string{"Conditions": "OSD,OSH", "IncludeOffsetting": "-1", "IncludePhantom": "-1", "MaxDollars": "10000000000", "MinVolume": "10000", "SignaturePrints": "0", "TradeCount": "3", "TradeRank": "100"}},
		commonPreset("Top-100 Rank; >=20x avg size; DP Only", map[string]string{"DarkPools": "1", "RelativeSize": "20", "SignaturePrints": "0", "TradeRank": "100"}),
		commonPreset("Top-30 Rank; >10x avg size; 99th %", map[string]string{"RelativeSize": "10", "SignaturePrints": "0", "TradeRank": "30", "VCD": "99.00"}),
		{name: "Phantom Trades", group: "Common", filters: map[string]string{"Conditions": "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH", "DarkPools": "1", "IncludeAH": "0", "IncludeClosing": "0", "IncludeOffsetting": "0", "IncludeOpening": "0", "IncludePremarket": "0", "IncludeRTH": "0", "MaxDollars": "100000000000", "RelativeSize": "0", "SignaturePrints": "0", "TradeCount": "3"}},
		{name: "Offsetting Trades", group: "Common", filters: map[string]string{"Conditions": "IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH", "IncludeAH": "0", "IncludeClosing": "0", "IncludeOpening": "0", "IncludePhantom": "0", "IncludePremarket": "0", "IncludeRTH": "0", "MaxDollars": "100000000000", "RelativeSize": "0", "SignaturePrints": "0", "TradeCount": "3"}},
		large("All Disproportionately Large Trades", nil),
		large("Bear Leverage", map[string]string{"SectorIndustry": "X Bear", "VCD": "97.00"}),
		large("Biotechnology", map[string]string{"SectorIndustry": "Biotech"}),
		large("Bonds", map[string]string{"SectorIndustry": "Bonds"}),
		large("Bull Leverage", map[string]string{"SectorIndustry": "X Bull", "VCD": "97.00"}),
		large("China", map[string]string{"SectorIndustry": "China", "MaxDollars": "100000000000"}),
		large("Communication Services", map[string]string{"SectorIndustry": "Comm Services"}),
		large("Consumer Discretionary", map[string]string{"SectorIndustry": "Consumer Disc"}),
		large("Consumer Staples", map[string]string{"SectorIndustry": "Consumer Staples"}),
		large("Crypto", map[string]string{"SectorIndustry": "Crypto", "VCD": "97.00"}),
		large("Emerging Markets", map[string]string{"SectorIndustry": "Emerging Markets"}),
		large("Energy", map[string]string{"SectorIndustry": "Energy"}),
		large("Financials", map[string]string{"SectorIndustry": "Financial"}),
		large("Healthcare", map[string]string{"SectorIndustry": "Healthcare"}),
		large("Industrials", map[string]string{"SectorIndustry": "Industrials"}),
		large("Materials", map[string]string{"SectorIndustry": "Materials"}),
		large("Metals and Mining", map[string]string{"SectorIndustry": "Metals and Mining"}),
		large("Real Estate", map[string]string{"SectorIndustry": "Real Estate"}),
		large("Semiconductors", map[string]string{"SectorIndustry": "Semis"}),
		large("Technology", map[string]string{"SectorIndustry": "Technology"}),
		large("Utilities", map[string]string{"SectorIndustry": "Utilities"}),
		large("Commodities", map[string]string{"Tickers": "AGQ,BOIL,CORN,COPX,CPER,DBC,DJP,GLD,GLDM,IAU,KOLD,PPLT,SCO,SLV,SOYB,UCO,UGL,UNG,URA,USO,UUP,WEAT,ZSL", "VCD": "97.00"}),
		large("Electric Vehicles", map[string]string{"Tickers": "BLNK,F,GM,LI,NIO,NKLA,TSLA,WKHS,QS,LCID,RIVN,TSLQ,TSLL,TSLS,TSLY,TSDD", "VCD": "97.00"}),
		large("Megacaps", map[string]string{"Tickers": "AAPL,AMZN,META,GOOG,GOOGL,MSFT,NFLX,NVDA,TSLA", "VCD": "97.00"}),
		large("Meme Stocks", map[string]string{"Tickers": "AMC,BB,CLF,GME,NOK,SAVA,SPCE,TLRY,LOGC,CLOV,SOFI,BKKT,PUBM", "VCD": "97.00"}),
		large("Sector ETFs", map[string]string{"Tickers": "DGRO,EEM,GLD,IBB,ITOT,IVE,IVW,IVV,IWM,IWY,MDY,QQQ,RSP,SLV,SMH,SPYD,SPY,SPYV,SPYG,TLT,USO,XBI,XLE,XLK,XLP,XLI,XLF,XLC,XLY,XLV,XLU", "VCD": "97.00"}),
		large("SPY/QQQ Surrogates", map[string]string{"Tickers": "ACWI,DGRO,FBCG,FBCV,IWL,IWB,IVW,IVV,IWF,IWX,IWV,IWY,MGC,MGK,MGV,MTUM,OEF,PSQ,QLD,QID,QQQE,QQQ,QQEW,RSP,SCHG,SCHK,SCHV,SCHX,SDS,SH,SPYM,SPXS,SPXL,SPYD,SPY,SQQQ,SPYV,SPXU,SPYG,SSO,SUSA,TCHP,TQQQ,UDOW,UPRO,VFVA,VOO,VOOG,VOOV,VUG,VV,VTV,XLK,CGGR,JGRO,SPYU", "MaxDollars": "100000000000", "RelativeSize": "0"}),
		large("Volatility", map[string]string{"Tickers": "SVXY,UVXY,VIXY,VXX,SVIX,UVIX", "VCD": "97.00"}),
	}
}

func findPreset(name string) (*tradePreset, error) {
	for i := range tradePresets {
		if strings.EqualFold(tradePresets[i].name, name) {
			return &tradePresets[i], nil
		}
	}
	return nil, fmt.Errorf("preset %q not found; available presets: %s", name, strings.Join(presetNames(), ", "))
}

func presetNames() []string {
	names := make([]string, len(tradePresets))
	for i, p := range tradePresets {
		names[i] = p.name
	}
	return names
}

func applyExplicitFlags(cmd *cobra.Command, filters map[string]string) {
	stringFlags := [][2]string{{"tickers", "Tickers"}, {"sector", "SectorIndustry"}}
	for _, sf := range stringFlags {
		if cmd.Flags().Changed(sf[0]) {
			filters[sf[1]] = getString(cmd, sf[0])
		}
	}
	intFlags := [][2]string{{"conditions", "Conditions"}, {"vcd", "VCD"}, {"security-type", "SecurityTypeKey"}, {"relative-size", "RelativeSize"}, {"dark-pools", "DarkPools"}, {"sweeps", "Sweeps"}, {"late-prints", "LatePrints"}, {"sig-prints", "SignaturePrints"}, {"even-shared", "EvenShared"}, {"trade-rank", "TradeRank"}, {"rank-snapshot", "TradeRankSnapshot"}, {"market-cap", "MarketCap"}, {"premarket", "IncludePremarket"}, {"rth", "IncludeRTH"}, {"ah", "IncludeAH"}, {"opening", "IncludeOpening"}, {"closing", "IncludeClosing"}, {"phantom", "IncludePhantom"}, {"offsetting", "IncludeOffsetting"}, {"min-volume", "MinVolume"}, {"max-volume", "MaxVolume"}}
	for _, inf := range intFlags {
		if cmd.Flags().Changed(inf[0]) {
			filters[inf[1]] = common.IntStr(getInt(cmd, inf[0]))
		}
	}
	floatFlags := [][2]string{{"min-price", "MinPrice"}, {"max-price", "MaxPrice"}, {"min-dollars", "MinDollars"}, {"max-dollars", "MaxDollars"}}
	for _, ff := range floatFlags {
		if cmd.Flags().Changed(ff[0]) {
			filters[ff[1]] = common.FormatFloat(getFloat(cmd, ff[0]))
		}
	}
}

func fetchWatchlistFilters(ctx context.Context, name string) (map[string]string, error) {
	vlClient, err := common.NewCommandClient(ctx)
	if err != nil {
		return nil, err
	}
	request := common.NewDataTablesRequest(datatables.WatchlistConfigColumns, common.DataTableOptions{Start: 0, Length: -1, OrderCol: 1, OrderDir: "asc"})
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
		filters["SecurityTypeKey"] = common.IntStr(cfg.SecurityTypeKey)
	}
	if cfg.MinVCD > 0 {
		filters["VCD"] = common.FormatFloat(cfg.MinVCD)
	}
	if cfg.MinRelativeSizeSelected != nil && *cfg.MinRelativeSizeSelected {
		filters["RelativeSize"] = common.IntStr(cfg.MinRelativeSize)
	}
	if cfg.MaxTradeRankSelected != nil && *cfg.MaxTradeRankSelected {
		filters["TradeRank"] = common.IntStr(cfg.MaxTradeRank)
	}
	toggles := []struct {
		selected, value bool
		key             string
	}{{cfg.DarkPoolsSelected, cfg.DarkPools, "DarkPools"}, {cfg.SweepsSelected, cfg.Sweeps, "Sweeps"}, {cfg.SignaturePrintsSelected, cfg.SignaturePrints, "SignaturePrints"}, {cfg.LatePrintsSelected, cfg.LatePrints, "LatePrints"}, {cfg.PremarketTradesSelected, cfg.PremarketTrades, "IncludePremarket"}, {cfg.RTHTradesSelected, cfg.RTHTrades, "IncludeRTH"}, {cfg.AHTradesSelected, cfg.AHTrades, "IncludeAH"}, {cfg.OpeningTradesSelected, cfg.OpeningTrades, "IncludeOpening"}, {cfg.ClosingTradesSelected, cfg.ClosingTrades, "IncludeClosing"}, {cfg.PhantomTradesSelected, cfg.PhantomTrades, "IncludePhantom"}, {cfg.OffsettingTradesSelected, cfg.OffsettingTrades, "IncludeOffsetting"}}
	for _, t := range toggles {
		if t.selected {
			if t.value {
				filters[t.key] = "1"
			} else {
				filters[t.key] = "0"
			}
		}
	}
	if cfg.MinVolume > 0 {
		filters["MinVolume"] = common.IntStr(cfg.MinVolume)
	}
	if cfg.MaxVolume > 0 && cfg.MaxVolume < 2000000000 {
		filters["MaxVolume"] = common.IntStr(cfg.MaxVolume)
	}
	if cfg.MinDollars > 0 {
		filters["MinDollars"] = common.FormatFloat(cfg.MinDollars)
	}
	if cfg.MaxDollars > 0 && cfg.MaxDollars < 30000000000 {
		filters["MaxDollars"] = common.FormatFloat(cfg.MaxDollars)
	}
	if cfg.MinPrice > 0 {
		filters["MinPrice"] = common.FormatFloat(cfg.MinPrice)
	}
	if cfg.MaxPrice > 0 && cfg.MaxPrice < 100000 {
		filters["MaxPrice"] = common.FormatFloat(cfg.MaxPrice)
	}
	return filters
}

func newTradePresetTickersCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "preset-tickers", Short: "Extract ticker symbols from a preset", Example: "volumeleaders-agent trade preset-tickers --preset NAME", RunE: runTradePresetTickers}
	cmd.Flags().String("preset", "", "Preset name (case-insensitive)")
	_ = cmd.MarkFlagRequired("preset")
	return cmd
}

func runTradePresetTickers(cmd *cobra.Command, _ []string) error {
	p, err := findPreset(getString(cmd, "preset"))
	if err != nil {
		return err
	}
	info := models.PresetTickersInfo{Preset: p.name, Group: p.group}
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
	return common.PrintJSON(cmd.OutOrStdout(), cmd.Context(), info)
}

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

func newTradePresetsCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "presets", Short: "List available trade filter presets", Example: "volumeleaders-agent trade presets", RunE: runTradePresets}
	common.AddOutputFormatFlags(cmd)
	return cmd
}

func runTradePresets(cmd *cobra.Command, _ []string) error {
	format, err := common.ParseOutputFormat(getString(cmd, "format"))
	if err != nil {
		return err
	}
	presets := make([]models.PresetInfo, len(tradePresets))
	for i, p := range tradePresets {
		presets[i] = models.PresetInfo{Name: p.name, Group: p.group, Filters: p.filters}
	}
	return common.PrintDataTablesResult(cmd.OutOrStdout(), cmd.Context(), presets, nil, format)
}
