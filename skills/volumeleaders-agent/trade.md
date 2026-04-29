# trade

Institutional trade discovery. See `SKILL.md` for global conventions, trade shared flags, and metric meanings.

## Quick Reference

| Command | Use when | Required | Output |
|---|---|---|---|
| `trade presets` | List built-in filters | none | JSON, CSV, TSV |
| `trade preset-tickers` | Inspect a preset universe | `--preset` | JSON |
| `trade list` | Query individual trades | none | JSON, CSV, TSV, or summary JSON |
| `trade sentiment` | Bull/bear leveraged ETF flow | date range or `--days` | JSON, CSV, TSV |
| `trade clusters` | Converging trades at price levels | date range or `--days` | JSON, CSV, TSV |
| `trade cluster-bombs` | Sudden aggressive bursts | date range or `--days` | JSON, CSV, TSV |
| `trade alerts` | System trade alerts | `--date` | JSON, CSV, TSV |
| `trade cluster-alerts` | System cluster alerts | `--date` | JSON, CSV, TSV |
| `trade levels` | Support/resistance by ticker | ticker flag or positional ticker | JSON, CSV, TSV |
| `trade level-touches` | Price revisits institutional levels | date range or `--days` | JSON, CSV, TSV |

## Presets

`trade presets` lists built-in filters. Groups: `Common`, `Disproportionately Large`.

`trade preset-tickers --preset NAME` returns `Preset`, `Group`, `Type`, and either `Tickers` or `SectorIndustry`. Types: `tickers`, `sector-filter`, `unfiltered`. Ticker output takes precedence if both ticker and sector filters exist.

```bash
volumeleaders-agent trade presets --format csv
volumeleaders-agent trade preset-tickers --preset "Megacaps"
```

## trade list

Primary individual-print query. Optional flags: `--start-date`, `--end-date`, `--days`, ticker aliases, positional tickers, `--sector`, `--preset`, `--watchlist`, `--fields`, `--format`, `--summary`, `--group-by`, trade shared flags, pagination.

Default JSON returns compact analysis rows, not the full raw API row. It keeps trade timing, ticker/company context, price/volume/dollar size, rank, relative-size metrics, key boolean traits, frequency windows, trade conditions, and RSI. Repetitive query dates, internal numeric keys, raw bid/ask, redundant volume leaderboard totals, `TotalRows`, and feed metadata are omitted to reduce tokens.

Date defaults: with tickers, 365-day lookback from today. Without tickers, today only. Explicit dates override defaults. `--days N` uses today or explicit `--end-date` as the range end and computes the start date; do not combine `--days` with `--start-date`. Presets and watchlists never supply dates.

Filter precedence: preset baseline, then watchlist merge, then explicit CLI flags override both.

Summary mode: `--summary` returns aggregate JSON instead of rows. Valid `--group-by`: `ticker`, `day`, `ticker,day`. Summary respects pagination, so use `--length -1` for all matching rows. `--fields` and non-JSON `--format` are invalid with `--summary`.

Fields mode: `--fields FIELD1,FIELD2` limits JSON keys or CSV/TSV columns using raw API field names. Use it to request fields omitted from default compact JSON, such as `TradeID`, `Bid`, `Ask`, `AverageDailyVolume`, or `TotalRows`. Names are case-sensitive and validated before the API query. CSV/TSV without `--fields` still use the full raw trade row columns.

```bash
volumeleaders-agent trade list --tickers AAPL --dark-pools 1 --min-dollars 1000000
volumeleaders-agent trade list XLE --days 5
volumeleaders-agent trade list --preset "Top-100 Rank" --start-date 2026-04-28 --end-date 2026-04-28
volumeleaders-agent trade list --watchlist "Magnificent 7" --start-date 2026-04-01 --end-date 2026-04-28
volumeleaders-agent trade list --tickers SPY,QQQ --start-date 2026-04-21 --end-date 2026-04-28 --summary --group-by ticker,day --length -1
volumeleaders-agent trade list --tickers AAPL,MSFT --start-date 2026-04-21 --end-date 2026-04-28 --fields Date,Ticker,Dollars --format csv
```

Default JSON row fields: `Date`, `FullDateTime`, `FullTimeString24`, `Ticker`, `Name`, `Sector`, `Industry`, `Price`, `Volume`, `Dollars`, `DollarsMultiplier`, `PercentDailyVolume`, `RelativeSize`, `CumulativeDistribution`, `TradeRank`, `TradeRankSnapshot`, `DarkPool`, `Sweep`, `LatePrint`, `SignaturePrint`, `OpeningTrade`, `ClosingTrade`, `PhantomPrint`, `TradeConditions`, `FrequencyLast30TD`, `FrequencyLast90TD`, `FrequencyLast1CY`, `RSIHour`, `RSIDay`.

## trade sentiment

Daily bull/bear leveraged ETF flow. The command always queries the combined leveraged ETF sector filter `SectorIndustry=X B`, classifies bull and bear ETFs locally, and cannot be constrained by ticker or sector flags.

Required: complete `--start-date`/`--end-date` range or `--days`. Optional: `--format json|csv|tsv`, trade shared flags except ticker/sector filters. Non-standard defaults: `--min-dollars 5000000`, `--vcd 97`; shared `--relative-size 5` still applies.

