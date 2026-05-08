---
applyTo: "internal/cli/**/*.go"
---

# CLI review instructions

- Command behavior should match README, `volumeleaders-agent --jsonschema=tree`, `volumeleaders-agent outputschema`, and root help conventions.
- If commands, flags, aliases, defaults, or examples change, verify JSON Schema output reflects the changes.
- If workflows, behavior, models, output formats, or output fields change, update relevant command Long descriptions and output contracts.
- MCP must keep JSON-RPC protocol output on stdout and never leak credentials in tool results or errors.
