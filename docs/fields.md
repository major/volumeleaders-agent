# Trade and cluster fields

This page explains the upstream VolumeLeaders trade and cluster JSON fields used by the CLI presets. Keep the CLI help and structcli metadata token-efficient; use this page for longer field notes from `annotated-trade.jsonc` and `annotated-cluster.jsonc`.

## Presets

- `core`: compact default for LLM workflows. Emits the most useful table fields plus derived `CalendarEvent` and `AuctionTrade` where available.
- `expanded`: all annotated non-internal signal fields. Still excludes raw upstream internals, always-zero fields, and always-null fields.
- `full`: raw upstream payload. Use only for debugging or when a field has not been classified yet.

Array output is the most token-efficient shape because field names appear once in `fields` and row values follow the same order in `rows`. Object output repeats field names per row but is easier to inspect manually.

## Derived fields

- `CalendarEvent`: joins true upstream calendar flags from `EOM`, `EOQ`, `EOY`, `OPEX`, and `VOLEX`. Array output uses `null` when no flag is true. Object output omits the field when no flag is true.
- `AuctionTrade`: derives from trade `OpeningTrade` and `ClosingTrade` `0`/`1` flags. Values are `"open"`, `"close"`, or `null` in array output. Object output omits it when neither flag is true. Clusters do not currently include upstream auction fields in annotated cluster records.

## Shared signal fields

- `Date`: upstream date wrapper for the trade or cluster date.
- `DateKey`: compact `YYYYMMDD` date key.
- `Ticker`, `Sector`, `Industry`, `Name`: symbol and company classification fields.
- `Price`: trade price, or the shared price for trades in a cluster.
- `Dollars`: total dollar value, calculated from price and volume.
- `Volume`: shares traded.
- `IPODate`: company IPO date.
- `CumulativeDistribution`: percentile-like distribution value for trade or cluster size relative to that ticker.
- `InsideBar`, `DoubleInsideBar`: daily bar compression signals. `DoubleInsideBar` means the inside-bar condition also applied to the previous day.
- `EOM`, `EOQ`, `EOY`, `OPEX`, `VOLEX`: raw calendar flags behind `CalendarEvent`.

## Trade-only signal fields

- `TimeKey`: compact `HHMMSS` trade time key.
- `TradeID`: unique upstream trade identifier. Useful for correlation, not as a trading signal by itself.
- `FullDateTime`, `FullTimeString24`: full trade timestamp and `HH:MM:SS` time.
- `LastComparibleTradeDate`: upstream spelling for the most recent comparable trade date.
- `OffsettingTradeDate`: most recent date with an offsetting trade.
- `TradeCount`: number of large trades for this ticker today, not the number of executions in the row.
- `TradeRank`: mutable current rank. `9999` means unranked.
- `TradeRankSnapshot`: immutable rank when the trade was processed. `9999` means unranked.
- `LatePrint`, `Sweep`, `DarkPool`, `PhantomPrint`, `SignaturePrint`: upstream print classifications. `SignaturePrint` marks very late reports that are usually less interesting.
- `OpeningTrade`, `ClosingTrade`: raw auction flags behind `AuctionTrade`.
- `NewPosition`: whether the trade appears to represent a new position.
- `RSIHour`, `RSIDay`: hourly and daily RSI values at the trade time.
- `TotalRows`: total upstream rows for the ticker/day result set.
- `FrequencyLast30TD`, `FrequencyLast90TD`, `FrequencyLast1CY`: count of trades this size or larger in the last 30 trading days, 90 trading days, or one calendar year.
- `Cancelled`: `1` means the trade was cancelled and should be ignored for analysis.

## Cluster-only signal fields

- `MinFullDateTime`, `MaxFullDateTime`: earliest and latest trade timestamps in the cluster.
- `MinFullTimeString24`, `MaxFullTimeString24`: earliest and latest cluster times in `HH:MM:SS` form.
- `TradeCount`: number of trades inside the cluster.
- `TradeClusterRank`: mutable current cluster rank. `9999` means unranked.
- `LastComparibleTradeClusterDate`: upstream spelling for the most recent comparable cluster date.

## Excluded from expanded

The `expanded` preset intentionally excludes fields annotated as internal, always zero, always null, or otherwise not useful as direct signal columns.

Trade exclusions include query echoes and internal dates (`StartDate`, `EndDate`, `TD30`, `TD90`, `TD1CY`), internal identifiers (`SecurityKey`, `SequenceNumber`), always-zero quote or aggregate fields (`Bid`, `Ask`, `AverageDailyVolume`, `PercentDailyVolume`, `AverageBlockSizeDollars`, `AverageBlockSizeShares`, institutional totals, closing totals, `ClosePrice`, `TotalDollars`, `TotalDollarsRank`, `TotalVolume`, `TotalTrades`), internal relative-size plumbing (`DollarsMultiplier`), always-null phantom fulfillment fields, `TradeConditions`, and `ExternalFeed`.

Cluster exclusions include `SecurityKey`, `ClosePrice`, `AverageBlockSizeShares`, `AverageBlockSizeDollars`, `AverageDailyVolume`, `DollarsMultiplier`, `TotalRows`, and `ExternalFeed`.
