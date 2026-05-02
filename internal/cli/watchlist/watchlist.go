package watchlist

import (
	"fmt"
	"log/slog"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
)

// NewCmd returns the "watchlist" command group with all subcommands.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:      "watchlist",
		Short:    "Watch list commands",
		GroupID:  "watchlists",
	}
	cmd.AddCommand(
		newConfigsCmd(),
		newTickersCmd(),
		newDeleteCmd(),
		newAddTickerCmd(),
		newCreateCmd(),
		newEditCmd(),
	)
	return cmd
}

// newConfigsCmd returns the "configs" subcommand.
func newConfigsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configs",
		Short:   "List saved watch list configurations",
		Example: "volumeleaders-agent watchlist configs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			format, _ := cmd.Flags().GetString("format")
			opts := common.DataTableOptions{
				Start:    0,
				Length:   -1,
				OrderCol: 1,
				OrderDir: "asc",
			}
			return common.RunDataTablesCommand[models.WatchListConfig](
				cmd,
				"/WatchListConfigs/GetWatchLists",
				datatables.WatchlistConfigColumns,
				opts,
				format,
				"query watchlist configs",
			)
		},
	}
	common.AddOutputFormatFlags(cmd)
	return cmd
}

// newTickersCmd returns the "tickers" subcommand.
func newTickersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tickers",
		Short:   "Query tickers for a selected watch list",
		Example: "volumeleaders-agent watchlist tickers --watchlist-key 1",
		RunE: func(cmd *cobra.Command, _ []string) error {
			watchlistKey, _ := cmd.Flags().GetInt("watchlist-key")
			format, _ := cmd.Flags().GetString("format")
			opts := common.DataTableOptions{
				Start:    0,
				Length:   -1,
				OrderCol: 0,
				OrderDir: "asc",
				Filters: map[string]string{
					"WatchListKey": strconv.Itoa(watchlistKey),
				},
			}
			return common.RunDataTablesCommand[models.WatchListTicker](
				cmd,
				"/WatchLists/GetWatchListTickers",
				datatables.WatchlistTickerColumns,
				opts,
				format,
				"query watchlist tickers",
			)
		},
	}
	cmd.Flags().Int("watchlist-key", -1, "Watch list key (-1 for all)")
	common.AddOutputFormatFlags(cmd)
	return cmd
}

// newDeleteCmd returns the "delete" subcommand.
func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete a watch list configuration",
		Example: "volumeleaders-agent watchlist delete --key 1",
		RunE: func(cmd *cobra.Command, _ []string) error {
			key, _ := cmd.Flags().GetInt("key")
			ctx := cmd.Context()

			vlClient, err := common.NewCommandClient(ctx)
			if err != nil {
				return err
			}

			payload := map[string]int{"WatchListKey": key}
			var result any
			if err := vlClient.PostJSON(ctx, "/WatchListConfigs/DeleteWatchList", payload, &result); err != nil {
				slog.Error("failed to delete watchlist", "error", err)
				return fmt.Errorf("delete watchlist: %w", err)
			}

			return common.PrintJSON(cmd.OutOrStdout(), ctx, result)
		},
	}
	cmd.Flags().Int("key", 0, "Watch list key to delete")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}

// newAddTickerCmd returns the "add-ticker" subcommand.
func newAddTickerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-ticker",
		Short:   "Add a ticker to an existing watch list",
		Example: "volumeleaders-agent watchlist add-ticker --watchlist-key 1 --ticker NVDA",
		RunE: func(cmd *cobra.Command, _ []string) error {
			watchlistKey, _ := cmd.Flags().GetInt("watchlist-key")
			ticker, _ := cmd.Flags().GetString("ticker")
			ctx := cmd.Context()

			vlClient, err := common.NewCommandClient(ctx)
			if err != nil {
				return err
			}

			values := url.Values{
				"WatchListKey": {strconv.Itoa(watchlistKey)},
				"Ticker":       {ticker},
			}
			var result any
			if err := vlClient.PostForm(ctx, "/Chart0/UpdateWatchList", values, &result); err != nil {
				slog.Error("failed to add ticker to watchlist", "error", err)
				return fmt.Errorf("add ticker to watchlist: %w", err)
			}

			return common.PrintJSON(cmd.OutOrStdout(), ctx, result)
		},
	}
	cmd.Flags().Int("watchlist-key", 0, "Watch list key")
	_ = cmd.MarkFlagRequired("watchlist-key")
	cmd.Flags().String("ticker", "", "Ticker symbol to add")
	_ = cmd.MarkFlagRequired("ticker")
	return cmd
}

// newCreateCmd returns the "create" subcommand.
func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new watch list configuration",
		Example: `volumeleaders-agent watchlist create --name "Tech stocks" --tickers AAPL,MSFT,GOOGL
volumeleaders-agent watchlist create --name "Large caps" --security-type 1 --min-dollars 10000000`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCreateEdit(cmd, 0)
		},
	}
	addWatchlistConfigFlags(cmd)
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

