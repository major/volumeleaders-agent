# volumeleaders-agent

volumeleaders-agent queries institutional trade data from VolumeLeaders. Use it for trades, volume leaderboards, market data, alerts, and watchlists.

Auth: reads browser cookies automatically. If auth fails with exit code 2 and "Authentication required: VolumeLeaders session has expired.", log in at https://www.volumeleaders.com in your browser, then retry.

Output: compact JSON to stdout by default. Use --pretty before the command group for indented JSON. Use --jsonschema on any command for machine-readable input JSON Schema output, --jsonschema=tree on the root for the full CLI tree, outputschema for machine-readable stdout contracts, or --mcp on the root to serve leaf commands as MCP tools over stdio. Errors and logs go to stderr.

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

Pagination: --start offset, --length count, --length -1 means all rows unless a capped endpoint rejects it. trade list, trade list --summary, trade clusters, and trade cluster-bombs fetch all rows internally in browser-sized 100-row pages and do not expose --length. trade level-touches only allows 1 to 50 rows. trade levels and trade level-touches only allow --trade-level-count values of 5, 10, 20, or 50.

Toggle filters: -1 means all/unfiltered, 0 means exclude, 1 means include/only.

Tickers: --tickers is comma-separated, --ticker is single-symbol. Commands that take tickers generally accept positional tickers too, for example: trade list XLE XLK. Trade and volume ticker filters also accept --symbol and --symbols aliases.

Output formats: list-style commands may support --format json/csv/tsv. CSV/TSV include headers, booleans render as true/false, null or missing values render as empty cells. Nested summaries and single-object commands are JSON-only unless the input schema shows a format flag. Use outputschema to inspect the success stdout shape for each command.

Performance: use explicit dates and tickers when possible. Start narrow, then expand. VolumeLeaders endpoints can be expensive, and trade retrieval endpoints intentionally match the browser's 100-row page size.

RECOVERY PLAYBOOK

Authentication failed or exit code 2: log in at https://www.volumeleaders.com in the same browser profile, confirm the site loads, then retry the exact command. Do not paste cookies or session values into commands.

Date validation failed: use YYYY-MM-DD. For required ranges, provide either --start-date D --end-date D or --days N. Do not combine --days with --start-date.

Pagination validation failed: reduce --length to the documented cap. trade level-touches accepts 1 to 50 rows per request. Do not add --length to trade list, trade clusters, or trade cluster-bombs because they page internally at 100 rows per request.

Unknown flag or enum value: run the same command with --help or --jsonschema to inspect supported flags, defaults, allowed values, and required fields before retrying.

Empty or too broad output: add tickers, explicit dates, min dollar filters, or a preset first. If JSON is too verbose, use --fields where supported or --format csv for list-style commands.

COMMAND SEQUENCES

Broad scan: volume institutional --date D, then trade list TICKER --days N, then trade levels TICKER --days N.

Event context: market earnings --days N, then trade list TICKER --start-date D --end-date D, then market exhaustion --date D.

Watchlist workflow: watchlist configs to find keys and names, watchlist tickers --watchlist-key K to inspect symbols, then trade list --watchlist NAME --days N.

## Installation

```bash
go install github.com/major/volumeleaders-agent/cmd/volumeleaders-agent@latest
```

## Commands

| Command | Description | Required Flags |
|---------|-------------|---------------|
| `volumeleaders-agent alert configs` | List all saved alert configurations with their keys, names, ticker filters, trade conditions, and notification settings. Outputs compact JSON or CSV/TSV with --format. Use --fields to select specific output fields. |  |
| `volumeleaders-agent alert create` | Create a new alert configuration with a name and filter settings for institutional trade activity. Requires --name. Specify filters such as trade rank, dollar thresholds, dark pool and sweep conditions, and ticker scope. Returns a success response with the new configuration key. | `--name` |
| `volumeleaders-agent alert delete` | Remove a saved alert configuration by its numeric key. Requires --key with the alert config key (visible in configs output). The deletion is permanent and cannot be undone. | `--key` |
| `volumeleaders-agent alert edit` | Modify an existing alert configuration identified by its numeric key. Requires --key with the alert config key. Specify the fields you want to set; unspecified fields are replaced with their default values. | `--key` |
| `volumeleaders-agent market earnings` | Query the earnings calendar for a date range, showing tickers with earnings dates and associated trade activity counts. Requires --start-date and --end-date (or --days). Outputs compact JSON or CSV/TSV with --format. PREREQUISITES: provide a date range with --days or explicit start and end dates. RECOVERY: if date validation fails, use --days N for the fastest retry or provide both --start-date and --end-date. NEXT STEPS: run trade list for tickers near earnings, then market exhaustion for broader reversal context. |  |
| `volumeleaders-agent market exhaustion` | Query exhaustion scores that indicate overbought or oversold market conditions based on institutional trade clustering patterns. Omit --date to query the current trading day. Outputs compact JSON with rank metrics at different lookback periods. |  |
| `volumeleaders-agent market snapshots` | Retrieve current price snapshot data for all symbols tracked by VolumeLeaders, returning the latest available price and volume data. No date filtering is available; always returns the most recent data. Outputs compact JSON by default. |  |
| `volumeleaders-agent outputschema` | Print machine-readable stdout contracts for executable commands. With no arguments it returns every contract as a JSON array. Pass a command path such as trade list to return one contract. This describes success output only; structured errors are documented by structcli flag errors. |  |
| `volumeleaders-agent trade alerts` | Query trade alerts fired on a specific date based on saved alert configurations. Requires --date. Returns alert records matching your configured filters. Outputs compact JSON or CSV/TSV with --format.

