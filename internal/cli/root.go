package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/leodido/structcli"
	"github.com/leodido/structcli/helptopics"
	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/alert"
	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/cli/market"
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

Output: compact JSON to stdout by default. Use --pretty before the command group for indented JSON. Use --jsonschema on any command for machine-readable JSON Schema output, or --jsonschema=tree on the root for the full CLI tree. Errors and logs go to stderr.

COMMAND CHOOSER

Goal                                          Start with                              Notes
--------------------------------------------  --------------------------------------  -----------------------------------------------
Find individual institutional prints          trade list X --days N                   Use ticker filters, presets, or watchlists
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

1. volume institutional --date D for top dollar movers.
2. trade list X --days N for individual prints.
3. trade levels X --days N for support/resistance.
4. trade clusters X --days N when prints appear concentrated around a price.
5. market earnings --days N and market exhaustion --date D for event and reversal context.

GLOBAL CONVENTIONS

Dates: YYYY-MM-DD. Commands with date ranges accept either --start-date D --end-date D or --days N. --days counts backward from today unless --end-date is also set, and cannot be combined with --start-date.

Pagination: --start offset, --length count, --length -1 means all rows unless a capped endpoint rejects it. trade list, trade list --summary, and trade level-touches only allow 1 to 50 rows. trade levels caps --trade-level-count at 50.

Toggle filters: -1 means all/unfiltered, 0 means exclude, 1 means include/only.

Tickers: --tickers is comma-separated, --ticker is single-symbol. Commands that take tickers generally accept positional tickers too, for example: trade list XLE XLK. Trade and volume ticker filters also accept --symbol and --symbols aliases.

Output formats: list-style commands may support --format json/csv/tsv. CSV/TSV include headers, booleans render as true/false, null or missing values render as empty cells. Nested summaries and single-object commands are JSON-only unless the schema shows a format flag.

Performance: use explicit dates and tickers when possible. Start narrow, then expand. VolumeLeaders endpoints can be expensive and some trade retrieval endpoints are intentionally capped.

RECOVERY PLAYBOOK

Authentication failed or exit code 2: log in at https://www.volumeleaders.com in the same browser profile, confirm the site loads, then retry the exact command. Do not paste cookies or session values into commands.

Date validation failed: use YYYY-MM-DD. For required ranges, provide either --start-date D --end-date D or --days N. Do not combine --days with --start-date.

Pagination validation failed: reduce --length to the documented cap. trade list, trade list --summary, and trade level-touches accept 1 to 50 rows per request. Use --start to page through more rows.

Unknown flag or enum value: run the same command with --help or --jsonschema to inspect supported flags, defaults, allowed values, and required fields before retrying.

Empty or too broad output: add tickers, explicit dates, min dollar filters, or a preset first. If JSON is too verbose, use --fields where supported or --format csv for list-style commands.

COMMAND SEQUENCES

Broad scan: volume institutional --date D, then trade list TICKER --days N, then trade levels TICKER --days N.

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
	)
	cmd.AddCommand(
		trade.NewCmd(),
		volume.NewVolumeCommand(),
		market.NewMarketCommand(),
		alert.NewAlertCommand(),
		watchlist.NewCmd(),
	)
	if err := structcli.Bind(cmd, opts); err != nil {
		panic(fmt.Sprintf("structcli.Bind root options: %v", err))
	}
	return cmd
}

// SetupCLI configures structcli features (JSON schema, structured flag errors)
// on the root command. Called from main after NewRootCmd; separated because
// WithJSONSchema uses cobra.OnInitialize (process-global) which races in
// parallel tests.
func SetupCLI(cmd *cobra.Command) {
	if err := structcli.Setup(
		cmd,
		structcli.WithAppName("volumeleaders-agent"),
		structcli.WithJSONSchema(),
		structcli.WithHelpTopics(helptopics.Options{ReferenceSection: true}),
		structcli.WithFlagErrors(),
	); err != nil {
		panic(fmt.Sprintf("structcli.Setup: %v", err))
	}
}