// newEditCmd returns the "edit" subcommand.
func newEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "edit",
		Short:   "Edit an existing watch list configuration",
		Example: `volumeleaders-agent watchlist edit --key 1 --name "Updated watchlist" --tickers AAPL,MSFT`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			key, _ := cmd.Flags().GetInt("key")
			return runCreateEdit(cmd, key)
		},
	}
	cmd.Flags().Int("key", 0, "Watch list key to edit")
	_ = cmd.MarkFlagRequired("key")
	addWatchlistConfigFlags(cmd)
	return cmd
}

// addWatchlistConfigFlags registers the shared CLI flags for watchlist create/edit.
func addWatchlistConfigFlags(cmd *cobra.Command) {
	cmd.Flags().String("name", "", "Watch list name")
	cmd.Flags().String("tickers", "", "Comma-separated ticker symbols (max 500)")
	cmd.Flags().Int("min-volume", 0, "Minimum volume filter")
	cmd.Flags().Int("max-volume", 2000000000, "Maximum volume filter")
	cmd.Flags().Float64("min-dollars", 0, "Minimum dollars filter")
	cmd.Flags().Float64("max-dollars", 30000000000, "Maximum dollars filter")
	cmd.Flags().Float64("min-price", 0, "Minimum price filter")
	cmd.Flags().Float64("max-price", 100000, "Maximum price filter")
	cmd.Flags().Float64("min-vcd", 0, "Minimum VCD percentile (0-100)")
	cmd.Flags().String("sector-industry", "", "Sector/industry filter (max 100 chars)")
	cmd.Flags().Int("security-type", -1, "Security type (-1=all, 1=stocks, 26=ETFs, 4=REITs)")
	cmd.Flags().Int("min-relative-size", 0, "Minimum relative size (0/5/10/25/50/100)")
	cmd.Flags().Int("max-trade-rank", -1, "Maximum trade rank (-1=all, 1/3/5/10/25/50/100)")
	cmd.Flags().Bool("normal-prints", true, "Include normal prints")
	cmd.Flags().Bool("signature-prints", true, "Include signature prints")
	cmd.Flags().Bool("late-prints", true, "Include late prints")
	cmd.Flags().Bool("timely-prints", true, "Include timely prints")
	cmd.Flags().Bool("dark-pools", true, "Include dark pool trades")
	cmd.Flags().Bool("lit-exchanges", true, "Include lit exchange trades")
	cmd.Flags().Bool("sweeps", true, "Include sweep trades")
	cmd.Flags().Bool("blocks", true, "Include block trades")
	cmd.Flags().Bool("premarket-trades", true, "Include premarket trades")
	cmd.Flags().Bool("rth-trades", true, "Include regular trading hours trades")
	cmd.Flags().Bool("ah-trades", true, "Include after-hours trades")
	cmd.Flags().Bool("opening-trades", true, "Include opening trades")
	cmd.Flags().Bool("closing-trades", true, "Include closing trades")
	cmd.Flags().Bool("phantom-trades", true, "Include phantom trades")
	cmd.Flags().Bool("offsetting-trades", true, "Include offsetting trades")
	cmd.Flags().Int("rsi-overbought-daily", -1, "RSI overbought daily (1=yes, 0=no, -1=ignore)")
	cmd.Flags().Int("rsi-overbought-hourly", -1, "RSI overbought hourly (1=yes, 0=no, -1=ignore)")
	cmd.Flags().Int("rsi-oversold-daily", -1, "RSI oversold daily (1=yes, 0=no, -1=ignore)")
	cmd.Flags().Int("rsi-oversold-hourly", -1, "RSI oversold hourly (1=yes, 0=no, -1=ignore)")
}

