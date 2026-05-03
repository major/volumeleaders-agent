package watchlist

import (
	"fmt"
	"log/slog"
	"net/url"
	"strconv"

	"github.com/leodido/structcli"
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

func init() {
	structcli.RegisterEnum[watchlistSecurityType](map[watchlistSecurityType][]string{
		watchlistSecurityAll:    {"-1"},
		watchlistSecurityStocks: {"1"},
		watchlistSecurityETFs:   {"26"},
		watchlistSecurityREITs:  {"4"},
	})
	structcli.RegisterEnum[watchlistRelativeSize](map[watchlistRelativeSize][]string{
		watchlistRelativeSizeAll: {"0"},
		watchlistRelativeSize5:   {"5"},
		watchlistRelativeSize10:  {"10"},
		watchlistRelativeSize25:  {"25"},
		watchlistRelativeSize50:  {"50"},
		watchlistRelativeSize100: {"100"},
	})
	structcli.RegisterEnum[watchlistTradeRank](map[watchlistTradeRank][]string{
		watchlistTradeRankAll: {"-1"},
		watchlistTradeRank1:   {"1"},
		watchlistTradeRank3:   {"3"},
		watchlistTradeRank5:   {"5"},
		watchlistTradeRank10:  {"10"},
		watchlistTradeRank25:  {"25"},
		watchlistTradeRank50:  {"50"},
		watchlistTradeRank100: {"100"},
	})
}

// watchlistConfigsOptions holds flags for the "watchlist configs" subcommand.
type watchlistConfigsOptions struct {
	Format common.OutputFormat `flag:"format" flaggroup:"Output" flagshort:"f" default:"json" flagdescr:"Output format: json, csv, or tsv"`
}

// watchlistTickersOptions holds flags for the "watchlist tickers" subcommand.
type watchlistTickersOptions struct {
	WatchlistKey int                 `flag:"watchlist-key" flaggroup:"Input" flagshort:"k" flagdescr:"Watch list key (-1 for all)"`
	Format       common.OutputFormat `flag:"format" flaggroup:"Output" flagshort:"f" default:"json" flagdescr:"Output format: json, csv, or tsv"`
}

// watchlistDeleteOptions holds flags for the "watchlist delete" subcommand.
type watchlistDeleteOptions struct {
	Key int `flag:"key" flaggroup:"Input" flagshort:"k" flagrequired:"true" flagdescr:"Watch list key to delete"`
}

// watchlistAddTickerOptions holds flags for the "watchlist add-ticker" subcommand.
type watchlistAddTickerOptions struct {
	WatchlistKey int    `flag:"watchlist-key" flaggroup:"Input" flagshort:"k" flagrequired:"true" flagdescr:"Watch list key"`
	Ticker       string `flag:"ticker" flaggroup:"Input" flagshort:"t" flagrequired:"true" flagdescr:"Ticker symbol to add"`
}

// watchlistConfigFlags holds the shared flag set for watchlist create/edit commands.
type watchlistConfigFlags struct {
	Name                string                `flag:"name" flaggroup:"Basic" flagdescr:"Watch list name"`
	Tickers             string                `flag:"tickers" flaggroup:"Basic" flagshort:"t" flagdescr:"Comma-separated ticker symbols (max 500)"`
	MinVolume           int                   `flag:"min-volume" flaggroup:"Ranges" flagdescr:"Minimum volume filter"`
	MaxVolume           int                   `flag:"max-volume" flaggroup:"Ranges" flagdescr:"Maximum volume filter"`
	MinDollars          float64               `flag:"min-dollars" flaggroup:"Ranges" flagdescr:"Minimum dollars filter"`
	MaxDollars          float64               `flag:"max-dollars" flaggroup:"Ranges" flagdescr:"Maximum dollars filter"`
	MinPrice            float64               `flag:"min-price" flaggroup:"Ranges" flagdescr:"Minimum price filter"`
	MaxPrice            float64               `flag:"max-price" flaggroup:"Ranges" flagdescr:"Maximum price filter"`
	MinVCD              float64               `flag:"min-vcd" flaggroup:"Filters" flagdescr:"Minimum VCD percentile (0-100)"`
	SectorIndustry      string                `flag:"sector-industry" flaggroup:"Filters" flagdescr:"Sector/industry filter (max 100 chars)"`
	SecurityType        watchlistSecurityType `flag:"security-type" flaggroup:"Filters" flagdescr:"Security type (-1=all, 1=stocks, 26=ETFs, 4=REITs)"`
	MinRelativeSize     watchlistRelativeSize `flag:"min-relative-size" flaggroup:"Filters" flagdescr:"Minimum relative size (0/5/10/25/50/100)"`
	MaxTradeRank        watchlistTradeRank    `flag:"max-trade-rank" flaggroup:"Filters" flagdescr:"Maximum trade rank (-1=all, 1/3/5/10/25/50/100)"`
	NormalPrints        bool                  `flag:"normal-prints" flaggroup:"Print Types" flagdescr:"Include normal prints"`
	SignaturePrints     bool                  `flag:"signature-prints" flaggroup:"Print Types" flagdescr:"Include signature prints"`
	LatePrints          bool                  `flag:"late-prints" flaggroup:"Print Types" flagdescr:"Include late prints"`
	TimelyPrints        bool                  `flag:"timely-prints" flaggroup:"Print Types" flagdescr:"Include timely prints"`
	DarkPools           bool                  `flag:"dark-pools" flaggroup:"Venues" flagdescr:"Include dark pool trades"`
	LitExchanges        bool                  `flag:"lit-exchanges" flaggroup:"Venues" flagdescr:"Include lit exchange trades"`
	Sweeps              bool                  `flag:"sweeps" flaggroup:"Venues" flagdescr:"Include sweep trades"`
	Blocks              bool                  `flag:"blocks" flaggroup:"Venues" flagdescr:"Include block trades"`
	PremarketTrades     bool                  `flag:"premarket-trades" flaggroup:"Sessions" flagdescr:"Include premarket trades"`
	RTHTrades           bool                  `flag:"rth-trades" flaggroup:"Sessions" flagdescr:"Include regular trading hours trades"`
	AHTrades            bool                  `flag:"ah-trades" flaggroup:"Sessions" flagdescr:"Include after-hours trades"`
	OpeningTrades       bool                  `flag:"opening-trades" flaggroup:"Sessions" flagdescr:"Include opening trades"`
	ClosingTrades       bool                  `flag:"closing-trades" flaggroup:"Sessions" flagdescr:"Include closing trades"`
	PhantomTrades       bool                  `flag:"phantom-trades" flaggroup:"Sessions" flagdescr:"Include phantom trades"`
	OffsettingTrades    bool                  `flag:"offsetting-trades" flaggroup:"Sessions" flagdescr:"Include offsetting trades"`
	RSIOverboughtDaily  common.TriStateFilter `flag:"rsi-overbought-daily" flaggroup:"RSI" flagdescr:"RSI overbought daily (-1=ignore, 0=no, 1=yes)"`
	RSIOverboughtHourly common.TriStateFilter `flag:"rsi-overbought-hourly" flaggroup:"RSI" flagdescr:"RSI overbought hourly (-1=ignore, 0=no, 1=yes)"`
	RSIOversoldDaily    common.TriStateFilter `flag:"rsi-oversold-daily" flaggroup:"RSI" flagdescr:"RSI oversold daily (-1=ignore, 0=no, 1=yes)"`
	RSIOversoldHourly   common.TriStateFilter `flag:"rsi-oversold-hourly" flaggroup:"RSI" flagdescr:"RSI oversold hourly (-1=ignore, 0=no, 1=yes)"`
}

// watchlistCreateOptions holds flags for the "watchlist create" subcommand.
type watchlistCreateOptions struct {
	watchlistConfigFlags
}

// watchlistEditOptions holds flags for the "watchlist edit" subcommand.
type watchlistEditOptions struct {
	Key int `flag:"key" flaggroup:"Input" flagshort:"k" flagrequired:"true" flagdescr:"Watch list key to edit"`
	watchlistConfigFlags
}

// presetWatchlistConfigDefaults sets non-zero default values on watchlistConfigFlags
// before structcli.Bind so they become pflag defaults without using default tags.
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
	opts := &watchlistConfigsOptions{}
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
	if err := structcli.Bind(cmd, opts); err != nil {
		panic(fmt.Sprintf("structcli.Bind configs: %v", err))
	}
	return cmd
}

