package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/leodido/structcli"
	"github.com/leodido/structcli/helptopics"
	structclimcp "github.com/leodido/structcli/mcp"
	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/alert"
	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/cli/market"
	"github.com/major/volumeleaders-agent/internal/cli/report"
	"github.com/major/volumeleaders-agent/internal/cli/trade"
	"github.com/major/volumeleaders-agent/internal/cli/volume"
	"github.com/major/volumeleaders-agent/internal/cli/watchlist"
)

// rootOptions holds flags bound to the root command via structcli.Bind.
// The bind pipeline populates these fields before PersistentPreRunE fires.
type rootOptions struct {
	Pretty bool `flag:"pretty" flaggroup:"Output" flagshort:"p" flagdescr:"Pretty-print JSON output with indentation"`
}

// NewRootCmd returns the root cobra command for volumeleaders-agent.
func NewRootCmd(version string) *cobra.Command {
	opts := &rootOptions{}
	cmd := &cobra.Command{
		Use:   "volumeleaders-agent",
		Short: "CLI tool for querying VolumeLeaders institutional trade data",
		Long: `volumeleaders-agent queries institutional trade data from VolumeLeaders. Use it for trades, volume leaderboards, market data, alerts, and watchlists.

Auth: reads browser cookies automatically. If auth fails with exit code 2 and "Authentication required: VolumeLeaders session has expired.", log in at https://www.volumeleaders.com in your browser, then retry.

Output: compact JSON to stdout by default. Use --pretty before the command group for indented JSON. Use --jsonschema on any command for machine-readable input JSON Schema output, --jsonschema=tree on the root for the full CLI tree, outputschema for machine-readable stdout contracts, or --mcp on the root to serve leaf commands as MCP tools over stdio. Errors and logs go to stderr.

COMMAND CHOOSER

Goal                                          Start with                              Notes
--------------------------------------------  --------------------------------------  -----------------------------------------------
Run safe preset trade scans                   report list                             Prefer reports before raw trade filters
Find ranked institutional prints              report top-100-rank                     Vetted browser preset, timeout-aware defaults
Find strongest ranked prints                  report top-10-rank                      Narrower ranked-trade preset
Find dark pool sweep activity                 report dark-pool-sweeps                 Vetted dark-pool sweep preset
Find unusually large prints                   report disproportionately-large          5x relative size browser preset
Find individual institutional prints          trade list X --days N                   Advanced path: use presets or tickers first
Get comprehensive ticker overview            trade dashboard X --days N              Fast chart-style trades, clusters, levels, bombs
Compare leveraged ETF bull/bear flow          trade sentiment --days N                Fixed leveraged ETF universe, not buy/sell flow
Find converging price-level activity          trade clusters --days N                 Cluster conviction around similar prices
Find sudden aggressive bursts                 trade cluster-bombs --days N            Burst detection, different defaults than clusters
Inspect trade or cluster alerts               trade alerts --date D                   System-generated alerts
Find support/resistance levels                trade levels X --days N                 One ticker, capped level count
Find revisits to institutional levels         trade level-touches X --days N          Level retests, capped pagination
See institutional volume leaders              volume institutional --date D            Same trade model, volume-ranked
See after-hours institutional leaders         volume ah-institutional --date D        After-hours institutional flow
See total volume leaders                      volume total --date D                   Total market volume across trade types
Get current prices                            market snapshots                        JSON object
Find earnings with prior institutional flow   market earnings --days N                CSV/TSV supported
Check exhaustion/reversal signals             market exhaustion --date D              Lower rank is stronger
Manage alert configs                          alert configs/create/edit/delete        Edit replaces unspecified values with defaults
Manage watchlists                             watchlist configs/create/edit/delete    Edit replaces unspecified values with defaults
Get watchlist tickers                         watchlist tickers --watchlist-key K     Key comes from watchlist configs

ANALYSIS WORKFLOW

1. report list to choose a vetted preset report before raw filters.
2. report top-100-rank or report disproportionately-large for the broad scan.
3. trade dashboard X --days N for a fast ticker overview before deeper drilling.
4. trade list --preset NAME only when report commands are not specific enough.
5. trade levels X --days N for support/resistance.
6. trade clusters X --days N when prints appear concentrated around a price.
7. market earnings --days N and market exhaustion --date D for event and reversal context.

GLOBAL CONVENTIONS

Dates: YYYY-MM-DD. Commands with date ranges accept either --start-date D --end-date D or --days N. --days counts backward from today unless --end-date is also set, and cannot be combined with --start-date.

Pagination: --start offset, --length count, --length -1 means all rows unless a capped endpoint rejects it. trade list does not expose --length; multi-day lookups whose effective filters include tickers return the top 10 long-period trades with VolumeLeaders' lightweight chart query shape, while trade list --summary, single-day trade scans, all-market trade scans, sector-only presets, trade clusters, and trade cluster-bombs fetch all rows internally in browser-sized 100-row pages. trade level-touches only allows 1 to 50 rows. trade levels and trade level-touches only allow --trade-level-count values of 5, 10, 20, or 50.

Toggle filters: -1 means all/unfiltered, 0 means exclude, 1 means include/only.

Tickers: --tickers is comma-separated, --ticker is single-symbol. Commands that take tickers generally accept positional tickers too, for example: trade list XLE XLK. Trade and volume ticker filters also accept --symbol and --symbols aliases.

Output formats: list-style commands may support --format json/csv/tsv. CSV/TSV include headers, booleans render as true/false, null or missing values render as empty cells. Nested summaries and single-object commands are JSON-only unless the input schema shows a format flag. Use outputschema to inspect the success stdout shape for each command.

Performance: use report commands and built-in presets first. Start with one vetted report, one day, and tickers when possible, then expand. VolumeLeaders endpoints can be expensive; broad custom trade list filters are easy to overdo. report commands reject broad multi-day scans without tickers, trade list uses a bounded chart-style request for multi-day ticker lookups, and full-result retrieval keeps the browser's 100-row page size.

RECOVERY PLAYBOOK

Authentication failed or exit code 2: log in at https://www.volumeleaders.com in the same browser profile, confirm the site loads, then retry the exact command. Do not paste cookies or session values into commands.

Date validation failed: use YYYY-MM-DD. For required ranges, provide either --start-date D --end-date D or --days N. Do not combine --days with --start-date.

Pagination validation failed: reduce --length to the documented cap. trade level-touches accepts 1 to 50 rows per request. Do not add --length to trade list, trade clusters, or trade cluster-bombs because they page internally at 100 rows per request.

Unknown flag or enum value: run the same command with --help or --jsonschema to inspect supported flags, defaults, allowed values, and required fields before retrying.

Empty or too broad output: use report list to pick a vetted preset report first, then add tickers or explicit dates. If JSON is too verbose, use --fields where supported or --format csv for list-style commands. Avoid hand-building raw filters unless report commands and trade list --preset cannot answer the question.

COMMAND SEQUENCES

Broad scan: report top-100-rank, then report disproportionately-large, then trade dashboard TICKER --days N, then trade levels TICKER --days N.

Preset workflow: report list, then report NAME for safe defaults, then trade list --preset NAME only if advanced customization is needed.

Ticker drilldown: trade dashboard TICKER --days N, then trade list TICKER --days N, then trade clusters TICKER --days N.

Event context: market earnings --days N, then trade list TICKER --start-date D --end-date D, then market exhaustion --date D.

Watchlist workflow: watchlist configs to find keys and names, watchlist tickers --watchlist-key K to inspect symbols, then trade list --watchlist NAME --days N.`,
		Version:          version,
		SilenceErrors:    true,
		SilenceUsage:     true,
		TraverseChildren: true,
		Args:             cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
			cmd.SetContext(context.WithValue(cmd.Context(), common.PrettyJSONKey, opts.Pretty))
			return nil
		},
	}
	cmd.AddGroup(
		&cobra.Group{ID: "trading", Title: "Trading Commands:"},
		&cobra.Group{ID: "volume", Title: "Volume Commands:"},
		&cobra.Group{ID: "market", Title: "Market Commands:"},
		&cobra.Group{ID: "alerts", Title: "Alert Commands:"},
		&cobra.Group{ID: "watchlists", Title: "Watchlist Commands:"},
		&cobra.Group{ID: "reference", Title: "Reference Commands:"},
	)
	cmd.AddCommand(
		report.NewCmd(),
		trade.NewCmd(),
		volume.NewVolumeCommand(),
		market.NewMarketCommand(),
		alert.NewAlertCommand(),
		watchlist.NewCmd(),
		newOutputSchemaCmd(),
	)
	common.BindOrPanic(cmd, opts, "root options")
	return cmd
}

// SetupCLI configures structcli features on the root command. Called from main
// after NewRootCmd; separated because WithJSONSchema uses cobra.OnInitialize
// (process-global) which races in parallel tests.
func SetupCLI(cmd *cobra.Command) {
	if err := structcli.Setup(
		cmd,
		structcli.WithAppName("volumeleaders-agent"),
		structcli.WithJSONSchema(),
		structcli.WithHelpTopics(helptopics.Options{ReferenceSection: true}),
		structcli.WithFlagErrors(),
		structcli.WithMCP(structclimcp.Options{
			Name:    "volumeleaders-agent",
			Version: cmd.Version,
			Exclude: []string{
				"completion-bash",
				"completion-fish",
				"completion-powershell",
				"completion-zsh",
			},
		}),
	); err != nil {
		panic(fmt.Sprintf("structcli.Setup: %v", err))
	}
}
