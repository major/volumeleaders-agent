# volumeleaders-agent

[![CI](https://github.com/major/volumeleaders-agent/actions/workflows/ci.yml/badge.svg)](https://github.com/major/volumeleaders-agent/actions/workflows/ci.yml)
[![CodeQL](https://github.com/major/volumeleaders-agent/actions/workflows/codeql.yml/badge.svg)](https://github.com/major/volumeleaders-agent/actions/workflows/codeql.yml)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/major/volumeleaders-agent/badge)](https://scorecard.dev/viewer/?uri=github.com/major/volumeleaders-agent)
[![codecov](https://codecov.io/gh/major/volumeleaders-agent/graph/badge.svg)](https://codecov.io/gh/major/volumeleaders-agent)
[![Go Report Card](https://goreportcard.com/badge/github.com/major/volumeleaders-agent)](https://goreportcard.com/report/github.com/major/volumeleaders-agent)
[![Go Reference](https://pkg.go.dev/badge/github.com/major/volumeleaders-agent.svg)](https://pkg.go.dev/github.com/major/volumeleaders-agent)

Go CLI for [VolumeLeaders](https://www.volumeleaders.com) market intelligence workflows. The new command surface uses [structcli](https://github.com/leodido/structcli) so commands are friendly to humans, LLM agents, JSON schema discovery, and MCP tool execution.

## Prerequisites

You must be logged into volumeleaders.com in a supported browser (Chrome, Firefox, Edge, etc.). The auth package extracts session cookies directly from the browser's cookie store, so no API keys or manual token management is needed.

## Build

```bash
make build      # Build binary
make test       # Run tests
make lint       # Run golangci-lint
```

## Disproportionately large trades

```bash
volumeleaders-agent trades --date 2026-04-30
volumeleaders-agent trades --date 2026-04-30 --limit 10
volumeleaders-agent trades --date 2026-04-30 --tickers AAPL,IONQ
volumeleaders-agent trades --date 2026-04-30 --tickers AAPL,IONQ,AMZN
```

The `trades` command fetches VolumeLeaders' default Disproportionately large trades preset for one trading day. The upstream filter is intended for a single day of data, so the CLI exposes only `--date` and sends the same value as both `StartDate` and `EndDate` in the `Trades/GetTrades` request. Add `--tickers` to filter the preset to one ticker or a comma-delimited ticker list without spaces.

Trade commands return compact, stable JSON with the requested date, DataTables record counts, field names, and array rows by default. The default `--preset-fields core --shape array` output is optimized for token-efficient LLM workflows by emitting field names once and then returning each trade as an array in that order. Use `--limit` to control the total number of returned rows. Limits must be between 1 and 100, and omitted limits default to 100 for broad trade presets so the CLI never requests oversized result sets from VolumeLeaders. Use `--pretty` when reading the JSON directly. JSON examples in this README are formatted for readability:

```json
{
  "status": "ok",
  "date": "2026-04-30",
  "recordsTotal": 1492,
  "recordsFiltered": 1492,
  "fields": ["Ticker", "TradeCount", "FullTimeString24", "ClosePrice", "Price", "Sector", "Industry", "Volume", "Dollars", "DollarsMultiplier", "CumulativeDistribution", "TradeRank", "LastComparibleTradeDate", "CalendarEvent", "AuctionTrade"],
  "rows": [
    ["KRE", 2, "17:47:49", 0.0, 69.85, "Financial Services", "Banks", 250565, 17501965.25, 5.01912515023852678249, 0.9673, 9999, "/Date(1777334400000)/", "OPEX", "open"]
  ]
}
```

All top-level trade commands support `--limit`, `--fields`, `--preset-fields core|expanded|full`, `--shape array|objects`, and `--pretty`. Use `--fields Ticker,Dollars,TradeRank` for a custom projection, `--preset-fields expanded` when an LLM needs all annotated non-internal signal fields, `--shape objects` when repeated key names are acceptable, or `--preset-fields full` to return raw upstream `trades` objects. Structcli features are available from the scaffold:

```bash
volumeleaders-agent --jsonschema=tree  # Full command schema for agents
volumeleaders-agent env-vars           # Environment variable reference
volumeleaders-agent config-keys        # Config key reference
volumeleaders-agent --mcp              # Run stdio MCP server
```

The date flag can also be set with the environment variable shown by `volumeleaders-agent env-vars`, for example `VOLUMELEADERS_AGENT_TRADES_DATE`.

## Trade levels

```bash
volumeleaders-agent trade-levels --ticker BAND
volumeleaders-agent trade-levels --ticker BAND --start-date 2025-05-01 --end-date 2026-05-01
volumeleaders-agent trade-levels --ticker SPY --trade-level-count 10 --min-dollars 500000 --relative-size 3
```

The `trade-levels` command fetches VolumeLeaders price levels for one ticker through `TradeLevels/GetTradeLevels`. A level groups that ticker's large-trade history by price across the requested date range, then returns aggregate dollars, shares, trade count, relative size, percentile, rank, and the upstream level date range. Omit `--start-date` and `--end-date` to use a one-year window ending today, or pass both dates explicitly when you need reproducible output.

Trade levels use the same compact output controls as trade commands: `--fields`, `--preset-fields core|expanded|full`, `--shape array|objects`, and `--pretty`. The default core preset focuses on the visible table values:

```json
{
  "status": "ok",
  "ticker": "BAND",
  "startDate": "2025-05-01",
  "endDate": "2026-05-01",
  "recordsTotal": 60,
  "recordsFiltered": 60,
  "fields": ["Ticker", "Price", "Dollars", "Volume", "Trades", "RelativeSize", "CumulativeDistribution", "TradeLevelRank", "Dates"],
  "rows": [
    ["BAND", 16.9, 25076379.52, 1485429, 17, 3.84, 0.9586, 44, "2022-07-27 - 2026-04-13"]
  ]
}
```

Captured filters include `--min-dollars`, `--max-dollars`, `--min-volume`, `--max-volume`, `--min-price`, `--max-price`, `--vcd`, `--relative-size`, `--trade-level-rank`, and `--trade-level-count`. Use `--preset-fields expanded` when an LLM needs company context or upstream timestamps. Use `--preset-fields full` only for raw debugging.

## Trade level touches

```bash
volumeleaders-agent trade-level-touches
volumeleaders-agent trade-level-touches --start-date 2026-04-24 --end-date 2026-05-01
volumeleaders-agent trade-level-touches --ticker MSFT --start-date 2025-05-01 --end-date 2026-05-01 --limit 25
```

The `trade-level-touches` command fetches the `TradeLevelTouches/GetTradeLevelTouches` table. A trade level touch records when price comes back from above or below to test a ranked large-trade level, which is useful for support and resistance workflows. VolumeLeaders enforces a 7-day maximum when querying all tickers or multiple comma-delimited tickers, so the command defaults those scans to the last 7 days and rejects wider ranges. Exactly one ticker can use a wider range; when `--tickers` contains one symbol and `--start-date` is omitted, the command defaults to a one-year lookback.

The captured browser default is `TradeLevelRank=10`, meaning ranks 1 through 10. Output rows are emitted under `levelTouches` for `--shape objects` or `--preset-fields full`; default array output uses the visible touch table fields:

```json
{
  "status": "ok",
  "startDate": "2026-04-24",
  "endDate": "2026-05-01",
  "recordsTotal": 11,
  "recordsFiltered": 11,
  "fields": ["Ticker", "FullDateTime", "Price", "Dollars", "Volume", "Trades", "RelativeSize", "CumulativeDistribution", "TradeLevelRank", "Dates", "Sector", "Industry"],
  "rows": [
    ["MSFT", "2026-05-01 : 10:59:00", 413.6, 13744031091.14, 33228468, 117, 13.8, 0.9926, 10, "2024-02-29 - 2026-02-09", "Technology", "Software"]
  ]
}
```

Captured filters include `--min-dollars`, `--max-dollars`, `--min-volume`, `--max-volume`, `--min-price`, `--max-price`, `--vcd`, `--relative-size`, `--trade-level-rank`, and `--sector-industry`. The date and filter flags can also be set with environment variables such as `VOLUMELEADERS_AGENT_TRADE_LEVEL_TOUCHES_START_DATE`, `VOLUMELEADERS_AGENT_TRADE_LEVEL_TOUCHES_END_DATE`, and `VOLUMELEADERS_AGENT_TRADE_LEVEL_TOUCHES_TICKERS`.

## Watchlists

```bash
volumeleaders-agent watchlists
volumeleaders-agent watchlists --search-template-key 4952 --preset-fields expanded
volumeleaders-agent watchlists --preset-fields expanded --pretty
volumeleaders-agent watchlists --fields Name,Tickers,MaxTradeRank,IncludedTradeTypes
volumeleaders-agent save-watchlist --name "Big dark-pool sweeps" --min-dollars 10000000 --min-relative-size 5
volumeleaders-agent save-watchlist --search-template-key 4952 --name BigOnes --tickers AAPL,MSFT --min-dollars 10000000 --max-trade-rank 10
volumeleaders-agent delete-watchlist --search-template-key 4952
```

The `watchlists` command lists saved watchlist filters configured in the authenticated VolumeLeaders account. These watchlists are saved criteria for trades or clusters, so they may filter by ticker, dollar size, price, relative size, rank, RSI conditions, venue, print type, session, auction status, phantom prints, and offsetting trades.

The `save-watchlist` command creates or fully replaces those saved filters through the same authenticated `WatchListConfig` browser form. Omit `--search-template-key` or set it to `0` to create a new watchlist. Set `--search-template-key` to an existing `SearchTemplateKey` from `volumeleaders-agent watchlists` to replace that watchlist with the complete criteria passed to the command. Browser defaults include all print/session/venue checkboxes, all security types, no rank limit, no RSI condition filters, and no ticker restriction unless you pass the matching flags.

The `delete-watchlist` command removes a saved filter by posting the browser-compatible `WatchListConfigs/DeleteWatchList` request with the selected `SearchTemplateKey`.

The default `--preset-fields summary --shape array` output is compact JSON for picking a specific watchlist before requesting details:

```json
{
  "status": "ok",
  "count": 1,
  "fields": ["SearchTemplateKey", "Name"],
  "rows": [[4952, "BigOnes"]]
}
```

Use `--search-template-key` with `--preset-fields expanded` to inspect one saved watchlist's configuration, `--preset-fields full` to return raw upstream objects, `--shape objects` for repeated key names, or `--pretty` for readable JSON. The output flags can also be set with environment variables shown by `volumeleaders-agent env-vars`, for example `VOLUMELEADERS_AGENT_WATCHLISTS_PRESET_FIELDS`. See [`docs/fields.md`](docs/fields.md#watchlist-fields) for the longer field reference derived from `Getwatchlists.jsonc`, including selected-field semantics, security type codes, and internal-only fields to ignore.

## LLM field guide for trade filters and signal fields

These names come from VolumeLeaders' browser forms and JSON responses, so some are terse UI labels rather than plain English API names. For trades, users and LLM callers should focus on the fields that appear in the VolumeLeaders table: time, ticker/count, CP, TP, sector, industry, Sh, $$, RS, PCT, R, and Last. `Ticker` is the stock ticker symbol, such as `TSLA` or `AMZN`. `TradeCount` is the `#T` count shown beside the ticker: the number of large trades for that ticker today, so `KRE (2)` means two large KRE trades today. Raw response fields outside that visible table are secondary debugging or correlation context unless this guide says otherwise. The same field guide is also embedded in the CLI command metadata so structcli JSON schema discovery and MCP callers can see it without reading this README:

- `RelativeSize` is a request filter for minimum relative size. Captured browser values are `0`, `5`, `10`, `25`, `50`, and `100`, where `0` means any size and the others mean at least that many times the ticker's average dollar trade size.
- `DarkPools` and `Sweeps` are request filters. Dark pool trades are done off exchange and reported later; lit exchange trades are done on exchange and reported immediately. Sweeps are orders spread across multiple exchanges to get done quickly; blocks are orders sent to one exchange. For the `trades` command, `--dark-pools=false --sweeps=false` shows everything, `--dark-pools` shows dark pools of all kinds, `--sweeps` shows sweeps from dark pools or lit exchanges, and both flags together show dark pool sweeps only. In raw output, `DarkPool` and `Sweep` describe the classification of each returned row.
- `DollarsMultiplier`, shown as `RS` in the UI, is the returned relative size value: trade dollars divided by average dollars for that ticker. VolumeLeaders highlights trades at or above `25x` average size.
- `CumulativeDistribution`, shown as `PCT` in the UI, is the trade's percentile rank relative to other trades for the same ticker.
- `Conditions` carries RSI condition filters. `OBD` means overbought daily, `OBH` means overbought hourly, `OSD` means oversold daily, and `OSH` means oversold hourly. Captured defaults use `-1` for no RSI condition filter. Code presets may also contain `IgnoreOBD`, `IgnoreOBH`, `IgnoreOSD`, and `IgnoreOSH`; treat those as “do not consider this RSI condition” values rather than “exclude matching rows.”
- `VCD` appears to carry the minimum `CumulativeDistribution` percentile. Captures use `0` for no percentile filter and `99` for the 99th percentile or above.
- `TradeID`, `SequenceNumber`, and `SecurityKey` are VolumeLeaders internal identifiers. `DateKey` and `TimeKey` are compact internal date/time keys. Treat these five fields as upstream metadata for correlation or debugging, not as trading-decision signals.
- `Date` is the trade date. `FullDateTime` is the full trade timestamp. `StartDate` and `EndDate` appear to be upstream query-range echoes or internal metadata rather than separate trade signals. `LastComparibleTradeDate` uses the upstream spelling and means the last date VolumeLeaders saw a trade close to this trade's size.
- `CalendarEvent` is a compact derived core field. It contains the true upstream calendar markers joined with commas, or `null` in array output when no marker is true. Source markers are `EOM` for end of month, `EOQ` for end of quarter, `EOY` for end of year, `OPEX` for a market options expiration date, and `VOLEX` for a market volatility expiration date such as VIX options expiration. In object output, `CalendarEvent` is omitted when no marker is true.
- `AuctionTrade` is a compact derived core field from upstream `OpeningTrade` and `ClosingTrade` `0` or `1` flags. It is `"open"` when the trade hit in the market opening auction, `"close"` when it hit in the market-on-close auction, and `null` in array output when neither flag is true. In object output, `AuctionTrade` is omitted when neither flag is true. In full output, the raw upstream `OpeningTrade` and `ClosingTrade` values are preserved.
- `--preset-fields expanded` keeps token-efficient projection while adding annotated non-internal fields. Trade expanded fields cover IDs/timestamps, ticker identity, price/size, comparable dates, rank snapshots, print classifications, RSI values, frequency counts, cancellation state, calendar flags, and auction flags. Cluster expanded fields cover the cluster time range, ticker identity, price/size, cluster count/rank, comparable cluster date, IPO date, distribution, calendar flags, and inside-bar flags. Use `full` only when raw upstream internals or always-empty fields are needed for debugging.
- See [`docs/fields.md`](docs/fields.md) for the longer trade and cluster field reference derived from the annotated JSON examples.
- `Ask` and `Bid` are the ask and bid prices in the bid/ask spread when the trade happened. `ClosePrice` is `CP` in the UI: the close price at the end of the day, or the current price if the market is still open. `Price` is `TP` in the UI: the trade price when the large trade hit. `AverageDailyVolume` is a moving-average measure of the stock's normal volume, and `PercentDailyVolume` compares today's volume with that moving average.
- `Volume`, shown as `Sh` in the UI, is how many shares were in the trade. `Dollars`, shown as `$$` in the UI, is how big the trade was in dollars: number of shares times the trade price.
- `TradeRank` is the trade's current rank among all current trades and can change when larger trades arrive. In the UI `R` column, a dash means the trade is not ranked in the top 100 trades, while a number such as `27` means the trade is currently ranked 27th. `TradeRankSnapshot` is immutable: it preserves how the trade ranked at the time it appeared.
- `TotalVolume` and `TotalDollars` are internal upstream values. Do not treat them as standalone trading-decision signals.

## RSI overbought and oversold trades

```bash
volumeleaders-agent overbought --date 2026-04-30
volumeleaders-agent oversold --date 2026-04-30
volumeleaders-agent overbought --date 2026-04-30 --limit 10
volumeleaders-agent oversold --date 2026-04-30 --tickers AAPL,MSFT
```

The `overbought` and `oversold` commands replay VolumeLeaders RSI-condition searches captured from the browser. `overbought` sends `Conditions=OBD,OBH,` with preset `84`, which requires daily or hourly overbought RSI matches. `oversold` sends `Conditions=OSD,OSH` with preset `85`, which requires daily or hourly oversold RSI matches. Both commands use the same compact trade output shape and flags as `trades`, including `--limit`, `--fields`, `--preset-fields core|expanded|full`, `--shape array|objects`, and `--pretty`.

Cluster equivalents are available as `overbought-clusters` and `oversold-clusters`. They send the same RSI filters to `TradeClusters/GetTradeClusters`, use `TradeClusterRank=100`, and support the cluster output flags described below.

Use `--fields` when an LLM needs specific raw signal columns behind these filters, such as `RSIHour`, `RSIDay`, `CumulativeDistribution`, and `DollarsMultiplier`, `--preset-fields expanded` when it needs all annotated signal fields, or `--preset-fields full` when the full upstream object is acceptable. The date flags can also be set with `VOLUMELEADERS_AGENT_OVERBOUGHT_DATE`, `VOLUMELEADERS_AGENT_OVERBOUGHT_CLUSTERS_DATE`, `VOLUMELEADERS_AGENT_OVERSOLD_DATE`, and `VOLUMELEADERS_AGENT_OVERSOLD_CLUSTERS_DATE`.


## Disproportionately large trade clusters

```bash
volumeleaders-agent trade-clusters --date 2026-04-30
volumeleaders-agent trade-clusters --date 2026-04-30 --limit 10
volumeleaders-agent trade-clusters --date 2026-04-30 --tickers AAPL,IONQ
```

The `trade-clusters` command fetches VolumeLeaders' disproportionately large trade clusters preset for one trading day. A trade cluster is a group of smaller trades that occur close together in time and add up to a larger dollar-volume event, so the command uses `TradeClusters/GetTradeClusters` rather than the single-trade `Trades/GetTrades` endpoint. The request mirrors the browser's cluster form, including the `TradeClusterRank`, relative size, price, dollar, and sector filters captured from the VolumeLeaders UI.

Cluster commands return the same compact JSON envelope as trade commands, but cluster rows are emitted under `clusters` when using `--shape objects` or `--preset-fields full`. The default core fields include cluster-specific values such as `MinFullTimeString24`, `MaxFullTimeString24`, `TradeCount`, and `TradeClusterRank`:

```json
{
  "status": "ok",
  "date": "2026-04-30",
  "recordsTotal": 3213,
  "recordsFiltered": 3213,
  "fields": ["Ticker", "MinFullTimeString24", "MaxFullTimeString24", "Price", "Dollars", "DollarsMultiplier", "Volume", "TradeCount", "TradeClusterRank", "Sector", "CalendarEvent", "AuctionTrade"],
  "rows": [
    ["AAPL", "10:01:04", "10:01:08", 203.25, 1250000, 4.2, 6150, 7, 14, "Technology", "EOM,VOLEX", "close"]
  ]
}
```

The date flag can also be set with `VOLUMELEADERS_AGENT_TRADE_CLUSTERS_DATE`.

## Trade cluster bombs

```bash
volumeleaders-agent trade-cluster-bombs
volumeleaders-agent trade-cluster-bombs --start-date 2026-04-24 --end-date 2026-05-01
volumeleaders-agent trade-cluster-bombs --tickers AAPL,AMZN --limit 25
```

The `trade-cluster-bombs` command fetches the `TradeClusterBombs/GetTradeClusterBombs` table. VolumeLeaders defines these rare stock-only events as at least three dark-pool sweeps in one day with at least $38M combined value. VolumeLeaders enforces a 7-day maximum when querying all tickers or a comma-delimited list of multiple tickers, so the command defaults those scans to the last 7 days and rejects wider ranges before posting the browser form. Exactly one ticker can use a wider range; when `--tickers` contains one symbol and `--start-date` is omitted, the command defaults to a one-year lookback because longer windows are more likely to time out.

Cluster bomb output uses a date-range envelope with rows emitted under `clusterBombs` for `--shape objects` or `--preset-fields full`. The default core fields focus on the browser table and cluster-bomb-specific rank/comparable-date fields:

```json
{
  "status": "ok",
  "startDate": "2026-04-24",
  "endDate": "2026-05-01",
  "recordsTotal": 20,
  "recordsFiltered": 20,
  "fields": ["Ticker", "MinFullTimeString24", "MaxFullTimeString24", "Dollars", "DollarsMultiplier", "Volume", "TradeCount", "TradeClusterBombRank", "Sector", "Industry", "LastComparableTradeClusterBombDate", "CalendarEvent"],
  "rows": [
    ["AAPL", "10:01:04", "10:01:08", 125000000, 42.5, 700000, 7, 14, "Technology", "Consumer Electronics", "/Date(1777420800000)/", "EOM,VOLEX"]
  ]
}
```

The date and filter flags can also be set with environment variables such as `VOLUMELEADERS_AGENT_TRADE_CLUSTER_BOMBS_START_DATE`, `VOLUMELEADERS_AGENT_TRADE_CLUSTER_BOMBS_END_DATE`, and `VOLUMELEADERS_AGENT_TRADE_CLUSTER_BOMBS_TICKERS`.

## All-time ranked trade clusters

```bash
volumeleaders-agent top10-clusters --date 2026-04-30
volumeleaders-agent top100-clusters --date 2026-04-30 --limit 25
volumeleaders-agent top10-clusters --date 2026-04-30 --tickers AAPL,MSFT
```

The `top10-clusters` and `top100-clusters` commands query `TradeClusters/GetTradeClusters` and return cluster rows for VolumeLeaders' all-time cluster rank filters (`TradeClusterRank=10` or `100`). Phantom and offsetting are trade-only presets and do not have cluster commands.

These commands support the same cluster output flags as `trade-clusters`, including `--limit`, `--fields`, `--preset-fields core|expanded|full`, `--shape array|objects`, and `--pretty`. Each command also exposes a matching date environment variable, for example `VOLUMELEADERS_AGENT_TOP10_CLUSTERS_DATE`.

## All-time ranked trades

```bash
volumeleaders-agent top10 --date 2026-04-30
volumeleaders-agent top100 --date 2026-04-30
volumeleaders-agent top100 --date 2026-04-30 --limit 25
volumeleaders-agent top10 --date 2026-04-30 --tickers AAPL,MSFT
```

The `top10` and `top100` commands fetch trades from one trading day where each trade ranks in the stock's all-time largest single trades. A `TradeRank` of `1` is the biggest single trade VolumeLeaders has recorded for that stock, while `10` means the tenth biggest. Both commands use the same `Trades/GetTrades` auth and response handling as `trades`, but they apply the ranked-trade presets captured from VolumeLeaders.

The ranked commands return the same token-efficient trade output shape as `trades`, with an added `rankLimit` value. Use `--limit` to override the command preset row count, up to the same hard maximum of 100 rows:

```json
{
  "status": "ok",
  "date": "2026-04-30",
  "rankLimit": 10,
  "recordsTotal": 76,
  "recordsFiltered": 76,
  "fields": ["Ticker", "TradeCount", "FullTimeString24", "ClosePrice", "Price", "Sector", "Industry", "Volume", "Dollars", "DollarsMultiplier", "CumulativeDistribution", "TradeRank", "LastComparibleTradeDate", "CalendarEvent", "AuctionTrade"],
  "rows": [
    ["SNDQ", 1, "09:54:09", 28.55, 28.07, "ETF", "ETF", 556520, 15623499.12, 29.4, 0.99, 1, "/Date(1777334400000)/", null, null]
  ]
}
```

## HAR-derived ranked filters

```bash
volumeleaders-agent top30-10x-99pct --date 2026-04-30
volumeleaders-agent top100-dark-pool-20x --date 2026-04-30
volumeleaders-agent top100-leveraged-etfs --date 2026-04-30
volumeleaders-agent top100-dark-pool-sweeps --date 2026-04-30
```

These commands replay additional `Trades/GetTrades` filters captured from browser HAR files:

- `top30-10x-99pct`: `TradeRank=30`, `RelativeSize=10`, and `VCD=99` for trades in the 99th percentile or above.
- `top100-dark-pool-20x`: `TradeRank=100`, `DarkPools=1`, and `RelativeSize=20`. It does not filter on `Sweeps`, so it includes dark-pool blocks and dark-pool sweeps.
- `top100-leveraged-etfs`: `TradeRank=100` and `SectorIndustry="X B"` for leveraged ETFs.
- `top100-dark-pool-sweeps`: `TradeRank=100`, `DarkPools=1`, `Sweeps=1`, and captured session filters that include premarket and regular-hours prints while excluding after-hours, opening, closing, and phantom prints. Because both dark-pool and sweep filters are set, this returns only dark-pool sweeps.

Each filter also has a trade-cluster equivalent: `top30-10x-99pct-clusters`, `top100-dark-pool-20x-clusters`, `top100-leveraged-etfs-clusters`, and `top100-dark-pool-sweeps-clusters`. The cluster commands send the same filters to `TradeClusters/GetTradeClusters` with `TradeClusterRank` matching the trade command's `TradeRank`.

## Phantom and offsetting trades

```bash
volumeleaders-agent phantom --date 2026-04-30
volumeleaders-agent offsetting --date 2026-04-30
volumeleaders-agent offsetting --date 2026-04-30 --limit 10
volumeleaders-agent phantom --date 2026-04-30 --tickers PLTR,NVDA
```

The `phantom` command fetches trades where VolumeLeaders marks the trade price as far from the current price. These prints can hint at where price may move later, but they are not guaranteed signals. The `offsetting` command fetches trades where nearly matching share sizes appear on different dates, which can hint that a trader entered and later exited a position.

Both commands use the same `Trades/GetTrades` auth and response handling as `trades`, and both return the same token-efficient trade output shape by default. Use `--preset-fields full` when you need signal-specific raw fields such as `PhantomPrint` or `OffsettingTradeDate` that are not part of the core default:

```json
{
  "status": "ok",
  "date": "2026-04-30",
  "recordsTotal": 12,
  "recordsFiltered": 12,
  "fields": ["Ticker", "TradeCount", "FullTimeString24", "ClosePrice", "Price", "Sector", "Industry", "Volume", "Dollars", "DollarsMultiplier", "CumulativeDistribution", "TradeRank", "LastComparibleTradeDate"],
  "rows": [
    ["PLTR", 1, "15:59:58", 113.20, 112.47, "Technology", "Software", 15465, 1739337.39, 6.7, 0.98, 54, "/Date(1777334400000)/"]
  ]
}
```

## Auth package

```go
import "github.com/major/volumeleaders-agent/internal/auth"
```

## License

See [LICENSE](LICENSE) for details.
