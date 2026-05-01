# Contributing

## Getting Started

```bash
git clone https://github.com/major/volumeleaders-agent.git
cd volumeleaders-agent
make test
make lint
```

## Development

**Requirements:** Go 1.26 or later.

Linting uses [golangci-lint](https://golangci-lint.run/) with the config in `.golangci.yml`. Run `make lint` before submitting.

A few conventions to follow when writing code:

- Tests use table-driven style with `t.Parallel()` at the top of each test and subtest.
- Authentication errors include useful troubleshooting context without exposing cookies, tokens, browser profile paths, or other secrets.
- Context cancellation must propagate through browser cookie extraction and token lookup paths.

## Pull Requests

Fork the repo, create a branch, and open a PR against `main`.

- Keep PRs focused on a single change. Unrelated fixes belong in separate PRs.
- All CI checks (tests and linting) must pass before merge.
- Add tests for any new functionality.

## Code Style

Follow the patterns already in the codebase. `make lint` catches most issues. When in doubt, match the style in `internal/auth/`.

## License

By contributing, you agree that your changes will be licensed under the [Apache-2.0 License](LICENSE).
