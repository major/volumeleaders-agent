# chart

Single-ticker price bars, quotes, chart-ready levels, and company metadata.

Chart filter names differ from trade commands: use `--signature-prints` not `--sig-prints`, `--trade-rank-snapshot` not `--rank-snapshot`, and `--include-premarket|rth|ah|opening|closing|phantom|offsetting` not bare session names.

## chart price-data

One-minute OHLCV bars with institutional trade overlays.

Required: `--ticker`, `--start-date`, `--end-date`. Supports `--format json|csv|tsv`.

Important options: `--volume-profile 0|1`, `--levels 5`, `--trade-count 3`, volume/price/dollar ranges, `--dark-pools`, `--sweeps`, `--late-prints`, `--signature-prints`, `--vcd`, `--trade-rank`, `--trade-rank-snapshot`, include-session flags.

```bash
volumeleaders-agent chart price-data --ticker AAPL --start-date 2026-04-28 --end-date 2026-04-28
volumeleaders-agent chart price-data --ticker AAPL --start-date 2026-04-28 --end-date 2026-04-28 --format tsv
```

Key fields: `DateKey`, `TimeKey`, `Date`, `FullDateTime`, `OpenPrice`, `ClosePrice`, `HighPrice`, `LowPrice`, `Volume`, `Dollars`, `Trades`, `CumulativeDistribution`, `TradeRank`, `TradeRankSnapshot`, `TradeLevelRank`, `DollarsMultiplier`, `RelativeSize`, `DarkPoolTrade`, `LatePrint`, `OpeningTrade`, `ClosingTrade`, `SignaturePrint`, `PhantomPrint`, `Sweep`, frequency fields.

## chart snapshot

Quick quote for one ticker. JSON-only single object.

Required: `--ticker`, `--date-key`.

```bash
volumeleaders-agent chart snapshot --ticker AAPL --date-key 2026-04-28
```

Fields: `ticker`, `lastQuote`, `lastTrade`, `todaysChange`, `todaysChangePerc`. `lastQuote` and `lastTrade` use compact single-letter keys from the source API.

## chart levels

Chart-ready institutional price levels with fewer filters than `trade levels`.

Required: `--ticker`, `--start-date`, `--end-date`. Optional: `--levels` default `5`, `--format json|csv|tsv`.

```bash
volumeleaders-agent chart levels --ticker AAPL --start-date 2026-01-01 --end-date 2026-04-28 --levels 10
```

Fields: same TradeLevel model as `trade levels`.

## chart company

Company metadata, fundamentals, sector/industry, and trading averages. JSON-only single object.

Required: `--ticker`.

```bash
volumeleaders-agent chart company --ticker AAPL
```

Key fields: `SecurityKey`, `Name`, `Ticker`, `Sector`, `Industry`, `Description`, `HomePageURL`, `MarketCap`, `CurrentPrice`, `OptionsEnabled`, `IPODate`, all-time/30-day/90-day average block, volume, trade-share, range, closing-trade, cluster-size, and level-size fields.
