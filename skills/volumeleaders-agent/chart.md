# chart

Price data, snapshots, levels, and company metadata for individual tickers.

**Flag name warning**: Chart commands use different flag names than trade commands for several filters. Differences are marked with bold annotations below.

## chart price-data

One-minute OHLCV bars with institutional trade overlays. Each bar includes trade metadata when institutional trades occurred during that minute.

Required: `--ticker`, `--start-date`, `--end-date`

Optional flags:

| Flag | Default | Note |
|---|---|---|
| `--volume-profile` | 0 | 0=no, 1=yes |
| `--levels` | 5 | Number of trade levels to overlay |
| `--min-volume` / `--max-volume` | 0 / 2000000000 | |
| `--min-price` / `--max-price` | 0 / 100000 | |
| `--min-dollars` / `--max-dollars` | 500000 / 30000000000 | |
| `--dark-pools` | -1 | Toggle |
| `--sweeps` | -1 | Toggle |
| `--late-prints` | -1 | Toggle |
| **`--signature-prints`** | -1 | **Not `--sig-prints`** |
| `--trade-count` | 3 | Min trades for overlay |
| `--vcd` | 0 | |
| `--trade-rank` | -1 | |
| **`--trade-rank-snapshot`** | -1 | **Not `--rank-snapshot`** |
| **`--include-premarket`** | 1 | **Not `--premarket`** |
| **`--include-rth`** | 1 | **Not `--rth`** |
| **`--include-ah`** | 1 | **Not `--ah`** |
| **`--include-opening`** | 1 | **Not `--opening`** |
| **`--include-closing`** | 1 | **Not `--closing`** |
| **`--include-phantom`** | 1 | **Not `--phantom`** |
| **`--include-offsetting`** | 1 | **Not `--offsetting`** |
| `--format` | json | `json`, `csv`, or `tsv` |

```bash
volumeleaders-agent chart price-data --ticker AAPL --start-date 2025-04-23 --end-date 2025-04-23
volumeleaders-agent chart price-data --ticker AAPL --start-date 2025-04-23 --end-date 2025-04-23 --format tsv
```

Output fields (PriceBar model): `DateKey`, `TimeKey`, `Date`, `FullDateTime`, `FullTimeString24`, `OpenPrice`, `ClosePrice`, `HighPrice`, `LowPrice`, `Volume`, `Dollars`, `Trades`, `CumulativeDistribution`, `TradeRank`, `TradeRankSnapshot`, `TradeLevelRank`, `DollarsMultiplier`, `RelativeSize`, `DarkPoolTrade` (note: NOT `DarkPool`), `LatePrint`, `OpeningTrade`, `ClosingTrade`, `SignaturePrint`, `PhantomPrint`, `Sweep`, `FrequencyLast30TD`, `FrequencyLast90TD`, `FrequencyLast1CY`

## chart snapshot

Quick quote with bid/ask/last for a single ticker.

Required: `--ticker`, `--date-key` (YYYY-MM-DD)

```bash
volumeleaders-agent chart snapshot --ticker AAPL --date-key 2025-04-23
```

Output fields: `ticker`, `lastQuote` (nested: `p`=bid price, `s`=bid size, `P`=ask price, `S`=ask size, `t`=timestamp), `lastTrade` (nested: `p`=last trade price), `todaysChange`, `todaysChangePerc`

Note: `lastQuote` and `lastTrade` use single-letter JSON keys, not descriptive names.

## chart levels

Institutional price levels for charting. Simpler interface than `trade levels` with fewer filter options.

Required: `--ticker`, `--start-date`, `--end-date`
Optional: `--levels` (default 5), `--format json|csv|tsv`

```bash
volumeleaders-agent chart levels --ticker AAPL --start-date 2025-01-01 --end-date 2025-04-23 --levels 10
```

Output fields: same TradeLevel model as `trade levels`.

Output format: `chart price-data` and `chart levels` default to JSON and support CSV/TSV. `chart snapshot` and `chart company` remain JSON-only single-object outputs.

## chart company

Company metadata: fundamentals, sector/industry, and trading averages.

Required: `--ticker`

```bash
volumeleaders-agent chart company --ticker AAPL
```

Output fields: `SecurityKey`, `Name`, `Ticker`, `Sector`, `Industry`, `Description`, `HomePageURL`, `MarketCap`, `CurrentPrice`, `OptionsEnabled`, `IPODate`, plus averages in all-time / 30-day / 90-day variants: `AverageBlockSizeDollars[30Days|90Days]`, `AverageDailyVolume[30Days|90Days]`, `AverageTradeShares[30Days|90Days]`, `AverageDailyRange[30Days|90Days]`, `AverageDailyRangePct[30Days|90Days]`, `AverageClosingTradeDollars[30Days|90Days]`, `AverageClusterSizeDollars`, `AverageLevelSizeDollars`
