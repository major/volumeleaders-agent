# volumeleaders-agent

CLI for VolumeLeaders institutional trade data. Use it for trades, daily summaries, volume leaderboards, charts, market data, alerts, and watchlists. Binary: `volumeleaders-agent`.

Auth: reads browser cookies. If auth fails, the user must log in at volumeleaders.com in their browser.

Output: compact JSON to stdout by default. Put `--pretty` before the command group for indented JSON. Errors/logs go to stderr.

## Command Chooser

| Goal | Command | Details |
|---|---|---|
| Find individual institutional trades | `trade list --tickers X --start-date D --end-date D` | `trade.md` |
| Start with a market-wide daily snapshot | `daily summary --date D` | `daily.md` |
| Compare leveraged ETF bull/bear flow | `trade sentiment --start-date D --end-date D` | Fixed leveraged ETF universe |
| Find price-level trade clusters | `trade clusters --start-date D --end-date D` | Cluster conviction |
| Find sudden aggressive bursts | `trade cluster-bombs --start-date D --end-date D` | Burst detection |
| Check individual trade alerts | `trade alerts --date D` | System alerts |
| Check cluster alerts | `trade cluster-alerts --date D` | System alerts |
| Find support/resistance levels | `trade levels --ticker X --start-date D --end-date D` | One ticker |
| Find price revisits to levels | `trade level-touches --start-date D --end-date D` | Level tests |
| See institutional volume leaders | `volume institutional --date D` | `volume.md` |
| See after-hours institutional leaders | `volume ah-institutional --date D` | `volume.md` |
| See total volume leaders | `volume total --date D` | `volume.md` |
| Get 1-min bars with trade overlays | `chart price-data --ticker X --start-date D --end-date D` | `chart.md` |
| Get bid/ask/last quote | `chart snapshot --ticker X --date-key D` | JSON only |
| Get chart-ready levels | `chart levels --ticker X --start-date D --end-date D` | Fewer filters than `trade levels` |
| Get company metadata | `chart company --ticker X` | JSON only |
| Get current prices | `market snapshots` | JSON object |
| Find earnings with prior institutional activity | `market earnings --start-date D --end-date D` | CSV/TSV supported |
| Check exhaustion/reversal signals | `market exhaustion [--date D]` | Lower rank = stronger signal |
| Manage alert configs | `alert configs/create/edit/delete` | `alert.md` |
| Manage watchlists | `watchlist configs/create/edit/delete` | `watchlist.md` |
| Get watchlist tickers | `watchlist tickers --watchlist-key K` | Key from `watchlist configs` |

## Analysis Workflow

1. `daily summary --date D` for the broad read.
2. `volume institutional --date D` for top dollar movers.
3. `trade list --tickers X --start-date D --end-date D` for individual prints.
4. `trade levels --ticker X --start-date D --end-date D` for support/resistance.
5. `chart company --ticker X` for company context.
6. `chart price-data --ticker X --start-date D --end-date D` for bar-level overlays.

## Global Conventions

- Dates: `YYYY-MM-DD`.
- Pagination: `--start` offset, `--length` count, `--length -1` means all rows, `--order-col` sort column index, `--order-dir asc|desc`.
- Toggle filters: `-1` all/unfiltered, `0` exclude, `1` include/only.
- Tickers: `--tickers` is comma-separated, `--ticker` is single-symbol. Trade ticker filters also accept `--symbol` and `--symbols` aliases.
- Output formats: list-style commands may support `--format json|csv|tsv`. CSV/TSV include headers, booleans render as `true`/`false`, null/missing values render as empty cells. Nested summaries and single-object commands are JSON-only unless their command file says otherwise.
- Performance: use explicit dates and tickers when possible. `--length -1` can return large datasets and slow summary aggregation.

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
