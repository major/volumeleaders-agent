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

type watchlistSecurityType string
type watchlistRelativeSize string
type watchlistTradeRank string

const (
	watchlistSecurityAll    watchlistSecurityType = "-1"
	watchlistSecurityStocks watchlistSecurityType = "1"
	watchlistSecurityETFs   watchlistSecurityType = "26"
	watchlistSecurityREITs  watchlistSecurityType = "4"
)

const (
	watchlistRelativeSizeAll watchlistRelativeSize = "0"
	watchlistRelativeSize5   watchlistRelativeSize = "5"
	watchlistRelativeSize10  watchlistRelativeSize = "10"
	watchlistRelativeSize25  watchlistRelativeSize = "25"
	watchlistRelativeSize50  watchlistRelativeSize = "50"
	watchlistRelativeSize100 watchlistRelativeSize = "100"
)

const (
	watchlistTradeRankAll watchlistTradeRank = "-1"
	watchlistTradeRank1   watchlistTradeRank = "1"
	watchlistTradeRank3   watchlistTradeRank = "3"
	watchlistTradeRank5   watchlistTradeRank = "5"
	watchlistTradeRank10  watchlistTradeRank = "10"
	watchlistTradeRank25  watchlistTradeRank = "25"
	watchlistTradeRank50  watchlistTradeRank = "50"
	watchlistTradeRank100 watchlistTradeRank = "100"
)

// Set implements pflag.Value for watchlistSecurityType.
func (v *watchlistSecurityType) Set(value string) error {
	switch watchlistSecurityType(value) {
	case watchlistSecurityAll, watchlistSecurityStocks, watchlistSecurityETFs, watchlistSecurityREITs:
		*v = watchlistSecurityType(value)
		return nil
	default:
		return fmt.Errorf("invalid value %q, expected one of -1, 1, 26, 4", value)
	}
}

// String implements pflag.Value for watchlistSecurityType.
func (v watchlistSecurityType) String() string {
	return string(v)
}

// Type implements pflag.Value for watchlistSecurityType.
func (v watchlistSecurityType) Type() string {
	return "string"
}

// Set implements pflag.Value for watchlistRelativeSize.
func (v *watchlistRelativeSize) Set(value string) error {
	switch watchlistRelativeSize(value) {
	case watchlistRelativeSizeAll, watchlistRelativeSize5, watchlistRelativeSize10,
		watchlistRelativeSize25, watchlistRelativeSize50, watchlistRelativeSize100:
		*v = watchlistRelativeSize(value)
		return nil
	default:
		return fmt.Errorf("invalid value %q, expected one of 0, 5, 10, 25, 50, 100", value)
	}
}

// String implements pflag.Value for watchlistRelativeSize.
func (v watchlistRelativeSize) String() string {
	return string(v)
}

// Type implements pflag.Value for watchlistRelativeSize.
func (v watchlistRelativeSize) Type() string {
	return "string"
}

// Set implements pflag.Value for watchlistTradeRank.
func (v *watchlistTradeRank) Set(value string) error {
	switch watchlistTradeRank(value) {
	case watchlistTradeRankAll, watchlistTradeRank1, watchlistTradeRank3, watchlistTradeRank5,
		watchlistTradeRank10, watchlistTradeRank25, watchlistTradeRank50, watchlistTradeRank100:
		*v = watchlistTradeRank(value)
		return nil
	default:
		return fmt.Errorf("invalid value %q, expected one of -1, 1, 3, 5, 10, 25, 50, 100", value)
	}
}

// String implements pflag.Value for watchlistTradeRank.
func (v watchlistTradeRank) String() string {
	return string(v)
}

// Type implements pflag.Value for watchlistTradeRank.
func (v watchlistTradeRank) Type() string {
	return "string"
}

// watchlistConfigsOptions holds flags for the "watchlist configs" subcommand.
type watchlistConfigsOptions struct {
	Format common.OutputFormat
}

