---
applyTo: "internal/**/*.go"
---

# Go review instructions

- Wrap errors with useful context and `%w` when returning underlying errors.
- Use `errors.As()` for typed error matching.
- Preserve context cancellation through HTTP and DataTables requests.
- Command behavior should match README, `--jsonschema=tree`, `outputschema`, and root help conventions.
- Do not suggest style-only changes handled by gofmt, go vet, or golangci-lint.
