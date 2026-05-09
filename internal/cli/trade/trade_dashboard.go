package trade

import (
	"fmt"
	"log/slog"
	"strings"

	vlgo "github.com/major/volumeleaders-go/volumeleaders"
	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/models"
)

const dashboardSummaryMaxRows = 3

type dashboardSection string

const (
	dashboardSectionTrades       dashboardSection = "trades"
	dashboardSectionClusters     dashboardSection = "clusters"
	dashboardSectionLevels       dashboardSection = "levels"
	dashboardSectionClusterBombs dashboardSection = "clusterBombs"
)

var allDashboardSections = []dashboardSection{dashboardSectionTrades, dashboardSectionClusters, dashboardSectionLevels, dashboardSectionClusterBombs}

type dashboardSectionSet map[dashboardSection]bool

// Chart column layouts mirroring the compact shapes VolumeLeaders uses
// in its browser dashboard charts. These are narrower than the standard
// leaderboard columns and are shared by dashboard and chart commands.

// tradeChartColumns is the compact chart DataTables layout VolumeLeaders uses
// for long-period ticker trade lookups.
var tradeChartColumns = []vlgo.DataTablesColumn{
	{Data: "FullTimeString24", Name: "FullTimeString24", Searchable: true},
	{Data: "Volume", Name: "Sh", Searchable: true},
	{Data: "Price", Name: "Price", Searchable: true},
	{Data: "Dollars", Name: "$$", Searchable: true},
	{Data: "DollarsMultiplier", Name: "RS", Searchable: true},
	{Data: "TradeRank", Name: "R", Searchable: true},
	{Data: "LastComparibleTradeDate", Name: "Last Comp", Searchable: true},
}

// tradeClusterChartColumns is the compact chart DataTables layout used by
// the browser dashboard for ticker-specific cluster summaries.
var tradeClusterChartColumns = []vlgo.DataTablesColumn{
	{Data: "MinFullTimeString24", Name: "MinFullTimeString24", Searchable: true},
	{Data: "Price", Name: "Price", Searchable: true},
	{Data: "TradeCount", Name: "TradeCount", Searchable: true},
	{Data: "Volume", Name: "Sh", Searchable: true},
	{Data: "Dollars", Name: "$$", Searchable: true},
	{Data: "DollarsMultiplier", Name: "RS", Searchable: true},
	{Data: "TradeClusterRank", Name: "R", Searchable: true},
	{Data: "LastComparibleTradeClusterDate", Name: "Last Comp", Searchable: true},
}

// tradeLevelChartColumns is the chart dashboard layout for level rows.
var tradeLevelChartColumns = []vlgo.DataTablesColumn{
	{Data: "Price", Name: "Price", Searchable: true},
	{Data: "Dollars", Name: "$$", Searchable: true},
	{Data: "Volume", Name: "Sh", Searchable: true},
	{Data: "Trades", Name: "Trades", Searchable: true},
	{Data: "RelativeSize", Name: "RS", Searchable: true},
	{Data: "CumulativeDistribution", Name: "PCT", Searchable: true},
	{Data: "TradeLevelRank", Name: "Rank", Searchable: true},
	{Data: "Dates", Name: "Dates", Searchable: true},
}

// tradeClusterBombChartColumns adapts the browser cluster table shape for the
// cluster-bomb endpoint so dashboard output can include the same burst context.
var tradeClusterBombChartColumns = []vlgo.DataTablesColumn{
	{Data: "MinFullTimeString24", Name: "MinFullTimeString24", Searchable: true},
	{Data: "TradeCount", Name: "TradeCount", Searchable: true},
	{Data: "Volume", Name: "Sh", Searchable: true},
	{Data: "Dollars", Name: "$$", Searchable: true},
	{Data: "DollarsMultiplier", Name: "RS", Searchable: true},
	{Data: "CumulativeDistribution", Name: "PCT", Searchable: true},
	{Data: "TradeClusterBombRank", Name: "R", Searchable: true},
	{Data: "LastComparableTradeClusterBombDate", Name: "Last Comp", Searchable: true},
}

// newDashboardDTRequest builds a vlgo DataTables request for dashboard/chart queries
// with compact column layouts, custom ordering, and search enabled.
func newDashboardDTRequest(columns []vlgo.DataTablesColumn, orderColumn int, orderName string, count int) vlgo.DataTablesRequest {
	return vlgo.DataTablesRequest{
		Draw:          1,
		Start:         0,
		Length:        count,
		Columns:       columns,
		Order:         []vlgo.DataTablesOrder{{Column: orderColumn, Dir: "DESC", Name: orderName}},
		IncludeSearch: true,
	}
}

