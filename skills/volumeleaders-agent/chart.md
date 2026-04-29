# chart

Single-ticker price bars, quotes, chart-ready levels, and company metadata.

Chart filter names differ from trade commands: use `--signature-prints` not `--sig-prints`, `--trade-rank-snapshot` not `--rank-snapshot`, and `--include-premarket|rth|ah|opening|closing|phantom|offsetting` not bare session names.

## chart price-data

One-minute OHLCV bars with institutional trade overlays.

Required: ticker flag or positional ticker, plus a complete `--start-date`/`--end-date` range or `--days`. Supports `--format json|csv|tsv` and `--fields` for custom output fields.

Important options: `--volume-profile 0|1`, `--levels 5`, `--trade-count 3`, volume/price/dollar ranges, `--dark-pools`, `--sweeps`, `--late-prints`, `--signature-prints`, `--vcd`, `--trade-rank`, `--trade-rank-snapshot`, include-session flags.

```bash
volumeleaders-agent chart price-data --ticker AAPL --start-date 2026-04-28 --end-date 2026-04-28
volumeleaders-agent chart price-data AAPL --days 2
volumeleaders-agent chart price-data --ticker AAPL --start-date 2026-04-28 --end-date 2026-04-28 --format tsv
volumeleaders-agent chart price-data AAPL --days 1 --fields FullDateTime,ClosePrice,Volume,Dollars
volumeleaders-agent chart price-data AAPL --days 1 --fields all
```

Default output is compact for LLM context windows. It keeps analytical bar and overlay fields: `DateKey`, `TimeKey`, `FullDateTime`, `OpenPrice`, `HighPrice`, `LowPrice`, `ClosePrice`, `Volume`, `Dollars`, `Trades`, `CumulativeDistribution`, `TradeRank`, `TradeRankSnapshot`, `TradeLevelRank`, `DollarsMultiplier`, `RelativeSize`, `DarkPoolTrade`, `LatePrint`, `OpeningTrade`, `ClosingTrade`, `SignaturePrint`, `PhantomPrint`, `Sweep`.

Use `--fields FieldA,FieldB` to select explicit `PriceBar` fields, or `--fields all` to include the full API model. Full-model fields include repeated or verbose values such as `SecurityKey`, `TradeID`, `Ticker`, duplicate date strings, `TradeConditions`, `Dates`, comparison dates, and frequency fields that are omitted by default.

## chart snapshot

Quick quote for one ticker. JSON-only single object.

Required: ticker flag or positional ticker, plus `--date-key`.

```bash
volumeleaders-agent chart snapshot AAPL --date-key 2026-04-28
```

Fields: `ticker`, `lastQuote`, `lastTrade`, `todaysChange`, `todaysChangePerc`. `lastQuote` and `lastTrade` use compact single-letter keys from the source API.

## chart levels

Chart-ready institutional price levels with fewer filters than `trade levels`.

Required: ticker flag or positional ticker, plus a complete `--start-date`/`--end-date` range or `--days`. Optional: `--levels` default `5`, `--format json|csv|tsv`, `--fields`.

```bash
volumeleaders-agent chart levels --ticker AAPL --start-date 2026-01-01 --end-date 2026-04-28 --levels 10
volumeleaders-agent chart levels AAPL --days 30 --levels 10
volumeleaders-agent chart levels AAPL --days 30 --fields Price,Dollars,TradeLevelRank
```

Default fields: `Price`, `Dollars`, `Volume`, `Trades`, `RelativeSize`, `CumulativeDistribution`, `TradeLevelRank`, `Dates`. Use `--fields all` for the full `TradeLevel` model, including repeated `Ticker`/`Name` and separate `MinDate`/`MaxDate` fields.

## chart company

Company metadata, fundamentals, sector/industry, and trading averages. JSON-only single object.

Required: ticker flag or positional ticker. Optional: `--fields`.

```bash
volumeleaders-agent chart company AAPL
volumeleaders-agent chart company AAPL --fields Name,Ticker,MarketCap,CurrentPrice
volumeleaders-agent chart company AAPL --fields all
```

Default output is compact company context: `Name`, `Ticker`, `Sector`, `Industry`, `MarketCap`, `CurrentPrice`, `OptionsEnabled`, `IPODate`, `AverageBlockSizeDollars`, `AverageDailyVolume`, `AverageTradeShares`, `AverageDailyRangePct`, `AverageClusterSizeDollars`, `AverageLevelSizeDollars`, `TotalTrades`, `FirstTradeDate`, `MaxDate`.

Use `--fields all` when you need the full company model, including keys, status flags, long text fields (`Description`, `News`, `Financials`, `Splits`), homepage URL, previous ticker data, and duplicate 30-day/90-day averages.
