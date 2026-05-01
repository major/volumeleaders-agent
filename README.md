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
volumeleaders-agent trades --date 2026-04-30 --tickers AAPL,IONQ
volumeleaders-agent trades --date 2026-04-30 --tickers AAPL,IONQ,AMZN
```

The `trades` command fetches VolumeLeaders' default Disproportionately large trades preset for one trading day. The upstream filter is intended for a single day of data, so the CLI exposes only `--date` and sends the same value as both `StartDate` and `EndDate` in the `Trades/GetTrades` request. Add `--tickers` to filter the preset to one ticker or a comma-delimited ticker list without spaces.

The command returns stable JSON with the requested date, DataTables record counts, and the first 100 raw trade objects from VolumeLeaders:

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

Structcli features are available from the scaffold:

```bash
volumeleaders-agent --jsonschema=tree  # Full command schema for agents
volumeleaders-agent env-vars           # Environment variable reference
volumeleaders-agent config-keys        # Config key reference
volumeleaders-agent --mcp              # Run stdio MCP server
```

The date flag can also be set with `VOLUMELEADERS_AGENT_TRADES_DATE`.

## Auth package

```go
import "github.com/major/volumeleaders-agent/internal/auth"
```

## License

See [LICENSE](LICENSE) for details.