func runTradeDashboard(cmd *cobra.Command, opts *tradeDashboardOptions) error {
	startDate, endDate, err := common.ResolveDateRange(cmd, tradeListTickerLookbackDays, false)
	if err != nil {
		return err
	}
	sections, err := parseDashboardSections(opts.Sections)
	if err != nil {
		return err
	}
	ticker, err := common.SingleTickerValue(cmd)
	if err != nil {
		return err
	}
	vlClient, err := common.NewVLClient(cmd.Context())
	if err != nil {
		return err
	}

	requestCount := opts.Count
	outputCount := dashboardOutputCount(opts)
	var trades []models.Trade
	var clusters []models.TradeCluster
	var levels []models.TradeLevel
	var clusterBombs []models.TradeClusterBomb
	returnedSections := make([]string, 0, len(sections))
	if sections.contains(dashboardSectionTrades) {
		trades, err = fetchDashboardTrades(cmd, vlClient, dashboardTradeFilters(opts, ticker, startDate, endDate), requestCount)
		if err != nil {
			return err
		}
		returnedSections = append(returnedSections, string(dashboardSectionTrades))
	}
	if sections.contains(dashboardSectionClusters) {
		clusters, err = fetchDashboardClusters(cmd, vlClient, dashboardClusterFilters(opts, ticker, startDate, endDate), requestCount)
		if err != nil {
			return err
		}
		returnedSections = append(returnedSections, string(dashboardSectionClusters))
	}
	if sections.contains(dashboardSectionLevels) {
		levels, err = fetchDashboardLevels(cmd, vlClient, dashboardLevelFilters(ticker, startDate, endDate, requestCount), requestCount)
		if err != nil {
			return err
		}
		returnedSections = append(returnedSections, string(dashboardSectionLevels))
	}
	if sections.contains(dashboardSectionClusterBombs) {
		clusterBombs, err = fetchDashboardClusterBombs(cmd, vlClient, dashboardClusterBombFilters(opts, ticker, startDate, endDate), requestCount)
		if err != nil {
			return err
		}
		returnedSections = append(returnedSections, string(dashboardSectionClusterBombs))
	}

	requestedSections := sections.names()
	dateRange := models.TradeDashboardDateRange{Start: startDate, End: endDate}
	if opts.Summary {
		summary := models.NewTradeDashboardSummary(
			ticker,
			dateRange,
			outputCount,
			requestedSections,
			returnedSections,
			limitDashboardRows(trades, outputCount),
			limitDashboardRows(clusters, outputCount),
			limitDashboardRows(levels, outputCount),
			limitDashboardRows(clusterBombs, outputCount),
		)
		return common.PrintJSON(cmd.OutOrStdout(), cmd.Context(), summary)
	}

	dashboard := models.NewTradeDashboard(
		ticker,
		dateRange,
		requestCount,
		requestedSections,
		returnedSections,
		trades,
		clusters,
		levels,
		clusterBombs,
	)
	return common.PrintJSON(cmd.OutOrStdout(), cmd.Context(), dashboard)
}

func parseDashboardSections(value string) (dashboardSectionSet, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || strings.EqualFold(trimmed, "all") {
		return newDashboardSectionSet(allDashboardSections), nil
	}
	sections := make(dashboardSectionSet)
	for part := range strings.SplitSeq(trimmed, ",") {
		section, err := parseDashboardSection(part)
		if err != nil {
			return nil, err
		}
		sections[section] = true
	}
	if len(sections) == 0 {
		return nil, fmt.Errorf("--sections must include at least one dashboard section")
	}
	return sections, nil
}

func parseDashboardSection(value string) (dashboardSection, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.NewReplacer("_", "-", " ", "-").Replace(normalized)
	switch normalized {
	case "trades":
		return dashboardSectionTrades, nil
	case "clusters":
		return dashboardSectionClusters, nil
	case "levels":
		return dashboardSectionLevels, nil
	case "cluster-bombs", "clusterbombs", "bombs":
		return dashboardSectionClusterBombs, nil
	default:
		return "", fmt.Errorf("invalid dashboard section %q; valid values: all, trades, clusters, levels, cluster-bombs", value)
	}
}

func newDashboardSectionSet(sections []dashboardSection) dashboardSectionSet {
	set := make(dashboardSectionSet, len(sections))
	for _, section := range sections {
		set[section] = true
	}
	return set
}

func (sections dashboardSectionSet) contains(section dashboardSection) bool {
	return sections[section]
}

func (sections dashboardSectionSet) names() []string {
	names := make([]string, 0, len(sections))
	for _, section := range allDashboardSections {
		if sections.contains(section) {
			names = append(names, string(section))
		}
	}
	return names
}

func dashboardOutputCount(opts *tradeDashboardOptions) int {
	if opts.Summary && opts.Count > dashboardSummaryMaxRows {
		return dashboardSummaryMaxRows
	}
	return opts.Count
}

func limitDashboardRows[T any](rows []T, count int) []T {
	if len(rows) <= count {
		return rows
	}
	return rows[:count]
}