Alert configs trigger when trades match thresholds. Threshold names follow the pattern CategoryMetricLTE or CategoryMetricGTE where LTE is maximum rank and GTE is minimum value. Use alert configs to see your configured thresholds. | `--date` |
| `volumeleaders-agent trade cluster-alerts` | Query trade cluster alerts fired on a specific date based on saved alert configurations that target cluster activity. Requires --date. Returns cluster alert records matching your configured filters.

Cluster alert rows use the full cluster-shaped model rather than the compact default from trade clusters. Use trade alerts for individual trade alert rows and this command for cluster-level alert rows. | `--date` |
| `volumeleaders-agent trade cluster-bombs` | Query trade cluster bombs, which are extreme-magnitude trade clusters that exceed normal institutional activity thresholds. Filterable by ticker, date range, dollar amounts, sector, and cluster bomb rank. Outputs compact JSON by default.

Results are fetched in browser-sized 100-row pages to match VolumeLeaders' frontend behavior. Cluster bombs find sudden aggressive bursts tightly grouped in time and price, with different defaults and rank fields than trade clusters. Use this command when looking for extreme concentration events, not general price-level clustering. |  |
| `volumeleaders-agent trade clusters` | Query aggregated trade clusters, which group multiple trades in a short window into a single cluster record. Filterable by ticker, date range, dollar amounts, sector, and trade cluster rank. Outputs compact JSON or CSV/TSV with --format.


Results are fetched in browser-sized 100-row pages to match VolumeLeaders' frontend behavior. Use clusters when the question is about price-level concentration, not single prints. This command uses larger default dollar thresholds than ordinary trade list. Use trade cluster-bombs instead when looking for sudden aggressive bursts tightly grouped in time and price. |  |
| `volumeleaders-agent trade level-touches` | Query institutional trade events that occurred at notable price levels for a ticker, showing how the market interacted with key support and resistance zones. Accepts a ticker as positional argument or via --ticker flag. Requires --start-date and --end-date (or --days).

Defaults to --trade-level-rank 5 and --length 50, rejects --length -1, --length 0, and values above 50, and only allows --trade-level-count values of 5, 10, 20, or 50. Use trade levels first to identify significant price zones, then use this command to find events where price revisited those levels.

PREREQUISITES: Provide exactly one ticker and a date range with --start-date and --end-date or --days.

RECOVERY: If --length is rejected, use 1 to 50. If --trade-level-count is rejected, use 5, 10, 20, or 50. If --trade-level-rank is rejected, use 5 or higher. If dates are missing, add --days N for a quick retry.

NEXT STEPS: Compare touched levels with fresh trade list output to see whether recent institutional prints confirm or reject the level. |  |
| `volumeleaders-agent trade levels` | Query significant price levels for a ticker, showing historical support and resistance zones identified by institutional trade clustering. Accepts a ticker as positional argument or via --ticker flag. Outputs compact JSON by default.

Defaults to a 365-day lookback when dates are omitted. Uses non-standard --relative-size 0 and only allows --trade-level-count values of 5, 10, 20, or 50. Default JSON is compact and omits repetitive ticker metadata and the verbose Dates list; use --fields all or CSV/TSV when raw fields are needed.

PREREQUISITES: Provide exactly one ticker as a positional argument or with --ticker.

RECOVERY: If ticker validation fails, use one ticker only. If --trade-level-count is rejected, use 5, 10, 20, or 50.

NEXT STEPS: Use trade level-touches with the same ticker and date range to find trades that revisited these levels. |  |
| `volumeleaders-agent trade list` | Query individual institutional trades from VolumeLeaders, filterable by ticker, date range, dollar amounts, volume, trade conditions, session type, and trade rank. Supports built-in filter presets (--preset) and watchlist-based filtering (--watchlist). Outputs compact JSON or CSV/TSV with --format; use --summary for aggregate metrics grouped by ticker or day.

Date defaults: 365-day lookback when tickers are provided, today-only without tickers. Preset and watchlist filters do not supply dates. Filter precedence is preset baseline, then watchlist merge, then explicit CLI flags override both.

Default JSON is compact and omits repetitive/internal fields. Use --fields FIELD1,FIELD2, CSV/TSV, or --fields all where supported when raw API fields are needed. --summary returns aggregate JSON with valid --group-by values of ticker, day, or ticker,day; do not combine summary mode with --fields or non-JSON formats.

KEY METRICS

