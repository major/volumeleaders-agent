# Trade, cluster, level, touch, and watchlist fields

This page explains the upstream VolumeLeaders trade, cluster, level, touch, and watchlist JSON fields used by the CLI presets. Keep the CLI help and structcli metadata token-efficient; use this page for longer field notes from `annotated-trade.jsonc`, `annotated-cluster.jsonc`, captured trade-level responses, captured trade-level-touch responses, and `Getwatchlists.jsonc`.

## Presets

- Trade, cluster, trade-cluster-bomb, trade-level, and trade-level-touch `core`: compact default for LLM workflows. Trades and clusters emit the most useful table fields plus derived `CalendarEvent` and `AuctionTrade` where available. Trade cluster bombs emit the visible cluster-bomb table fields plus derived `CalendarEvent`. Trade levels and level touches emit the visible level table fields.
- Watchlist `summary`: compact watchlist default with only `SearchTemplateKey` and `Name`, so callers can choose a saved filter before requesting details.
- `expanded`: all annotated non-internal signal fields for trades and clusters, or the saved filter configuration fields for watchlists. Still excludes raw upstream internals, always-zero fields, and always-null fields.
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

## Trade cluster bomb fields

The `trade-cluster-bombs` command fetches `TradeClusterBombs/GetTradeClusterBombs` across a date range. VolumeLeaders defines cluster bombs as rare stock-only rows with at least three dark-pool sweeps in one day and at least $38M combined value. VolumeLeaders accepts at most 7 days when querying all tickers or multiple comma-delimited tickers, so those scans default to 7 days and reject wider requested ranges. Single-ticker scans default to a one-year lookback when `--start-date` is omitted.

- `MinFullDateTime`, `MaxFullDateTime`: earliest and latest trade timestamps in the cluster bomb.
- `MinFullTimeString24`, `MaxFullTimeString24`: earliest and latest cluster bomb times in `HH:MM:SS` form.
- `TradeCount`: number of trades inside the cluster bomb.
- `DollarsMultiplier`: returned relative size value shown as `RS` in the browser table.
- `TradeClusterBombRank`: mutable current cluster bomb rank. Captured browser defaults use `-1` as the no-rank request filter.
- `LastComparableTradeClusterBombDate`: upstream spelling for the most recent comparable cluster bomb date. Unlike older trade and cluster fields, this endpoint spells `Comparable` correctly.
- `TotalRows`: upstream total row count used as a fallback when DataTables record totals are omitted.

## Excluded from expanded

The `expanded` preset intentionally excludes fields annotated as internal, always zero, always null, or otherwise not useful as direct signal columns.

Trade exclusions include query echoes and internal dates (`StartDate`, `EndDate`, `TD30`, `TD90`, `TD1CY`), internal identifiers (`SecurityKey`, `SequenceNumber`), always-zero quote or aggregate fields (`Bid`, `Ask`, `AverageDailyVolume`, `PercentDailyVolume`, `AverageBlockSizeDollars`, `AverageBlockSizeShares`, institutional totals, closing totals, `ClosePrice`, `TotalDollars`, `TotalDollarsRank`, `TotalVolume`, `TotalTrades`), internal relative-size plumbing (`DollarsMultiplier`), always-null phantom fulfillment fields, `TradeConditions`, and `ExternalFeed`.

Cluster exclusions include `SecurityKey`, `ClosePrice`, `AverageBlockSizeShares`, `AverageBlockSizeDollars`, `AverageDailyVolume`, `DollarsMultiplier`, `TotalRows`, and `ExternalFeed`.

Trade cluster bomb exclusions include `SecurityKey`, `ClosePrice`, `AverageBlockSizeShares`, `AverageBlockSizeDollars`, `AverageDailyVolume`, `TotalRows`, and `ExternalFeed`.

## Trade level fields

The `trade-levels` command fetches one ticker's level table from `TradeLevels/GetTradeLevels`. A trade level groups large prints by price across a date range. Default `core` fields are `Ticker`, `Price`, `Dollars`, `Volume`, `Trades`, `RelativeSize`, `CumulativeDistribution`, `TradeLevelRank`, and `Dates`.

- `Ticker`, `Sector`, `Industry`, `Name`: symbol and company classification fields.
- `Date`, `MinDate`, `MaxDate`, `FullDateTime`, `FullTimeString24`: upstream date/timestamp context for the level result.
- `Price`: the price level.
- `Dollars`: aggregate dollars for large prints at that level.
- `Volume`: aggregate shares for large prints at that level.
- `Trades`: count of prints contributing to the level.
- `RelativeSize`: the level's relative size versus the ticker's average dollar trade size. Captured filter values are `0`, `3`, `5`, and `10`, where `0` means any size.
- `CumulativeDistribution`: percentile-like size distribution for the level, shown as `PCT` in the browser table.
- `TradeLevelRank`: all-time level rank when VolumeLeaders marks the level as ranked. Captured BAND output included ranks such as `44`, `81`, `89`, and `98`, while unranked rows used `0`.
- `TradeLevelTouches`: upstream touch count for the level.
- `Dates`: upstream level date range string, for example `2022-07-27 - 2026-04-13`.
- `TotalRows`: upstream total row count used as a fallback when DataTables record totals are omitted.

## Trade level touch fields

The `trade-level-touches` command fetches `TradeLevelTouches/GetTradeLevelTouches` across a date range. A touch happens when price returns from above or below to test a ranked large-trade level, so these rows are support/resistance context tied to the rank of the touched level on that day. VolumeLeaders accepts at most 7 days when querying all tickers or multiple comma-delimited tickers, so those scans default to 7 days and reject wider requested ranges. Single-ticker scans default to a one-year lookback when `--start-date` is omitted. The captured browser default is `TradeLevelRank=10`, meaning ranks 1 through 10.