// newTickersCmd returns the "tickers" subcommand.
func newTickersCmd() *cobra.Command {
	opts := &watchlistTickersOptions{}
	opts.WatchlistKey = -1
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
	if err := structcli.Bind(cmd, opts); err != nil {
		panic(fmt.Sprintf("structcli.Bind tickers: %v", err))
	}
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
	if err := structcli.Bind(cmd, opts); err != nil {
		panic(fmt.Sprintf("structcli.Bind delete: %v", err))
	}
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
	if err := structcli.Bind(cmd, opts); err != nil {
		panic(fmt.Sprintf("structcli.Bind add-ticker: %v", err))
	}
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
	if err := structcli.Bind(cmd, opts); err != nil {
		panic(fmt.Sprintf("structcli.Bind create: %v", err))
	}
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

// newEditCmd returns the "edit" subcommand.
func newEditCmd() *cobra.Command {
	opts := &watchlistEditOptions{}
	presetWatchlistConfigDefaults(&opts.watchlistConfigFlags)
	cmd := &cobra.Command{
		Use:        "edit",
		Short:      "Edit an existing watch list configuration",
		Long:       "Modify an existing watchlist configuration identified by its numeric key. Requires --key with the watchlist key. Specify only the fields you want to change; unspecified fields retain their current values.",
		Example:    `volumeleaders-agent watchlist edit --key 1 --name "Updated watchlist" --tickers AAPL,MSFT`,
		Args:       cobra.NoArgs,
		SuggestFor: []string{"edt", "modify"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCreateEdit(cmd, &opts.watchlistConfigFlags, opts.Key)
		},
	}
	if err := structcli.Bind(cmd, opts); err != nil {
		panic(fmt.Sprintf("structcli.Bind edit: %v", err))
	}
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
