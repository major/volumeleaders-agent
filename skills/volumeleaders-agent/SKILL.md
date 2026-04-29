---
name: volumeleaders-agent
description: |
  CLI for querying institutional trade data from VolumeLeaders. Use when:
  - Looking up institutional block trades, clusters, or price levels for a ticker
  - Checking daily institutional activity summaries
  - Finding volume leaders (institutional, after-hours, total)
  - Getting price bars with trade overlays or company metadata
  - Checking market-wide snapshots, earnings calendar, or exhaustion signals
  - Managing trade alerts or watchlists
  Triggers: "volumeleaders", "institutional trades", "block trades", "volume leaders", "trade clusters", "market exhaustion", "dark pool", "sweep"
---

# volumeleaders-agent

CLI for VolumeLeaders institutional trade data. Use it for trades, daily summaries, volume leaderboards, charts, market data, alerts, and watchlists. Binary: `volumeleaders-agent`.

Auth: reads browser cookies. If auth fails with exit code 2 and `Authentication required: VolumeLeaders session has expired.`, the user must log in at https://www.volumeleaders.com in their browser, then retry.

Output: compact JSON to stdout by default. Put `--pretty` before the command group for indented JSON. Errors/logs go to stderr.

## Companion Files

This is the entry point. Command details are split by command group:

| File | Scope |
|------|-------|
| [trade.md](trade.md) | Trade list, sentiment, clusters, cluster bombs, alerts, levels, level touches |
| [daily.md](daily.md) | Daily institutional activity summaries |
| [volume.md](volume.md) | Volume leaderboards (institutional, after-hours, total) |
| [chart.md](chart.md) | Price bars with trade overlays, snapshots, levels, company metadata |
| [market.md](market.md) | Market snapshots, earnings calendar, exhaustion signals |
| [alert.md](alert.md) | Alert configuration management |
| [watchlist.md](watchlist.md) | Watchlist management and ticker retrieval |

## Command Chooser

| Goal | Command | Details |
|---|---|---|
| Find individual institutional trades | `trade list X --days N` | [trade.md](trade.md) |
| Start with a market-wide daily snapshot | `daily summary --date D` | [daily.md](daily.md) |
| Compare leveraged ETF bull/bear flow | `trade sentiment --days N` | Fixed leveraged ETF universe |
| Find price-level trade clusters | `trade clusters --days N` | Cluster conviction |
| Find sudden aggressive bursts | `trade cluster-bombs --days N` | Burst detection |
| Check individual trade alerts | `trade alerts --date D` | System alerts |
| Check cluster alerts | `trade cluster-alerts --date D` | System alerts |
| Find support/resistance levels | `trade levels X --days N` | One ticker |
| Find price revisits to levels | `trade level-touches X --days N` | Level tests |
| See institutional volume leaders | `volume institutional --date D` | [volume.md](volume.md) |
| See after-hours institutional leaders | `volume ah-institutional --date D` | [volume.md](volume.md) |
| See total volume leaders | `volume total --date D` | [volume.md](volume.md) |
| Get 1-min bars with trade overlays | `chart price-data X --days N` | [chart.md](chart.md) |
| Get bid/ask/last quote | `chart snapshot X --date-key D` | JSON only |
| Get chart-ready levels | `chart levels X --days N` | Fewer filters than `trade levels` |
| Get company metadata | `chart company X` | JSON only |
| Get current prices | `market snapshots` | JSON object |
| Find earnings with prior institutional activity | `market earnings --days N` | CSV/TSV supported |
| Check exhaustion/reversal signals | `market exhaustion [--date D]` | Lower rank = stronger signal |
| Manage alert configs | `alert configs/create/edit/delete` | [alert.md](alert.md) |
| Manage watchlists | `watchlist configs/create/edit/delete` | [watchlist.md](watchlist.md) |
| Get watchlist tickers | `watchlist tickers --watchlist-key K` | Key from `watchlist configs` |

## Analysis Workflow

1. `daily summary --date D` for the broad read.
2. `volume institutional --date D` for top dollar movers.
3. `trade list X --days N` for individual prints.
4. `trade levels X --days N` for support/resistance.
5. `chart company X` for company context.
6. `chart price-data X --days N` for bar-level overlays.

## Global Conventions

- Dates: `YYYY-MM-DD`. Commands with date ranges accept either `--start-date D --end-date D` or `--days N`; `--days` counts backward from today unless `--end-date` is also set, and cannot be combined with `--start-date`.
- Pagination: `--start` offset, `--length` count, `--length -1` means all rows unless the command file documents a stricter endpoint limit, `--order-col` sort column index, `--order-dir asc|desc`.
- Toggle filters: `-1` all/unfiltered, `0` exclude, `1` include/only.
- Tickers: `--tickers` is comma-separated, `--ticker` is single-symbol. Commands that take tickers generally accept positional tickers too, for example `trade list XLE XLK` or `chart company AAPL`. Trade and volume ticker filters also accept `--symbol` and `--symbols` aliases.
- Output formats: list-style commands may support `--format json|csv|tsv`. CSV/TSV include headers, booleans render as `true`/`false`, null/missing values render as empty cells. Nested summaries and single-object commands are JSON-only unless their command file says otherwise.
- Performance: use explicit dates and tickers when possible. Individual trade and trade-level retrieval commands are capped at 50 rows per request to protect the VolumeLeaders backend.

## Trade Shared Flags

These defaults apply to trade commands only unless a subcommand overrides them. Chart commands use different names for some filters, see `chart.md`.

| Flag | Default | Notes |
|---|---|---|
| `--min-volume` / `--max-volume` | 0 / 2000000000 | Share range |
| `--min-price` / `--max-price` | 0 / 100000 | Price range |
| `--min-dollars` / `--max-dollars` | 500000 / 30000000000 | Dollar range |
| `--conditions` | -1 | Trade condition filter |
| `--vcd` | 0 | Volume Confirmation Distribution minimum |
| `--relative-size` | 5 | Trade size vs normal activity |
| `--security-type` | -1 | Security type filter |
| `--market-cap` | 0 | Market cap filter |
| `--trade-rank` / `--rank-snapshot` | -1 / -1 | Lower rank = more significant |
| `--dark-pools`, `--sweeps`, `--late-prints`, `--sig-prints` | -1 | Toggle filters |
| `--even-shared` | -1 | Even-share filter |
| `--premarket`, `--rth`, `--ah`, `--opening`, `--closing`, `--phantom`, `--offsetting` | 1 | Session/event toggles |

Trade sentiment overrides: `--min-dollars 5000000`, `--vcd 97`, and fixed `SectorIndustry=X B`; it keeps the shared `--relative-size 5` default. Users cannot override the sentiment universe with tickers or sector flags.

## Key Metrics

| Field | Meaning |
|---|---|
| `CumulativeDistribution` | Volume percentile, 0 to 1, higher = more accumulation |
| `DollarsMultiplier` | Trade dollars relative to average block size |
| `TradeRank`, `TradeRankSnapshot` | VL significance rank now vs at print time, lower = stronger |
| `RelativeSize`, `PercentDailyVolume` | Trade size vs normal volume |
| `VCD` | Volume Confirmation Distribution score |
| `FrequencyLast30TD`, `FrequencyLast90TD`, `FrequencyLast1CY` | Similar-magnitude trade frequency windows |
| `RSIHour`, `RSIDay` | Hourly and daily RSI |
| `DarkPool`, `Sweep`, `LatePrint`, `SignaturePrint`, `PhantomPrint`, `InsideBar` | Boolean trade/bar traits |