// watchlistTickersOptions holds flags for the "watchlist tickers" subcommand.
type watchlistTickersOptions struct {
	WatchlistKey int
	Format       common.OutputFormat
}

// watchlistDeleteOptions holds flags for the "watchlist delete" subcommand.
type watchlistDeleteOptions struct {
	Key int
}

// watchlistAddTickerOptions holds flags for the "watchlist add-ticker" subcommand.
type watchlistAddTickerOptions struct {
	WatchlistKey int
	Ticker       string
}

// watchlistConfigFlags holds the shared flag set for watchlist create/edit commands.
type watchlistConfigFlags struct {
	Name                string
	Tickers             string
	MinVolume           int
	MaxVolume           int
	MinDollars          float64
	MaxDollars          float64
	MinPrice            float64
	MaxPrice            float64
	MinVCD              float64
	SectorIndustry      string
	SecurityType        watchlistSecurityType
	MinRelativeSize     watchlistRelativeSize
	MaxTradeRank        watchlistTradeRank
	NormalPrints        bool
	SignaturePrints     bool
	LatePrints          bool
	TimelyPrints        bool
	DarkPools           bool
	LitExchanges        bool
	Sweeps              bool
	Blocks              bool
	PremarketTrades     bool
	RTHTrades           bool
	AHTrades            bool
	OpeningTrades       bool
	ClosingTrades       bool
	PhantomTrades       bool
	OffsettingTrades    bool
	RSIOverboughtDaily  common.TriStateFilter
	RSIOverboughtHourly common.TriStateFilter
	RSIOversoldDaily    common.TriStateFilter
	RSIOversoldHourly   common.TriStateFilter
}

// watchlistCreateOptions holds flags for the "watchlist create" subcommand.
type watchlistCreateOptions struct {
	watchlistConfigFlags
}

// watchlistEditOptions holds flags for the "watchlist edit" subcommand.
type watchlistEditOptions struct {
	Key int
	watchlistConfigFlags
}

// presetWatchlistConfigDefaults sets non-zero default values on watchlistConfigFlags
// before flag registration so they become pflag defaults.
func presetWatchlistConfigDefaults(cfg *watchlistConfigFlags) {
	cfg.MaxVolume = 2000000000
	cfg.MaxDollars = 30000000000
	cfg.MaxPrice = 100000
	cfg.SecurityType = watchlistSecurityAll
	cfg.MinRelativeSize = watchlistRelativeSizeAll
	cfg.MaxTradeRank = watchlistTradeRankAll
	cfg.NormalPrints = true
	cfg.SignaturePrints = true
	cfg.LatePrints = true
	cfg.TimelyPrints = true
	cfg.DarkPools = true
	cfg.LitExchanges = true
	cfg.Sweeps = true
	cfg.Blocks = true
	cfg.PremarketTrades = true
	cfg.RTHTrades = true
	cfg.AHTrades = true
	cfg.OpeningTrades = true
	cfg.ClosingTrades = true
	cfg.PhantomTrades = true
	cfg.OffsettingTrades = true
	cfg.RSIOverboughtDaily = common.TriStateAll
	cfg.RSIOverboughtHourly = common.TriStateAll
	cfg.RSIOversoldDaily = common.TriStateAll
	cfg.RSIOversoldHourly = common.TriStateAll
}

