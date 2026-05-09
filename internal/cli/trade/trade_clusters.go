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
	if opts.Fields == "" && opts.Format == common.OutputFormatJSON {
		return runTradeClustersCompact(cmd, dtOpts)
	}
	return common.RunVLDataTablesCommandWithPageSize(cmd, dtOpts, opts.Format, tradeBrowserPageLength, "query trade clusters", fetchTradeClusters, common.MapVLTradeCluster)
}

func runTradeClusterBombs(cmd *cobra.Command, opts *tradeClusterBombsOptions) error {
	startDate, endDate, err := common.ResolveDateRange(cmd, 0, true)
	if err != nil {
		return err
	}
	rangeFilters := tradeClusterBombRangeFilters(cmd, opts, startDate, endDate)
	filters := rangeFilters.clusterBombMap(opts.SecurityType, opts.RelativeSize, opts.TradeClusterBombRank)
	dtOpts := common.NewDataTableOptions(common.DataTableRequestConfig{Start: opts.Start, Length: -1, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Filters: filters})
	if opts.Format != common.OutputFormatJSON {
		return common.RunVLDataTablesCommandWithPageSize(cmd, dtOpts, opts.Format, tradeBrowserPageLength, "query trade cluster bombs", fetchTradeClusterBombs, common.MapVLTradeClusterBomb)
	}
	return common.RunVLDataTablesCommandWithPageSize(cmd, dtOpts, opts.Format, tradeBrowserPageLength, "query trade cluster bombs", fetchTradeClusterBombs, mapVLTradeClusterBombRow)
}

func runTradeAlerts(cmd *cobra.Command, opts *tradeAlertsOptions) error {
	dtOpts := common.NewDataTableOptions(common.DataTableRequestConfig{Start: opts.Start, Length: opts.Length, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Filters: map[string]string{"Date": opts.Date}})
	if opts.Format != common.OutputFormatJSON {
		return common.RunVLDataTablesCommand(cmd, dtOpts, opts.Format, "query trade alerts", fetchTradeAlerts, common.MapVLTradeAlert)
	}
	return common.RunVLDataTablesCommand(cmd, dtOpts, opts.Format, "query trade alerts", fetchTradeAlerts, mapVLTradeAlertRow)
}

func runTradeClusterAlerts(cmd *cobra.Command, opts *tradeClusterAlertsOptions) error {
	dtOpts := common.NewDataTableOptions(common.DataTableRequestConfig{Start: opts.Start, Length: opts.Length, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Filters: map[string]string{"Date": opts.Date}})
	if opts.Format != common.OutputFormatJSON {
		return common.RunVLDataTablesCommand(cmd, dtOpts, opts.Format, "query trade cluster alerts", fetchTradeClusterAlerts, common.MapVLTradeCluster)
	}
	return common.RunVLDataTablesCommand(cmd, dtOpts, opts.Format, "query trade cluster alerts", fetchTradeClusterAlerts, mapVLTradeClusterRow)
}

func runTradeClustersCompact(cmd *cobra.Command, opts common.DataTableOptions) error {
	vlClient, err := common.NewVLClient(cmd.Context())
	if err != nil {
		return fmt.Errorf("creating VolumeLeaders client: %w", err)
	}
	clusters, err := common.FetchVLPages(cmd.Context(), vlClient, opts, tradeBrowserPageLength, "query trade clusters", fetchTradeClusters)
	if err != nil {
		return fmt.Errorf("query trade clusters: %w", err)
	}
	mapped := common.MapSlice(clusters, common.MapVLTradeCluster)
	return common.PrintJSON(cmd.OutOrStdout(), cmd.Context(), models.NewTradeClusterRows(mapped))
}

func mapVLTradeClusterRow(cluster *vlgo.TradeCluster) models.TradeClusterRow {
	mapped := common.MapVLTradeCluster(cluster)
	return models.NewTradeClusterRow(&mapped)
}

func mapVLTradeClusterBombRow(bomb *vlgo.TradeClusterBomb) models.TradeClusterBombRow {
	mapped := common.MapVLTradeClusterBomb(bomb)
	return models.NewTradeClusterBombRow(&mapped)
}

func mapVLTradeAlertRow(alert *vlgo.TradeAlert) models.TradeAlertRow {
	mapped := common.MapVLTradeAlert(alert)
	return models.NewTradeAlertRow(&mapped)
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
