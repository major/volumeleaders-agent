# volumeleaders-agent review instructions

Review this repository as a Go CLI for querying institutional trade data from VolumeLeaders. It uses browser cookie extraction for authentication, compact JSON stdout by default, generated LLM discovery files, JSON Schema discovery, output schema discovery, and MCP for local trusted clients.

Focus on correctness, credential safety, stdout and stderr discipline, CLI compatibility, output contracts, and repository conventions. Do not nitpick formatting or style that gofmt, go vet, or golangci-lint already handles.

## Project invariants

- Commands emit compact JSON to stdout by default. Diagnostics and errors belong on stderr through `slog`.
- Use `volumeleaders-agent --jsonschema=tree` as the command contract and `volumeleaders-agent outputschema` as the success output contract.
- Generated discovery files come from `make docs` or `make generate-discovery` and should stay in sync with command changes.
- Browser cookies, XSRF tokens, session values, and API responses containing credentials must never leak in logs, errors, docs, or MCP results.
- Keep `AGENTS.md`, nested AGENTS files, README, CodeRabbit, and Copilot review guidance aligned when review priorities or behavior change.

## Behavior checks

- Treat broken stdout and stderr discipline as P1.
- Treat CLI compatibility breaks as P1 unless the PR clearly documents an intentional breaking change.
- Check flag names, positional args, date handling, boolean toggle values, pagination limits, and default output fields.
- MCP output must keep JSON-RPC protocol data on stdout and must not expose shell completion or parent routing commands as tools.