func dashboardTradeFilters(opts *tradeDashboardOptions, ticker, startDate, endDate string) map[string]string {
	filters := buildTradeFilters(&tradesOptions{tickers: ticker, startDate: startDate, endDate: endDate, minVolume: opts.MinVolume, maxVolume: opts.MaxVolume, minPrice: opts.MinPrice, maxPrice: opts.MaxPrice, minDollars: opts.MinDollars, maxDollars: opts.MaxDollars, conditions: opts.Conditions, vcd: opts.VCD, relativeSize: opts.RelativeSize, darkPools: opts.DarkPools.Int(), sweeps: opts.Sweeps.Int(), latePrints: opts.LatePrints.Int(), sigPrints: opts.SigPrints.Int(), tradeRank: opts.TradeRank, premarket: opts.Premarket.Int(), rth: opts.RTH.Int(), ah: opts.AH.Int(), opening: opts.Opening.Int(), closing: opts.Closing.Int(), phantom: opts.Phantom.Int(), offsetting: opts.Offsetting.Int()})
	filters["Sort"] = "Dollars"
	delete(filters, "SecurityTypeKey")
	delete(filters, "EvenShared")
	delete(filters, "TradeRankSnapshot")
	delete(filters, "MarketCap")
	return filters
}

func dashboardClusterFilters(opts *tradeDashboardOptions, ticker, startDate, endDate string) map[string]string {
	return map[string]string{"Tickers": ticker, "StartDate": startDate, "EndDate": endDate, "MinVolume": common.IntStr(opts.MinVolume), "MaxVolume": common.IntStr(opts.MaxVolume), "MinPrice": common.FormatFloat(opts.MinPrice), "MaxPrice": common.FormatFloat(opts.MaxPrice), "MinDollars": common.FormatFloat(opts.MinDollars), "MaxDollars": common.FormatFloat(opts.MaxDollars), "VCD": common.IntStr(opts.VCD), "RelativeSize": common.IntStr(opts.RelativeSize), "TradeClusterRank": "-1", "Sort": "Dollars"}
}

func dashboardClusterBombFilters(opts *tradeDashboardOptions, ticker, startDate, endDate string) map[string]string {
	filters := dashboardClusterFilters(opts, ticker, startDate, endDate)
	delete(filters, "MinPrice")
	delete(filters, "MaxPrice")
	delete(filters, "TradeClusterRank")
	filters["TradeClusterBombRank"] = "-1"
	return filters
}

func dashboardLevelFilters(ticker, startDate, endDate string, count int) map[string]string {
	return map[string]string{"Ticker": ticker, "StartDate": startDate, "EndDate": endDate, "Levels": common.IntStr(count)}
}

func fetchDashboardTrades(cmd *cobra.Command, vlClient *vlgo.Client, filters map[string]string, count int) ([]models.Trade, error) {
	dtReq := newDashboardDTRequest(tradeChartColumns, 0, "FullTimeString24", count)
	resp, err := vlClient.GetTrades(cmd.Context(), vlgo.TradesRequest{DataTables: dtReq, Filters: common.FiltersToValues(filters)})
	if err != nil {
		slog.Error("failed to query dashboard trades", "error", err)
		return nil, fmt.Errorf("query dashboard trades: %w", err)
	}
	return common.MapSlice(resp.Data, common.MapVLTrade), nil
}

func fetchDashboardClusters(cmd *cobra.Command, vlClient *vlgo.Client, filters map[string]string, count int) ([]models.TradeCluster, error) {
	dtReq := newDashboardDTRequest(tradeClusterChartColumns, 3, "Sh", count)
	resp, err := vlClient.GetTradeClusters(cmd.Context(), vlgo.TradeClustersRequest{DataTables: dtReq, Filters: common.FiltersToValues(filters)})
	if err != nil {
		slog.Error("failed to query dashboard trade clusters", "error", err)
		return nil, fmt.Errorf("query dashboard trade clusters: %w", err)
	}
	return common.MapSlice(resp.Data, common.MapVLTradeCluster), nil
}

func fetchDashboardLevels(cmd *cobra.Command, vlClient *vlgo.Client, filters map[string]string, count int) ([]models.TradeLevel, error) {
	dtReq := newDashboardDTRequest(tradeLevelChartColumns, 0, "Price", count)
	dtReq.Length = -1
	resp, err := vlClient.GetChart0TradeLevels(cmd.Context(), vlgo.TradeLevelsRequest{DataTables: dtReq, Filters: common.FiltersToValues(filters)})
	if err != nil {
		slog.Error("failed to query dashboard trade levels", "error", err)
		return nil, fmt.Errorf("query dashboard trade levels: %w", err)
	}
	result := common.MapSlice(resp.Data, common.MapVLTradeLevel)
	if len(result) > count {
		result = result[:count]
	}
	return result, nil
}

func fetchDashboardClusterBombs(cmd *cobra.Command, vlClient *vlgo.Client, filters map[string]string, count int) ([]models.TradeClusterBomb, error) {
	dtReq := newDashboardDTRequest(tradeClusterBombChartColumns, 2, "Sh", count)
	resp, err := vlClient.GetTradeClusterBombs(cmd.Context(), vlgo.TradeClusterBombsRequest{DataTables: dtReq, Filters: common.FiltersToValues(filters)})
	if err != nil {
		slog.Error("failed to query dashboard trade cluster bombs", "error", err)
		return nil, fmt.Errorf("query dashboard trade cluster bombs: %w", err)
	}
	return common.MapSlice(resp.Data, common.MapVLTradeClusterBomb), nil
}
