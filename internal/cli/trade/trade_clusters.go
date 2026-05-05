package trade

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
)

func runTradeClusters(cmd *cobra.Command, opts *tradeClustersOptions) error {
	startDate, endDate, err := common.ResolveDateRange(cmd, 0, true)
	if err != nil {
		return err
	}
	fields, err := common.OutputFields[models.TradeCluster](opts.Fields, tradeClusterDefaultFields)
	if err != nil {
		return fmt.Errorf("parsing fields flag: %w", err)
	}
	rangeFilters := tradeClusterRangeFilters(cmd, opts, startDate, endDate)
	filters := rangeFilters.clusterMap(opts.SecurityType, opts.RelativeSize, opts.TradeClusterRank)
	dtOpts := common.NewDataTableOptions(common.DataTableRequestConfig{Start: opts.Start, Length: -1, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Fields: fields, Filters: filters})
	return common.RunDataTablesCommandWithPageSize[models.TradeCluster](cmd, "/TradeClusters/GetTradeClusters", datatables.TradeClusterColumns, dtOpts, opts.Format, tradeBrowserPageLength, "query trade clusters")
}

func runTradeClusterBombs(cmd *cobra.Command, opts *tradeClusterBombsOptions) error {
	startDate, endDate, err := common.ResolveDateRange(cmd, 0, true)
	if err != nil {
		return err
	}
	rangeFilters := tradeClusterBombRangeFilters(cmd, opts, startDate, endDate)
	filters := rangeFilters.clusterBombMap(opts.SecurityType, opts.RelativeSize, opts.TradeClusterBombRank)
	dtOpts := common.NewDataTableOptions(common.DataTableRequestConfig{Start: opts.Start, Length: -1, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Filters: filters})
	return common.RunDataTablesCommandWithPageSize[models.TradeClusterBomb](cmd, "/TradeClusterBombs/GetTradeClusterBombs", datatables.TradeClusterBombColumns, dtOpts, opts.Format, tradeBrowserPageLength, "query trade cluster bombs")
}

func runTradeAlerts(cmd *cobra.Command, opts *tradeAlertsOptions) error {
	dtOpts := common.NewDataTableOptions(common.DataTableRequestConfig{Start: opts.Start, Length: opts.Length, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Filters: map[string]string{"Date": opts.Date}})
	return common.RunDataTablesCommand[models.TradeAlert](cmd, "/TradeAlerts/GetTradeAlerts", datatables.TradeColumns, dtOpts, opts.Format, "query trade alerts")
}

func runTradeClusterAlerts(cmd *cobra.Command, opts *tradeClusterAlertsOptions) error {
	dtOpts := common.NewDataTableOptions(common.DataTableRequestConfig{Start: opts.Start, Length: opts.Length, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Filters: map[string]string{"Date": opts.Date}})
	return common.RunDataTablesCommand[models.TradeClusterAlert](cmd, "/TradeClusterAlerts/GetTradeClusterAlerts", datatables.TradeClusterColumns, dtOpts, opts.Format, "query trade cluster alerts")
}

func tradeClusterRangeFilters(cmd *cobra.Command, opts *tradeClustersOptions, startDate, endDate string) tradeRangeFilters {
	return tradeRangeFilters{Tickers: common.MultiTickerValue(cmd), StartDate: startDate, EndDate: endDate, MinVolume: opts.MinVolume, MaxVolume: opts.MaxVolume, MinPrice: opts.MinPrice, MaxPrice: opts.MaxPrice, MinDollars: opts.MinDollars, MaxDollars: opts.MaxDollars, VCD: opts.VCD, Sector: opts.Sector}
}

func tradeClusterBombRangeFilters(cmd *cobra.Command, opts *tradeClusterBombsOptions, startDate, endDate string) tradeRangeFilters {
	return tradeRangeFilters{Tickers: common.MultiTickerValue(cmd), StartDate: startDate, EndDate: endDate, MinVolume: opts.MinVolume, MaxVolume: opts.MaxVolume, MinDollars: opts.MinDollars, MaxDollars: opts.MaxDollars, VCD: opts.VCD, Sector: opts.Sector}
}