Field                      Meaning
-------------------------  ---------------------------------------------------------------
CumulativeDistribution     Volume percentile, 0 to 1, higher means more accumulation
DollarsMultiplier          Trade dollars relative to average block size
TradeRank                  VL significance rank now, lower is stronger
TradeRankSnapshot          VL significance rank at print time, lower is stronger
TradeClusterRank           Rank for cluster significance, lower is stronger
TradeClusterBombRank       Rank for burst significance, lower is stronger
TradeLevelRank             Rank for level significance, lower is stronger
RelativeSize               Trade size vs normal activity
PercentDailyVolume         Trade volume as percent of average daily volume
VCD                        Volume Confirmation Distribution score
FrequencyLast30TD          Similar-magnitude trade frequency over last 30 trading days
FrequencyLast90TD          Similar-magnitude trade frequency over last 90 trading days
FrequencyLast1CY           Similar-magnitude trade frequency over last calendar year
RSIHour                    Hourly RSI at time of trade
RSIDay                     Daily RSI at time of trade
DarkPool                   Boolean: trade printed on a dark pool
Sweep                      Boolean: trade was a sweep order
LatePrint                  Boolean: trade was a late print
SignaturePrint             Boolean: trade matched a signature print pattern
PhantomPrint               Boolean: trade was a phantom print
InsideBar                  Boolean: bar was an inside bar

Shared trade filters include volume, price, dollars, conditions, VCD, relative size, security type, market cap, trade rank, dark pools, sweeps, late prints, signature prints, even-share prints, and session/event toggles.

PREREQUISITES: Browser authentication. For reproducible scans, pass explicit dates or --days plus tickers, preset, watchlist, or sector filters.

RECOVERY: Results are fetched in browser-sized 100-row pages. If --summary rejects --fields or --format, rerun summary as JSON without --fields. If date flags conflict, use either --days or --start-date with --end-date.

NEXT STEPS: Use trade levels for support/resistance after finding a ticker, trade clusters when prints concentrate near a price, or trade sentiment for leveraged ETF bull/bear context. |  |
| `volumeleaders-agent trade preset-tickers` | Extract the ticker symbols configured in a named trade filter preset, showing whether the preset uses an explicit ticker list, a sector/industry filter, or is unfiltered. Requires --preset with the preset name (case-insensitive). Outputs JSON with the preset name, group, type, and ticker details. | `--preset` |
| `volumeleaders-agent trade presets` | List all built-in trade filter presets with their names, groups, and filter configurations. Each preset defines a named set of filters that can be applied to trade list queries via --preset. Outputs compact JSON by default; use --format csv or tsv for tabular output. |  |
| `volumeleaders-agent trade sentiment` | Summarize leveraged ETF bull and bear flow by trading day, showing aggregate institutional dollar volume on the bull and bear side. Requires --start-date and --end-date (or --days). Outputs one record per day with bull and bear totals.

This command always queries the combined leveraged ETF sector filter SectorIndustry=X B, classifies bull and bear ETFs locally, and cannot be constrained by ticker or sector flags. Non-standard defaults include --min-dollars 5000000 and --vcd 97; shared --relative-size 5 still applies.

Ratio is bull dollars divided by bear dollars and is null when bear flow is zero. Treat the output as leveraged ETF proxy flow, not signed buy/sell flow for the broader market. |  |
| `volumeleaders-agent volume ah-institutional` | Query the after-hours institutional volume leaderboard, ranking tickers by total institutional trade volume during after-hours sessions for a given date. Accepts optional ticker positional arguments; also accepts --tickers flag. Requires --date. | `--date` |
| `volumeleaders-agent volume institutional` | Query the regular-hours institutional volume leaderboard, ranking tickers by total institutional trade volume for a given date. Accepts optional ticker positional arguments to filter results; also accepts --tickers flag. Requires --date. Outputs compact JSON or CSV/TSV with --format. PREREQUISITES: choose a trading date in YYYY-MM-DD format. RECOVERY: if --date is missing or invalid, retry with an explicit trading day. NEXT STEPS: run trade list for interesting tickers, then trade levels for support and resistance context. | `--date` |
| `volumeleaders-agent volume total` | Query the total volume leaderboard combining all session types, ranking tickers by total institutional trade volume for a given date. Accepts optional ticker positional arguments; also accepts --tickers flag. Requires --date. | `--date` |
| `volumeleaders-agent watchlist add-ticker` | Add a ticker symbol to an existing watchlist. Requires --watchlist-key with the watchlist key and --ticker with the symbol to add. The ticker is appended to the watchlist without affecting existing symbols. | `--ticker`, `--watchlist-key` |
| `volumeleaders-agent watchlist configs` | List all saved watchlist configurations with their keys and names. Outputs compact JSON or CSV/TSV with --format. Each row shows the watchlist key and name; use the tickers subcommand to view symbols in a specific watchlist. |  |
| `volumeleaders-agent watchlist create` | Create a new watchlist configuration with a name and optional filter settings such as minimum volume, price range, sector, and trade conditions. Requires --name. Use --tickers to specify an explicit ticker list or leave unset for a filter-based watchlist. | `--name` |
| `volumeleaders-agent watchlist delete` | Remove a saved watchlist configuration by its numeric key. Requires --key with the watchlist key (visible in configs output). The deletion is permanent and removes the watchlist and all its tickers. | `--key` |
| `volumeleaders-agent watchlist edit` | Modify an existing watchlist configuration identified by its numeric key. Requires --key with the watchlist key. Specify the fields you want to set; unspecified fields are replaced with their default values. | `--key` |
| `volumeleaders-agent watchlist tickers` | Query the ticker symbols belonging to a specific watchlist identified by --watchlist-key. Returns all tickers in the watchlist with their metadata. Outputs compact JSON or CSV/TSV with --format. |  |

## Configuration

### Flags