```bash
volumeleaders-agent trade sentiment --start-date 2026-04-21 --end-date 2026-04-28
volumeleaders-agent trade sentiment --days 7
volumeleaders-agent trade sentiment --start-date 2026-04-21 --end-date 2026-04-28 --format csv
```

JSON: `dateRange`, `daily`, `totals`. Daily rows include `date`, `bear`, `bull`, `ratio`, `signal`. Ratio is bull dollars divided by bear dollars, `null` when bear flow is zero. Signals: `<0.2 extreme_bear`, `<0.5 moderate_bear`, `0.5-2.0 neutral`, `>2.0-5.0 moderate_bull`, `>5.0 extreme_bull`. CSV/TSV flatten daily rows and add a final `date=total` row.

## Clusters and cluster bombs

`trade clusters` finds multiple institutional trades converging at similar prices. Requires a complete date range or `--days`. Defaults: `--min-dollars 10000000`, `--length 1000`, `--order-col 1`, `--order-dir desc`. Optional filters: ticker aliases, positional tickers, `--sector`, volume/price/dollar ranges, `--vcd`, `--security-type`, `--relative-size`, `--trade-cluster-rank`, format, pagination.

`trade cluster-bombs` finds sudden, aggressive bursts tightly grouped in time and price. Requires a complete date range or `--days`. Defaults: `--min-dollars 0`, `--security-type 0`, `--relative-size 0`, `--length 100`. It accepts ticker aliases and positional tickers, has no price range filters, and uses `--trade-cluster-bomb-rank`.

```bash
volumeleaders-agent trade clusters --start-date 2026-04-01 --end-date 2026-04-28 --min-dollars 50000000
volumeleaders-agent trade clusters AAPL MSFT --days 5
volumeleaders-agent trade cluster-bombs --start-date 2026-04-28 --end-date 2026-04-28
```

Key fields: `Ticker`, `Name`, `Sector`, `Industry`, `Date`, `Price`, `Dollars`, `Volume`, `TradeCount`, `DollarsMultiplier`, `CumulativeDistribution`, `TradeClusterRank` or `TradeClusterBombRank`, `MinFullDateTime`, `MaxFullDateTime`, `TotalRows`.

## Alerts

`trade alerts --date D` returns system-generated individual trade alerts. `trade cluster-alerts --date D` returns system-generated cluster alerts. Both support `--format json|csv|tsv` and pagination.

```bash
volumeleaders-agent trade alerts --date 2026-04-28
volumeleaders-agent trade cluster-alerts --date 2026-04-28 --format tsv
```

Trade alert fields include `Ticker`, `Name`, `AlertType`, `Price`, `TradeRank`, `VolumeCumulativeDistribution`, `DollarsMultiplier`, `Volume`, `Dollars`, booleans, `FullDateTime`, `InProcess`, `Complete`. Cluster alert fields match `trade clusters`.

## Levels and level touches

`trade levels` finds significant institutional prices for one ticker. Required: `--ticker` (aliases accepted) or one positional ticker. Optional: `--start-date`, `--end-date`, `--days`, shared ranges, `--vcd`, `--trade-level-rank`, `--trade-level-count`, `--fields`, `--format`. Defaults to a 1-year lookback when dates are omitted, `--trade-level-count 10`, and non-standard `--relative-size 0`. It does not expose pagination flags, but sends one request with `start=0` and `length=-1` to match the observed VolumeLeaders levels request and return all matching levels.

Default `trade levels` JSON returns compact analysis rows. It keeps level price, dollars, volume, trade count, relative size, cumulative distribution, rank, and min/max dates. Repetitive single-ticker metadata (`Ticker`, `Name`) and the verbose `Dates` list are omitted to reduce tokens. Use `--fields Ticker,Name,Dates` or another explicit field list when those raw API fields are needed. CSV/TSV without `--fields` still use the full raw trade level row columns.

`trade level-touches` finds events where price revisits institutional levels. Required: complete `--start-date`/`--end-date` range or `--days`. Optional: ticker aliases, positional tickers, volume/price/dollar ranges, `--vcd`, `--trade-level-rank`, format, pagination. Defaults: `--relative-size 0`, `--trade-level-rank 10`, `--order-col 0`, `--order-dir desc`.

```bash
volumeleaders-agent trade levels AAPL --days 30
volumeleaders-agent trade level-touches XLE --days 5
```

Default `trade levels` JSON row fields: `Price`, `Dollars`, `Volume`, `Trades`, `RelativeSize`, `CumulativeDistribution`, `TradeLevelRank`, `MinDate`, `MaxDate`.

Raw level fields available through `--fields`, CSV, or TSV: `Ticker`, `Name`, `Price`, `Dollars`, `Volume`, `Trades`, `RelativeSize`, `CumulativeDistribution`, `TradeLevelRank`, `MinDate`, `MaxDate`, `Dates`. Level-touch rows include those context fields plus `Sector`, `Industry`, `Date`, `FullDateTime`, `FullTimeString24`, `TotalRows`, and `TradeLevelTouches`.
