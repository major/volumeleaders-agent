# volumeleaders-agent

Go CLI tool for querying institutional trade data from [VolumeLeaders](https://www.volumeleaders.com). Uses browser cookie extraction for authentication.

## Documentation Maintenance

When modifying CLI commands, flags, models, or behavior:

- Update `AGENTS.md` if the change affects project structure, build process, or conventions.
- Keep `.coderabbit.yaml` and `.github/copilot-instructions.md` plus `.github/instructions/*.instructions.md` aligned with current repo conventions when review-relevant behavior changes.
- Use `volumeleaders-agent --jsonschema=tree` as the source of truth for command names, flags, aliases, defaults, and examples. Use `volumeleaders-agent outputschema` as the source of truth for success stdout contracts, formats, fields, and variants. Command Long descriptions contain embedded domain knowledge (workflows, recovery steps, conventions, gotchas).
- Run `make docs` or `make generate-discovery` when commands, flags, defaults, examples, or Long descriptions change. The generated `SKILL.md` lives at the repository root for consistent tool discovery; extended generated LLM discovery files live in `docs/llm/` so they do not overwrite this hand-maintained root `AGENTS.md`.

Command documentation mapping:

- All command groups -> `volumeleaders-agent --jsonschema=tree` for command shape and Long descriptions for semantic guidance. Use `volumeleaders-agent --mcp` when validating the MCP tool surface exposed to LLM clients.
- All command outputs -> `volumeleaders-agent outputschema` for success stdout contracts.
- Generated LLM discovery files -> `make docs` writes root `SKILL.md` plus `docs/llm/AGENTS.md` and `docs/llm/llms.txt` from structcli metadata.
- Shared conventions, workflows, recovery behavior, output behavior, auth guidance, and domain gotchas -> root command Long description (run `volumeleaders-agent --help`).

## Project Layout

```text
cmd/volumeleaders-agent/main.go    Entry point
cmd/smoke-test/main.go             Local-only live smoke test harness
internal/auth/                     Browser cookie + XSRF token extraction
internal/client/                   HTTP client (DataTables + JSON requests)
internal/cli/                      CLI command definitions, MCP surface, output contracts, and generated discovery metadata
internal/discovery/                Generated LLM discovery file writer
internal/datatables/               DataTables protocol encoding + column definitions
internal/models/                   Response type definitions
internal/update/                   GitHub release update checks, updater settings, and self-update logic
SKILL.md                           Generated skill file for Claude Code and skill-aware tools
docs/llm/                          Generated AGENTS.md and llms.txt for extended LLM discovery
```

## Build and Test

```bash
make build      # Build binary
make test       # Run tests
make lint       # Run linters
make install    # Install to $GOPATH/bin
make docs       # Refresh root SKILL.md and docs/llm discovery files
make generate-discovery # Same as make docs
make smoke      # Run local live smoke tests against the built binary
```

## Conventions

- Commands output compact JSON to stdout by default. List-style commands may support `--format json|csv|tsv`; CSV/TSV include a header row, render booleans as `true`/`false`, and render null or missing values as empty cells. Use `--pretty` for indented JSON output. Use `outputschema` to inspect machine-readable success output contracts. Use `--mcp` on the root command to serve leaf commands as MCP tools over stdio for trusted local LLM clients. Errors go to stderr via `slog`.
- Root help must keep the recovery playbook actionable for LLMs: auth failures, date validation, pagination caps, unknown flags/enums, and empty or broad outputs should each have a concrete retry path.
- Dates use `YYYY-MM-DD` format on the CLI, converted internally as needed.
- Boolean/toggle filters use integers: `-1` = all/unfiltered, `0` = exclude, `1` = include/only.
- Pagination uses `--start` (offset) and `--length` (count) where commands expose count selection. `--length -1` means all results for generic DataTables commands. `trade list` does not expose `--length`; multi-day lookups whose effective filters include tickers return the top 10 long-period trades with VolumeLeaders' lightweight chart query shape, while `trade list --summary`, single-day trade scans, all-market trade scans, sector-only presets, `trade clusters`, and `trade cluster-bombs` fetch all rows internally in browser-sized 100-row pages because VolumeLeaders expects those endpoints to use 100 results per request. `trade levels` and `trade level-touches` only allow `--trade-level-count` values of 5, 10, 20, or 50. `trade level-touches` defaults to `--trade-level-rank 5`, requires rank 5 or higher, and only allows `--length` values from 1 to 50.
- The binary name is `volumeleaders-agent`.
- Update notifications are on by default for interactive release builds, cached for one day, skipped in CI and non-interactive output, and controlled only by the updater-specific settings file managed through `volumeleaders-agent update config`. Do not replace this with structcli config/env loading. The `update` command must verify downloaded release assets against GoReleaser checksums before replacing the binary.
- `make smoke` is a local-only live test harness, not part of CI. It builds `./volumeleaders-agent`, discovers command coverage from `--jsonschema=tree`, validates stdout JSON, and may create/update/delete smoke-owned alert and watchlist records named with a `smoke-*` prefix. Smoke mutations must only target keys created during the same smoke run, and cleanup must be attempted even after failures.
- structcli environment-variable and config-file features are intentionally out of scope for this project. Do not add or recommend `flagenv`, `flagenv:"only"`, `structcli.WithConfig`, `--config`, YAML/JSON/TOML config loading, or environment-variable driven CLI defaults unless this guidance is explicitly changed later.

## Review guidelines

- Focus review comments on correctness, safety, maintainability, and repository conventions. Do not nitpick formatting or style that `gofmt`, `go vet`, or `golangci-lint` already enforce.
- Treat any change that can leak browser cookies, XSRF tokens, session values, API responses containing credentials, or other secrets as P1. Authentication failures must degrade gracefully and must not expose sensitive values in logs or errors.
- Treat broken stdout/stderr discipline as P1. Command output belongs on stdout as compact JSON by default, optional CSV/TSV where supported; diagnostics and errors belong on stderr through `slog`.
- Treat changes that break CLI compatibility as P1 unless the PR clearly documents an intentional breaking change. Check flag names, positional arguments, date handling, boolean toggle values, pagination limits, and default output fields.
- For Go code under `internal/**/*.go`, verify errors are wrapped with useful context and `%w` when returning underlying errors, typed error matching uses `errors.As`, and context cancellation is propagated through HTTP and DataTables requests.
- For `internal/auth/**/*.go`, check cookie extraction, browser profile handling, and token lookup paths for credential safety, useful error messages, and graceful behavior when browsers or cookies are unavailable.
- For `internal/client/**/*.go`, check HTTP status handling, request context propagation, response body closure, DataTables encoding, and errors that explain the endpoint or operation that failed.
- For `internal/models/**/*.go`, verify JSON tags match VolumeLeaders response fields and that model changes do not silently drop data needed by commands, summaries, or CSV/TSV output.
- For `internal/cli/**/*.go`, verify command behavior matches README, `volumeleaders-agent --jsonschema=tree`, `volumeleaders-agent outputschema`, and the conventions above. If commands, flags, aliases, defaults, or examples change, verify `--jsonschema=tree` output reflects the changes and run `make generate-discovery`. If workflows, behavior, models, output formats, or output fields change, update the relevant command Long descriptions and output contracts.
- For MCP changes, verify `volumeleaders-agent --mcp` keeps JSON-RPC protocol output on stdout, does not expose shell completion or parent routing commands as tools, and never leaks cookies, XSRF tokens, session values, or other credentials in tool results or errors.
- For `internal/discovery/**/*.go`, verify generated files are deterministic, do not overwrite the root `AGENTS.md`, and stay in sync with `docs/llm/` through tests.
- For `internal/update/**/*.go`, verify update checks do not poll more often than intended, never write update notices to stdout, skip dev or non-interactive runs, preserve context cancellation, and validate downloads against the release checksum asset before applying a binary replacement.
- For tests, expect table-driven subtests with `t.Run`, parallelization where safe, `t.TempDir()` for filesystem work, deterministic fixtures, and assertions on observable behavior rather than implementation details. Do not request arbitrary coverage percentage changes.
- For GitHub Actions workflows, treat unpinned actions, excessive permissions, secret exposure in logs, or unsafe pull request execution patterns as P1.
- For `Makefile`, check that non-file targets are declared `.PHONY` and avoid adding flags that duplicate tool defaults.
- For `.goreleaser.yml`, verify GoReleaser v2 compatibility, CGO settings, signing/release behavior, and the intended platform matrix.
- For documentation-only changes, flag factual inaccuracies or stale command examples as P1 when they would cause users or LLM agents to run the wrong command.

## Maintenance notes

- Keep the review guidelines in this file and nested `AGENTS.md` files aligned with current project behavior. Update them when command behavior, security assumptions, output formats, CI/release workflows, or review priorities change.
- Prefer updating the closest nested `AGENTS.md` when guidance only applies to one package or directory. Keep this root file focused on cross-repository rules.
- When adding a new high-risk package or workflow area, add a nearby `AGENTS.md` with a `## Review guidelines` section so Codex receives the most specific instructions for changed files.