// registerWatchlistConfigFlags registers all watchlistConfigFlags fields on cmd.
// Call presetWatchlistConfigDefaults before this function so the preset values
// become the flag defaults.
func registerWatchlistConfigFlags(cmd *cobra.Command, opts *watchlistConfigFlags) {
	// Basic
	cmd.Flags().StringVar(&opts.Name, "name", opts.Name, "Watch list name")
	cmd.Flags().StringVarP(&opts.Tickers, "tickers", "t", opts.Tickers, "Comma-separated ticker symbols (max 500)")
	common.AnnotateFlagGroup(cmd, "name", "Basic")
	common.AnnotateFlagGroup(cmd, "tickers", "Basic")

	// Ranges
	cmd.Flags().IntVar(&opts.MinVolume, "min-volume", opts.MinVolume, "Minimum volume filter")
	cmd.Flags().IntVar(&opts.MaxVolume, "max-volume", opts.MaxVolume, "Maximum volume filter")
	cmd.Flags().Float64Var(&opts.MinDollars, "min-dollars", opts.MinDollars, "Minimum dollars filter")
	cmd.Flags().Float64Var(&opts.MaxDollars, "max-dollars", opts.MaxDollars, "Maximum dollars filter")
	cmd.Flags().Float64Var(&opts.MinPrice, "min-price", opts.MinPrice, "Minimum price filter")
	cmd.Flags().Float64Var(&opts.MaxPrice, "max-price", opts.MaxPrice, "Maximum price filter")
	common.AnnotateFlagGroup(cmd, "min-volume", "Ranges")
	common.AnnotateFlagGroup(cmd, "max-volume", "Ranges")
	common.AnnotateFlagGroup(cmd, "min-dollars", "Ranges")
	common.AnnotateFlagGroup(cmd, "max-dollars", "Ranges")
	common.AnnotateFlagGroup(cmd, "min-price", "Ranges")
	common.AnnotateFlagGroup(cmd, "max-price", "Ranges")

	// Filters
	cmd.Flags().Float64Var(&opts.MinVCD, "min-vcd", opts.MinVCD, "Minimum VCD percentile (0-100)")
	cmd.Flags().StringVar(&opts.SectorIndustry, "sector-industry", opts.SectorIndustry, "Sector/industry filter (max 100 chars)")
	cmd.Flags().Var(&opts.SecurityType, "security-type", "Security type (-1=all, 1=stocks, 26=ETFs, 4=REITs)")
	cmd.Flags().Var(&opts.MinRelativeSize, "min-relative-size", "Minimum relative size (0/5/10/25/50/100)")
	cmd.Flags().Var(&opts.MaxTradeRank, "max-trade-rank", "Maximum trade rank (-1=all, 1/3/5/10/25/50/100)")
	common.AnnotateFlagGroup(cmd, "min-vcd", "Filters")
	common.AnnotateFlagGroup(cmd, "sector-industry", "Filters")
	common.AnnotateFlagGroup(cmd, "security-type", "Filters")
	common.AnnotateFlagGroup(cmd, "min-relative-size", "Filters")
	common.AnnotateFlagGroup(cmd, "max-trade-rank", "Filters")
	common.AnnotateFlagEnum(cmd, "security-type", []string{"-1", "1", "26", "4"})
	common.AnnotateFlagEnum(cmd, "min-relative-size", []string{"0", "5", "10", "25", "50", "100"})
	common.AnnotateFlagEnum(cmd, "max-trade-rank", []string{"-1", "1", "3", "5", "10", "25", "50", "100"})

	// Print Types
	cmd.Flags().BoolVar(&opts.NormalPrints, "normal-prints", opts.NormalPrints, "Include normal prints")
	cmd.Flags().BoolVar(&opts.SignaturePrints, "signature-prints", opts.SignaturePrints, "Include signature prints")
	cmd.Flags().BoolVar(&opts.LatePrints, "late-prints", opts.LatePrints, "Include late prints")
	cmd.Flags().BoolVar(&opts.TimelyPrints, "timely-prints", opts.TimelyPrints, "Include timely prints")
	common.AnnotateFlagGroup(cmd, "normal-prints", "Print Types")
	common.AnnotateFlagGroup(cmd, "signature-prints", "Print Types")
	common.AnnotateFlagGroup(cmd, "late-prints", "Print Types")
	common.AnnotateFlagGroup(cmd, "timely-prints", "Print Types")

	// Venues
	cmd.Flags().BoolVar(&opts.DarkPools, "dark-pools", opts.DarkPools, "Include dark pool trades")
	cmd.Flags().BoolVar(&opts.LitExchanges, "lit-exchanges", opts.LitExchanges, "Include lit exchange trades")
	cmd.Flags().BoolVar(&opts.Sweeps, "sweeps", opts.Sweeps, "Include sweep trades")
	cmd.Flags().BoolVar(&opts.Blocks, "blocks", opts.Blocks, "Include block trades")
	common.AnnotateFlagGroup(cmd, "dark-pools", "Venues")
	common.AnnotateFlagGroup(cmd, "lit-exchanges", "Venues")
	common.AnnotateFlagGroup(cmd, "sweeps", "Venues")
	common.AnnotateFlagGroup(cmd, "blocks", "Venues")

	// Sessions
	cmd.Flags().BoolVar(&opts.PremarketTrades, "premarket-trades", opts.PremarketTrades, "Include premarket trades")
	cmd.Flags().BoolVar(&opts.RTHTrades, "rth-trades", opts.RTHTrades, "Include regular trading hours trades")
	cmd.Flags().BoolVar(&opts.AHTrades, "ah-trades", opts.AHTrades, "Include after-hours trades")
	cmd.Flags().BoolVar(&opts.OpeningTrades, "opening-trades", opts.OpeningTrades, "Include opening trades")
	cmd.Flags().BoolVar(&opts.ClosingTrades, "closing-trades", opts.ClosingTrades, "Include closing trades")
	cmd.Flags().BoolVar(&opts.PhantomTrades, "phantom-trades", opts.PhantomTrades, "Include phantom trades")
	cmd.Flags().BoolVar(&opts.OffsettingTrades, "offsetting-trades", opts.OffsettingTrades, "Include offsetting trades")
	common.AnnotateFlagGroup(cmd, "premarket-trades", "Sessions")
	common.AnnotateFlagGroup(cmd, "rth-trades", "Sessions")
	common.AnnotateFlagGroup(cmd, "ah-trades", "Sessions")
	common.AnnotateFlagGroup(cmd, "opening-trades", "Sessions")
	common.AnnotateFlagGroup(cmd, "closing-trades", "Sessions")
	common.AnnotateFlagGroup(cmd, "phantom-trades", "Sessions")
	common.AnnotateFlagGroup(cmd, "offsetting-trades", "Sessions")

	// RSI
	cmd.Flags().Var(&opts.RSIOverboughtDaily, "rsi-overbought-daily", "RSI overbought daily (-1=ignore, 0=no, 1=yes)")
	cmd.Flags().Var(&opts.RSIOverboughtHourly, "rsi-overbought-hourly", "RSI overbought hourly (-1=ignore, 0=no, 1=yes)")
	cmd.Flags().Var(&opts.RSIOversoldDaily, "rsi-oversold-daily", "RSI oversold daily (-1=ignore, 0=no, 1=yes)")
	cmd.Flags().Var(&opts.RSIOversoldHourly, "rsi-oversold-hourly", "RSI oversold hourly (-1=ignore, 0=no, 1=yes)")
	common.AnnotateFlagGroup(cmd, "rsi-overbought-daily", "RSI")
	common.AnnotateFlagGroup(cmd, "rsi-overbought-hourly", "RSI")
	common.AnnotateFlagGroup(cmd, "rsi-oversold-daily", "RSI")
	common.AnnotateFlagGroup(cmd, "rsi-oversold-hourly", "RSI")
	common.AnnotateFlagEnum(cmd, "rsi-overbought-daily", []string{"-1", "0", "1"})
	common.AnnotateFlagEnum(cmd, "rsi-overbought-hourly", []string{"-1", "0", "1"})
	common.AnnotateFlagEnum(cmd, "rsi-oversold-daily", []string{"-1", "0", "1"})
	common.AnnotateFlagEnum(cmd, "rsi-oversold-hourly", []string{"-1", "0", "1"})
}

