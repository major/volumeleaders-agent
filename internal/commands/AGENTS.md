# internal/commands

CLI command definitions and output formatting.

## Review guidelines

- Treat broken stdout/stderr discipline as P1. Command data belongs on stdout as compact JSON by default, CSV/TSV where supported, or raw indented JSON for `schema`, while diagnostics and errors belong on stderr through `slog`.
- Treat unintentional CLI compatibility breaks as P1. Check flag names, positional arguments, default values, date parsing, boolean toggle values, pagination limits, and output fields.
- Verify command behavior matches README, `volumeleaders-agent schema`, and `skills/volumeleaders-agent/SKILL.md`. If commands, flags, aliases, defaults, or examples change, require schema coverage. If behavior, models, gotchas, or output formats change, require matching `SKILL.md` updates.
- Verify date inputs use `YYYY-MM-DD`, boolean toggles use `-1`, `0`, and `1`, and capped trade retrieval commands still enforce backend-safe limits.
- For tests, prefer observable command behavior, stdout/stderr separation, and deterministic fixtures over implementation details.

## Maintenance notes

- Update these guidelines whenever commands, flags, defaults, positional arguments, output formats, pagination limits, or skill documentation requirements change.
