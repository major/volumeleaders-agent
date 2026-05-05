---
applyTo: "internal/{cli,discovery}/**/*.go"
---

# CLI and discovery review instructions

- Command behavior should match README, `volumeleaders-agent --jsonschema=tree`, `volumeleaders-agent outputschema`, and root help conventions.
- If commands, flags, aliases, defaults, or examples change, verify JSON Schema output reflects the changes and run `make generate-discovery`.
- If workflows, behavior, models, output formats, or output fields change, update relevant command Long descriptions and output contracts.
- Generated files should be deterministic and must not overwrite the root `AGENTS.md`.
- MCP must keep JSON-RPC protocol output on stdout and never leak credentials in tool results or errors.
