---
name: volumeleaders-agent
description: |
  CLI for querying institutional trade data from VolumeLeaders. Use when:
  - Looking up institutional block trades, clusters, or price levels for a ticker
  - Finding volume leaders (institutional, after-hours, total)
  - Checking market-wide snapshots, earnings calendar, or exhaustion signals
  - Managing trade alerts or watchlists
  Triggers: "volumeleaders", "institutional trades", "block trades", "volume leaders", "trade clusters", "market exhaustion", "dark pool", "sweep"
---

# volumeleaders-agent

CLI for VolumeLeaders institutional trade data. Use it for trades, volume leaderboards, market data, alerts, and watchlists. Binary: `volumeleaders-agent`.

Auth: reads browser cookies. If auth fails with exit code 2 and `Authentication required: VolumeLeaders session has expired.`, the user must log in at https://www.volumeleaders.com in their browser, then retry.

Output: compact JSON to stdout by default. Put `--pretty` before the command group for indented JSON. The `schema` command always emits raw indented JSON for machine-readable CLI introspection. Errors/logs go to stderr.

## Schema-First Usage

Use the schema command as the source of truth for command names, flags, defaults, aliases, examples, and global flags:

```bash
volumeleaders-agent schema
volumeleaders-agent schema --command "trade list"
volumeleaders-agent schema --command "market earnings"
```

The schema output has `commands` and `global_flags` keys. Each command entry includes `description`, `flags`, `args`, and `examples`. Filter with `--command "group subcommand"` when you only need one command. The schema command does not call VolumeLeaders and does not require auth.

Use this skill for domain meaning, workflows, and gotchas that the schema cannot express.

## Command Chooser

| Goal | Start with | Notes |
|---|---|---|
| Find individual institutional prints | `trade list X --days N` | Use ticker filters, presets, or watchlists to stay inside caps |
| Compare leveraged ETF bull/bear flow | `trade sentiment --days N` | Fixed leveraged ETF universe, not signed buy/sell flow |
| Find converging price-level activity | `trade clusters --days N` | Cluster conviction around similar prices |
| Find sudden aggressive bursts | `trade cluster-bombs --days N` | Burst detection, different defaults than clusters |
| Inspect trade or cluster alerts | `trade alerts --date D`, `trade cluster-alerts --date D` | System-generated alerts |
| Find support/resistance levels | `trade levels X --days N` | One ticker, capped level count |
| Find revisits to institutional levels | `trade level-touches X --days N` | Level retests, capped pagination |
| See institutional volume leaders | `volume institutional --date D` | Same trade model, volume-ranked |
| See after-hours institutional leaders | `volume ah-institutional --date D` | After-hours institutional flow |
| See total volume leaders | `volume total --date D` | Total market volume across trade types |
| Get current prices | `market snapshots` | JSON object |
| Find earnings with prior institutional activity | `market earnings --days N` | CSV/TSV supported |
| Check exhaustion/reversal signals | `market exhaustion [--date D]` | Lower rank is stronger |
| Manage alert configs | `alert configs/create/edit/delete` | Edit replaces unspecified values with defaults |
| Manage watchlists | `watchlist configs/create/edit/delete` | Edit replaces unspecified values with defaults |
| Get watchlist tickers | `watchlist tickers --watchlist-key K` | Key comes from `watchlist configs` |

## Analysis Workflow

1. `volume institutional --date D` for top dollar movers.
2. `trade list X --days N` for individual prints.
3. `trade levels X --days N` for support/resistance.
4. `trade clusters X --days N` when prints appear concentrated around a price.
5. `market earnings --days N` and `market exhaustion [--date D]` for event and reversal context.

## Global Conventions

- Dates: `YYYY-MM-DD`. Commands with date ranges accept either `--start-date D --end-date D` or `--days N`; `--days` counts backward from today unless `--end-date` is also set, and cannot be combined with `--start-date`.
- Pagination: `--start` offset, `--length` count, `--length -1` means all rows unless a capped endpoint rejects it. `trade list`, `trade list --summary`, and `trade level-touches` only allow 1 to 50 rows. `trade levels` caps `--trade-level-count` at 50.
- Toggle filters: `-1` means all/unfiltered, `0` means exclude, `1` means include/only.
- Tickers: `--tickers` is comma-separated, `--ticker` is single-symbol. Commands that take tickers generally accept positional tickers too, for example `trade list XLE XLK`. Trade and volume ticker filters also accept `--symbol` and `--symbols` aliases.
- Output formats: list-style commands may support `--format json|csv|tsv`. CSV/TSV include headers, booleans render as `true`/`false`, null or missing values render as empty cells. Nested summaries and single-object commands are JSON-only unless the schema shows a format flag.
- Performance: use explicit dates and tickers when possible. Start narrow, then expand. VolumeLeaders endpoints can be expensive and some trade retrieval endpoints are intentionally capped.

## Trade Guidance

Shared trade filters include volume, price, dollars, conditions, VCD, relative size, security type, market cap, trade rank, dark pools, sweeps, late prints, signature prints, even-share prints, and session/event toggles. Check `volumeleaders-agent schema --command "trade list"` for exact flag names, aliases, defaults, and examples.