Default `core` fields are `Ticker`, `FullDateTime`, `Price`, `Dollars`, `Volume`, `Trades`, `RelativeSize`, `CumulativeDistribution`, `TradeLevelRank`, `Dates`, `Sector`, and `Industry`.

- `FullDateTime`, `FullTimeString24`: timestamp for the price touch event.
- `Date`, `MinDate`, `MaxDate`, and `Dates`: date context for the touched level. `Dates` is the browser's level date range string.
- `Price`: the support or resistance level touched by price.
- `Dollars`, `Volume`, and `Trades`: aggregate large-trade size behind the touched level.
- `RelativeSize`: the level's relative size versus the ticker's average dollar trade size.
- `CumulativeDistribution`: percentile-like size distribution for the touched level, shown as `PCT` in the browser table.
- `TradeLevelRank`: touched level rank on the touch date. The command defaults the request filter to `10`, which returns ranks 1 through 10.
- `TradeLevelTouches`: upstream touch count for the level in the captured row.
- `TotalRows`: upstream total row count used as a fallback when DataTables record totals are omitted.

## Watchlist fields

The `watchlists` command lists saved account-level watchlist filters from `WatchListConfigs/GetWatchLists`. Its default `summary` preset returns only `SearchTemplateKey` and `Name` so callers can pick a watchlist before requesting criteria fields. The `save-watchlist` command creates or updates those filters by posting the authenticated `WatchListConfig` form, and `delete-watchlist` removes one by posting `WatchListConfigs/DeleteWatchList` with the selected key. These are not necessarily traditional ticker watchlists. They are saved criteria for which trades or clusters should appear in the VolumeLeaders watchlist UI.

- `SearchTemplateKey`: unique identifier for the saved watchlist template.
- `UserKey`: unique identifier for the user who created the watchlist.
- `SearchTemplateTypeKey`: internal upstream template type. Captured trade-filter watchlists use `0`; omit it from normal analysis.
- `Name`: watchlist display name.
- `Tickers`: comma-delimited ticker filter. Empty means the watchlist applies to all tickers.
- `SortOrder`: internal upstream UI ordering field. Omit it from normal analysis.
- `MinVolume`, `MaxVolume`: minimum and maximum share-volume bounds for a trade to be included.
- `MinDollars`, `MaxDollars`: minimum and maximum dollar-value bounds for a trade to be included.
- `MinPrice`, `MaxPrice`: minimum and maximum trade-price bounds for a trade to be included.
- `RSIOverboughtHourly`, `RSIOverboughtDaily`, `RSIOversoldHourly`, `RSIOversoldDaily`: RSI thresholds when configured. `null` means the condition is ignored.
- `Conditions`: comma-delimited RSI condition flags. `IgnoreOBD`, `IgnoreOBH`, `IgnoreOSD`, and `IgnoreOSH` mean the corresponding daily or hourly RSI condition is ignored. Required-condition examples include `OBD` for daily overbought and `OBH` for hourly overbought.
- `RSIOverboughtHourlySelected`, `RSIOverboughtDailySelected`, `RSIOversoldHourlySelected`, `RSIOversoldDailySelected`: non-null when the user set the matching RSI threshold in the UI.
- `MinRelativeSize`: minimum relative size for a trade to be included. Relative size is the ratio of the trade's volume to average volume for that ticker, so `5` means at least 5 times average volume.
- `MinRelativeSizeSelected`: non-null when the user set a minimum relative-size threshold in the UI.
- `MaxTradeRank`: maximum all-time rank to include. `10` means the trade must rank 1 through 10. `-1` means no rank limit.
- `MaxTradeRankSelected`: non-null when the user set a maximum trade-rank threshold in the UI.
- `SecurityTypeKey`: security type filter. `-1` means all security types, `1` means stocks only, `26` means ETFs only, and `4` means REITs only.
- `SecurityType`: non-null when the user selected a specific security type.
- `SectorIndustry`: comma-delimited sector or industry filter, or `null` when not configured.
- `MinVCD`: internal upstream field. Omit it from normal analysis.
- `NormalPrints`, `SignaturePrints`, `LatePrints`, `TimelyPrints`, `DarkPools`, `LitExchanges`, `Sweeps`, `Blocks`, `PremarketTrades`, `RTHTrades`, `AHTrades`, `OpeningTrades`, `ClosingTrades`, `PhantomTrades`, and `OffsettingTrades`: booleans for which print, venue, session, auction, or special trade categories are included.
- `IncludedTradeTypes`: CLI-derived array of enabled trade-type booleans from the fields above. Array output uses `null` when none are enabled. Object output omits it when none are enabled.
- `NormalPrintsSelected`, `SignaturePrintsSelected`, `LatePrintsSelected`, `TimelyPrintsSelected`, `DarkPoolsSelected`, `LitExchangesSelected`, `SweepsSelected`, `BlocksSelected`, `PremarketTradesSelected`, `RTHTradesSelected`, `AHTradesSelected`, `OpeningTradesSelected`, `ClosingTradesSelected`, `PhantomTradesSelected`, and `OffsettingTradesSelected`: upstream UI state indicating whether the user explicitly selected the matching inclusion flag.
- `APIKey`: internal upstream field. Omit it from normal analysis.

The `watchlists` command's `expanded` preset includes annotated non-internal filter fields and the derived `IncludedTradeTypes` helper. It intentionally excludes `SearchTemplateTypeKey`, `SortOrder`, `MinVCD`, and `APIKey`; use `--preset-fields full` only when raw upstream payloads are needed for debugging.
