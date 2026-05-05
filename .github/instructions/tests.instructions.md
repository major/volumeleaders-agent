---
applyTo: "**/*_test.go"
---

# Test review instructions

- Prefer table-driven subtests with `t.Run()`.
- Use parallelization where safe.
- Use `t.TempDir()` for filesystem work.
- Keep fixtures deterministic.
- Assert observable behavior rather than implementation details.
- Do not request arbitrary coverage percentage changes.
