# volumeleaders-agent

CLI for querying institutional trade data from VolumeLeaders. Binary: `volumeleaders-agent`.

Auth: extracts browser cookies automatically. User must be logged into volumeleaders.com. Auth errors mean the user needs to log in via their browser.

Output: compact JSON to stdout by default. List-style commands support `--format json|csv|tsv`; CSV/TSV include a header row, render booleans as `true`/`false`, and render null/missing values as empty cells. Add `--pretty` before the command group for indented JSON output. Errors go to stderr.

## Command Chooser

| I want to... | Command |
|---|---|
| Find large institutional trades in specific stocks | `trade list --tickers X --start-date D --end-date D` |
| Compare leveraged ETF bull/bear flow | `trade sentiment --start-date D --end-date D` |
| See where institutions are clustering trades | `trade clusters --start-date D --end-date D` |
| Detect sudden aggressive institutional bursts | `trade cluster-bombs --start-date D --end-date D` |
| Check system-flagged individual trade alerts | `trade alerts --date D` |
| Check system-flagged cluster alerts | `trade cluster-alerts --date D` |
| Find institutional support/resistance levels | `trade levels --ticker X --start-date D --end-date D` |
| Track price revisiting institutional levels | `trade level-touches --start-date D --end-date D` |
| See top institutional volume movers for a day | `volume institutional --date D` |
| See after-hours institutional activity | `volume ah-institutional --date D` |
| See overall volume leaders | `volume total --date D` |
| Get 1-min price bars with trade overlays | `chart price-data --ticker X --start-date D --end-date D` |
| Get a quick quote (bid/ask/last) | `chart snapshot --ticker X --date-key D` |
| Get chart-ready institutional levels | `chart levels --ticker X --start-date D --end-date D` |
| Get company metadata and averages | `chart company --ticker X` |
| Get all current market prices | `market snapshots` |
| Find upcoming earnings with institutional positioning | `market earnings --start-date D --end-date D` |
| Check market exhaustion/reversal signals | `market exhaustion` |
| View/create/edit/delete alert configs | `alert configs/create/edit/delete` |
| View/create/edit/delete watchlists | `watchlist configs/create/edit/delete` |
| Get tickers from a watchlist | `watchlist tickers` |

## Analysis Workflow

Chain commands for deeper analysis:

1. `volume institutional --date D` - find top institutional movers
2. `trade list --tickers X --start-date D --end-date D` - drill into individual trades
3. `trade levels --ticker X --start-date D --end-date D` - find key support/resistance levels
4. `chart company --ticker X` - get company context
5. `chart price-data --ticker X --start-date D --end-date D` - detailed price action with overlays

## Conventions

**Dates**: `YYYY-MM-DD` on all flags.

**Toggle flags** (-1/0/1): `-1` = all/unfiltered, `0` = exclude, `1` = only. Applies to trade type and session flags.

**Pagination**: `--start` (offset, default 0), `--length` (count, default varies, `-1` = all), `--order-col` (sort column index), `--order-dir` (`asc` or `desc`).

**Output formats**: list-style object outputs accept `--format json|csv|tsv` (default `json`). CSV/TSV use output field names as headers. `--pretty` only affects JSON.

**Ticker flags**: `--tickers` takes comma-separated list (multi-ticker commands). `--ticker` takes a single symbol (single-ticker commands). Ticker-based trade subcommands accept `--ticker`, `--tickers`, `--symbol`, and `--symbols` aliases, so any form works there. `trade sentiment` intentionally does not accept ticker flags because it always analyzes the leveraged ETF universe.

## Shared Flag Defaults

These defaults apply across trade commands unless overridden (overrides noted per-subcommand).

| Flag | Default | Notes |
|---|---|---|
| `--min-volume` / `--max-volume` | 0 / 2000000000 | |
| `--min-price` / `--max-price` | 0 / 100000 | |
| `--min-dollars` / `--max-dollars` | 500000 / 30000000000 | |
| `--conditions` | -1 | |
| `--vcd` | 0 | |
| `--relative-size` | 5 | |
| `--security-type` | -1 | |
| `--market-cap` | 0 | |
| `--trade-rank` | -1 | |
| `--rank-snapshot` | -1 | |
| `--dark-pools` | -1 | Toggle |
| `--sweeps` | -1 | Toggle |
| `--late-prints` | -1 | Toggle |
| `--sig-prints` | -1 | Toggle |
| `--even-shared` | -1 | |
| `--premarket` | 1 | Session toggle |
| `--rth` | 1 | Session toggle |
| `--ah` | 1 | Session toggle |
| `--opening` | 1 | Session toggle |
| `--closing` | 1 | Session toggle |
| `--phantom` | 1 | Session toggle |
| `--offsetting` | 1 | Session toggle |

**Chart commands use different flag names** for some of these. See chart.md for the exact mapping.

**Trade sentiment overrides**: `trade sentiment` defaults to `--min-dollars 5000000` and `--vcd 97` because leveraged ETF sentiment is meant to highlight unusually large, high-confirmation flow.

## Key Metrics Glossary

| Field | Meaning |
|---|---|
| `CumulativeDistribution` | Percentile position in volume distribution (0-1 scale, higher = more accumulation) |
| `DollarsMultiplier` | Trade dollar value relative to average block size (higher = bigger than usual) |
| `TradeRank` | VL proprietary significance ranking (lower = more significant) |
| `TradeRankSnapshot` | TradeRank at time of trade (vs current recalculated rank) |
| `RelativeSize` | Trade size vs average daily volume (higher = more unusual). Present on trade list, levels, and price-data outputs |
| `PercentDailyVolume` | Trade volume as % of average daily volume. On trade list output |
| `VCD` | Volume Confirmation Distribution score |
| `FrequencyLast30TD` | Count of similar-magnitude trades in last 30 trading days |
| `FrequencyLast90TD` | Count of similar-magnitude trades in last 90 trading days |
| `FrequencyLast1CY` | Count of similar-magnitude trades in last calendar year (~252 trading days) |
| `RSIHour` / `RSIDay` | Relative Strength Index on hourly / daily timeframe |
| `DarkPool` | Trade executed on a dark pool (boolean) |
| `Sweep` | Aggressive order sweeping multiple price levels (boolean) |
| `LatePrint` | Trade reported late to the tape (boolean) |
| `SignaturePrint` | VL proprietary notable trade pattern (boolean) |
| `PhantomPrint` | Trade that may not settle / gets cancelled (boolean) |
| `InsideBar` | Price bar contained within prior bar's range (boolean) |
