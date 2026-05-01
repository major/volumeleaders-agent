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
  "fields": ["Ticker", "FullTimeString24", "Price", "Dollars", "DollarsMultiplier", "Volume", "TradeRank", "DarkPool", "Sweep", "LatePrint", "SignaturePrint", "Sector"],
  "rows": [
    ["KRE", "17:47:49", 59.12, 17501965.25, 5.019755999966191, 296050, 16, false, false, false, true, "Financials"]
  ]
}
```

All top-level trade commands support `--limit`, `--fields`, `--preset-fields core|signals|full`, `--shape array|objects`, and `--pretty`. Use `--preset-fields signals` for a richer LLM-friendly projection, `--fields Ticker,Dollars,TradeRank` for a custom projection, `--shape objects` when repeated key names are acceptable, or `--preset-fields full` to return raw upstream `trades` objects. Structcli features are available from the scaffold:

```bash
volumeleaders-agent --jsonschema=tree  # Full command schema for agents
volumeleaders-agent env-vars           # Environment variable reference
volumeleaders-agent config-keys        # Config key reference
volumeleaders-agent --mcp              # Run stdio MCP server
```

The date flag can also be set with the environment variable shown by `volumeleaders-agent env-vars`, for example `VOLUMELEADERS_AGENT_TRADES_DATE`.

## LLM field guide for trade filters and signal fields

These names come from VolumeLeaders' browser forms and JSON responses, so some are terse UI labels rather than plain English API names. The same field guide is also embedded in the CLI command metadata so structcli JSON schema discovery and MCP callers can see it without reading this README:

- `RelativeSize` is a request filter for minimum relative size. Captured browser values are `0`, `5`, `10`, `25`, `50`, and `100`, where `0` means any size and the others mean at least that many times the ticker's average dollar trade size.
- `DollarsMultiplier`, shown as `RS` in the UI, is the returned relative size value: trade dollars divided by average dollars for that ticker. VolumeLeaders highlights trades at or above `25x` average size.
- `CumulativeDistribution`, shown as `PCT` in the UI, is the trade's percentile rank relative to other trades for the same ticker.
- `Conditions` carries RSI condition filters. `OBD` means overbought daily, `OBH` means overbought hourly, `OSD` means oversold daily, and `OSH` means oversold hourly. Captured defaults use `-1` for no RSI condition filter. Code presets may also contain `IgnoreOBD`, `IgnoreOBH`, `IgnoreOSD`, and `IgnoreOSH`; treat those as “do not consider this RSI condition” values rather than “exclude matching rows.”
- `VCD` appears to carry the minimum `CumulativeDistribution` percentile. Captures use `0` for no percentile filter and `99` for the 99th percentile or above.

## RSI overbought and oversold trades

```bash
volumeleaders-agent overbought --date 2026-04-30
volumeleaders-agent oversold --date 2026-04-30
volumeleaders-agent overbought --date 2026-04-30 --limit 10
volumeleaders-agent oversold --date 2026-04-30 --tickers AAPL,MSFT
```

The `overbought` and `oversold` commands replay VolumeLeaders RSI-condition searches captured from the browser. `overbought` sends `Conditions=OBD,OBH,` with preset `84`, which requires daily or hourly overbought RSI matches. `oversold` sends `Conditions=OSD,OSH` with preset `85`, which requires daily or hourly oversold RSI matches. Both commands use the same compact trade output shape and flags as `trades`, including `--limit`, `--fields`, `--preset-fields core|signals|full`, `--shape array|objects`, and `--pretty`.

Cluster equivalents are available as `overbought-clusters` and `oversold-clusters`. They send the same RSI filters to `TradeClusters/GetTradeClusters`, use `TradeClusterRank=100`, and support the cluster output flags described below.

Use `--preset-fields signals` when an LLM needs the raw signal columns behind these filters, especially `RSIHour`, `RSIDay`, `CumulativeDistribution`, and `DollarsMultiplier`. The date flags can also be set with `VOLUMELEADERS_AGENT_OVERBOUGHT_DATE`, `VOLUMELEADERS_AGENT_OVERBOUGHT_CLUSTERS_DATE`, `VOLUMELEADERS_AGENT_OVERSOLD_DATE`, and `VOLUMELEADERS_AGENT_OVERSOLD_CLUSTERS_DATE`.


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
  "fields": ["Ticker", "MinFullTimeString24", "MaxFullTimeString24", "Price", "Dollars", "DollarsMultiplier", "Volume", "TradeCount", "TradeClusterRank", "Sector"],
  "rows": [
    ["AAPL", "10:01:04", "10:01:08", 203.25, 1250000, 4.2, 6150, 7, 14, "Technology"]
  ]
}
```

The date flag can also be set with `VOLUMELEADERS_AGENT_TRADE_CLUSTERS_DATE`.

## All-time ranked trade clusters

```bash
volumeleaders-agent top10-clusters --date 2026-04-30
volumeleaders-agent top100-clusters --date 2026-04-30 --limit 25
volumeleaders-agent top10-clusters --date 2026-04-30 --tickers AAPL,MSFT
```

The `top10-clusters` and `top100-clusters` commands query `TradeClusters/GetTradeClusters` and return cluster rows for VolumeLeaders' all-time cluster rank filters (`TradeClusterRank=10` or `100`). Phantom and offsetting are trade-only presets and do not have cluster commands.

These commands support the same cluster output flags as `trade-clusters`, including `--limit`, `--fields`, `--preset-fields core|signals|full`, `--shape array|objects`, and `--pretty`. Each command also exposes a matching date environment variable, for example `VOLUMELEADERS_AGENT_TOP10_CLUSTERS_DATE`.

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
  "fields": ["Ticker", "FullTimeString24", "Price", "Dollars", "DollarsMultiplier", "Volume", "TradeRank", "DarkPool", "Sweep", "LatePrint", "SignaturePrint", "Sector"],
  "rows": [
    ["SNDQ", "09:54:09", 28.07, 15623499.12, 29.4, 556520, 1, true, false, false, true, "ETF"]
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
- `top100-dark-pool-20x`: `TradeRank=100`, `DarkPools=1`, and `RelativeSize=20`.
- `top100-leveraged-etfs`: `TradeRank=100` and `SectorIndustry="X B"` for leveraged ETFs.
- `top100-dark-pool-sweeps`: `TradeRank=100`, `DarkPools=1`, `Sweeps=1`, and captured session filters that include premarket and regular-hours prints while excluding after-hours, opening, closing, and phantom prints.

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
  "fields": ["Ticker", "FullTimeString24", "Price", "Dollars", "DollarsMultiplier", "Volume", "TradeRank", "DarkPool", "Sweep", "LatePrint", "SignaturePrint", "Sector"],
  "rows": [
    ["PLTR", "15:59:58", 112.47, 1739337.39, 6.7, 15465, 54, true, false, false, false, "Technology"]
  ]
}
```

## Auth package

```go
import "github.com/major/volumeleaders-agent/internal/auth"
```

## License

See [LICENSE](LICENSE) for details.