// NewCmd returns the "watchlist" command group with all subcommands.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "watchlist",
		Short:   "Watch list commands",
		Long:    "watchlist contains subcommands for managing saved watchlist configurations that group and track ticker symbols for use in trade filtering. Use create to define new watchlists, edit to update them, delete to remove them, configs to list all watchlists, and tickers to view the symbols in a watchlist.",
		GroupID: "watchlists",
		Args:    cobra.NoArgs,
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
	opts := &watchlistConfigsOptions{Format: common.OutputFormatJSON}
	cmd := &cobra.Command{
		Use:        "configs",
		Short:      "List saved watch list configurations",
		Long:       "List all saved watchlist configurations with their keys and names. Outputs compact JSON or CSV/TSV with --format. Each row shows the watchlist key and name; use the tickers subcommand to view symbols in a specific watchlist.",
		Example:    "volumeleaders-agent watchlist configs",
		Args:       cobra.NoArgs,
		Aliases:    []string{"ls"},
		SuggestFor: []string{"config", "cfg"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			dtOpts := common.DataTableOptions{
				Start:    0,
				Length:   -1,
				OrderCol: 1,
				OrderDir: "asc",
			}
			return common.RunDataTablesCommand[models.WatchListConfig](
				cmd,
				"/WatchListConfigs/GetWatchLists",
				datatables.WatchlistConfigColumns,
				dtOpts,
				opts.Format,
				"query watchlist configs",
			)
		},
	}
	cmd.Flags().VarP(&opts.Format, "format", "f", "Output format: json, csv, or tsv")
	common.AnnotateFlagGroup(cmd, "format", "Output")
	common.AnnotateFlagEnum(cmd, "format", []string{"json", "csv", "tsv"})
	common.WrapValidation(cmd, opts)
	return cmd
}

