package trade

import (
	"context"
	"fmt"
	"net/url"

	vlgo "github.com/major/volumeleaders-go/volumeleaders"
	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
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
	return common.RunVLDataTablesCommandWithPageSize[vlgo.TradeCluster, models.TradeCluster](cmd, dtOpts, opts.Format, tradeBrowserPageLength, "query trade clusters", fetchTradeClusters, common.MapVLTradeCluster)
}

func runTradeClusterBombs(cmd *cobra.Command, opts *tradeClusterBombsOptions) error {
	startDate, endDate, err := common.ResolveDateRange(cmd, 0, true)
	if err != nil {
		return err
	}
	rangeFilters := tradeClusterBombRangeFilters(cmd, opts, startDate, endDate)
	filters := rangeFilters.clusterBombMap(opts.SecurityType, opts.RelativeSize, opts.TradeClusterBombRank)
	dtOpts := common.NewDataTableOptions(common.DataTableRequestConfig{Start: opts.Start, Length: -1, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Filters: filters})
	return common.RunVLDataTablesCommandWithPageSize[vlgo.TradeClusterBomb, models.TradeClusterBomb](cmd, dtOpts, opts.Format, tradeBrowserPageLength, "query trade cluster bombs", fetchTradeClusterBombs, common.MapVLTradeClusterBomb)
}

func runTradeAlerts(cmd *cobra.Command, opts *tradeAlertsOptions) error {
	dtOpts := common.NewDataTableOptions(common.DataTableRequestConfig{Start: opts.Start, Length: opts.Length, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Filters: map[string]string{"Date": opts.Date}})
	return common.RunVLDataTablesCommand[vlgo.TradeAlert, models.TradeAlert](cmd, dtOpts, opts.Format, "query trade alerts", fetchTradeAlerts, common.MapVLTradeAlert)
}

func runTradeClusterAlerts(cmd *cobra.Command, opts *tradeClusterAlertsOptions) error {
	dtOpts := common.NewDataTableOptions(common.DataTableRequestConfig{Start: opts.Start, Length: opts.Length, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Filters: map[string]string{"Date": opts.Date}})
	return common.RunVLDataTablesCommand[vlgo.TradeClusterAlert, models.TradeClusterAlert](cmd, dtOpts, opts.Format, "query trade cluster alerts", fetchTradeClusterAlerts, common.MapVLTradeCluster)
}

// fetchTradeClusters wraps vlgo.Client.GetTradeClusters as a VLFetcher.
func fetchTradeClusters(ctx context.Context, c *vlgo.Client, dt *vlgo.DataTablesRequest, filters url.Values) (*vlgo.DataTablesResponse[vlgo.TradeCluster], error) {
	return c.GetTradeClusters(ctx, vlgo.TradeClustersRequest{DataTables: *dt, Filters: filters})
}

// fetchTradeClusterBombs wraps vlgo.Client.GetTradeClusterBombs as a VLFetcher.
func fetchTradeClusterBombs(ctx context.Context, c *vlgo.Client, dt *vlgo.DataTablesRequest, filters url.Values) (*vlgo.DataTablesResponse[vlgo.TradeClusterBomb], error) {
	return c.GetTradeClusterBombs(ctx, vlgo.TradeClusterBombsRequest{DataTables: *dt, Filters: filters})
}

// fetchTradeAlerts wraps vlgo.Client.GetTradeAlerts as a VLFetcher.
func fetchTradeAlerts(ctx context.Context, c *vlgo.Client, dt *vlgo.DataTablesRequest, filters url.Values) (*vlgo.DataTablesResponse[vlgo.TradeAlert], error) {
	return c.GetTradeAlerts(ctx, vlgo.TradeAlertsRequest{DataTables: *dt, Filters: filters})
}

// fetchTradeClusterAlerts wraps vlgo.Client.GetTradeClusterAlerts as a VLFetcher.
func fetchTradeClusterAlerts(ctx context.Context, c *vlgo.Client, dt *vlgo.DataTablesRequest, filters url.Values) (*vlgo.DataTablesResponse[vlgo.TradeClusterAlert], error) {
	return c.GetTradeClusterAlerts(ctx, vlgo.TradeClusterAlertsRequest{DataTables: *dt, Filters: filters})
}

func tradeClusterRangeFilters(cmd *cobra.Command, opts *tradeClustersOptions, startDate, endDate string) tradeRangeFilters {
	return tradeRangeFilters{Tickers: common.MultiTickerValue(cmd), StartDate: startDate, EndDate: endDate, MinVolume: opts.MinVolume, MaxVolume: opts.MaxVolume, MinPrice: opts.MinPrice, MaxPrice: opts.MaxPrice, MinDollars: opts.MinDollars, MaxDollars: opts.MaxDollars, VCD: opts.VCD, Sector: opts.Sector}
}

func tradeClusterBombRangeFilters(cmd *cobra.Command, opts *tradeClusterBombsOptions, startDate, endDate string) tradeRangeFilters {
	return tradeRangeFilters{Tickers: common.MultiTickerValue(cmd), StartDate: startDate, EndDate: endDate, MinVolume: opts.MinVolume, MaxVolume: opts.MaxVolume, MinDollars: opts.MinDollars, MaxDollars: opts.MaxDollars, VCD: opts.VCD, Sector: opts.Sector}
}
