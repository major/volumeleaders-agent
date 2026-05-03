# volumeleaders-agent

[![CI](https://github.com/major/volumeleaders-agent/actions/workflows/ci.yml/badge.svg)](https://github.com/major/volumeleaders-agent/actions/workflows/ci.yml)
[![CodeQL](https://github.com/major/volumeleaders-agent/actions/workflows/codeql.yml/badge.svg)](https://github.com/major/volumeleaders-agent/actions/workflows/codeql.yml)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/major/volumeleaders-agent/badge)](https://scorecard.dev/viewer/?uri=github.com/major/volumeleaders-agent)
[![codecov](https://codecov.io/gh/major/volumeleaders-agent/graph/badge.svg)](https://codecov.io/gh/major/volumeleaders-agent)
[![Go Report Card](https://goreportcard.com/badge/github.com/major/volumeleaders-agent)](https://goreportcard.com/report/github.com/major/volumeleaders-agent)
[![Go Reference](https://pkg.go.dev/badge/github.com/major/volumeleaders-agent.svg)](https://pkg.go.dev/github.com/major/volumeleaders-agent)

Go CLI for querying institutional trade data from [VolumeLeaders](https://www.volumeleaders.com). Surfaces large block trades, trade clusters, price levels, volume leaderboards, and market-wide signals that indicate institutional activity.

## Prerequisites

You must be logged into volumeleaders.com in a supported browser (Chrome, Firefox, Edge, etc.). The CLI extracts session cookies directly from the browser's cookie store, so no API keys or manual token management is needed.

## Install

```bash
go install github.com/major/volumeleaders-agent/cmd/volumeleaders-agent@latest
```

Pre-built binaries (signed with [Sigstore](https://www.sigstore.dev/)) are attached to each [GitHub release](https://github.com/major/volumeleaders-agent/releases).

## Quick start

```bash
# Today's top institutional volume movers
volumeleaders-agent volume institutional --date 2026-04-28

# Large trades in a specific ticker
volumeleaders-agent trade list --tickers NVDA --start-date 2025-01-01 --end-date 2025-04-24

# Market exhaustion signals
volumeleaders-agent --pretty market exhaustion
```

Commands emit compact JSON to stdout by default. Use `--pretty` for indented output. Errors go to stderr. Use `--jsonschema=tree` on the root command for a machine-readable JSON Schema of commands and flags, `volumeleaders-agent outputschema` for machine-readable stdout contracts, or `--mcp` to serve leaf commands as MCP tools over stdio for trusted local LLM clients. Root help includes a recovery playbook for authentication, date validation, pagination, unknown flags, and broad result sets.

## Commands

| Group | Purpose |
|---|---|
| `trade` | Institutional trades, clusters, cluster bombs, price levels |
| `volume` | Volume leaderboards (institutional, after-hours, total) |
| `market` | Market-wide snapshots, earnings calendar, exhaustion scores |
| `alert` | Saved alert configurations |
| `watchlist` | Saved watchlists and their tickers |
| `outputschema` | Machine-readable success output contracts |

Use `volumeleaders-agent --jsonschema=tree` for the machine-readable JSON Schema of all commands and flags. Use `volumeleaders-agent outputschema trade list` for the stdout contract of a specific command, including formats, fields, and conditional variants. Use `volumeleaders-agent --mcp` to expose the same leaf-command surface to MCP clients over stdio. Run `volumeleaders-agent --help` for embedded domain knowledge, filter conventions, recovery steps, workflows, and domain gotchas.

## Build

```bash
make build      # Build binary
make test       # Run tests
make lint       # Run golangci-lint
make install    # Install to $GOPATH/bin
```

## License

See [LICENSE](LICENSE) for details.