#### `volumeleaders-agent alert configs`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--fields` | string | - | Comma-separated fields to include (use 'all' for every field) |
| `--format` | string | json | Output format: json, csv, or tsv (csv, json, tsv) |

#### `volumeleaders-agent alert create`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ah-dollars-gte` | int | 0 | After-hours dollars >= |
| `--ah-rank-lte` | int | 0 | After-hours rank <= |
| `--ah-volume-gte` | int | 0 | After-hours volume >= |
| `--closing-trade-conditions` | string | 0 | Closing trade conditions |
| `--closing-trade-dollars-gte` | int | 0 | Closing trade dollars >= |
| `--closing-trade-mult-gte` | int | 0 | Closing trade multiplier >= |
| `--closing-trade-rank-lte` | int | 0 | Closing trade rank <= |
| `--closing-trade-vcd-gte` | int | 0 | Closing trade VCD >= (0/97/98/99/100) |
| `--closing-trade-volume-gte` | int | 0 | Closing trade volume >= |
| `--cluster-dollars-gte` | int | 0 | Trade cluster dollars >= |
| `--cluster-mult-gte` | int | 0 | Trade cluster multiplier >= |
| `--cluster-rank-lte` | int | 0 | Trade cluster rank <= |
| `--cluster-vcd-gte` | int | 0 | Trade cluster VCD >= (0/97/98/99/100) |
| `--cluster-volume-gte` | int | 0 | Trade cluster volume >= |
| `--dark-pool` | bool | false | Dark pool filter |
| `--name` | string | - | Alert name (max 50 chars) |
| `--offsetting-print` | bool | false | Offsetting print filter |
| `--phantom-print` | bool | false | Phantom print filter |
| `--sweep` | bool | false | Sweep filter |
| `--ticker-group` | string | AllTickers | Ticker group: AllTickers or SelectedTickers (AllTickers, SelectedTickers) |
| `--tickers` | string | - | Comma-separated ticker symbols (max 500, used with SelectedTickers) |
| `--total-dollars-gte` | int | 0 | Total dollars >= |
| `--total-rank-lte` | int | 0 | Total rank <= (0/1/3/10/25/50/100) |
| `--total-volume-gte` | int | 0 | Total volume >= |
| `--trade-conditions` | string | 0 | Trade conditions (0=N/A, OBH/OBD/OSH/OSD combos) |
| `--trade-dollars-gte` | int | 0 | Trade dollars >= (0=N/A, 1000000/10000000/...) |
| `--trade-mult-gte` | int | 0 | Trade multiplier >= (0=N/A, 5/10/25/50/100) |
| `--trade-rank-lte` | int | 0 | Trade rank <= (0=N/A, 1/3/5/10/25/50/100) |
| `--trade-vcd-gte` | int | 0 | Trade VCD >= (0=N/A, 99/100) |
| `--trade-volume-gte` | int | 0 | Trade volume >= (0=N/A, 1000000/2000000/5000000/10000000) |

#### `volumeleaders-agent alert delete`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--key` | int | 0 | Alert config key to delete |

#### `volumeleaders-agent alert edit`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ah-dollars-gte` | int | 0 | After-hours dollars >= |
| `--ah-rank-lte` | int | 0 | After-hours rank <= |
| `--ah-volume-gte` | int | 0 | After-hours volume >= |
| `--closing-trade-conditions` | string | 0 | Closing trade conditions |
| `--closing-trade-dollars-gte` | int | 0 | Closing trade dollars >= |
| `--closing-trade-mult-gte` | int | 0 | Closing trade multiplier >= |
| `--closing-trade-rank-lte` | int | 0 | Closing trade rank <= |
| `--closing-trade-vcd-gte` | int | 0 | Closing trade VCD >= (0/97/98/99/100) |
| `--closing-trade-volume-gte` | int | 0 | Closing trade volume >= |
| `--cluster-dollars-gte` | int | 0 | Trade cluster dollars >= |
| `--cluster-mult-gte` | int | 0 | Trade cluster multiplier >= |
| `--cluster-rank-lte` | int | 0 | Trade cluster rank <= |
| `--cluster-vcd-gte` | int | 0 | Trade cluster VCD >= (0/97/98/99/100) |
| `--cluster-volume-gte` | int | 0 | Trade cluster volume >= |
| `--dark-pool` | bool | false | Dark pool filter |
| `--key` | int | 0 | Alert config key to edit |
| `--name` | string | - | Alert name (max 50 chars) |
| `--offsetting-print` | bool | false | Offsetting print filter |
| `--phantom-print` | bool | false | Phantom print filter |
| `--sweep` | bool | false | Sweep filter |
| `--ticker-group` | string | AllTickers | Ticker group: AllTickers or SelectedTickers (AllTickers, SelectedTickers) |
| `--tickers` | string | - | Comma-separated ticker symbols (max 500, used with SelectedTickers) |
| `--total-dollars-gte` | int | 0 | Total dollars >= |
| `--total-rank-lte` | int | 0 | Total rank <= (0/1/3/10/25/50/100) |
| `--total-volume-gte` | int | 0 | Total volume >= |
| `--trade-conditions` | string | 0 | Trade conditions (0=N/A, OBH/OBD/OSH/OSD combos) |
| `--trade-dollars-gte` | int | 0 | Trade dollars >= (0=N/A, 1000000/10000000/...) |
| `--trade-mult-gte` | int | 0 | Trade multiplier >= (0=N/A, 5/10/25/50/100) |
| `--trade-rank-lte` | int | 0 | Trade rank <= (0=N/A, 1/3/5/10/25/50/100) |
| `--trade-vcd-gte` | int | 0 | Trade VCD >= (0=N/A, 99/100) |
| `--trade-volume-gte` | int | 0 | Trade volume >= (0=N/A, 1000000/2000000/5000000/10000000) |

