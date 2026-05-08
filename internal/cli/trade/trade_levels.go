package trade

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	vlgo "github.com/major/volumeleaders-go/volumeleaders"
	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/models"
)

func runTradeLevels(cmd *cobra.Command, opts *tradeLevelsOptions) error {
	format, err := common.ParseOutputFormat(opts.Format)
	if err != nil {
		return err
	}
	startDate, endDate, err := common.ResolveDateRange(cmd, 365, false)
	if err != nil {
		return err
	}
	fields, err := common.OutputFields[models.TradeLevel](opts.Fields, nil)
	if err != nil {
		return fmt.Errorf("parsing fields flag: %w", err)
	}
	ticker, err := common.SingleTickerValue(cmd)
	if err != nil {
		return err
	}
	filters := dashboardLevelFilters(ticker, startDate, endDate, opts.TradeLevelCount)
	if format == common.OutputFormatJSON && len(fields) == 0 {
		levels, err := fetchTradeLevels(cmd, filters, opts.TradeLevelCount)
		if err != nil {
			return err
		}
		return common.PrintJSON(cmd.OutOrStdout(), cmd.Context(), models.NewTradeLevelRows(levels))
	}
	levels, err := fetchTradeLevels(cmd, filters, opts.TradeLevelCount)
	if err != nil {
		return err
	}
	return common.PrintDataTablesResult(cmd.OutOrStdout(), cmd.Context(), levels, fields, format)
}

func fetchTradeLevels(cmd *cobra.Command, filters map[string]string, count int) ([]models.TradeLevel, error) {
	ctx := cmd.Context()
	vlClient, err := common.NewVLClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create VL client: %w", err)
	}
	dtReq := newDashboardDTRequest(tradeLevelChartColumns, 0, "Price", count)
	dtReq.Length = -1
	resp, err := vlClient.GetChart0TradeLevels(ctx, vlgo.TradeLevelsRequest{DataTables: dtReq, Filters: common.FiltersToValues(filters)})
	if err != nil {
		slog.Error("failed to query trade levels", "error", err)
		return nil, fmt.Errorf("query trade levels: %w", err)
	}
	result := common.MapSlice(resp.Data, common.MapVLTradeLevel)
	if len(result) > count {
		result = result[:count]
	}
	return result, nil
}

func runTradeLevelTouches(cmd *cobra.Command, opts *tradeLevelTouchesOptions) error {
	startDate, endDate, err := common.ResolveDateRange(cmd, 0, true)
	if err != nil {
		return err
	}
	ticker, err := common.SingleTickerValue(cmd)
	if err != nil {
		return err
	}
	rangeFilters := tradeLevelTouchRangeFilters(ticker, opts, startDate, endDate)
	filters := rangeFilters.levelTouchMap(opts.RelativeSize, opts.TradeLevelRank, opts.TradeLevelCount)
	dtOpts := common.NewDataTableOptions(common.DataTableRequestConfig{Start: opts.Start, Length: opts.Length, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Filters: filters})
	return common.RunVLDataTablesCommand[vlgo.TradeLevel, models.TradeLevelTouch](cmd, dtOpts, opts.Format, "query trade level touches", fetchTradeLevelTouches, common.MapVLTradeLevelTouch)
}

// fetchTradeLevelTouches wraps vlgo.Client.GetTradeLevelTouches as a VLFetcher.
func fetchTradeLevelTouches(ctx context.Context, c *vlgo.Client, dt *vlgo.DataTablesRequest, filters url.Values) (*vlgo.DataTablesResponse[vlgo.TradeLevel], error) {
	return c.GetTradeLevelTouches(ctx, vlgo.TradeLevelTouchesRequest{DataTables: *dt, Filters: filters})
}

func tradeLevelTouchRangeFilters(ticker string, opts *tradeLevelTouchesOptions, startDate, endDate string) tradeRangeFilters {
	return tradeRangeFilters{Tickers: ticker, StartDate: startDate, EndDate: endDate, MinVolume: opts.MinVolume, MaxVolume: opts.MaxVolume, MinPrice: opts.MinPrice, MaxPrice: opts.MaxPrice, MinDollars: opts.MinDollars, MaxDollars: opts.MaxDollars, VCD: opts.VCD}
}
