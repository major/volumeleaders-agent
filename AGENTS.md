# volumeleaders-agent

Go CLI tool for querying institutional trade data from [VolumeLeaders](https://www.volumeleaders.com). Uses browser cookie extraction for authentication.

## Documentation Maintenance

When modifying CLI commands, flags, models, or behavior:

- Update `AGENTS.md` if the change affects project structure, build process, or conventions.
- Use `volumeleaders-agent --jsonschema=tree` as the source of truth for command names, flags, aliases, defaults, and examples. Use `volumeleaders-agent outputschema` as the source of truth for success stdout contracts, formats, fields, and variants. Command Long descriptions contain embedded domain knowledge (workflows, recovery steps, conventions, gotchas).

Command documentation mapping:

- All command groups -> `volumeleaders-agent --jsonschema=tree` for command shape and Long descriptions for semantic guidance. Use `volumeleaders-agent --mcp` when validating the MCP tool surface exposed to LLM clients.
- All command outputs -> `volumeleaders-agent outputschema` for success stdout contracts.
- Shared conventions, workflows, recovery behavior, output behavior, auth guidance, and domain gotchas -> root command Long description (run `volumeleaders-agent --help`).

## Project Layout

```text
cmd/volumeleaders-agent/main.go    Entry point
internal/auth/                     Browser cookie + XSRF token extraction
internal/client/                   HTTP client (DataTables + JSON requests)
internal/cli/                      CLI command definitions, MCP surface, and output contracts
internal/datatables/               DataTables protocol encoding + column definitions
internal/models/                   Response type definitions
```

## Build and Test

```bash
make build      # Build binary
make test       # Run tests
make lint       # Run linters
make install    # Install to $GOPATH/bin
```

## Conventions

- Commands output compact JSON to stdout by default. List-style commands may support `--format json|csv|tsv`; CSV/TSV include a header row, render booleans as `true`/`false`, and render null or missing values as empty cells. Use `--pretty` for indented JSON output. Use `outputschema` to inspect machine-readable success output contracts. Use `--mcp` on the root command to serve leaf commands as MCP tools over stdio for trusted local LLM clients. Errors go to stderr via `slog`.
- Root help must keep the recovery playbook actionable for LLMs: auth failures, date validation, pagination caps, unknown flags/enums, and empty or broad outputs should each have a concrete retry path.
- Dates use `YYYY-MM-DD` format on the CLI, converted internally as needed.
- Boolean/toggle filters use integers: `-1` = all/unfiltered, `0` = exclude, `1` = include/only.
- Pagination uses `--start` (offset) and `--length` (count). `--length -1` means all results except for capped trade retrieval endpoints. `trade list`, including `--summary`, defaults to `--length 10` and only allows `--length` values from 1 to 50 because the VolumeLeaders backend cannot safely retrieve more than 50 individual trades per request. `trade levels` caps `--trade-level-count` at 50, and `trade level-touches` only allows `--length` values from 1 to 50.
- The binary name is `volumeleaders-agent`.
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
- For `internal/cli/**/*.go`, verify command behavior matches README, `volumeleaders-agent --jsonschema=tree`, `volumeleaders-agent outputschema`, and the conventions above. If commands, flags, aliases, defaults, or examples change, verify `--jsonschema=tree` output reflects the changes. If workflows, behavior, models, output formats, or output fields change, update the relevant command Long descriptions and output contracts.
- For MCP changes, verify `volumeleaders-agent --mcp` keeps JSON-RPC protocol output on stdout, does not expose shell completion or parent routing commands as tools, and never leaks cookies, XSRF tokens, session values, or other credentials in tool results or errors.
- For tests, expect table-driven subtests with `t.Run`, parallelization where safe, `t.TempDir()` for filesystem work, deterministic fixtures, and assertions on observable behavior rather than implementation details. Do not request arbitrary coverage percentage changes.
- For GitHub Actions workflows, treat unpinned actions, excessive permissions, secret exposure in logs, or unsafe pull request execution patterns as P1.
- For `Makefile`, check that non-file targets are declared `.PHONY` and avoid adding flags that duplicate tool defaults.
- For `.goreleaser.yml`, verify GoReleaser v2 compatibility, CGO settings, signing/release behavior, and the intended platform matrix.
- For documentation-only changes, flag factual inaccuracies or stale command examples as P1 when they would cause users or LLM agents to run the wrong command.

## Maintenance notes

- Keep the review guidelines in this file and nested `AGENTS.md` files aligned with current project behavior. Update them when command behavior, security assumptions, output formats, CI/release workflows, or review priorities change.
- Prefer updating the closest nested `AGENTS.md` when guidance only applies to one package or directory. Keep this root file focused on cross-repository rules.
- When adding a new high-risk package or workflow area, add a nearby `AGENTS.md` with a `## Review guidelines` section so Codex receives the most specific instructions for changed files.