#### `volumeleaders-agent market earnings`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--days` | int | 0 | Look back this many days from --end-date or today |
| `--end-date` | string | - | End date YYYY-MM-DD (required unless --days is set) |
| `--fields` | string | - | Comma-separated fields to include (use 'all' for every field) |
| `--format` | string | json | Output format: json, csv, or tsv (csv, json, tsv) |
| `--start-date` | string | - | Start date YYYY-MM-DD (required unless --days is set) |

#### `volumeleaders-agent market exhaustion`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--date` | string | - | Date YYYY-MM-DD (empty for current day) |

#### `volumeleaders-agent trade alerts`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--date` | string | - | Date YYYY-MM-DD |
| `--format` | string | json | Output format: json, csv, or tsv (csv, json, tsv) |
| `--length` | int | 100 | Number of results |
| `--order-col` | int | 1 | Order column index |
| `--order-dir` | string | desc | Order direction (asc, desc) |
| `--start` | int | 0 | DataTables start offset |

#### `volumeleaders-agent trade cluster-alerts`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--date` | string | - | Date YYYY-MM-DD |
| `--format` | string | json | Output format: json, csv, or tsv (csv, json, tsv) |
| `--length` | int | 100 | Number of results |
| `--order-col` | int | 1 | Order column index |
| `--order-dir` | string | desc | Order direction (asc, desc) |
| `--start` | int | 0 | DataTables start offset |

#### `volumeleaders-agent trade cluster-bombs`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--days` | int | 0 | Look back this many days from --end-date or today |
| `--end-date` | string | - | End date YYYY-MM-DD (required unless --days is set) |
| `--format` | string | json | Output format: json, csv, or tsv (csv, json, tsv) |
| `--max-dollars` | float64 | 30000000000 | Maximum dollar value |
| `--max-volume` | int | 2000000000 | Maximum volume |
| `--min-dollars` | float64 | 0 | Minimum dollar value |
| `--min-volume` | int | 0 | Minimum volume |
| `--order-col` | int | 1 | Order column index |
| `--order-dir` | string | desc | Order direction (asc, desc) |
| `--relative-size` | int | 0 | Relative size threshold |
| `--sector` | string | - | Sector/Industry filter |
| `--security-type` | int | 0 | Security type key |
| `--start` | int | 0 | DataTables start offset |
| `--start-date` | string | - | Start date YYYY-MM-DD (required unless --days is set) |
| `--tickers` | string | - | Comma-separated ticker symbols |
| `--trade-cluster-bomb-rank` | int | -1 | Trade cluster bomb rank filter |
| `--vcd` | int | 0 | VCD filter |

#### `volumeleaders-agent trade clusters`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--days` | int | 0 | Look back this many days from --end-date or today |
| `--end-date` | string | - | End date YYYY-MM-DD (required unless --days is set) |
| `--fields` | string | - | Comma-separated TradeCluster fields to include in output, or 'all' for every field |
| `--format` | string | json | Output format: json, csv, or tsv (csv, json, tsv) |
| `--max-dollars` | float64 | 30000000000 | Maximum dollar value |
| `--max-price` | float64 | 100000 | Maximum price |
| `--max-volume` | int | 2000000000 | Maximum volume |
| `--min-dollars` | float64 | 10000000 | Minimum dollar value |
| `--min-price` | float64 | 0 | Minimum price |
| `--min-volume` | int | 0 | Minimum volume |
| `--order-col` | int | 1 | Order column index |
| `--order-dir` | string | desc | Order direction (asc, desc) |
| `--relative-size` | int | 5 | Relative size threshold |
| `--sector` | string | - | Sector/Industry filter |
| `--security-type` | int | -1 | Security type key |
| `--start` | int | 0 | DataTables start offset |
| `--start-date` | string | - | Start date YYYY-MM-DD (required unless --days is set) |
| `--tickers` | string | - | Comma-separated ticker symbols |
| `--trade-cluster-rank` | int | -1 | Trade cluster rank filter |
| `--vcd` | int | 0 | VCD filter |

#### `volumeleaders-agent trade level-touches`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--days` | int | 0 | Look back this many days from --end-date or today |
| `--end-date` | string | - | End date YYYY-MM-DD (required unless --days is set) |
| `--format` | string | json | Output format: json, csv, or tsv (csv, json, tsv) |
| `--length` | int | 50 | Number of results |
| `--max-dollars` | float64 | 30000000000 | Maximum dollar value |
| `--max-price` | float64 | 100000 | Maximum price |
| `--max-volume` | int | 2000000000 | Maximum volume |
| `--min-dollars` | float64 | 500000 | Minimum dollar value |
| `--min-price` | float64 | 0 | Minimum price |
| `--min-volume` | int | 0 | Minimum volume |
| `--order-col` | int | 0 | Order column index |
| `--order-dir` | string | desc | Order direction (asc, desc) |
| `--relative-size` | int | 0 | Relative size threshold |
| `--start` | int | 0 | DataTables start offset |
| `--start-date` | string | - | Start date YYYY-MM-DD (required unless --days is set) |
| `--ticker` | string | - | Ticker symbol |
| `--trade-level-count` | int | 50 | Number of price levels to include (5, 10, 20, or 50) |
| `--trade-level-rank` | int | 5 | Trade level rank filter |
| `--vcd` | int | 0 | VCD filter |

