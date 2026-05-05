package trade

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/datatables"
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
	vlClient, err := common.NewCommandClient(ctx)
	if err != nil {
		return nil, err
	}
	request := dashboardRequest(datatables.TradeLevelChartColumns, 0, "Price", count, filters)
	request.Length = -1
	var result []models.TradeLevel
	if err := vlClient.PostDataTables(ctx, "/Chart0/GetTradeLevels", request.Encode(), &result); err != nil {
		slog.Error("failed to query trade levels", "error", err)
		return nil, fmt.Errorf("query trade levels: %w", err)
	}
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
	return common.RunDataTablesCommand[models.TradeLevelTouch](cmd, "/TradeLevelTouches/GetTradeLevelTouches", datatables.TradeLevelTouchColumns, common.DataTableOptions{Start: opts.Start, Length: opts.Length, OrderCol: opts.OrderCol, OrderDir: opts.OrderDir, Filters: map[string]string{"Tickers": ticker, "StartDate": startDate, "EndDate": endDate, "MinVolume": common.IntStr(opts.MinVolume), "MaxVolume": common.IntStr(opts.MaxVolume), "MinPrice": common.FormatFloat(opts.MinPrice), "MaxPrice": common.FormatFloat(opts.MaxPrice), "MinDollars": common.FormatFloat(opts.MinDollars), "MaxDollars": common.FormatFloat(opts.MaxDollars), "VCD": common.IntStr(opts.VCD), "RelativeSize": common.IntStr(opts.RelativeSize), "TradeLevelRank": common.IntStr(opts.TradeLevelRank), "Levels": common.IntStr(opts.TradeLevelCount)}}, opts.Format, "query trade level touches")
}