// buildWatchlistConfigFields maps CLI flag values to form field names for the
// WatchListConfig create/edit multipart POST.
func buildWatchlistConfigFields(cmd *cobra.Command, key int) map[string]string {
	getName, _ := cmd.Flags().GetString("name")
	getTickers, _ := cmd.Flags().GetString("tickers")
	getMinVolume, _ := cmd.Flags().GetInt("min-volume")
	getMaxVolume, _ := cmd.Flags().GetInt("max-volume")
	getMinDollars, _ := cmd.Flags().GetFloat64("min-dollars")
	getMaxDollars, _ := cmd.Flags().GetFloat64("max-dollars")
	getMinPrice, _ := cmd.Flags().GetFloat64("min-price")
	getMaxPrice, _ := cmd.Flags().GetFloat64("max-price")
	getMinVCD, _ := cmd.Flags().GetFloat64("min-vcd")
	getSectorIndustry, _ := cmd.Flags().GetString("sector-industry")
	getSecurityType, _ := cmd.Flags().GetInt("security-type")
	getMinRelativeSize, _ := cmd.Flags().GetInt("min-relative-size")
	getMaxTradeRank, _ := cmd.Flags().GetInt("max-trade-rank")
	getNormalPrints, _ := cmd.Flags().GetBool("normal-prints")
	getSignaturePrints, _ := cmd.Flags().GetBool("signature-prints")
	getLatePrints, _ := cmd.Flags().GetBool("late-prints")
	getTimelyPrints, _ := cmd.Flags().GetBool("timely-prints")
	getDarkPools, _ := cmd.Flags().GetBool("dark-pools")
	getLitExchanges, _ := cmd.Flags().GetBool("lit-exchanges")
	getSweeps, _ := cmd.Flags().GetBool("sweeps")
	getBlocks, _ := cmd.Flags().GetBool("blocks")
	getPremarketTrades, _ := cmd.Flags().GetBool("premarket-trades")
	getRTHTrades, _ := cmd.Flags().GetBool("rth-trades")
	getAHTrades, _ := cmd.Flags().GetBool("ah-trades")
	getOpeningTrades, _ := cmd.Flags().GetBool("opening-trades")
	getClosingTrades, _ := cmd.Flags().GetBool("closing-trades")
	getPhantomTrades, _ := cmd.Flags().GetBool("phantom-trades")
	getOffsettingTrades, _ := cmd.Flags().GetBool("offsetting-trades")
	getRSIOverboughtDaily, _ := cmd.Flags().GetInt("rsi-overbought-daily")
	getRSIOverboughtHourly, _ := cmd.Flags().GetInt("rsi-overbought-hourly")
	getRSIOversoldDaily, _ := cmd.Flags().GetInt("rsi-oversold-daily")
	getRSIOversoldHourly, _ := cmd.Flags().GetInt("rsi-oversold-hourly")

	return map[string]string{
		"SearchTemplateKey":           strconv.Itoa(key),
		"Name":                        getName,
		"Tickers":                     getTickers,
		"MinVolume":                   strconv.Itoa(getMinVolume),
		"MaxVolume":                   strconv.Itoa(getMaxVolume),
		"MinDollars":                  common.FormatFloat(getMinDollars),
		"MaxDollars":                  common.FormatFloat(getMaxDollars),
		"MinPrice":                    common.FormatFloat(getMinPrice),
		"MaxPrice":                    common.FormatFloat(getMaxPrice),
		"MinVCD":                      common.FormatFloat(getMinVCD),
		"SectorIndustry":              getSectorIndustry,
		"SecurityTypeKey":             strconv.Itoa(getSecurityType),
		"MinRelativeSizeSelected":     strconv.Itoa(getMinRelativeSize),
		"MaxTradeRankSelected":        strconv.Itoa(getMaxTradeRank),
		"NormalPrintsSelected":        common.BoolString(getNormalPrints),
		"SignaturePrintsSelected":     common.BoolString(getSignaturePrints),
		"LatePrintsSelected":          common.BoolString(getLatePrints),
		"TimelyPrintsSelected":        common.BoolString(getTimelyPrints),
		"DarkPoolsSelected":           common.BoolString(getDarkPools),
		"LitExchangesSelected":        common.BoolString(getLitExchanges),
		"SweepsSelected":              common.BoolString(getSweeps),
		"BlocksSelected":              common.BoolString(getBlocks),
		"PremarketTradesSelected":     common.BoolString(getPremarketTrades),
		"RTHTradesSelected":           common.BoolString(getRTHTrades),
		"AHTradesSelected":            common.BoolString(getAHTrades),
		"OpeningTradesSelected":       common.BoolString(getOpeningTrades),
		"ClosingTradesSelected":       common.BoolString(getClosingTrades),
		"PhantomTradesSelected":       common.BoolString(getPhantomTrades),
		"OffsettingTradesSelected":    common.BoolString(getOffsettingTrades),
		"RSIOverboughtDailySelected":  strconv.Itoa(getRSIOverboughtDaily),
		"RSIOverboughtHourlySelected": strconv.Itoa(getRSIOverboughtHourly),
		"RSIOversoldDailySelected":    strconv.Itoa(getRSIOversoldDaily),
		"RSIOversoldHourlySelected":   strconv.Itoa(getRSIOversoldHourly),
	}
}

// runCreateEdit handles both the create and edit subcommands. A key of 0
// indicates a new watchlist; a non-zero key indicates an edit.
func runCreateEdit(cmd *cobra.Command, key int) error {
	ctx := cmd.Context()

	vlClient, err := common.NewCommandClient(ctx)
	if err != nil {
		return err
	}

	fields := buildWatchlistConfigFields(cmd, key)
	if err := vlClient.PostMultipart(ctx, "/WatchListConfig", fields); err != nil {
		slog.Error("failed to save watchlist config", "error", err)
		return fmt.Errorf("save watchlist config: %w", err)
	}

	action := "created"
	if key != 0 {
		action = "updated"
	}
	return common.PrintJSON(cmd.OutOrStdout(), ctx, map[string]any{"success": true, "action": action, "key": key})
}