#### `volumeleaders-agent trade levels`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--days` | int | 0 | Look back this many days from --end-date or today |
| `--end-date` | string | - | End date YYYY-MM-DD (default: today) |
| `--fields` | string | - | Comma-separated TradeLevel fields to include in output, or 'all' for every field |
| `--format` | string | json | Output format: json, csv, or tsv (csv, json, tsv) |
| `--max-dollars` | float64 | 30000000000 | Maximum dollar value |
| `--max-price` | float64 | 100000 | Maximum price |
| `--max-volume` | int | 2000000000 | Maximum volume |
| `--min-dollars` | float64 | 500000 | Minimum dollar value |
| `--min-price` | float64 | 0 | Minimum price |
| `--min-volume` | int | 0 | Minimum volume |
| `--relative-size` | int | 0 | Relative size threshold |
| `--start-date` | string | - | Start date YYYY-MM-DD (default: auto) |
| `--ticker` | string | - | Ticker symbol |
| `--trade-level-count` | int | 10 | Number of price levels to return (5, 10, 20, or 50) |
| `--trade-level-rank` | int | -1 | Trade level rank filter |
| `--vcd` | int | 0 | VCD filter |

#### `volumeleaders-agent trade list`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ah` | string | 1 | After-hours session filter (-1=all, 0=exclude, 1=include) (-1, 0, 1) |
| `--closing` | string | 1 | Closing trade filter (-1=all, 0=exclude, 1=include) (-1, 0, 1) |
| `--conditions` | int | -1 | Trade conditions filter |
| `--dark-pools` | string | -1 | Dark pool filter (-1=all, 0=exclude, 1=only) (-1, 0, 1) |
| `--days` | int | 0 | Look back this many days from --end-date or today |
| `--end-date` | string | - | End date YYYY-MM-DD (default: today) |
| `--even-shared` | string | -1 | Even shared filter (-1=all, 0=exclude, 1=only) (-1, 0, 1) |
| `--fields` | string | - | Comma-separated trade fields to include in output |
| `--format` | string | json | Output format: json, csv, or tsv (csv, json, tsv) |
| `--group-by` | string | ticker | Summary grouping (requires --summary): ticker, day, or ticker,day (day, ticker, ticker,day) |
| `--late-prints` | string | -1 | Late print filter (-1=all, 0=exclude, 1=only) (-1, 0, 1) |
| `--market-cap` | int | 0 | Market cap filter |
| `--max-dollars` | float64 | 30000000000 | Maximum dollar value |
| `--max-price` | float64 | 100000 | Maximum price |
| `--max-volume` | int | 2000000000 | Maximum volume |
| `--min-dollars` | float64 | 500000 | Minimum dollar value |
| `--min-price` | float64 | 0 | Minimum price |
| `--min-volume` | int | 0 | Minimum volume |
| `--offsetting` | string | 1 | Offsetting trade filter (-1=all, 0=exclude, 1=include) (-1, 0, 1) |
| `--opening` | string | 1 | Opening trade filter (-1=all, 0=exclude, 1=include) (-1, 0, 1) |
| `--order-col` | int | 1 | Order column index |
| `--order-dir` | string | desc | Order direction (asc, desc) |
| `--phantom` | string | 1 | Phantom print filter (-1=all, 0=exclude, 1=include) (-1, 0, 1) |
| `--premarket` | string | 1 | Premarket session filter (-1=all, 0=exclude, 1=include) (-1, 0, 1) |
| `--preset` | string | - | Apply a built-in filter preset (see: trade presets) |
| `--rank-snapshot` | int | -1 | Trade rank snapshot filter |
| `--relative-size` | int | 5 | Relative size threshold |
| `--rth` | string | 1 | Regular trading hours filter (-1=all, 0=exclude, 1=include) (-1, 0, 1) |
| `--sector` | string | - | Sector/Industry filter |
| `--security-type` | int | -1 | Security type key |
| `--sig-prints` | string | -1 | Signature print filter (-1=all, 0=exclude, 1=only) (-1, 0, 1) |
| `--start` | int | 0 | DataTables start offset |
| `--start-date` | string | - | Start date YYYY-MM-DD (default: auto) |
| `--summary` | bool | false | Return aggregate metrics instead of individual trades |
| `--sweeps` | string | -1 | Sweep filter (-1=all, 0=exclude, 1=only) (-1, 0, 1) |
| `--tickers` | string | - | Comma-separated ticker symbols |
| `--trade-rank` | int | -1 | Trade rank filter |
| `--vcd` | int | 97 | VCD filter |
| `--watchlist` | string | - | Apply filters from a saved watchlist by name |

#### `volumeleaders-agent trade preset-tickers`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--preset` | string | - | Preset name (case-insensitive) |

