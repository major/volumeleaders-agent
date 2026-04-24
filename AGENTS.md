# volumeleaders-agent

Go CLI tool for querying institutional trade data from [VolumeLeaders](https://www.volumeleaders.com). Uses browser cookie extraction for authentication.

## Documentation Maintenance

When modifying CLI commands, flags, models, or behavior:

- Update `AGENTS.md` if the change affects project structure, build process, or conventions.
- Update the relevant skill file(s) in `skills/` to reflect new/changed/removed commands, flags, defaults, or output formats. Skill files are the primary reference for LLM agents using this tool.

Command-to-skill mapping:

- `internal/commands/trade.go`, `internal/commands/presets.go` -> `skills/trade.md`
- `internal/commands/volume.go` -> `skills/volume.md`
- `internal/commands/chart.go` -> `skills/chart.md`
- `internal/commands/market.go` -> `skills/market.md`
- `internal/commands/alert.go` -> `skills/alert.md`
- `internal/commands/watchlist.go` -> `skills/watchlist.md`
- Shared conventions and command chooser updates -> `skills/SKILL.md`

## Project Layout

```text
cmd/volumeleaders-agent/main.go    Entry point
internal/auth/                     Browser cookie + XSRF token extraction
internal/client/                   HTTP client (DataTables + JSON requests)
internal/commands/                 CLI command definitions (6 groups, 21 subcommands)
internal/datatables/               DataTables protocol encoding + column definitions
internal/models/                   Response type definitions
skills/                            LLM skill files for agent integration
```

## Build and Test

```bash
make build      # Build binary
make test       # Run tests
make lint       # Run linters
make install    # Install to $GOPATH/bin
```

## Conventions

- All commands output compact JSON to stdout by default. Use `--pretty` for indented output. Errors go to stderr via `slog`.
- Dates use `YYYY-MM-DD` format on the CLI, converted internally as needed.
- Boolean/toggle filters use integers: `-1` = all/unfiltered, `0` = exclude, `1` = include/only.
- Pagination uses `--start` (offset) and `--length` (count). `--length -1` means all results.
- The binary name is `volumeleaders-agent`.
