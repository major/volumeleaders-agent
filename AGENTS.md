# volumeleaders-agent

Go authentication package for [VolumeLeaders](https://www.volumeleaders.com). Uses browser cookie extraction and XSRF token retrieval for authentication.

## Documentation Maintenance

When modifying authentication behavior:

- Update `AGENTS.md` if the change affects project structure, build process, or conventions.
- Keep README examples aligned with the exported auth package behavior.

## Project Layout

```text
internal/auth/                     Browser cookie + XSRF token extraction
```

## Build and Test

```bash
make test       # Run tests
make lint       # Run linters
```

## Conventions

- Authentication errors must include useful troubleshooting context without exposing browser cookies, XSRF tokens, session values, profile paths, or other secrets.
- Context cancellation must propagate through browser cookie extraction and token lookup paths.

## Review guidelines

- Focus review comments on correctness, safety, maintainability, and repository conventions. Do not nitpick formatting or style that `gofmt`, `go vet`, or `golangci-lint` already enforce.
- Treat any change that can leak browser cookies, XSRF tokens, session values, API responses containing credentials, or other secrets as P1. Authentication failures must degrade gracefully and must not expose sensitive values in logs or errors.
- For Go code under `internal/**/*.go`, verify errors are wrapped with useful context and `%w` when returning underlying errors, typed error matching uses `errors.As`, and context cancellation is propagated through HTTP requests.
- For `internal/auth/**/*.go`, check cookie extraction, browser profile handling, and token lookup paths for credential safety, useful error messages, and graceful behavior when browsers or cookies are unavailable.
- For tests, expect table-driven subtests with `t.Run`, parallelization where safe, `t.TempDir()` for filesystem work, deterministic fixtures, and assertions on observable behavior rather than implementation details. Do not request arbitrary coverage percentage changes.
- For GitHub Actions workflows, treat unpinned actions, excessive permissions, secret exposure in logs, or unsafe pull request execution patterns as P1.
- For `Makefile`, check that non-file targets are declared `.PHONY` and avoid adding flags that duplicate tool defaults.
- For documentation-only changes, flag factual inaccuracies or stale command examples as P1 when they would cause users or LLM agents to run the wrong command.

## Maintenance notes

- Keep the review guidelines in this file and nested `AGENTS.md` files aligned with current project behavior. Update them when authentication behavior, security assumptions, CI workflows, or review priorities change.
- Prefer updating the closest nested `AGENTS.md` when guidance only applies to one package or directory. Keep this root file focused on cross-repository rules.
- When adding a new high-risk package or workflow area, add a nearby `AGENTS.md` with a `## Review guidelines` section so Codex receives the most specific instructions for changed files.