#### `volumeleaders-agent trade presets`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | json | Output format: json, csv, or tsv (csv, json, tsv) |

#### `volumeleaders-agent trade sentiment`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ah` | string | 1 | After-hours session filter (-1=all, 0=exclude, 1=include) (-1, 0, 1) |
| `--closing` | string | 1 | Closing trade filter (-1=all, 0=exclude, 1=include) (-1, 0, 1) |
| `--conditions` | int | -1 | Trade conditions filter |
| `--dark-pools` | string | -1 | Dark pool filter (-1=all, 0=exclude, 1=only) (-1, 0, 1) |
| `--days` | int | 0 | Look back this many days from --end-date or today |
| `--end-date` | string | - | End date YYYY-MM-DD (required unless --days is set) |
| `--even-shared` | string | -1 | Even shared filter (-1=all, 0=exclude, 1=only) (-1, 0, 1) |
| `--format` | string | json | Output format: json, csv, or tsv (csv, json, tsv) |
| `--late-prints` | string | -1 | Late print filter (-1=all, 0=exclude, 1=only) (-1, 0, 1) |
| `--market-cap` | int | 0 | Market cap filter |
| `--max-dollars` | float64 | 30000000000 | Maximum dollar value |
| `--max-price` | float64 | 100000 | Maximum price |
| `--max-volume` | int | 2000000000 | Maximum volume |
| `--min-dollars` | float64 | 5000000 | Minimum dollar value |
| `--min-price` | float64 | 0 | Minimum price |
| `--min-volume` | int | 0 | Minimum volume |
| `--offsetting` | string | 1 | Offsetting trade filter (-1=all, 0=exclude, 1=include) (-1, 0, 1) |
| `--opening` | string | 1 | Opening trade filter (-1=all, 0=exclude, 1=include) (-1, 0, 1) |
| `--phantom` | string | 1 | Phantom print filter (-1=all, 0=exclude, 1=include) (-1, 0, 1) |
| `--premarket` | string | 1 | Premarket session filter (-1=all, 0=exclude, 1=include) (-1, 0, 1) |
| `--rank-snapshot` | int | -1 | Trade rank snapshot filter |
| `--relative-size` | int | 5 | Relative size threshold |
| `--rth` | string | 1 | Regular trading hours filter (-1=all, 0=exclude, 1=include) (-1, 0, 1) |
| `--security-type` | int | -1 | Security type key |
| `--sig-prints` | string | -1 | Signature print filter (-1=all, 0=exclude, 1=only) (-1, 0, 1) |
| `--start-date` | string | - | Start date YYYY-MM-DD (required unless --days is set) |
| `--sweeps` | string | -1 | Sweep filter (-1=all, 0=exclude, 1=only) (-1, 0, 1) |
| `--trade-rank` | int | -1 | Trade rank filter |
| `--vcd` | int | 97 | VCD filter |

#### `volumeleaders-agent volume ah-institutional`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--date` | string | - | Date YYYY-MM-DD |
| `--format` | string | json | Output format: json, csv, or tsv (csv, json, tsv) |
| `--length` | int | 100 | Number of results |
| `--order-col` | int | 1 | Order column index |
| `--order-dir` | string | asc | Order direction (asc, desc) |
| `--start` | int | 0 | DataTables start offset |
| `--tickers` | string | - | Comma-separated ticker symbols |

#### `volumeleaders-agent volume institutional`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--date` | string | - | Date YYYY-MM-DD |
| `--format` | string | json | Output format: json, csv, or tsv (csv, json, tsv) |
| `--length` | int | 100 | Number of results |
| `--order-col` | int | 1 | Order column index |
| `--order-dir` | string | asc | Order direction (asc, desc) |
| `--start` | int | 0 | DataTables start offset |
| `--tickers` | string | - | Comma-separated ticker symbols |

#### `volumeleaders-agent volume total`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--date` | string | - | Date YYYY-MM-DD |
| `--format` | string | json | Output format: json, csv, or tsv (csv, json, tsv) |
| `--length` | int | 100 | Number of results |
| `--order-col` | int | 1 | Order column index |
| `--order-dir` | string | asc | Order direction (asc, desc) |
| `--start` | int | 0 | DataTables start offset |
| `--tickers` | string | - | Comma-separated ticker symbols |

#### `volumeleaders-agent watchlist add-ticker`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ticker` | string | - | Ticker symbol to add |
| `--watchlist-key` | int | 0 | Watch list key |

#### `volumeleaders-agent watchlist configs`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | json | Output format: json, csv, or tsv (csv, json, tsv) |

