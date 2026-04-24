# trade

Institutional trade discovery. 7 subcommands. See SKILL.md for shared flag defaults and metrics glossary.

## trade list

Query individual institutional block trades. The primary trade discovery command.

Required: `--start-date`, `--end-date`
Optional: `--tickers` (aliases: `--ticker`, `--symbol`, `--symbols`), `--sector`, all shared flags (volume/price/dollar ranges, trade filters, trade type toggles, session toggles), pagination (`--length 100 --order-col 1 --order-dir desc`)

```bash
volumeleaders-agent trade list --tickers AAPL --start-date 2025-04-16 --end-date 2025-04-23 --dark-pools 1 --min-dollars 1000000
```

Output fields: `Ticker`, `Name`, `Sector`, `Industry`, `Date`, `Price`, `Bid`, `Ask`, `Dollars`, `DollarsMultiplier`, `Volume`, `AverageDailyVolume`, `PercentDailyVolume`, `TradeCount`, `CumulativeDistribution`, `TradeRank`, `TradeRankSnapshot`, `DarkPool`, `Sweep`, `LatePrint`, `SignaturePrint`, `OpeningTrade`, `ClosingTrade`, `PhantomPrint`, `InsideBar`, `DoubleInsideBar`, `NewPosition`, `Cancelled`, `TotalInstitutionalDollars`, `TotalInstitutionalDollarsRank`, `TotalInstitutionalVolume`, `AHInstitutionalDollars`, `AHInstitutionalDollarsRank`, `AHInstitutionalVolume`, `ClosingTradeDollars`, `ClosingTradeDollarsRank`, `ClosingTradeVolume`, `TotalDollars`, `TotalDollarsRank`, `TotalVolume`, `ClosePrice`, `RSIHour`, `RSIDay`, `FrequencyLast30TD`, `FrequencyLast90TD`, `FrequencyLast1CY`, `LastComparibleTradeDate`, `IPODate`, `TotalRows`

## trade clusters

Aggregated clusters where multiple institutional trades converge at similar price levels. Clusters signal stronger conviction than individual trades.

Required: `--start-date`, `--end-date`
Optional: `--tickers` (aliases: `--ticker`, `--symbol`, `--symbols`), `--sector`, volume/price/dollar ranges, `--vcd`, `--security-type`, `--relative-size`, `--trade-cluster-rank` (-1), pagination (`--length 1000 --order-col 1 --order-dir desc`)
Non-standard defaults: `--min-dollars 10000000`, `--length 1000`

```bash
volumeleaders-agent trade clusters --start-date 2025-04-01 --end-date 2025-04-23 --min-dollars 50000000
```

Output fields: `Ticker`, `Name`, `Sector`, `Industry`, `Date`, `Price`, `Dollars`, `Volume`, `TradeCount`, `DollarsMultiplier`, `CumulativeDistribution`, `AverageDailyVolume`, `TradeClusterRank`, `MinFullDateTime`, `MaxFullDateTime`, `ClosePrice`, `InsideBar`, `DoubleInsideBar`, `LastComparibleTradeClusterDate`, `TotalRows`

## trade cluster-bombs

Sudden, aggressive institutional positioning bursts. Many trades clustered tightly in time and price.

Required: `--start-date`, `--end-date`
Optional: `--tickers` (aliases: `--ticker`, `--symbol`, `--symbols`), `--sector`, volume/dollar ranges (no price range), `--vcd`, `--security-type`, `--relative-size`, `--trade-cluster-bomb-rank` (-1), pagination (`--length 1000 --order-col 1 --order-dir desc`)
Non-standard defaults: `--min-dollars 0`, `--security-type 0`, `--relative-size 0`

```bash
volumeleaders-agent trade cluster-bombs --start-date 2025-04-23 --end-date 2025-04-23
```

Output fields: same as clusters but `TradeClusterBombRank` instead of `TradeClusterRank`, `LastComparableTradeClusterBombDate` instead of `LastComparibleTradeClusterDate`.

## trade alerts

System-generated notifications about notable individual trades for a single date.

Required: `--date`
Optional: pagination

```bash
volumeleaders-agent trade alerts --date 2025-04-23
```

Output fields: `Ticker`, `Name`, `Sector`, `Industry`, `Date`, `AlertType`, `Price`, `TradeRank`, `VolumeCumulativeDistribution`, `DollarsMultiplier`, `Volume`, `Dollars`, `RSIHour`, `RSIDay`, `DarkPool`, `Sweep`, `LatePrint`, `SignaturePrint`, `ClosingTrade`, `PhantomPrint`, `FullDateTime`, `InProcess`, `Complete`

## trade cluster-alerts

System-generated notifications about notable trade clusters for a single date.

Required: `--date`
Optional: pagination

```bash
volumeleaders-agent trade cluster-alerts --date 2025-04-23
```

Output fields: same as trade clusters.

## trade levels

Significant institutional price levels for a single ticker. These levels often act as support/resistance.

Required: `--ticker` (single; aliases: `--tickers`, `--symbol`, `--symbols`), `--start-date`, `--end-date`
Optional: volume/price/dollar ranges, `--vcd`, `--trade-level-rank` (-1), `--trade-level-count` (10)
Non-standard defaults: `--relative-size 0`. No pagination flags.

```bash
volumeleaders-agent trade levels --ticker AAPL --start-date 2025-01-01 --end-date 2025-04-23
```

Output fields: `Ticker`, `Name`, `Price`, `Dollars`, `Volume`, `Trades`, `RelativeSize`, `CumulativeDistribution`, `TradeLevelRank`, `MinDate`, `MaxDate`, `Dates`

## trade level-touches

Events where price revisits significant institutional price levels. Signals support/resistance tests or institutional re-engagement.

Required: `--start-date`, `--end-date`
Optional: `--tickers` (aliases: `--ticker`, `--symbol`, `--symbols`), volume/price/dollar ranges, `--vcd`, `--trade-level-rank` (10), pagination (`--length 100 --order-col 0 --order-dir desc`)
Non-standard defaults: `--relative-size 0`, `--order-col 0`, `--trade-level-rank 10`

```bash
volumeleaders-agent trade level-touches --start-date 2025-04-23 --end-date 2025-04-23
```

Output fields: `Ticker`, `Name`, `Sector`, `Industry`, `Date`, `FullDateTime`, `Price`, `Dollars`, `Volume`, `Trades`, `RelativeSize`, `CumulativeDistribution`, `TradeLevelRank`, `TradeLevelTouches`, `MinDate`, `MaxDate`, `Dates`, `TotalRows`
