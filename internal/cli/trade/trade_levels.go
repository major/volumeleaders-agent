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
	levelOpts := tradeLevelOptionsFromLevelsOptions(opts, ticker, startDate, endDate)
	dataOpts := common.DataTableOptions{Start: 0, Length: levelOpts.tradeLevelCount, OrderCol: 1, OrderDir: "desc", Filters: buildTradeLevelFilters(levelOpts), Fields: fields}
	if format == common.OutputFormatJSON && len(fields) == 0 {
		levels, err := fetchTradeLevels(cmd, dataOpts)
		if err != nil {
			return err
		}
		return common.PrintJSON(cmd.OutOrStdout(), cmd.Context(), models.NewTradeLevelRows(levels))
	}
	return common.RunDataTablesSingleRequestCommand[models.TradeLevel](cmd, "/TradeLevels/GetTradeLevels", datatables.TradeLevelColumns, dataOpts, opts.Format, "query trade levels")
}

func tradeLevelOptionsFromLevelsOptions(opts *tradeLevelsOptions, ticker, startDate, endDate string) *tradeLevelOptions {
	return &tradeLevelOptions{ticker: ticker, startDate: startDate, endDate: endDate, minVolume: opts.MinVolume, maxVolume: opts.MaxVolume, minPrice: opts.MinPrice, maxPrice: opts.MaxPrice, minDollars: opts.MinDollars, maxDollars: opts.MaxDollars, vcd: opts.VCD, relativeSize: opts.RelativeSize, tradeLevelRank: opts.TradeLevelRank, tradeLevelCount: opts.TradeLevelCount}
}

func fetchTradeLevels(cmd *cobra.Command, opts common.DataTableOptions) ([]models.TradeLevel, error) {
	ctx := cmd.Context()
	vlClient, err := common.NewCommandClient(ctx)
	if err != nil {
		return nil, err
	}
	request := common.NewDataTablesRequest(datatables.TradeLevelColumns, opts)
	var result []models.TradeLevel
	if err := vlClient.PostDataTables(ctx, "/TradeLevels/GetTradeLevels", request.Encode(), &result); err != nil {
		slog.Error("failed to query trade levels", "error", err)
		return nil, fmt.Errorf("query trade levels: %w", err)
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
