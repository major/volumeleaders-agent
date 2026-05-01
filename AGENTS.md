# volumeleaders-agent

Go CLI for [VolumeLeaders](https://www.volumeleaders.com). Uses structcli for human-friendly command help, LLM-friendly JSON schema discovery, MCP support, and browser cookie/XSRF authentication.

## Documentation Maintenance

When modifying CLI commands, flags, output behavior, or authentication behavior:

- Update `AGENTS.md` if the change affects project structure, build process, or conventions.
- Keep `AGENTS.md` updated whenever project behavior, command behavior, authentication behavior, verification expectations, or contributor/agent conventions change.
- Keep README examples aligned with the exported CLI and auth package behavior.

## Project Layout

```text
cmd/volumeleaders-agent/              CLI entry point
internal/cmd/                         structcli command tree
internal/cmd/trades/                  large unusual trades, trade clusters, cluster bombs, trade levels, trade level touches, and ranked trade commands
internal/cmd/watchlists/              configured watchlist list, save, and delete commands
internal/auth/                     Browser cookie + XSRF token extraction
```

## Build and Test

```bash
make build      # Build binary
make test       # Run tests
make lint       # Run linters
```

## Verification

- Run local verification such as `make test`, `make lint`, and `make build` for implementation work.
- Do not run Oracle review or other external AI review at the end of a task unless the user explicitly requests it.

## Conventions

- Commands should use structcli options structs for flags so `--jsonschema=tree`, `env-vars`, `config-keys`, and `--mcp` stay accurate for humans and agents.
- Treat structcli-visible code metadata as the source of truth for command behavior. When adding or changing commands, flags, filters, defaults, examples, output shapes, field meanings, or caveats, update option struct tags (`flagdescr`, `flagenv`, `flaggroup`, `flagrequired`), command `Short`/`Long` text, examples, and any embedded field guides so `--jsonschema=tree`, help output, `env-vars`, `config-keys`, and MCP discovery remain understandable without loading README files or skill files.
- Keep `docs/fields.md` aligned with annotated trade and cluster payloads when field meanings, preset inclusion, or ignored-field decisions change.
- Err on the side of adding too many helpful comments rather than too few, especially around non-obvious API filters, command grouping, authentication edge cases, and behavior copied from browser captures.
- Command output should be stable JSON by default. Errors should rely on structcli/cobra error handling and include actionable context without leaking credentials.
- Trade and cluster `core` presets use compact derived fields. `CalendarEvent` contains true upstream calendar markers (`EOM`, `EOQ`, `EOY`, `OPEX`, `VOLEX`); array output returns `null` when none are true and object output omits it. `AuctionTrade` derives from upstream `OpeningTrade` and `ClosingTrade` `0`/`1` flags; array output returns `"open"`, `"close"`, or `null`, and object output omits it when neither flag is true. Trade and cluster `expanded` presets include all annotated non-internal signal fields while still excluding raw upstream internals, always-zero fields, and always-null fields.
- The `trade-cluster-bombs` command posts the browser-compatible `TradeClusterBombs/GetTradeClusterBombs` DataTables form across a date range. VolumeLeaders allows at most 7 days when querying all tickers or multiple comma-delimited tickers, while single-ticker scans default to a one-year lookback when `--start-date` is omitted. Default output is compact core array JSON with visible cluster-bomb table fields, including `TradeClusterBombRank` and `LastComparableTradeClusterBombDate`.
- The `trade-levels` command posts the browser-compatible `TradeLevels/GetTradeLevels` DataTables form for one ticker across a date range. Default output is compact core array JSON with visible level table fields: `Ticker`, `Price`, `Dollars`, `Volume`, `Trades`, `RelativeSize`, `CumulativeDistribution`, `TradeLevelRank`, and `Dates`. `TradeLevelRank=0` means the returned row was not marked as ranked in captured browser responses.
- The `trade-level-touches` command posts the browser-compatible `TradeLevelTouches/GetTradeLevelTouches` DataTables form across a date range. VolumeLeaders allows at most 7 days when querying all tickers or multiple comma-delimited tickers, while single-ticker scans default to a one-year lookback when `--start-date` is omitted. Default `TradeLevelRank=10` returns ranks 1 through 10, and default output is compact core array JSON with visible touch table fields including `FullDateTime`, `Price`, `Dollars`, `RelativeSize`, `CumulativeDistribution`, `TradeLevelRank`, and `Dates`.
- The `watchlists` command lists account-level saved VolumeLeaders watchlist filter definitions through `WatchListConfigs/GetWatchLists`. Default output is compact summary array JSON with `SearchTemplateKey` and `Name`; `--preset-fields expanded` shows the saved filter configuration fields and `IncludedTradeTypes`, which derives enabled print/session booleans such as `DarkPools`, `Sweeps`, `RTHTrades`, and `OffsettingTrades`. The `save-watchlist` command creates or fully replaces those definitions by posting the browser-compatible `WatchListConfig` form; `SearchTemplateKey=0` creates a new watchlist, and a positive key replaces an existing one with the complete criteria in the request. The `delete-watchlist` command removes a definition by posting JSON `WatchListKey` to `WatchListConfigs/DeleteWatchList`.
- Authentication errors must include useful troubleshooting context without exposing browser cookies, XSRF tokens, session values, profile paths, or other secrets.
- Context cancellation must propagate through browser cookie extraction and token lookup paths.

## Review guidelines

- Focus review comments on correctness, safety, maintainability, and repository conventions. Do not nitpick formatting or style that `gofmt`, `go vet`, or `golangci-lint` already enforce.
- Treat any change that can leak browser cookies, XSRF tokens, session values, API responses containing credentials, or other secrets as P1. Authentication failures must degrade gracefully and must not expose sensitive values in logs or errors.
- For Go code under `internal/**/*.go`, verify errors are wrapped with useful context and `%w` when returning underlying errors, typed error matching uses `errors.As`, and context cancellation is propagated through HTTP requests.
- For `internal/cmd/**/*.go`, verify structcli tags, required flags, env var bindings, JSON schema output, MCP compatibility, and command output shape remain accurate.
- For `internal/auth/**/*.go`, check cookie extraction, browser profile handling, and token lookup paths for credential safety, useful error messages, and graceful behavior when browsers or cookies are unavailable.
- For tests, expect table-driven subtests with `t.Run`, parallelization where safe, `t.TempDir()` for filesystem work, deterministic fixtures, and assertions on observable behavior rather than implementation details. Do not request arbitrary coverage percentage changes.
- For GitHub Actions workflows, treat unpinned actions, excessive permissions, secret exposure in logs, or unsafe pull request execution patterns as P1.
- For `Makefile`, check that non-file targets are declared `.PHONY` and avoid adding flags that duplicate tool defaults.
- For documentation-only changes, flag factual inaccuracies or stale command examples as P1 when they would cause users or LLM agents to run the wrong command.

## Maintenance notes

- Keep the review guidelines in this file and nested `AGENTS.md` files aligned with current project behavior. Update them when command behavior, authentication behavior, security assumptions, CI workflows, or review priorities change.
- Prefer updating the closest nested `AGENTS.md` when guidance only applies to one package or directory. Keep this root file focused on cross-repository rules.
- When adding a new high-risk package or workflow area, add a nearby `AGENTS.md` with a `## Review guidelines` section so Codex receives the most specific instructions for changed files.
