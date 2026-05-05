---
applyTo: "internal/models/**/*.go"
---

# Model review instructions

- JSON tags should match VolumeLeaders response fields.
- Model changes must not silently drop data needed by commands, summaries, CSV or TSV output, MCP tools, or output schemas.
- Prefer observable output tests when model changes affect command results.
