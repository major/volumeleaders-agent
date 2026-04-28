# trade

Institutional trade discovery. See `SKILL.md` for global conventions, trade shared flags, and metric meanings.

## Quick Reference

| Command | Use when | Required | Output |
|---|---|---|---|
| `trade presets` | List built-in filters | none | JSON, CSV, TSV |
| `trade preset-tickers` | Inspect a preset universe | `--preset` | JSON |
| `trade list` | Query individual trades | none | JSON, CSV, TSV, or summary JSON |
| `trade sentiment` | Bull/bear leveraged ETF flow | `--start-date`, `--end-date` | JSON, CSV, TSV |
| `trade clusters` | Converging trades at price levels | `--start-date`, `--end-date` | JSON, CSV, TSV |
| `trade cluster-bombs` | Sudden aggressive bursts | `--start-date`, `--end-date` | JSON, CSV, TSV |
| `trade alerts` | System trade alerts | `--date` | JSON, CSV, TSV |
| `trade cluster-alerts` | System cluster alerts | `--date` | JSON, CSV, TSV |
| `trade levels` | Support/resistance by ticker | `--ticker` | JSON, CSV, TSV |
| `trade level-touches` | Price revisits institutional levels | `--start-date`, `--end-date` | JSON, CSV, TSV |

## Presets

`trade presets` lists built-in filters. Groups: `Common`, `Disproportionately Large`.

`trade preset-tickers --preset NAME` returns `Preset`, `Group`, `Type`, and either `Tickers` or `SectorIndustry`. Types: `tickers`, `sector-filter`, `unfiltered`. Ticker output takes precedence if both ticker and sector filters exist.

```bash
volumeleaders-agent trade presets --format csv
volumeleaders-agent trade preset-tickers --preset "Megacaps"
```

## trade list

Primary individual-print query. Optional flags: `--start-date`, `--end-date`, ticker aliases, `--sector`, `--preset`, `--watchlist`, `--fields`, `--format`, `--summary`, `--group-by`, trade shared flags, pagination.

Date defaults: with tickers, 90-day lookback from today. Without tickers, today only. Explicit dates override defaults. Presets and watchlists never supply dates.

Filter precedence: preset baseline, then watchlist merge, then explicit CLI flags override both.

Summary mode: `--summary` returns aggregate JSON instead of rows. Valid `--group-by`: `ticker`, `day`, `ticker,day`. Summary respects pagination, so use `--length -1` for all matching rows. `--fields` and non-JSON `--format` are invalid with `--summary`.

Fields mode: `--fields FIELD1,FIELD2` limits JSON keys or CSV/TSV columns. Names are case-sensitive and validated before the API query.

```bash
volumeleaders-agent trade list --tickers AAPL --dark-pools 1 --min-dollars 1000000
volumeleaders-agent trade list --preset "Top-100 Rank" --start-date 2026-04-28 --end-date 2026-04-28
volumeleaders-agent trade list --watchlist "Magnificent 7" --start-date 2026-04-01 --end-date 2026-04-28
volumeleaders-agent trade list --tickers SPY,QQQ --start-date 2026-04-21 --end-date 2026-04-28 --summary --group-by ticker,day --length -1
volumeleaders-agent trade list --tickers AAPL,MSFT --start-date 2026-04-21 --end-date 2026-04-28 --fields Date,Ticker,Dollars --format csv
```

Key row fields: `Date`, `FullDateTime`, `Ticker`, `Name`, `Sector`, `Industry`, `Price`, `Volume`, `Dollars`, `DollarsMultiplier`, `CumulativeDistribution`, `RelativeSize`, `PercentDailyVolume`, `TradeRank`, `TradeRankSnapshot`, `VCD`, `DarkPool`, `Sweep`, `LatePrint`, `SignaturePrint`, `PhantomPrint`, `OpeningTrade`, `ClosingTrade`, `RSIHour`, `RSIDay`, `TotalRows`.

## trade sentiment

Daily bull/bear leveraged ETF flow. The command always queries the combined leveraged ETF sector filter `SectorIndustry=X B`, classifies bull and bear ETFs locally, and cannot be constrained by ticker or sector flags.

Required: `--start-date`, `--end-date`. Optional: `--format json|csv|tsv`, trade shared flags except ticker/sector filters. Non-standard defaults: `--min-dollars 5000000`, `--vcd 97`; shared `--relative-size 5` still applies.

```bash
volumeleaders-agent trade sentiment --start-date 2026-04-21 --end-date 2026-04-28
volumeleaders-agent trade sentiment --start-date 2026-04-21 --end-date 2026-04-28 --format csv
```

JSON: `dateRange`, `daily`, `totals`. Daily rows include `date`, `bear`, `bull`, `ratio`, `signal`. Ratio is bull dollars divided by bear dollars, `null` when bear flow is zero. Signals: `<0.2 extreme_bear`, `<0.5 moderate_bear`, `0.5-2.0 neutral`, `>2.0-5.0 moderate_bull`, `>5.0 extreme_bull`. CSV/TSV flatten daily rows and add a final `date=total` row.

## Clusters and cluster bombs

`trade clusters` finds multiple institutional trades converging at similar prices. Defaults: `--min-dollars 10000000`, `--length 1000`, `--order-col 1`, `--order-dir desc`. Optional filters: ticker aliases, `--sector`, volume/price/dollar ranges, `--vcd`, `--security-type`, `--relative-size`, `--trade-cluster-rank`, format, pagination.

`trade cluster-bombs` finds sudden, aggressive bursts tightly grouped in time and price. Defaults: `--min-dollars 0`, `--security-type 0`, `--relative-size 0`, `--length 100`. It has no price range filters and uses `--trade-cluster-bomb-rank`.

```bash
volumeleaders-agent trade clusters --start-date 2026-04-01 --end-date 2026-04-28 --min-dollars 50000000
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

`trade levels` finds significant institutional prices for one ticker. Required: `--ticker` (aliases accepted). Optional: `--start-date`, `--end-date`, shared ranges, `--vcd`, `--trade-level-rank`, `--trade-level-count`, `--format`. Defaults to a 1-year lookback when dates are omitted. Non-standard default: `--relative-size 0`. No pagination.

`trade level-touches` finds events where price revisits institutional levels. Required: `--start-date`, `--end-date`. Optional: ticker aliases, volume/price/dollar ranges, `--vcd`, `--trade-level-rank`, format, pagination. Defaults: `--relative-size 0`, `--trade-level-rank 10`, `--order-col 0`, `--order-dir desc`.

```bash
volumeleaders-agent trade levels --ticker AAPL
volumeleaders-agent trade level-touches --start-date 2026-04-28 --end-date 2026-04-28
```

Key fields: `Ticker`, `Name`, `Price`, `Dollars`, `Volume`, `Trades`, `RelativeSize`, `CumulativeDistribution`, `TradeLevelRank`, `MinDate`, `MaxDate`, `Dates`, plus `TradeLevelTouches` on level-touch rows.
