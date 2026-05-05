---
applyTo: "Makefile"
---

# Makefile review instructions

- Non-file targets should have `.PHONY` declarations.
- Avoid flags that duplicate tool defaults.
- Keep docs, discovery generation, smoke, build, test, and lint targets aligned with README and `AGENTS.md`.
