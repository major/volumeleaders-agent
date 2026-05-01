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

The command returns compact, stable JSON with the requested date, DataTables record counts, and raw trade objects from VolumeLeaders. Use `--limit 1-100` to reduce the number of returned rows for token-efficient LLM workflows, or `--pretty` when reading the JSON directly. JSON examples in this README are formatted for readability:

```json
{
  "status": "ok",
  "date": "2026-04-30",
  "recordsTotal": 1492,
  "recordsFiltered": 1492,
  "trades": [
    {
      "Ticker": "KRE",
      "FullTimeString24": "17:47:49",
      "Dollars": 17501965.25,
      "DollarsMultiplier": 5.019755999966191
    }
  ]
}
```

All trade commands support `--limit 1-100` and `--pretty`. Structcli features are available from the scaffold:

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

The ranked commands return compact, stable JSON with the requested date, rank limit, DataTables record counts, and raw trade objects. Use `--limit 1-100` to override the command preset row count, or `--pretty` when reading the JSON directly:

```json
{
  "status": "ok",
  "date": "2026-04-30",
  "rankLimit": 10,
  "recordsTotal": 76,
  "recordsFiltered": 76,
  "trades": [
    {
      "Ticker": "SNDQ",
      "TradeRank": 1,
      "Dollars": 15623499.12
    }
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

Both commands use the same `Trades/GetTrades` auth and response handling as `trades`, and both return compact, stable JSON with the requested date, DataTables record counts, and raw trade objects:

```json
{
  "status": "ok",
  "date": "2026-04-30",
  "recordsTotal": 12,
  "recordsFiltered": 12,
  "trades": [
    {
      "Ticker": "PLTR",
      "PhantomPrint": 1,
      "OffsettingTradeDate": "/Date(-2208988800000)/",
      "Dollars": 1739337.39
    }
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

Both commands return compact, stable JSON with the requested date, DataTables record counts, and raw trade objects:

```json
{
  "status": "ok",
  "date": "2026-04-30",
  "recordsTotal": 8,
  "recordsFiltered": 8,
  "trades": [
    {
      "Ticker": "TQQQ",
      "Sector": "3x Bull Nasdaq",
      "Dollars": 14846502.24,
      "DollarsMultiplier": 11.96
    }
  ]
}
```

## Auth package

```go
import "github.com/major/volumeleaders-agent/internal/auth"
```

## License

See [LICENSE](LICENSE) for details.