// newTickersCmd returns the "tickers" subcommand.
func newTickersCmd() *cobra.Command {
	opts := &watchlistTickersOptions{Format: common.OutputFormatJSON}
	cmd := &cobra.Command{
		Use:        "tickers",
		Short:      "Query tickers for a selected watch list",
		Long:       "Query the ticker symbols belonging to a specific watchlist identified by --watchlist-key. Returns all tickers in the watchlist with their metadata. Outputs compact JSON or CSV/TSV with --format.",
		Example:    "volumeleaders-agent watchlist tickers --watchlist-key 1",
		Args:       cobra.NoArgs,
		SuggestFor: []string{"ticker", "tkrs"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			dtOpts := common.DataTableOptions{
				Start:    0,
				Length:   -1,
				OrderCol: 0,
				OrderDir: "asc",
				Filters: map[string]string{
					"WatchListKey": strconv.Itoa(opts.WatchlistKey),
				},
			}
			return common.RunDataTablesCommand[models.WatchListTicker](
				cmd,
				"/WatchLists/GetWatchListTickers",
				datatables.WatchlistTickerColumns,
				dtOpts,
				opts.Format,
				"query watchlist tickers",
			)
		},
	}
	cmd.Flags().IntVarP(&opts.WatchlistKey, "watchlist-key", "k", -1, "Watch list key (-1 for all)")
	cmd.Flags().VarP(&opts.Format, "format", "f", "Output format: json, csv, or tsv")
	common.AnnotateFlagGroup(cmd, "watchlist-key", "Input")
	common.AnnotateFlagGroup(cmd, "format", "Output")
	common.AnnotateFlagEnum(cmd, "format", []string{"json", "csv", "tsv"})
	common.WrapValidation(cmd, opts)
	return cmd
}

// newDeleteCmd returns the "delete" subcommand.
func newDeleteCmd() *cobra.Command {
	opts := &watchlistDeleteOptions{}
	cmd := &cobra.Command{
		Use:        "delete",
		Short:      "Delete a watch list configuration",
		Long:       "Remove a saved watchlist configuration by its numeric key. Requires --key with the watchlist key (visible in configs output). The deletion is permanent and removes the watchlist and all its tickers.",
		Example:    "volumeleaders-agent watchlist delete --key 1",
		Args:       cobra.NoArgs,
		Aliases:    []string{"rm"},
		SuggestFor: []string{"del", "remove"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			vlClient, err := common.NewCommandClient(ctx)
			if err != nil {
				return err
			}

			payload := map[string]int{"WatchListKey": opts.Key}
			var result any
			if err := vlClient.PostJSON(ctx, "/WatchListConfigs/DeleteWatchList", payload, &result); err != nil {
				slog.Error("failed to delete watchlist", "error", err)
				return fmt.Errorf("delete watchlist: %w", err)
			}

			return common.PrintJSON(cmd.OutOrStdout(), ctx, result)
		},
	}
	cmd.Flags().IntVarP(&opts.Key, "key", "k", 0, "Watch list key to delete")
	common.AnnotateFlagGroup(cmd, "key", "Input")
	common.MarkFlagRequired(cmd, "key")
	common.WrapValidation(cmd, opts)
	return cmd
}

