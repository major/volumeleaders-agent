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

## Bullish and bearish leveraged ETF trades

```bash
volumeleaders-agent bull-leverage --date 2026-04-30
volumeleaders-agent bear-leverage --date 2026-04-30
volumeleaders-agent bull-leverage --date 2026-04-30 --limit 5
volumeleaders-agent bull-leverage --date 2026-04-30 --tickers TQQQ
```

The `bull-leverage` and `bear-leverage` commands fetch one day of leveraged ETF trades from VolumeLeaders. They use the same `Trades/GetTrades` auth and response handling as `trades`, but apply the upstream bullish (`X Bull`) or bearish (`X Bear`) leveraged ETF preset captured from VolumeLeaders.

Both commands return the same token-efficient trade output shape by default:

```json
{
  "status": "ok",
  "date": "2026-04-30",
  "recordsTotal": 8,
  "recordsFiltered": 8,
  "fields": ["Ticker", "FullTimeString24", "Price", "Dollars", "DollarsMultiplier", "Volume", "TradeRank", "DarkPool", "Sweep", "LatePrint", "SignaturePrint", "Sector"],
  "rows": [
    ["TQQQ", "10:08:39", 75.91, 14846502.24, 11.96, 195579, 22, true, false, false, true, "3x Bull Nasdaq"]
  ]
}
```

## Bonds and biotechnology stock trades

```bash
volumeleaders-agent bonds --date 2026-04-30
volumeleaders-agent bonds --date 2026-04-30 --tickers HYG,TLT
volumeleaders-agent biotech --date 2026-04-30
volumeleaders-agent biotech --date 2026-04-30 --tickers IBB,XBI
```

The `bonds` and `biotech` commands fetch one day of sector-filtered trades from VolumeLeaders. They use the same `Trades/GetTrades` auth, pagination, and output handling as `trades`, but apply the upstream bonds (`SectorIndustry=Bonds`, preset `90`) or biotechnology (`SectorIndustry=Biotech`, preset `89`) presets captured from VolumeLeaders.

Both commands return the same token-efficient trade output shape by default:

```json
{
  "status": "ok",
  "date": "2026-04-30",
  "recordsTotal": 70,
  "recordsFiltered": 70,
  "fields": ["Ticker", "FullTimeString24", "Price", "Dollars", "DollarsMultiplier", "Volume", "TradeRank", "DarkPool", "Sweep", "LatePrint", "SignaturePrint", "Sector"],
  "rows": [
    ["USHY", "16:38:31", 37.23, 44115018.36, 8.004659259542926, 1184932, 9999, true, false, false, false, "Bonds"]
  ]
}
```

## Auth package

```go
import "github.com/major/volumeleaders-agent/internal/auth"
```

## License

See [LICENSE](LICENSE) for details.