`trade list` is the primary individual-print query. Date defaults are 365-day lookback when tickers are provided and today-only without tickers. Preset and watchlist filters do not supply dates. Filter precedence is preset baseline, then watchlist merge, then explicit CLI flags override both. Default JSON is compact and omits repetitive/internal fields; use `--fields FIELD1,FIELD2`, CSV/TSV, or `--fields all` where supported when raw API fields are needed. `--summary` returns aggregate JSON with valid `--group-by` values of `ticker`, `day`, or `ticker,day`; do not combine summary mode with `--fields` or non-JSON formats.

`trade sentiment` always queries the combined leveraged ETF sector filter `SectorIndustry=X B`, classifies bull and bear ETFs locally, and cannot be constrained by ticker or sector flags. Non-standard defaults include `--min-dollars 5000000` and `--vcd 97`; shared `--relative-size 5` still applies. Ratio is bull dollars divided by bear dollars and is `null` when bear flow is zero. Treat it as leveraged ETF proxy flow, not signed buy/sell flow.

`trade clusters` finds multiple institutional trades converging at similar prices. It uses larger default retrieval and dollar thresholds than ordinary trade list. `trade cluster-bombs` finds sudden aggressive bursts tightly grouped in time and price, with different defaults and rank fields. Use cluster commands when the question is about price-level concentration, not single prints.

`trade alerts` and `trade cluster-alerts` return system-generated alert rows for one date. Cluster alert rows use the full cluster-shaped model rather than the compact default from `trade clusters`.

`trade levels` finds significant institutional prices for one ticker. It defaults to a 1-year lookback when dates are omitted, uses non-standard `--relative-size 0`, and caps level count from 1 to 50. Default JSON is compact and omits repetitive ticker metadata and the verbose `Dates` list. `trade level-touches` finds events where price revisits institutional levels, defaults to `--length 50`, and rejects `--length -1`, `--length 0`, and values above 50.

## Volume and Market Guidance

Volume commands rank stocks by trading activity and require `--date`. The default sort is `--order-dir asc`, unlike most trade commands. Output uses the same Trade model as `trade list`; key fields include institutional dollars/volume, after-hours institutional dollars/volume, closing trade dollars/volume, total dollars/volume, average daily volume, percent daily volume, close price, and RSI.

`market earnings` defaults to compact rows with ticker, earnings date, after-market-close flag, and counts for trades, clusters, and cluster bombs. Repetitive company descriptors are omitted by default to reduce tokens; use fields options from the schema when needed. `market exhaustion` returns compact exhaustion ranks as `date_key`, `rank`, `rank_30d`, `rank_90d`, and `rank_365d`. Lower rank means stronger exhaustion or reversal signal, and multiple low ranks across timeframes reinforce reversal risk.

## Alert and Watchlist Guidance

Alert configs trigger when trades or clusters match thresholds. `alert configs` default output is compact; use field selection from the schema for full threshold details. Threshold names follow `{Category}{Metric}{LTE|GTE}` where LTE is maximum rank and GTE is minimum value. Important edit gotcha: `alert edit` replaces unspecified flags with defaults, so include every value that must remain set.

Watchlist configs define reusable filters and ticker groups. `watchlist tickers` can return all watchlists with `--watchlist-key -1` or one watchlist by key. Create/edit options cover tickers, volume, dollars, price, VCD, relative size, trade rank, security/sector, trade types, sessions, and RSI toggles. Important edit gotcha: `watchlist edit` replaces unspecified flags with defaults, so include every value that must remain set.

## Key Metrics

| Field | Meaning |
|---|---|
| `CumulativeDistribution` | Volume percentile, 0 to 1, higher means more accumulation |
| `DollarsMultiplier` | Trade dollars relative to average block size |
| `TradeRank`, `TradeRankSnapshot` | VL significance rank now vs at print time, lower is stronger |
| `TradeClusterRank`, `TradeClusterBombRank`, `TradeLevelRank` | Rank for cluster, burst, or level significance, lower is stronger |
| `RelativeSize`, `PercentDailyVolume` | Trade size vs normal activity |
| `VCD` | Volume Confirmation Distribution score |
| `FrequencyLast30TD`, `FrequencyLast90TD`, `FrequencyLast1CY` | Similar-magnitude trade frequency windows |
| `RSIHour`, `RSIDay` | Hourly and daily RSI |
| `DarkPool`, `Sweep`, `LatePrint`, `SignaturePrint`, `PhantomPrint`, `InsideBar` | Boolean trade/bar traits |

## Examples

```bash
volumeleaders-agent schema --command "trade list"
volumeleaders-agent volume institutional --date 2026-04-28
volumeleaders-agent trade list XLE --days 5
volumeleaders-agent trade list --tickers SPY,QQQ --start-date 2026-04-21 --end-date 2026-04-28 --summary --group-by ticker,day --length 50
volumeleaders-agent trade sentiment --days 7
volumeleaders-agent trade clusters AAPL MSFT --days 5
volumeleaders-agent trade levels AAPL --days 30
volumeleaders-agent market earnings --days 7 --format csv
volumeleaders-agent market exhaustion --date 2026-04-28
volumeleaders-agent alert configs --format csv
volumeleaders-agent watchlist tickers --watchlist-key 12345 --format tsv
```