// newAddTickerCmd returns the "add-ticker" subcommand.
func newAddTickerCmd() *cobra.Command {
	opts := &watchlistAddTickerOptions{}
	cmd := &cobra.Command{
		Use:        "add-ticker",
		Short:      "Add a ticker to an existing watch list",
		Long:       "Add a ticker symbol to an existing watchlist. Requires --watchlist-key with the watchlist key and --ticker with the symbol to add. The ticker is appended to the watchlist without affecting existing symbols.",
		Example:    "volumeleaders-agent watchlist add-ticker --watchlist-key 1 --ticker NVDA",
		Args:       cobra.NoArgs,
		SuggestFor: []string{"addticker", "add-tkr"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			vlClient, err := common.NewCommandClient(ctx)
			if err != nil {
				return err
			}

			values := url.Values{
				"WatchListKey": {strconv.Itoa(opts.WatchlistKey)},
				"Ticker":       {opts.Ticker},
			}
			var result any
			if err := vlClient.PostForm(ctx, "/Chart0/UpdateWatchList", values, &result); err != nil {
				slog.Error("failed to add ticker to watchlist", "error", err)
				return fmt.Errorf("add ticker to watchlist: %w", err)
			}

			return common.PrintJSON(cmd.OutOrStdout(), ctx, result)
		},
	}
	cmd.Flags().IntVarP(&opts.WatchlistKey, "watchlist-key", "k", 0, "Watch list key")
	cmd.Flags().StringVarP(&opts.Ticker, "ticker", "t", "", "Ticker symbol to add")
	common.AnnotateFlagGroup(cmd, "watchlist-key", "Input")
	common.AnnotateFlagGroup(cmd, "ticker", "Input")
	common.MarkFlagRequired(cmd, "watchlist-key")
	common.MarkFlagRequired(cmd, "ticker")
	common.WrapValidation(cmd, opts)
	return cmd
}

// newCreateCmd returns the "create" subcommand.
func newCreateCmd() *cobra.Command {
	opts := &watchlistCreateOptions{}
	presetWatchlistConfigDefaults(&opts.watchlistConfigFlags)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new watch list configuration",
		Long:  "Create a new watchlist configuration with a name and optional filter settings such as minimum volume, price range, sector, and trade conditions. Requires --name. Use --tickers to specify an explicit ticker list or leave unset for a filter-based watchlist.",
		Example: `volumeleaders-agent watchlist create --name "Tech stocks" --tickers AAPL,MSFT,GOOGL
volumeleaders-agent watchlist create --name "Large caps" --security-type 1 --min-dollars 10000000`,
		Args:       cobra.NoArgs,
		Aliases:    []string{"new"},
		SuggestFor: []string{"crate", "creat"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCreateEdit(cmd, &opts.watchlistConfigFlags, 0)
		},
	}
	registerWatchlistConfigFlags(cmd, &opts.watchlistConfigFlags)
	common.MarkFlagRequired(cmd, "name")
	common.WrapValidation(cmd, opts)
	return cmd
}

