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
volumeleaders-agent volume institutional

# Large trades in a specific ticker
volumeleaders-agent trade list --tickers NVDA --start-date 2025-01-01 --end-date 2025-04-24

# Market exhaustion signals
volumeleaders-agent --pretty market exhaustion
```

All commands emit compact JSON to stdout by default. Use `--pretty` for indented output. Errors go to stderr.

## Commands

| Group | Purpose | Docs |
|---|---|---|
| `trade` | Institutional trades, clusters, cluster bombs, price levels | [skills/trade.md](skills/trade.md) |
| `daily` | Compact daily institutional activity summaries | [skills/daily.md](skills/daily.md) |
| `volume` | Volume leaderboards (institutional, after-hours, total) | [skills/volume.md](skills/volume.md) |
| `chart` | Price bars with trade overlays, snapshots, company metadata | [skills/chart.md](skills/chart.md) |
| `market` | Market-wide snapshots, earnings calendar, exhaustion scores | [skills/market.md](skills/market.md) |
| `alert` | Saved alert configurations | [skills/alert.md](skills/alert.md) |
| `watchlist` | Saved watchlists and their tickers | [skills/watchlist.md](skills/watchlist.md) |

See [skills/SKILL.md](skills/SKILL.md) for full flag reference, filter conventions, and a decision tree for picking the right command.

## Build

```bash
make build      # Build binary
make test       # Run tests
make lint       # Run golangci-lint
make install    # Install to $GOPATH/bin
```

## License

See [LICENSE](LICENSE) for details.
