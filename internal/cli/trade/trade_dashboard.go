package trade

import (
	"fmt"
	"log/slog"
	"maps"

	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/client"
	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
)

func runTradeDashboard(cmd *cobra.Command, opts *tradeDashboardOptions) error {
	startDate, endDate, err := common.ResolveDateRange(cmd, tradeListTickerLookbackDays, false)
	if err != nil {
		return err
	}
	ticker, err := common.SingleTickerValue(cmd)
	if err != nil {
		return err
	}
	vlClient, err := common.NewCommandClient(cmd.Context())
	if err != nil {
		return err
	}

	tradeFilters := dashboardTradeFilters(opts, ticker, startDate, endDate)
	trades, err := fetchDashboardTrades(cmd, vlClient, tradeFilters, opts.Count)
	if err != nil {
		return err
	}
	clusters, err := fetchDashboardClusters(cmd, vlClient, dashboardClusterFilters(opts, ticker, startDate, endDate), opts.Count)
	if err != nil {
		return err
	}
	levels, err := fetchDashboardLevels(cmd, vlClient, dashboardLevelFilters(ticker, startDate, endDate, opts.Count), opts.Count)
	if err != nil {
		return err
	}
	clusterBombs, err := fetchDashboardClusterBombs(cmd, vlClient, dashboardClusterBombFilters(opts, ticker, startDate, endDate), opts.Count)
	if err != nil {
		return err
	}

	dashboard := models.TradeDashboard{
		Ticker:       ticker,
		DateRange:    models.TradeDashboardDateRange{Start: startDate, End: endDate},
		Count:        opts.Count,
		Trades:       models.NewTradeListRows(trades),
		Clusters:     clusters,
		Levels:       models.NewTradeLevelRows(levels),
		ClusterBombs: clusterBombs,
	}
	return common.PrintJSON(cmd.OutOrStdout(), cmd.Context(), dashboard)
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

func fetchDashboardTrades(cmd *cobra.Command, vlClient *client.Client, filters map[string]string, count int) ([]models.Trade, error) {
	request := dashboardRequest(datatables.TradeChartColumns, 0, "FullTimeString24", count, filters)
	var result []models.Trade
	if err := postDashboardDataTables(cmd, vlClient, "/Trades/GetTrades", &request, &result, "query dashboard trades"); err != nil {
		return nil, err
	}
	return result, nil
}

func fetchDashboardClusters(cmd *cobra.Command, vlClient *client.Client, filters map[string]string, count int) ([]models.TradeCluster, error) {
	request := dashboardRequest(datatables.TradeClusterChartColumns, 3, "Sh", count, filters)
	var result []models.TradeCluster
	if err := postDashboardDataTables(cmd, vlClient, "/TradeClusters/GetTradeClusters", &request, &result, "query dashboard trade clusters"); err != nil {
		return nil, err
	}
	return result, nil
}

func fetchDashboardLevels(cmd *cobra.Command, vlClient *client.Client, filters map[string]string, count int) ([]models.TradeLevel, error) {
	request := dashboardRequest(datatables.TradeLevelChartColumns, 0, "Price", count, filters)
	request.Length = -1
	var result []models.TradeLevel
	if err := postDashboardDataTables(cmd, vlClient, "/Chart0/GetTradeLevels", &request, &result, "query dashboard trade levels"); err != nil {
		return nil, err
	}
	if len(result) > count {
		result = result[:count]
	}
	return result, nil
}

func fetchDashboardClusterBombs(cmd *cobra.Command, vlClient *client.Client, filters map[string]string, count int) ([]models.TradeClusterBomb, error) {
	request := dashboardRequest(datatables.TradeClusterBombChartColumns, 2, "Sh", count, filters)
	var result []models.TradeClusterBomb
	if err := postDashboardDataTables(cmd, vlClient, "/TradeClusterBombs/GetTradeClusterBombs", &request, &result, "query dashboard trade cluster bombs"); err != nil {
		return nil, err
	}
	return result, nil
}

func dashboardRequest(columns []datatables.Column, orderColumn int, orderName string, count int, filters map[string]string) datatables.Request {
	return datatables.Request{ColumnDefs: columns, Start: 0, Length: count, OrderColumnIndex: orderColumn, OrderDirection: "DESC", OrderName: orderName, IncludeSearch: true, CustomFilters: maps.Clone(filters), Draw: 1}
}

func postDashboardDataTables[T any](cmd *cobra.Command, vlClient *client.Client, path string, request *datatables.Request, result *[]T, label string) error {
	if err := vlClient.PostDataTables(cmd.Context(), path, request.Encode(), result); err != nil {
		slog.Error("failed to "+label, "error", err)
		return fmt.Errorf("%s: %w", label, err)
	}
	return nil
}
