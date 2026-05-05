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

Commands emit compact JSON to stdout by default. Use `--pretty` for indented output. Errors go to stderr. Use `--jsonschema=tree` on the root command for a machine-readable JSON Schema of commands and flags, `volumeleaders-agent outputschema` for machine-readable stdout contracts, or `--mcp` to serve leaf commands as MCP tools over stdio for trusted local LLM clients. Root help includes a recovery playbook for authentication, date validation, pagination, unknown flags, and broad result sets. Generated LLM discovery files can be refreshed with `make docs`.

## Commands

| Group | Purpose |
| --- | --- |
| `trade` | Institutional trades, clusters, cluster bombs, price levels |
| `volume` | Volume leaderboards (institutional, after-hours, total) |
| `market` | Market-wide earnings calendar and exhaustion scores |
| `alert` | Saved alert configurations |
| `watchlist` | Saved watchlists and their tickers |
| `update` | Check for releases, self-update the binary, and manage update notifications |
| `outputschema` | Machine-readable success output contracts |

Use `volumeleaders-agent --jsonschema=tree` for the machine-readable JSON Schema of all commands and flags. Use `volumeleaders-agent outputschema trade list` for the stdout contract of a specific command, including formats, fields, and conditional variants. Use `volumeleaders-agent --mcp` to expose the same leaf-command surface to MCP clients over stdio. Run `volumeleaders-agent --help` for embedded domain knowledge, filter conventions, recovery steps, workflows, and domain gotchas.

## Local smoke tests

Run local smoke tests against the live VolumeLeaders API with your browser session:

```bash
make smoke
```

The smoke harness builds `./volumeleaders-agent`, discovers commands with `--jsonschema=tree`, verifies every discovered command has an explicit fixture, and checks that each command returns valid JSON. It is intentionally local-only and is not part of `make test` or GitHub Actions because it needs live network access and browser authentication.

By default, `make smoke` runs read-only commands and smoke-owned mutation checks. Mutating checks create alert and watchlist records with `smoke-*` names, resolve the keys for those exact records, then attempt to delete them before exiting. Use read-only mode when you want to avoid mutations entirely:

```bash
go run ./cmd/smoke-test --mode read-only
```

## Updates

Interactive release builds check for newer GitHub releases at most once per day and write update notices to stderr so command JSON on stdout stays machine-readable. Disable or re-enable those notices with the updater-specific settings command:

```bash
volumeleaders-agent update config --check-notifications=false
volumeleaders-agent update config --check-notifications=true
```

Check without installing, or install the latest release for your platform:

```bash
volumeleaders-agent update check
volumeleaders-agent update
```

Self-updates download the matching GitHub release archive, verify it against the GoReleaser checksum file, then replace the current binary with rollback support from the updater library.

## Agent integration

volumeleaders-agent ships a generated root `SKILL.md` for coding agents and LLM tools, with extended discovery files in `docs/llm/`. Use the file that matches your tool, then use live schema output from the binary as the authoritative command and flag contract.

### Claude Code and other skill-aware tools

Use the checked-in root `SKILL.md` when your tool supports Agent Skills, including Claude Code. It contains trigger phrases, command descriptions, flag tables, examples, and workflow guidance generated from the live Cobra and structcli command tree.

For Claude Code, copy or symlink the generated skill into your local skills directory if you want it available outside this repository:

```bash
mkdir -p ~/.claude/skills/volumeleaders-agent
ln -sf "$(pwd)/SKILL.md" ~/.claude/skills/volumeleaders-agent/SKILL.md
```

If you already have a custom skill at that path, review or back it up first because `ln -sf` replaces the existing file or symlink.

### OpenCode and Codex

OpenCode and Codex use `AGENTS.md` as project instructions. Start in the repository root so the tool can load the hand-maintained root `AGENTS.md`, then point it at `SKILL.md`, `docs/llm/AGENTS.md`, or `docs/llm/llms.txt` when it needs the generated command reference.

The root `AGENTS.md` stays hand-written because it captures architecture, package conventions, review guidance, and safety rules that cannot be generated from the CLI tree. The root `SKILL.md` and `docs/llm/` files are generated artifacts.

### Regenerating discovery files

The root `SKILL.md` plus `docs/llm/AGENTS.md` and `docs/llm/llms.txt` are built from the Cobra and structcli command tree. Refresh them after command, flag, default, or example changes:

```bash
make docs
```

Commit the regenerated files with the code change so agents see the same command surface as the binary.

## Build

```bash
make build      # Build binary
make test       # Run tests
make lint       # Run golangci-lint
make install    # Install to $GOPATH/bin
make docs       # Refresh root SKILL.md and docs/llm discovery files
make generate-discovery # Same as make docs
make smoke      # Run local live smoke tests
```

## License

See [LICENSE](LICENSE) for details.
