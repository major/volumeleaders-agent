# Contributing

## Getting Started

```bash
git clone https://github.com/major/volumeleaders-agent.git
cd volumeleaders-agent
make build
make test
make lint
```

## Development

**Requirements:** Go 1.24 or later.

Linting uses [golangci-lint](https://golangci-lint.run/) with the config in `.golangci.yml`. Run `make lint` before submitting.

A few conventions to follow when writing code:

- Tests use table-driven style with `t.Parallel()` at the top of each test and subtest.
- All commands write compact JSON to stdout. Errors go to stderr via `slog`. Use `--pretty` for indented output.
- Dates on the CLI use `YYYY-MM-DD` format and are converted internally as needed.
- Boolean/toggle filters use integers: `-1` means all/unfiltered, `0` means exclude, `1` means include.

## Pull Requests

Fork the repo, create a branch, and open a PR against `main`.

- Keep PRs focused on a single change. Unrelated fixes belong in separate PRs.
- All CI checks (tests and linting) must pass before merge.
- Add tests for any new functionality.

## Code Style

Follow the patterns already in the codebase. `make lint` catches most issues. When in doubt, look at how existing commands in `internal/commands/` are structured and match that style.

## License

By contributing, you agree that your changes will be licensed under the [Apache-2.0 License](LICENSE).
