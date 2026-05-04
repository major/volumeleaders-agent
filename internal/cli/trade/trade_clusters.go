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
	return common.RunDataTablesCommandWithPageSize[models.TradeCluster](cmd, "/TradeClusters/GetTradeClusters", datatables.TradeClusterColumns, common.DataTableOptions{Start: opts.Start, Length: -1, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Fields: fields, Filters: map[string]string{"Tickers": common.MultiTickerValue(cmd), "StartDate": startDate, "EndDate": endDate, "MinVolume": common.IntStr(opts.MinVolume), "MaxVolume": common.IntStr(opts.MaxVolume), "MinPrice": common.FormatFloat(opts.MinPrice), "MaxPrice": common.FormatFloat(opts.MaxPrice), "MinDollars": common.FormatFloat(opts.MinDollars), "MaxDollars": common.FormatFloat(opts.MaxDollars), "VCD": common.IntStr(opts.VCD), "SecurityTypeKey": common.IntStr(opts.SecurityType), "RelativeSize": common.IntStr(opts.RelativeSize), "TradeClusterRank": common.IntStr(opts.TradeClusterRank), "SectorIndustry": opts.Sector}}, opts.Format, tradeBrowserPageLength, "query trade clusters")
}

func runTradeClusterBombs(cmd *cobra.Command, opts *tradeClusterBombsOptions) error {
	startDate, endDate, err := common.ResolveDateRange(cmd, 0, true)
	if err != nil {
		return err
	}
	return common.RunDataTablesCommandWithPageSize[models.TradeClusterBomb](cmd, "/TradeClusterBombs/GetTradeClusterBombs", datatables.TradeClusterBombColumns, common.DataTableOptions{Start: opts.Start, Length: -1, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Filters: map[string]string{"Tickers": common.MultiTickerValue(cmd), "StartDate": startDate, "EndDate": endDate, "MinVolume": common.IntStr(opts.MinVolume), "MaxVolume": common.IntStr(opts.MaxVolume), "MinDollars": common.FormatFloat(opts.MinDollars), "MaxDollars": common.FormatFloat(opts.MaxDollars), "VCD": common.IntStr(opts.VCD), "SecurityTypeKey": common.IntStr(opts.SecurityType), "RelativeSize": common.IntStr(opts.RelativeSize), "TradeClusterBombRank": common.IntStr(opts.TradeClusterBombRank), "SectorIndustry": opts.Sector}}, opts.Format, tradeBrowserPageLength, "query trade cluster bombs")
}

func runTradeAlerts(cmd *cobra.Command, opts *tradeAlertsOptions) error {
	return common.RunDataTablesCommand[models.TradeAlert](cmd, "/TradeAlerts/GetTradeAlerts", datatables.TradeColumns, common.DataTableOptions{Start: opts.Start, Length: opts.Length, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Filters: map[string]string{"Date": opts.Date}}, opts.Format, "query trade alerts")
}

func runTradeClusterAlerts(cmd *cobra.Command, opts *tradeClusterAlertsOptions) error {
	return common.RunDataTablesCommand[models.TradeClusterAlert](cmd, "/TradeClusterAlerts/GetTradeClusterAlerts", datatables.TradeClusterColumns, common.DataTableOptions{Start: opts.Start, Length: opts.Length, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Filters: map[string]string{"Date": opts.Date}}, opts.Format, "query trade cluster alerts")
}
