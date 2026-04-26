# trade

Institutional trade discovery. 8 subcommands. See SKILL.md for shared flag defaults and metrics glossary.

## trade presets

List all built-in trade filter presets. No authentication required. Output is a JSON array of objects with `Name`, `Group`, and `Filters` fields.

Groups: "Common" (general-purpose filters), "Disproportionately Large" (>=5x avg size, sector/ticker-based).

```bash
volumeleaders-agent trade presets
volumeleaders-agent trade presets --pretty | jq '.[].Name'
```

## trade list

Query individual institutional block trades. The primary trade discovery command.

Required: `--start-date`, `--end-date`
Optional: `--tickers` (aliases: `--ticker`, `--symbol`, `--symbols`), `--sector`, `--preset`, `--watchlist`, `--fields`, `--summary`, `--group-by`, all shared flags (volume/price/dollar ranges, trade filters, trade type toggles, session toggles), pagination (`--length 100 --order-col 1 --order-dir desc`)

**`--preset NAME`**: Apply a built-in filter preset by name (case-insensitive). The preset sets baseline filters; any explicitly-provided CLI flags override the preset values. Use `trade presets` to list available names.

**`--watchlist NAME`**: Apply filters from a saved user watchlist by name (case-insensitive). Fetches the watchlist config at runtime and converts its settings to trade filters. Use `watchlist configs` to list available names.

**`--fields FIELD1,FIELD2`**: Return only the listed trade fields in each JSON object. Field names are case-sensitive and must match the output field names below. Invalid names fail before querying the API and include the valid field list in the error.

**`--summary`**: Return aggregate metrics instead of individual trade rows. The summary includes `totalTrades`, `totalDollars`, `dateRange`, and one grouped section. Metrics per group are `trades`, `dollars`, `avgDollarsMultiplier`, `pctDarkPool`, `pctSweep`, and `avgCumulativeDistribution`. Summaries respect pagination, use `--length -1` to aggregate all matching rows. `--fields` cannot be used with `--summary`.

**`--group-by VALUE`**: Select the summary grouping. Valid values are `ticker` (default, outputs `byTicker`), `day` (outputs `byDay`), and `ticker,day` (outputs `byTickerDay` keys in `TICKER|YYYY-MM-DD` format). Only applies with `--summary`.

Preset and watchlist filters can be combined: watchlist filters merge on top of preset filters, and explicit CLI flags override both.

```bash
volumeleaders-agent trade list --tickers AAPL --start-date 2025-04-16 --end-date 2025-04-23 --dark-pools 1 --min-dollars 1000000
volumeleaders-agent trade list --preset "Top-100 Rank" --start-date 2025-04-01 --end-date 2025-04-24
volumeleaders-agent trade list --preset "Megacaps" --start-date 2025-04-01 --end-date 2025-04-24 --trade-rank 10
volumeleaders-agent trade list --watchlist "Magnificent 7" --start-date 2025-04-01 --end-date 2025-04-24
volumeleaders-agent trade list --tickers SPY,QQQ --start-date 2025-04-21 --end-date 2025-04-25 --fields Date,Ticker,Dollars,DollarsMultiplier,DarkPool,CumulativeDistribution
volumeleaders-agent trade list --tickers SPY,QQQ --start-date 2025-04-21 --end-date 2025-04-25 --summary --group-by ticker,day
```

Output fields: `AHInstitutionalDollars`, `AHInstitutionalDollarsRank`, `AHInstitutionalVolume`, `Ask`, `AverageBlockSizeDollars`, `AverageBlockSizeShares`, `AverageDailyVolume`, `Bid`, `Cancelled`, `ClosePrice`, `ClosingTrade`, `ClosingTradeDollars`, `ClosingTradeDollarsRank`, `ClosingTradeVolume`, `CumulativeDistribution`, `DarkPool`, `Date`, `DateKey`, `Dollars`, `DollarsMultiplier`, `DoubleInsideBar`, `EOM`, `EOQ`, `EOY`, `EndDate`, `ExternalFeed`, `FrequencyLast1CY`, `FrequencyLast30TD`, `FrequencyLast90TD`, `FullDateTime`, `FullTimeString24`, `IPODate`, `Industry`, `InsideBar`, `LastComparibleTradeDate`, `LatePrint`, `Name`, `NewPosition`, `OPEX`, `OffsettingTradeDate`, `OpeningTrade`, `PercentDailyVolume`, `PhantomPrint`, `PhantomPrintFulfillmentDate`, `PhantomPrintFulfillmentDays`, `Price`, `RSIDay`, `RSIHour`, `Sector`, `SecurityKey`, `SequenceNumber`, `SignaturePrint`, `StartDate`, `Sweep`, `TD1CY`, `TD30`, `TD90`, `Ticker`, `TimeKey`, `TotalDollars`, `TotalDollarsRank`, `TotalInstitutionalDollars`, `TotalInstitutionalDollarsRank`, `TotalInstitutionalVolume`, `TotalRows`, `TotalTrades`, `TotalVolume`, `TradeConditions`, `TradeCount`, `TradeID`, `TradeRank`, `TradeRankSnapshot`, `VOLEX`, `Volume`

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