#### `volumeleaders-agent watchlist create`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ah-trades` | bool | true | Include after-hours trades |
| `--blocks` | bool | true | Include block trades |
| `--closing-trades` | bool | true | Include closing trades |
| `--dark-pools` | bool | true | Include dark pool trades |
| `--late-prints` | bool | true | Include late prints |
| `--lit-exchanges` | bool | true | Include lit exchange trades |
| `--max-dollars` | float64 | 3e+10 | Maximum dollars filter |
| `--max-price` | float64 | 100000 | Maximum price filter |
| `--max-trade-rank` | string | -1 | Maximum trade rank (-1=all, 1/3/5/10/25/50/100) (-1, 1, 10, 100, 25, 3, 5, 50) |
| `--max-volume` | int | 2000000000 | Maximum volume filter |
| `--min-dollars` | float64 | 0 | Minimum dollars filter |
| `--min-price` | float64 | 0 | Minimum price filter |
| `--min-relative-size` | string | 0 | Minimum relative size (0/5/10/25/50/100) (0, 10, 100, 25, 5, 50) |
| `--min-vcd` | float64 | 0 | Minimum VCD percentile (0-100) |
| `--min-volume` | int | 0 | Minimum volume filter |
| `--name` | string | - | Watch list name |
| `--normal-prints` | bool | true | Include normal prints |
| `--offsetting-trades` | bool | true | Include offsetting trades |
| `--opening-trades` | bool | true | Include opening trades |
| `--phantom-trades` | bool | true | Include phantom trades |
| `--premarket-trades` | bool | true | Include premarket trades |
| `--rsi-overbought-daily` | string | -1 | RSI overbought daily (-1=ignore, 0=no, 1=yes) (-1, 0, 1) |
| `--rsi-overbought-hourly` | string | -1 | RSI overbought hourly (-1=ignore, 0=no, 1=yes) (-1, 0, 1) |
| `--rsi-oversold-daily` | string | -1 | RSI oversold daily (-1=ignore, 0=no, 1=yes) (-1, 0, 1) |
| `--rsi-oversold-hourly` | string | -1 | RSI oversold hourly (-1=ignore, 0=no, 1=yes) (-1, 0, 1) |
| `--rth-trades` | bool | true | Include regular trading hours trades |
| `--sector-industry` | string | - | Sector/industry filter (max 100 chars) |
| `--security-type` | string | -1 | Security type (-1=all, 1=stocks, 26=ETFs, 4=REITs) (-1, 1, 26, 4) |
| `--signature-prints` | bool | true | Include signature prints |
| `--sweeps` | bool | true | Include sweep trades |
| `--tickers` | string | - | Comma-separated ticker symbols (max 500) |
| `--timely-prints` | bool | true | Include timely prints |

#### `volumeleaders-agent watchlist delete`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--key` | int | 0 | Watch list key to delete |

#### `volumeleaders-agent watchlist edit`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ah-trades` | bool | true | Include after-hours trades |
| `--blocks` | bool | true | Include block trades |
| `--closing-trades` | bool | true | Include closing trades |
| `--dark-pools` | bool | true | Include dark pool trades |
| `--key` | int | 0 | Watch list key to edit |
| `--late-prints` | bool | true | Include late prints |
| `--lit-exchanges` | bool | true | Include lit exchange trades |
| `--max-dollars` | float64 | 3e+10 | Maximum dollars filter |
| `--max-price` | float64 | 100000 | Maximum price filter |
| `--max-trade-rank` | string | -1 | Maximum trade rank (-1=all, 1/3/5/10/25/50/100) (-1, 1, 10, 100, 25, 3, 5, 50) |
| `--max-volume` | int | 2000000000 | Maximum volume filter |
| `--min-dollars` | float64 | 0 | Minimum dollars filter |
| `--min-price` | float64 | 0 | Minimum price filter |
| `--min-relative-size` | string | 0 | Minimum relative size (0/5/10/25/50/100) (0, 10, 100, 25, 5, 50) |
| `--min-vcd` | float64 | 0 | Minimum VCD percentile (0-100) |
| `--min-volume` | int | 0 | Minimum volume filter |
| `--name` | string | - | Watch list name |
| `--normal-prints` | bool | true | Include normal prints |
| `--offsetting-trades` | bool | true | Include offsetting trades |
| `--opening-trades` | bool | true | Include opening trades |
| `--phantom-trades` | bool | true | Include phantom trades |
| `--premarket-trades` | bool | true | Include premarket trades |
| `--rsi-overbought-daily` | string | -1 | RSI overbought daily (-1=ignore, 0=no, 1=yes) (-1, 0, 1) |
| `--rsi-overbought-hourly` | string | -1 | RSI overbought hourly (-1=ignore, 0=no, 1=yes) (-1, 0, 1) |
| `--rsi-oversold-daily` | string | -1 | RSI oversold daily (-1=ignore, 0=no, 1=yes) (-1, 0, 1) |
| `--rsi-oversold-hourly` | string | -1 | RSI oversold hourly (-1=ignore, 0=no, 1=yes) (-1, 0, 1) |
| `--rth-trades` | bool | true | Include regular trading hours trades |
| `--sector-industry` | string | - | Sector/industry filter (max 100 chars) |
| `--security-type` | string | -1 | Security type (-1=all, 1=stocks, 26=ETFs, 4=REITs) (-1, 1, 26, 4) |
| `--signature-prints` | bool | true | Include signature prints |
| `--sweeps` | bool | true | Include sweep trades |
| `--tickers` | string | - | Comma-separated ticker symbols (max 500) |
| `--timely-prints` | bool | true | Include timely prints |

#### `volumeleaders-agent watchlist tickers`

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format` | string | json | Output format: json, csv, or tsv (csv, json, tsv) |
| `--watchlist-key` | int | -1 | Watch list key (-1 for all) |

## Machine Interface

- JSON Schema: `volumeleaders-agent --jsonschema`
- Structured errors: JSON on stderr with semantic exit codes