// newEditCmd returns the "edit" subcommand.
func newEditCmd() *cobra.Command {
	opts := &watchlistEditOptions{}
	presetWatchlistConfigDefaults(&opts.watchlistConfigFlags)
	cmd := &cobra.Command{
		Use:        "edit",
		Short:      "Edit an existing watch list configuration",
		Long:       "Modify an existing watchlist configuration identified by its numeric key. Requires --key with the watchlist key. Specify the fields you want to set; unspecified fields are replaced with their default values.",
		Example:    `volumeleaders-agent watchlist edit --key 1 --name "Updated watchlist" --tickers AAPL,MSFT`,
		Args:       cobra.NoArgs,
		SuggestFor: []string{"edt", "modify"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCreateEdit(cmd, &opts.watchlistConfigFlags, opts.Key)
		},
	}
	cmd.Flags().IntVarP(&opts.Key, "key", "k", 0, "Watch list key to edit")
	common.AnnotateFlagGroup(cmd, "key", "Input")
	registerWatchlistConfigFlags(cmd, &opts.watchlistConfigFlags)
	common.MarkFlagRequired(cmd, "key")
	common.WrapValidation(cmd, opts)
	return cmd
}

// buildWatchlistConfigFields maps struct field values to form field names for the
// WatchListConfig create/edit multipart POST.
func buildWatchlistConfigFields(opts *watchlistConfigFlags, key int) map[string]string {
	return map[string]string{
		"SearchTemplateKey":           strconv.Itoa(key),
		"Name":                        opts.Name,
		"Tickers":                     opts.Tickers,
		"MinVolume":                   strconv.Itoa(opts.MinVolume),
		"MaxVolume":                   strconv.Itoa(opts.MaxVolume),
		"MinDollars":                  common.FormatFloat(opts.MinDollars),
		"MaxDollars":                  common.FormatFloat(opts.MaxDollars),
		"MinPrice":                    common.FormatFloat(opts.MinPrice),
		"MaxPrice":                    common.FormatFloat(opts.MaxPrice),
		"MinVCD":                      common.FormatFloat(opts.MinVCD),
		"SectorIndustry":              opts.SectorIndustry,
		"SecurityTypeKey":             string(opts.SecurityType),
		"MinRelativeSizeSelected":     string(opts.MinRelativeSize),
		"MaxTradeRankSelected":        string(opts.MaxTradeRank),
		"NormalPrintsSelected":        common.BoolString(opts.NormalPrints),
		"SignaturePrintsSelected":     common.BoolString(opts.SignaturePrints),
		"LatePrintsSelected":          common.BoolString(opts.LatePrints),
		"TimelyPrintsSelected":        common.BoolString(opts.TimelyPrints),
		"DarkPoolsSelected":           common.BoolString(opts.DarkPools),
		"LitExchangesSelected":        common.BoolString(opts.LitExchanges),
		"SweepsSelected":              common.BoolString(opts.Sweeps),
		"BlocksSelected":              common.BoolString(opts.Blocks),
		"PremarketTradesSelected":     common.BoolString(opts.PremarketTrades),
		"RTHTradesSelected":           common.BoolString(opts.RTHTrades),
		"AHTradesSelected":            common.BoolString(opts.AHTrades),
		"OpeningTradesSelected":       common.BoolString(opts.OpeningTrades),
		"ClosingTradesSelected":       common.BoolString(opts.ClosingTrades),
		"PhantomTradesSelected":       common.BoolString(opts.PhantomTrades),
		"OffsettingTradesSelected":    common.BoolString(opts.OffsettingTrades),
		"RSIOverboughtDailySelected":  string(opts.RSIOverboughtDaily),
		"RSIOverboughtHourlySelected": string(opts.RSIOverboughtHourly),
		"RSIOversoldDailySelected":    string(opts.RSIOversoldDaily),
		"RSIOversoldHourlySelected":   string(opts.RSIOversoldHourly),
	}
}

// runCreateEdit handles both the create and edit subcommands. A key of 0
// indicates a new watchlist; a non-zero key indicates an edit.
func runCreateEdit(cmd *cobra.Command, opts *watchlistConfigFlags, key int) error {
	ctx := cmd.Context()

	vlClient, err := common.NewCommandClient(ctx)
	if err != nil {
		return err
	}

	fields := buildWatchlistConfigFields(opts, key)
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
