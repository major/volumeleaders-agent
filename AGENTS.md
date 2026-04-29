# volumeleaders-agent

Go CLI tool for querying institutional trade data from [VolumeLeaders](https://www.volumeleaders.com). Uses browser cookie extraction for authentication.

## Documentation Maintenance

When modifying CLI commands, flags, models, or behavior:

- Update `AGENTS.md` if the change affects project structure, build process, or conventions.
- Update the relevant skill file(s) in `skills/volumeleaders-agent/` to reflect new/changed/removed commands, flags, defaults, or output formats. Skill files are the primary reference for LLM agents using this tool.

Command-to-skill mapping:

- `internal/commands/trade.go`, `internal/commands/presets.go` -> `skills/volumeleaders-agent/trade.md`
- `internal/commands/daily.go` -> `skills/volumeleaders-agent/daily.md`
- `internal/commands/volume.go` -> `skills/volumeleaders-agent/volume.md`
- `internal/commands/chart.go` -> `skills/volumeleaders-agent/chart.md`
- `internal/commands/market.go` -> `skills/volumeleaders-agent/market.md`
- `internal/commands/alert.go` -> `skills/volumeleaders-agent/alert.md`
- `internal/commands/watchlist.go` -> `skills/volumeleaders-agent/watchlist.md`
- Shared conventions and command chooser updates -> `skills/volumeleaders-agent/SKILL.md`

## Project Layout

```text
cmd/volumeleaders-agent/main.go    Entry point
internal/auth/                     Browser cookie + XSRF token extraction
internal/client/                   HTTP client (DataTables + JSON requests)
internal/commands/                 CLI command definitions (7 groups, 22 subcommands)
internal/datatables/               DataTables protocol encoding + column definitions
internal/models/                   Response type definitions
skills/volumeleaders-agent/        LLM skill files for agent integration
```

## Build and Test

```bash
make build      # Build binary
make test       # Run tests
make lint       # Run linters
make install    # Install to $GOPATH/bin
```

## Conventions

- All commands output compact JSON to stdout by default. List-style commands may support `--format json|csv|tsv`; CSV/TSV include a header row, render booleans as `true`/`false`, and render null or missing values as empty cells. Use `--pretty` for indented JSON output. Errors go to stderr via `slog`.
- Dates use `YYYY-MM-DD` format on the CLI, converted internally as needed.
- Boolean/toggle filters use integers: `-1` = all/unfiltered, `0` = exclude, `1` = include/only.
- Pagination uses `--start` (offset) and `--length` (count). `--length -1` means all results except for capped trade retrieval endpoints. `trade list`, including `--summary`, only allows `--length` values from 1 to 50 because the VolumeLeaders backend cannot safely retrieve more than 50 individual trades per request. `trade levels` caps `--trade-level-count` at 50, and `trade level-touches` only allows `--length` values from 1 to 50.
- The binary name is `volumeleaders-agent`.
