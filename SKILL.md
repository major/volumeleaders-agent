---
name: volumeleaders-agent
description: |
  volumeleaders-agent queries institutional trade data from VolumeLeaders. Use it for trades, volume leaderboards, market data, alerts, and watchlists.

  Auth: reads browser cookies automatically. If auth fails with exit code 2 and "Authentication required: VolumeLeaders session has expired.", log in at https://www.volumeleaders.com in your browser, then retry.

  Output: compact JSON to stdout by default. Use --pretty before the command group for indented JSON. Use --jsonschema on any command for machine-readable input JSON Schema output, --jsonschema=tree on the root for the full CLI tree, outputschema for machine-readable stdout contracts, or --mcp on the root to serve leaf commands as MCP tools over stdio. Errors and logs go to stderr.
metadata:
  author: major
  version: dev
---

# volumeleaders-agent

## Instructions

### Available Commands

#### `volumeleaders-agent alert configs`

List all saved alert configurations with their keys, names, ticker filters, trade conditions, and notification settings. Outputs compact JSON or CSV/TSV with --format. Use --fields to select specific output fields.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--fields` | string | - | no | Comma-separated fields to include (use 'all' for every field) |
| `--format` | string | json | no | Output format: json, csv, or tsv |

**Example:**

```bash
volumeleaders-agent alert configs
```

#### `volumeleaders-agent alert create`

Create a new alert configuration with a name and filter settings for institutional trade activity. Requires --name. Specify filters such as trade rank, dollar thresholds, dark pool and sweep conditions, and ticker scope. Returns a success response with the new configuration key.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--ah-dollars-gte` | int | 0 | no | After-hours dollars >= |
| `--ah-rank-lte` | int | 0 | no | After-hours rank <= |
| `--ah-volume-gte` | int | 0 | no | After-hours volume >= |
| `--closing-trade-conditions` | string | 0 | no | Closing trade conditions |
| `--closing-trade-dollars-gte` | int | 0 | no | Closing trade dollars >= |
| `--closing-trade-mult-gte` | int | 0 | no | Closing trade multiplier >= |
| `--closing-trade-rank-lte` | int | 0 | no | Closing trade rank <= |
| `--closing-trade-vcd-gte` | int | 0 | no | Closing trade VCD >= (0/97/98/99/100) |
| `--closing-trade-volume-gte` | int | 0 | no | Closing trade volume >= |
| `--cluster-dollars-gte` | int | 0 | no | Trade cluster dollars >= |
| `--cluster-mult-gte` | int | 0 | no | Trade cluster multiplier >= |
| `--cluster-rank-lte` | int | 0 | no | Trade cluster rank <= |
| `--cluster-vcd-gte` | int | 0 | no | Trade cluster VCD >= (0/97/98/99/100) |
| `--cluster-volume-gte` | int | 0 | no | Trade cluster volume >= |
| `--dark-pool` | bool | false | no | Dark pool filter |
| `--name` | string | - | yes | Alert name (max 50 chars) |
| `--offsetting-print` | bool | false | no | Offsetting print filter |
| `--phantom-print` | bool | false | no | Phantom print filter |
| `--sweep` | bool | false | no | Sweep filter |
| `--ticker-group` | string | AllTickers | no | Ticker group: AllTickers or SelectedTickers |
| `--tickers` | string | - | no | Comma-separated ticker symbols (max 500, used with SelectedTickers) |
| `--total-dollars-gte` | int | 0 | no | Total dollars >= |
| `--total-rank-lte` | int | 0 | no | Total rank <= (0/1/3/10/25/50/100) |
| `--total-volume-gte` | int | 0 | no | Total volume >= |
| `--trade-conditions` | string | 0 | no | Trade conditions (0=N/A, OBH/OBD/OSH/OSD combos) |
| `--trade-dollars-gte` | int | 0 | no | Trade dollars >= (0=N/A, 1000000/10000000/...) |
| `--trade-mult-gte` | int | 0 | no | Trade multiplier >= (0=N/A, 5/10/25/50/100) |
| `--trade-rank-lte` | int | 0 | no | Trade rank <= (0=N/A, 1/3/5/10/25/50/100) |
| `--trade-vcd-gte` | int | 0 | no | Trade VCD >= (0=N/A, 99/100) |
| `--trade-volume-gte` | int | 0 | no | Trade volume >= (0=N/A, 1000000/2000000/5000000/10000000) |

**Example:**

```bash
volumeleaders-agent alert create --name "Big trades" --tickers AAPL,MSFT --trade-rank-lte 5
volumeleaders-agent alert create --name "Dark pool sweeps" --sweep --dark-pool --trade-volume-gte 1000000
```

#### `volumeleaders-agent alert delete`

Remove a saved alert configuration by its numeric key. Requires --key with the alert config key (visible in configs output). The deletion is permanent and cannot be undone.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--key` | int | 0 | yes | Alert config key to delete |

**Example:**

```bash
volumeleaders-agent alert delete --key 42
```

#### `volumeleaders-agent alert edit`

Modify an existing alert configuration identified by its numeric key. Requires --key with the alert config key. Specify the fields you want to set; unspecified fields are replaced with their default values.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--ah-dollars-gte` | int | 0 | no | After-hours dollars >= |
| `--ah-rank-lte` | int | 0 | no | After-hours rank <= |
| `--ah-volume-gte` | int | 0 | no | After-hours volume >= |
| `--closing-trade-conditions` | string | 0 | no | Closing trade conditions |
| `--closing-trade-dollars-gte` | int | 0 | no | Closing trade dollars >= |
| `--closing-trade-mult-gte` | int | 0 | no | Closing trade multiplier >= |
| `--closing-trade-rank-lte` | int | 0 | no | Closing trade rank <= |
| `--closing-trade-vcd-gte` | int | 0 | no | Closing trade VCD >= (0/97/98/99/100) |
| `--closing-trade-volume-gte` | int | 0 | no | Closing trade volume >= |
| `--cluster-dollars-gte` | int | 0 | no | Trade cluster dollars >= |
| `--cluster-mult-gte` | int | 0 | no | Trade cluster multiplier >= |
| `--cluster-rank-lte` | int | 0 | no | Trade cluster rank <= |
| `--cluster-vcd-gte` | int | 0 | no | Trade cluster VCD >= (0/97/98/99/100) |
| `--cluster-volume-gte` | int | 0 | no | Trade cluster volume >= |
| `--dark-pool` | bool | false | no | Dark pool filter |
| `--key` | int | 0 | yes | Alert config key to edit |
| `--name` | string | - | no | Alert name (max 50 chars) |
| `--offsetting-print` | bool | false | no | Offsetting print filter |
| `--phantom-print` | bool | false | no | Phantom print filter |
| `--sweep` | bool | false | no | Sweep filter |
| `--ticker-group` | string | AllTickers | no | Ticker group: AllTickers or SelectedTickers |
| `--tickers` | string | - | no | Comma-separated ticker symbols (max 500, used with SelectedTickers) |
| `--total-dollars-gte` | int | 0 | no | Total dollars >= |
| `--total-rank-lte` | int | 0 | no | Total rank <= (0/1/3/10/25/50/100) |
| `--total-volume-gte` | int | 0 | no | Total volume >= |
| `--trade-conditions` | string | 0 | no | Trade conditions (0=N/A, OBH/OBD/OSH/OSD combos) |
| `--trade-dollars-gte` | int | 0 | no | Trade dollars >= (0=N/A, 1000000/10000000/...) |
| `--trade-mult-gte` | int | 0 | no | Trade multiplier >= (0=N/A, 5/10/25/50/100) |
| `--trade-rank-lte` | int | 0 | no | Trade rank <= (0=N/A, 1/3/5/10/25/50/100) |
| `--trade-vcd-gte` | int | 0 | no | Trade VCD >= (0=N/A, 99/100) |
| `--trade-volume-gte` | int | 0 | no | Trade volume >= (0=N/A, 1000000/2000000/5000000/10000000) |

**Example:**

```bash
volumeleaders-agent alert edit --key 42 --name "Updated alert" --trade-rank-lte 3
```

#### `volumeleaders-agent market earnings`

Query the earnings calendar for a date range, showing tickers with earnings dates and associated trade activity counts. Requires --start-date and --end-date (or --days). Outputs compact JSON or CSV/TSV with --format. PREREQUISITES: provide a date range with --days or explicit start and end dates. RECOVERY: if date validation fails, use --days N for the fastest retry or provide both --start-date and --end-date. NEXT STEPS: run trade list for tickers near earnings, then market exhaustion for broader reversal context.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--days` | int | 0 | no | Look back this many days from --end-date or today |
| `--end-date` | string | - | no | End date YYYY-MM-DD (required unless --days is set) |
| `--fields` | string | - | no | Comma-separated fields to include (use 'all' for every field) |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (required unless --days is set) |

**Example:**

```bash
volumeleaders-agent market earnings --days 5
```

#### `volumeleaders-agent market exhaustion`

Query exhaustion scores that indicate overbought or oversold market conditions based on institutional trade clustering patterns. Omit --date to query the current trading day. Outputs compact JSON with rank metrics at different lookback periods.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--date` | string | - | no | Date YYYY-MM-DD (empty for current day) |

**Example:**

```bash
volumeleaders-agent market exhaustion --date 2025-01-15
```

#### `volumeleaders-agent outputschema`

Print machine-readable stdout contracts for executable commands. With no arguments it returns every contract as a JSON array. Pass a command path such as trade list to return one contract. This describes success output only; structured errors are documented by structcli flag errors.

**Example:**

```bash
volumeleaders-agent outputschema
volumeleaders-agent outputschema trade list
```

#### `volumeleaders-agent report dark-pool-20x`

Run the 20x Dark Pool Only report with fixed VolumeLeaders browser-preset filters.

Returns the site-vetted top 100 ranked dark-pool-only preset for trades at least twenty times average size. Use this for unusually large dark-pool prints without adding raw dark-pool filters.

Reports are the recommended entry point for users and LLMs. They expose only safe overrides: tickers, dates, fields, summary grouping, and output format. Do not hand-build low-level filters unless this curated report cannot answer the question; use trade list --preset as the advanced escape hatch.

Defaults to today only. Multi-day broad scans without tickers are rejected to avoid expensive requests and backend timeouts. Results are fetched in browser-sized 100-row pages and ordered by time descending.

RECOVERY: If the report is too broad, add --tickers or query one day at a time. If a custom filter is truly required, run report list to inspect the vetted filters, then use trade list --preset rather than assembling raw filter flags.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--days` | int | 0 | no | Look back this many days from --end-date or today; broad scans require a single day |
| `--end-date` | string | - | no | End date YYYY-MM-DD (default: today) |
| `--fields` | string | - | no | Comma-separated raw Trade fields to include, or omit for compact JSON |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--group-by` | string | ticker | no | Summary grouping (requires --summary): ticker, day, or ticker,day |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (default: today) |
| `--summary` | bool | false | no | Return aggregate metrics instead of individual trades |
| `--tickers` | string | - | no | Comma-separated ticker symbols; use this for multi-day report lookbacks |

**Example:**

```bash
volumeleaders-agent report dark-pool-20x
volumeleaders-agent report dark-pool-20x --tickers SPY,QQQ --days 5
```

#### `volumeleaders-agent report dark-pool-sweeps`

Run the Dark Pool Sweeps report with fixed VolumeLeaders browser-preset filters.

Returns the site-vetted dark pool sweep preset: top 100 ranked dark pool sweeps during premarket and regular trading hours, excluding after-hours, opening, closing, phantom, and signature prints.

Reports are the recommended entry point for users and LLMs. They expose only safe overrides: tickers, dates, fields, summary grouping, and output format. Do not hand-build low-level filters unless this curated report cannot answer the question; use trade list --preset as the advanced escape hatch.

Defaults to today only. Multi-day broad scans without tickers are rejected to avoid expensive requests and backend timeouts. Results are fetched in browser-sized 100-row pages and ordered by time descending.

RECOVERY: If the report is too broad, add --tickers or query one day at a time. If a custom filter is truly required, run report list to inspect the vetted filters, then use trade list --preset rather than assembling raw filter flags.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--days` | int | 0 | no | Look back this many days from --end-date or today; broad scans require a single day |
| `--end-date` | string | - | no | End date YYYY-MM-DD (default: today) |
| `--fields` | string | - | no | Comma-separated raw Trade fields to include, or omit for compact JSON |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--group-by` | string | ticker | no | Summary grouping (requires --summary): ticker, day, or ticker,day |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (default: today) |
| `--summary` | bool | false | no | Return aggregate metrics instead of individual trades |
| `--tickers` | string | - | no | Comma-separated ticker symbols; use this for multi-day report lookbacks |

**Example:**

```bash
volumeleaders-agent report dark-pool-sweeps
volumeleaders-agent report dark-pool-sweeps --tickers AAPL,TSLA --days 5
```

#### `volumeleaders-agent report disproportionately-large`

Run the Disproportionately Large report with fixed VolumeLeaders browser-preset filters.

Returns the site-vetted 5x relative size scan. Use this when the user asks for unusually large prints, disproportionate activity, or trades that are at least five times normal block size.

Reports are the recommended entry point for users and LLMs. They expose only safe overrides: tickers, dates, fields, summary grouping, and output format. Do not hand-build low-level filters unless this curated report cannot answer the question; use trade list --preset as the advanced escape hatch.

Defaults to today only. Multi-day broad scans without tickers are rejected to avoid expensive requests and backend timeouts. Results are fetched in browser-sized 100-row pages and ordered by time descending.

RECOVERY: If the report is too broad, add --tickers or query one day at a time. If a custom filter is truly required, run report list to inspect the vetted filters, then use trade list --preset rather than assembling raw filter flags.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--days` | int | 0 | no | Look back this many days from --end-date or today; broad scans require a single day |
| `--end-date` | string | - | no | End date YYYY-MM-DD (default: today) |
| `--fields` | string | - | no | Comma-separated raw Trade fields to include, or omit for compact JSON |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--group-by` | string | ticker | no | Summary grouping (requires --summary): ticker, day, or ticker,day |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (default: today) |
| `--summary` | bool | false | no | Return aggregate metrics instead of individual trades |
| `--tickers` | string | - | no | Comma-separated ticker symbols; use this for multi-day report lookbacks |

**Example:**

```bash
volumeleaders-agent report disproportionately-large
volumeleaders-agent report disproportionately-large --tickers XLE,XLK --days 5
```

#### `volumeleaders-agent report leveraged-etfs`

Run the Leveraged ETFs report with fixed VolumeLeaders browser-preset filters.

Returns the site-vetted top 100 ranked leveraged ETF preset. Use this for broad ranked activity in leveraged and inverse ETF products without hand-building sector filters.

Reports are the recommended entry point for users and LLMs. They expose only safe overrides: tickers, dates, fields, summary grouping, and output format. Do not hand-build low-level filters unless this curated report cannot answer the question; use trade list --preset as the advanced escape hatch.

Defaults to today only. Multi-day broad scans without tickers are rejected to avoid expensive requests and backend timeouts. Results are fetched in browser-sized 100-row pages and ordered by time descending.

RECOVERY: If the report is too broad, add --tickers or query one day at a time. If a custom filter is truly required, run report list to inspect the vetted filters, then use trade list --preset rather than assembling raw filter flags.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--days` | int | 0 | no | Look back this many days from --end-date or today; broad scans require a single day |
| `--end-date` | string | - | no | End date YYYY-MM-DD (default: today) |
| `--fields` | string | - | no | Comma-separated raw Trade fields to include, or omit for compact JSON |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--group-by` | string | ticker | no | Summary grouping (requires --summary): ticker, day, or ticker,day |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (default: today) |
| `--summary` | bool | false | no | Return aggregate metrics instead of individual trades |
| `--tickers` | string | - | no | Comma-separated ticker symbols; use this for multi-day report lookbacks |

**Example:**

```bash
volumeleaders-agent report leveraged-etfs
volumeleaders-agent report leveraged-etfs --tickers TQQQ,SQQQ --days 5
```

#### `volumeleaders-agent report list`

List curated report commands, their source VolumeLeaders preset names, and their fixed filter configurations. Use these reports before raw trade list filters because they avoid expensive, timeout-prone filter combinations and expose only the safe override surface.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--format` | string | json | no | Output format: json, csv, or tsv |

**Example:**

```bash
volumeleaders-agent report list
```

#### `volumeleaders-agent report offsetting-trades`

Run the Offsetting Trades report with fixed VolumeLeaders browser-preset filters.

Returns the site-vetted offsetting trades preset, excluding normal trading sessions and phantom trades. Use this when the user specifically asks for offsetting trade activity.

Reports are the recommended entry point for users and LLMs. They expose only safe overrides: tickers, dates, fields, summary grouping, and output format. Do not hand-build low-level filters unless this curated report cannot answer the question; use trade list --preset as the advanced escape hatch.

Defaults to today only. Multi-day broad scans without tickers are rejected to avoid expensive requests and backend timeouts. Results are fetched in browser-sized 100-row pages and ordered by time descending.

RECOVERY: If the report is too broad, add --tickers or query one day at a time. If a custom filter is truly required, run report list to inspect the vetted filters, then use trade list --preset rather than assembling raw filter flags.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--days` | int | 0 | no | Look back this many days from --end-date or today; broad scans require a single day |
| `--end-date` | string | - | no | End date YYYY-MM-DD (default: today) |
| `--fields` | string | - | no | Comma-separated raw Trade fields to include, or omit for compact JSON |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--group-by` | string | ticker | no | Summary grouping (requires --summary): ticker, day, or ticker,day |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (default: today) |
| `--summary` | bool | false | no | Return aggregate metrics instead of individual trades |
| `--tickers` | string | - | no | Comma-separated ticker symbols; use this for multi-day report lookbacks |

**Example:**

```bash
volumeleaders-agent report offsetting-trades
volumeleaders-agent report offsetting-trades --tickers SPY,QQQ --days 5
```

#### `volumeleaders-agent report phantom-trades`

Run the Phantom Trades report with fixed VolumeLeaders browser-preset filters.

Returns the site-vetted phantom trades preset, excluding normal trading sessions and offsetting trades. Use this when the user specifically asks for phantom print activity.

Reports are the recommended entry point for users and LLMs. They expose only safe overrides: tickers, dates, fields, summary grouping, and output format. Do not hand-build low-level filters unless this curated report cannot answer the question; use trade list --preset as the advanced escape hatch.

Defaults to today only. Multi-day broad scans without tickers are rejected to avoid expensive requests and backend timeouts. Results are fetched in browser-sized 100-row pages and ordered by time descending.

RECOVERY: If the report is too broad, add --tickers or query one day at a time. If a custom filter is truly required, run report list to inspect the vetted filters, then use trade list --preset rather than assembling raw filter flags.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--days` | int | 0 | no | Look back this many days from --end-date or today; broad scans require a single day |
| `--end-date` | string | - | no | End date YYYY-MM-DD (default: today) |
| `--fields` | string | - | no | Comma-separated raw Trade fields to include, or omit for compact JSON |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--group-by` | string | ticker | no | Summary grouping (requires --summary): ticker, day, or ticker,day |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (default: today) |
| `--summary` | bool | false | no | Return aggregate metrics instead of individual trades |
| `--tickers` | string | - | no | Comma-separated ticker symbols; use this for multi-day report lookbacks |

**Example:**

```bash
volumeleaders-agent report phantom-trades
volumeleaders-agent report phantom-trades --tickers AAPL,MSFT --days 5
```

#### `volumeleaders-agent report rsi-overbought`

Run the RSI Overbought report with fixed VolumeLeaders browser-preset filters.

Returns the site-vetted top 100 ranked RSI overbought preset with trades at least five times average size. Use this when looking for high-rank prints in overbought names.

Reports are the recommended entry point for users and LLMs. They expose only safe overrides: tickers, dates, fields, summary grouping, and output format. Do not hand-build low-level filters unless this curated report cannot answer the question; use trade list --preset as the advanced escape hatch.

Defaults to today only. Multi-day broad scans without tickers are rejected to avoid expensive requests and backend timeouts. Results are fetched in browser-sized 100-row pages and ordered by time descending.

RECOVERY: If the report is too broad, add --tickers or query one day at a time. If a custom filter is truly required, run report list to inspect the vetted filters, then use trade list --preset rather than assembling raw filter flags.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--days` | int | 0 | no | Look back this many days from --end-date or today; broad scans require a single day |
| `--end-date` | string | - | no | End date YYYY-MM-DD (default: today) |
| `--fields` | string | - | no | Comma-separated raw Trade fields to include, or omit for compact JSON |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--group-by` | string | ticker | no | Summary grouping (requires --summary): ticker, day, or ticker,day |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (default: today) |
| `--summary` | bool | false | no | Return aggregate metrics instead of individual trades |
| `--tickers` | string | - | no | Comma-separated ticker symbols; use this for multi-day report lookbacks |

**Example:**

```bash
volumeleaders-agent report rsi-overbought
volumeleaders-agent report rsi-overbought --tickers NVDA,AMD --days 5
```

#### `volumeleaders-agent report rsi-oversold`

Run the RSI Oversold report with fixed VolumeLeaders browser-preset filters.

Returns the site-vetted top 100 ranked RSI oversold preset with trades at least five times average size. Use this when looking for high-rank prints in oversold names.

Reports are the recommended entry point for users and LLMs. They expose only safe overrides: tickers, dates, fields, summary grouping, and output format. Do not hand-build low-level filters unless this curated report cannot answer the question; use trade list --preset as the advanced escape hatch.

Defaults to today only. Multi-day broad scans without tickers are rejected to avoid expensive requests and backend timeouts. Results are fetched in browser-sized 100-row pages and ordered by time descending.

RECOVERY: If the report is too broad, add --tickers or query one day at a time. If a custom filter is truly required, run report list to inspect the vetted filters, then use trade list --preset rather than assembling raw filter flags.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--days` | int | 0 | no | Look back this many days from --end-date or today; broad scans require a single day |
| `--end-date` | string | - | no | End date YYYY-MM-DD (default: today) |
| `--fields` | string | - | no | Comma-separated raw Trade fields to include, or omit for compact JSON |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--group-by` | string | ticker | no | Summary grouping (requires --summary): ticker, day, or ticker,day |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (default: today) |
| `--summary` | bool | false | no | Return aggregate metrics instead of individual trades |
| `--tickers` | string | - | no | Comma-separated ticker symbols; use this for multi-day report lookbacks |

**Example:**

```bash
volumeleaders-agent report rsi-oversold
volumeleaders-agent report rsi-oversold --tickers IWM,QQQ --days 5
```

#### `volumeleaders-agent report top-10-rank`

Run the Top 10 Ranked Trades report with fixed VolumeLeaders browser-preset filters.

Returns the strongest ranked institutional prints using the site-vetted top 10 preset. Use this when the user asks for the highest-conviction trades without needing a broader top 100 scan.

Reports are the recommended entry point for users and LLMs. They expose only safe overrides: tickers, dates, fields, summary grouping, and output format. Do not hand-build low-level filters unless this curated report cannot answer the question; use trade list --preset as the advanced escape hatch.

Defaults to today only. Multi-day broad scans without tickers are rejected to avoid expensive requests and backend timeouts. Results are fetched in browser-sized 100-row pages and ordered by time descending.

RECOVERY: If the report is too broad, add --tickers or query one day at a time. If a custom filter is truly required, run report list to inspect the vetted filters, then use trade list --preset rather than assembling raw filter flags.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--days` | int | 0 | no | Look back this many days from --end-date or today; broad scans require a single day |
| `--end-date` | string | - | no | End date YYYY-MM-DD (default: today) |
| `--fields` | string | - | no | Comma-separated raw Trade fields to include, or omit for compact JSON |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--group-by` | string | ticker | no | Summary grouping (requires --summary): ticker, day, or ticker,day |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (default: today) |
| `--summary` | bool | false | no | Return aggregate metrics instead of individual trades |
| `--tickers` | string | - | no | Comma-separated ticker symbols; use this for multi-day report lookbacks |

**Example:**

```bash
volumeleaders-agent report top-10-rank
volumeleaders-agent report top-10-rank --tickers SPY,QQQ --days 3
```

#### `volumeleaders-agent report top-100-rank`

Run the Top 100 Ranked Trades report with fixed VolumeLeaders browser-preset filters.

Returns the site-vetted top 100 ranked institutional trades preset. Use this before manual TradeRank filters because it preserves the browser preset shape and avoids oversized custom queries.

Reports are the recommended entry point for users and LLMs. They expose only safe overrides: tickers, dates, fields, summary grouping, and output format. Do not hand-build low-level filters unless this curated report cannot answer the question; use trade list --preset as the advanced escape hatch.

Defaults to today only. Multi-day broad scans without tickers are rejected to avoid expensive requests and backend timeouts. Results are fetched in browser-sized 100-row pages and ordered by time descending.

RECOVERY: If the report is too broad, add --tickers or query one day at a time. If a custom filter is truly required, run report list to inspect the vetted filters, then use trade list --preset rather than assembling raw filter flags.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--days` | int | 0 | no | Look back this many days from --end-date or today; broad scans require a single day |
| `--end-date` | string | - | no | End date YYYY-MM-DD (default: today) |
| `--fields` | string | - | no | Comma-separated raw Trade fields to include, or omit for compact JSON |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--group-by` | string | ticker | no | Summary grouping (requires --summary): ticker, day, or ticker,day |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (default: today) |
| `--summary` | bool | false | no | Return aggregate metrics instead of individual trades |
| `--tickers` | string | - | no | Comma-separated ticker symbols; use this for multi-day report lookbacks |

**Example:**

```bash
volumeleaders-agent report top-100-rank
volumeleaders-agent report top-100-rank --tickers NVDA,MSFT --days 5
```

#### `volumeleaders-agent report top-30-rank-10x-99th`

Run the Top 30 Rank, 10x Average Size, 99th Percentile report with fixed VolumeLeaders browser-preset filters.

Returns the site-vetted top 30 ranked preset for trades above ten times average size and in the 99th cumulative distribution percentile. Use this when the user asks for the strongest extreme-size prints.

Reports are the recommended entry point for users and LLMs. They expose only safe overrides: tickers, dates, fields, summary grouping, and output format. Do not hand-build low-level filters unless this curated report cannot answer the question; use trade list --preset as the advanced escape hatch.

Defaults to today only. Multi-day broad scans without tickers are rejected to avoid expensive requests and backend timeouts. Results are fetched in browser-sized 100-row pages and ordered by time descending.

RECOVERY: If the report is too broad, add --tickers or query one day at a time. If a custom filter is truly required, run report list to inspect the vetted filters, then use trade list --preset rather than assembling raw filter flags.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--days` | int | 0 | no | Look back this many days from --end-date or today; broad scans require a single day |
| `--end-date` | string | - | no | End date YYYY-MM-DD (default: today) |
| `--fields` | string | - | no | Comma-separated raw Trade fields to include, or omit for compact JSON |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--group-by` | string | ticker | no | Summary grouping (requires --summary): ticker, day, or ticker,day |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (default: today) |
| `--summary` | bool | false | no | Return aggregate metrics instead of individual trades |
| `--tickers` | string | - | no | Comma-separated ticker symbols; use this for multi-day report lookbacks |

**Example:**

```bash
volumeleaders-agent report top-30-rank-10x-99th
volumeleaders-agent report top-30-rank-10x-99th --tickers XLK,XLF --days 5
```

#### `volumeleaders-agent trade alerts`

Query trade alerts fired on a specific date based on saved alert configurations. Requires --date. Returns alert records matching your configured filters. Outputs compact JSON or CSV/TSV with --format.

Alert configs trigger when trades match thresholds. Threshold names follow the pattern CategoryMetricLTE or CategoryMetricGTE where LTE is maximum rank and GTE is minimum value. Use alert configs to see your configured thresholds.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--date` | string | - | yes | Date YYYY-MM-DD |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--length` | int | 100 | no | Number of results |
| `--order-col` | int | 1 | no | Order column index |
| `--order-dir` | string | desc | no | Order direction |
| `--start` | int | 0 | no | DataTables start offset |

**Example:**

```bash
volumeleaders-agent trade alerts --date 2025-01-15
```

#### `volumeleaders-agent trade cluster-alerts`

Query trade cluster alerts fired on a specific date based on saved alert configurations that target cluster activity. Requires --date. Returns cluster alert records matching your configured filters.

Cluster alert rows use the full cluster-shaped model rather than the compact default from trade clusters. Use trade alerts for individual trade alert rows and this command for cluster-level alert rows.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--date` | string | - | yes | Date YYYY-MM-DD |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--length` | int | 100 | no | Number of results |
| `--order-col` | int | 1 | no | Order column index |
| `--order-dir` | string | desc | no | Order direction |
| `--start` | int | 0 | no | DataTables start offset |

**Example:**

```bash
volumeleaders-agent trade cluster-alerts --date 2025-01-15
```

#### `volumeleaders-agent trade cluster-bombs`

Query trade cluster bombs, which are extreme-magnitude trade clusters that exceed normal institutional activity thresholds. Filterable by ticker, date range, dollar amounts, sector, and cluster bomb rank. Outputs compact JSON by default.

Results are fetched in browser-sized 100-row pages to match VolumeLeaders' frontend behavior. Cluster bombs find sudden aggressive bursts tightly grouped in time and price, with different defaults and rank fields than trade clusters. Use this command when looking for extreme concentration events, not general price-level clustering.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--days` | int | 0 | no | Look back this many days from --end-date or today |
| `--end-date` | string | - | no | End date YYYY-MM-DD (required unless --days is set) |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--max-dollars` | float64 | 30000000000 | no | Maximum dollar value |
| `--max-volume` | int | 2000000000 | no | Maximum volume |
| `--min-dollars` | float64 | 0 | no | Minimum dollar value |
| `--min-volume` | int | 0 | no | Minimum volume |
| `--order-col` | int | 1 | no | Order column index |
| `--order-dir` | string | desc | no | Order direction |
| `--relative-size` | int | 0 | no | Relative size threshold |
| `--sector` | string | - | no | Sector/Industry filter |
| `--security-type` | int | 0 | no | Security type key |
| `--start` | int | 0 | no | DataTables start offset |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (required unless --days is set) |
| `--tickers` | string | - | no | Comma-separated ticker symbols |
| `--trade-cluster-bomb-rank` | int | -1 | no | Trade cluster bomb rank filter |
| `--vcd` | int | 0 | no | VCD filter |

**Example:**

```bash
volumeleaders-agent trade cluster-bombs TSLA --days 3
```

#### `volumeleaders-agent trade clusters`

Query aggregated trade clusters, which group multiple trades in a short window into a single cluster record. Filterable by ticker, date range, dollar amounts, sector, and trade cluster rank. Outputs compact JSON or CSV/TSV with --format.


Results are fetched in browser-sized 100-row pages to match VolumeLeaders' frontend behavior. Use clusters when the question is about price-level concentration, not single prints. This command uses larger default dollar thresholds than ordinary trade list. Use trade cluster-bombs instead when looking for sudden aggressive bursts tightly grouped in time and price.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--days` | int | 0 | no | Look back this many days from --end-date or today |
| `--end-date` | string | - | no | End date YYYY-MM-DD (required unless --days is set) |
| `--fields` | string | - | no | Comma-separated TradeCluster fields to include in output, or 'all' for every field |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--max-dollars` | float64 | 30000000000 | no | Maximum dollar value |
| `--max-price` | float64 | 100000 | no | Maximum price |
| `--max-volume` | int | 2000000000 | no | Maximum volume |
| `--min-dollars` | float64 | 10000000 | no | Minimum dollar value |
| `--min-price` | float64 | 0 | no | Minimum price |
| `--min-volume` | int | 0 | no | Minimum volume |
| `--order-col` | int | 1 | no | Order column index |
| `--order-dir` | string | desc | no | Order direction |
| `--relative-size` | int | 5 | no | Relative size threshold |
| `--sector` | string | - | no | Sector/Industry filter |
| `--security-type` | int | -1 | no | Security type key |
| `--start` | int | 0 | no | DataTables start offset |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (required unless --days is set) |
| `--tickers` | string | - | no | Comma-separated ticker symbols |
| `--trade-cluster-rank` | int | -1 | no | Trade cluster rank filter |
| `--vcd` | int | 0 | no | VCD filter |

**Example:**

```bash
volumeleaders-agent trade clusters AAPL --days 7
```

#### `volumeleaders-agent trade dashboard`

Query a fast ticker dashboard with the same chart-optimized institutional context VolumeLeaders shows in the browser. The dashboard fetches the largest trades, trade clusters, trade levels, and cluster bombs for one ticker in a single JSON object.

Defaults to a 365-day lookback, 10 rows per section, --vcd 0, --relative-size 0, and the same broad trade/session filters used by the browser chart page. Use this command as the first stop for any single-ticker investigation, including institutional levels, largest trades, clustered activity, or sudden bursts, then drill into trade list, trade clusters, trade levels, or trade cluster-bombs only when a section needs deeper pagination, CSV/TSV output, or explicit field selection.

PREREQUISITES: Provide exactly one ticker as a positional argument or with --ticker. Browser authentication must be available.

RECOVERY: If ticker validation fails, use one ticker only. If --count is rejected, use 5, 10, 20, or 50. If date flags conflict, use either --days or --start-date with --end-date.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--ah` | string | 1 | no | After-hours session filter (-1=all, 0=exclude, 1=include) |
| `--closing` | string | 1 | no | Closing trade filter (-1=all, 0=exclude, 1=include) |
| `--conditions` | int | -1 | no | Trade conditions filter |
| `--count` | int | 10 | no | Rows to return per dashboard section (5, 10, 20, or 50) |
| `--dark-pools` | string | -1 | no | Dark pool filter (-1=all, 0=exclude, 1=only) |
| `--days` | int | 0 | no | Look back this many days from --end-date or today |
| `--end-date` | string | - | no | End date YYYY-MM-DD (default: today) |
| `--even-shared` | string | -1 | no | Even shared filter (-1=all, 0=exclude, 1=only) |
| `--late-prints` | string | -1 | no | Late print filter (-1=all, 0=exclude, 1=only) |
| `--market-cap` | int | 0 | no | Market cap filter |
| `--max-dollars` | float64 | 30000000000 | no | Maximum dollar value |
| `--max-price` | float64 | 100000 | no | Maximum price |
| `--max-volume` | int | 2000000000 | no | Maximum volume |
| `--min-dollars` | float64 | 500000 | no | Minimum dollar value |
| `--min-price` | float64 | 0 | no | Minimum price |
| `--min-volume` | int | 0 | no | Minimum volume |
| `--offsetting` | string | 1 | no | Offsetting trade filter (-1=all, 0=exclude, 1=include) |
| `--opening` | string | 1 | no | Opening trade filter (-1=all, 0=exclude, 1=include) |
| `--phantom` | string | 1 | no | Phantom print filter (-1=all, 0=exclude, 1=include) |
| `--premarket` | string | 1 | no | Premarket session filter (-1=all, 0=exclude, 1=include) |
| `--rank-snapshot` | int | -1 | no | Trade rank snapshot filter |
| `--relative-size` | int | 0 | no | Relative size threshold |
| `--rth` | string | 1 | no | Regular trading hours filter (-1=all, 0=exclude, 1=include) |
| `--security-type` | int | -1 | no | Security type key |
| `--sig-prints` | string | -1 | no | Signature print filter (-1=all, 0=exclude, 1=only) |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (default: auto) |
| `--sweeps` | string | -1 | no | Sweep filter (-1=all, 0=exclude, 1=only) |
| `--ticker` | string | - | no | Ticker symbol |
| `--trade-rank` | int | -1 | no | Trade rank filter |
| `--vcd` | int | 0 | no | VCD filter |

**Example:**

```bash
volumeleaders-agent trade dashboard IGV
```

#### `volumeleaders-agent trade level-touches`

Query institutional trade events that occurred at notable price levels for a ticker, showing how the market interacted with key support and resistance zones. Accepts a ticker as positional argument or via --ticker flag. Requires --start-date and --end-date (or --days).

Defaults to --trade-level-rank 5 and --length 50, rejects --length -1, --length 0, and values above 50, and only allows --trade-level-count values of 5, 10, 20, or 50. Use trade levels first to identify significant price zones, then use this command to find events where price revisited those levels.

PREREQUISITES: Provide exactly one ticker and a date range with --start-date and --end-date or --days.

RECOVERY: If --length is rejected, use 1 to 50. If --trade-level-count is rejected, use 5, 10, 20, or 50. If --trade-level-rank is rejected, use 5 or higher. If dates are missing, add --days N for a quick retry.

NEXT STEPS: Compare touched levels with fresh trade list output to see whether recent institutional prints confirm or reject the level.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--days` | int | 0 | no | Look back this many days from --end-date or today |
| `--end-date` | string | - | no | End date YYYY-MM-DD (required unless --days is set) |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--length` | int | 50 | no | Number of results |
| `--max-dollars` | float64 | 30000000000 | no | Maximum dollar value |
| `--max-price` | float64 | 100000 | no | Maximum price |
| `--max-volume` | int | 2000000000 | no | Maximum volume |
| `--min-dollars` | float64 | 500000 | no | Minimum dollar value |
| `--min-price` | float64 | 0 | no | Minimum price |
| `--min-volume` | int | 0 | no | Minimum volume |
| `--order-col` | int | 0 | no | Order column index |
| `--order-dir` | string | desc | no | Order direction |
| `--relative-size` | int | 0 | no | Relative size threshold |
| `--start` | int | 0 | no | DataTables start offset |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (required unless --days is set) |
| `--ticker` | string | - | no | Ticker symbol |
| `--trade-level-count` | int | 50 | no | Number of price levels to include (5, 10, 20, or 50) |
| `--trade-level-rank` | int | 5 | no | Trade level rank filter |
| `--vcd` | int | 0 | no | VCD filter |

**Example:**

```bash
volumeleaders-agent trade level-touches AAPL --days 14
```

#### `volumeleaders-agent trade levels`

Query significant price levels for a ticker, showing historical support and resistance zones identified by institutional trade clustering. Accepts a ticker as positional argument or via --ticker flag. Outputs compact JSON by default.

Defaults to a 365-day lookback when dates are omitted and shares the chart-optimized VolumeLeaders level request used by trade dashboard. This command intentionally exposes a reduced CLI surface: ticker, dates, --trade-level-count, --fields, and --format. For any single-ticker investigation, run trade dashboard TICKER first because it returns trades, clusters, levels, and cluster bombs together; use trade levels only when you need level-only output, CSV/TSV, or explicit field selection. Only --trade-level-count values of 5, 10, 20, or 50 are accepted. Default JSON is compact and omits repetitive ticker metadata and the verbose Dates list; use --fields all or CSV/TSV when raw fields are needed.

PREREQUISITES: Provide exactly one ticker as a positional argument or with --ticker.

RECOVERY: If ticker validation fails, use one ticker only. If --trade-level-count is rejected, use 5, 10, 20, or 50.

NEXT STEPS: Use trade dashboard as the first single-ticker overview, or use trade level-touches with the same ticker and date range to find trades that revisited these levels.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--days` | int | 0 | no | Look back this many days from --end-date or today |
| `--end-date` | string | - | no | End date YYYY-MM-DD (default: today) |
| `--fields` | string | - | no | Comma-separated TradeLevel fields to include in output, or 'all' for every field |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (default: auto) |
| `--ticker` | string | - | no | Ticker symbol |
| `--trade-level-count` | int | 10 | no | Number of price levels to return (5, 10, 20, or 50) |

**Example:**

```bash
volumeleaders-agent trade levels AAPL
```

#### `volumeleaders-agent trade list`

Query individual institutional trades from VolumeLeaders, filterable by ticker, date range, dollar amounts, volume, trade conditions, session type, and trade rank. Supports built-in filter presets (--preset) and watchlist-based filtering (--watchlist). Start with report list for curated preset-backed reports; use trade list when custom raw trade filters are needed. Outputs compact JSON or CSV/TSV with --format; use --summary for aggregate metrics grouped by ticker or day.

Date defaults: 365-day lookback when tickers are provided, today-only without tickers. Preset and watchlist filters do not supply dates. Filter precedence is preset baseline, then watchlist merge, then explicit CLI flags override both.

Default JSON is compact and omits repetitive/internal fields. Use --fields FIELD1,FIELD2, CSV/TSV, or --fields all where supported when raw API fields are needed. --summary returns aggregate JSON with valid --group-by values of ticker, day, or ticker,day; do not combine summary mode with --fields or non-JSON formats.

KEY METRICS

Field                      Meaning
-------------------------  ---------------------------------------------------------------
CumulativeDistribution     Volume percentile, 0 to 1, higher means more accumulation
DollarsMultiplier          Trade dollars relative to average block size
TradeRank                  VL significance rank now, lower is stronger
TradeRankSnapshot          VL significance rank at print time, lower is stronger
TradeClusterRank           Rank for cluster significance, lower is stronger
TradeClusterBombRank       Rank for burst significance, lower is stronger
TradeLevelRank             Rank for level significance, lower is stronger
RelativeSize               Trade size vs normal activity
PercentDailyVolume         Trade volume as percent of average daily volume
VCD                        Volume Confirmation Distribution score
FrequencyLast30TD          Similar-magnitude trade frequency over last 30 trading days
FrequencyLast90TD          Similar-magnitude trade frequency over last 90 trading days
FrequencyLast1CY           Similar-magnitude trade frequency over last calendar year
RSIHour                    Hourly RSI at time of trade
RSIDay                     Daily RSI at time of trade
DarkPool                   Boolean: trade printed on a dark pool
Sweep                      Boolean: trade was a sweep order
LatePrint                  Boolean: trade was a late print
SignaturePrint             Boolean: trade matched a signature print pattern
PhantomPrint               Boolean: trade was a phantom print
InsideBar                  Boolean: bar was an inside bar

Shared trade filters include volume, price, dollars, conditions, VCD, relative size, security type, market cap, trade rank, dark pools, sweeps, late prints, signature prints, even-share prints, and session/event toggles.

PREREQUISITES: Browser authentication. For reproducible scans, pass explicit dates or --days plus tickers, preset, watchlist, or sector filters.

RECOVERY: Multi-day lookups whose effective filters include tickers return the top 10 long-period trades with the same lightweight chart query shape VolumeLeaders uses in the browser. Single-day scans, all-market scans, sector-only presets, and --summary still fetch all matching rows in browser-sized 100-row pages. If --summary rejects --fields or --format, rerun summary as JSON without --fields. If date flags conflict, use either --days or --start-date with --end-date.

NEXT STEPS: Use trade dashboard first for any single-ticker investigation, then trade levels for level-only support/resistance output, trade clusters when prints concentrate near a price, or trade sentiment for leveraged ETF bull/bear context.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--ah` | string | 1 | no | After-hours session filter (-1=all, 0=exclude, 1=include) |
| `--closing` | string | 1 | no | Closing trade filter (-1=all, 0=exclude, 1=include) |
| `--conditions` | int | -1 | no | Trade conditions filter |
| `--dark-pools` | string | -1 | no | Dark pool filter (-1=all, 0=exclude, 1=only) |
| `--days` | int | 0 | no | Look back this many days from --end-date or today |
| `--end-date` | string | - | no | End date YYYY-MM-DD (default: today) |
| `--even-shared` | string | -1 | no | Even shared filter (-1=all, 0=exclude, 1=only) |
| `--fields` | string | - | no | Comma-separated trade fields to include in output |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--group-by` | string | ticker | no | Summary grouping (requires --summary): ticker, day, or ticker,day |
| `--late-prints` | string | -1 | no | Late print filter (-1=all, 0=exclude, 1=only) |
| `--market-cap` | int | 0 | no | Market cap filter |
| `--max-dollars` | float64 | 30000000000 | no | Maximum dollar value |
| `--max-price` | float64 | 100000 | no | Maximum price |
| `--max-volume` | int | 2000000000 | no | Maximum volume |
| `--min-dollars` | float64 | 500000 | no | Minimum dollar value |
| `--min-price` | float64 | 0 | no | Minimum price |
| `--min-volume` | int | 0 | no | Minimum volume |
| `--offsetting` | string | 1 | no | Offsetting trade filter (-1=all, 0=exclude, 1=include) |
| `--opening` | string | 1 | no | Opening trade filter (-1=all, 0=exclude, 1=include) |
| `--order-col` | int | 1 | no | Order column index |
| `--order-dir` | string | desc | no | Order direction |
| `--phantom` | string | 1 | no | Phantom print filter (-1=all, 0=exclude, 1=include) |
| `--premarket` | string | 1 | no | Premarket session filter (-1=all, 0=exclude, 1=include) |
| `--preset` | string | - | no | Apply a built-in filter preset by name; use report list for curated preset-backed reports |
| `--rank-snapshot` | int | -1 | no | Trade rank snapshot filter |
| `--relative-size` | int | 5 | no | Relative size threshold |
| `--rth` | string | 1 | no | Regular trading hours filter (-1=all, 0=exclude, 1=include) |
| `--sector` | string | - | no | Sector/Industry filter |
| `--security-type` | int | -1 | no | Security type key |
| `--sig-prints` | string | -1 | no | Signature print filter (-1=all, 0=exclude, 1=only) |
| `--start` | int | 0 | no | DataTables start offset |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (default: auto) |
| `--summary` | bool | false | no | Return aggregate metrics instead of individual trades |
| `--sweeps` | string | -1 | no | Sweep filter (-1=all, 0=exclude, 1=only) |
| `--tickers` | string | - | no | Comma-separated ticker symbols |
| `--trade-rank` | int | -1 | no | Trade rank filter |
| `--vcd` | int | 97 | no | VCD filter |
| `--watchlist` | string | - | no | Apply filters from a saved watchlist by name |

**Example:**

```bash
volumeleaders-agent trade list AAPL MSFT
volumeleaders-agent trade list --tickers AAPL,MSFT
volumeleaders-agent trade list --tickers NVDA --dark-pools 1 --min-dollars 1000000
volumeleaders-agent trade list --sector Technology --relative-size 10
volumeleaders-agent trade list --preset "Top-100 Rank" --start-date 2025-04-01 --end-date 2025-04-24
volumeleaders-agent trade list --watchlist "Magnificent 7" --start-date 2025-04-01 --end-date 2025-04-24
```

#### `volumeleaders-agent trade sentiment`

Summarize leveraged ETF bull and bear flow by trading day, showing aggregate institutional dollar volume on the bull and bear side. Requires --start-date and --end-date (or --days). Outputs one record per day with bull and bear totals.

This command always queries the combined leveraged ETF sector filter SectorIndustry=X B, classifies bull and bear ETFs locally, and cannot be constrained by ticker or sector flags. Non-standard defaults include --min-dollars 5000000 and --vcd 97; shared --relative-size 5 still applies.

Ratio is bull dollars divided by bear dollars and is null when bear flow is zero. Treat the output as leveraged ETF proxy flow, not signed buy/sell flow for the broader market.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--ah` | string | 1 | no | After-hours session filter (-1=all, 0=exclude, 1=include) |
| `--closing` | string | 1 | no | Closing trade filter (-1=all, 0=exclude, 1=include) |
| `--conditions` | int | -1 | no | Trade conditions filter |
| `--dark-pools` | string | -1 | no | Dark pool filter (-1=all, 0=exclude, 1=only) |
| `--days` | int | 0 | no | Look back this many days from --end-date or today |
| `--end-date` | string | - | no | End date YYYY-MM-DD (required unless --days is set) |
| `--even-shared` | string | -1 | no | Even shared filter (-1=all, 0=exclude, 1=only) |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--late-prints` | string | -1 | no | Late print filter (-1=all, 0=exclude, 1=only) |
| `--market-cap` | int | 0 | no | Market cap filter |
| `--max-dollars` | float64 | 30000000000 | no | Maximum dollar value |
| `--max-price` | float64 | 100000 | no | Maximum price |
| `--max-volume` | int | 2000000000 | no | Maximum volume |
| `--min-dollars` | float64 | 5000000 | no | Minimum dollar value |
| `--min-price` | float64 | 0 | no | Minimum price |
| `--min-volume` | int | 0 | no | Minimum volume |
| `--offsetting` | string | 1 | no | Offsetting trade filter (-1=all, 0=exclude, 1=include) |
| `--opening` | string | 1 | no | Opening trade filter (-1=all, 0=exclude, 1=include) |
| `--phantom` | string | 1 | no | Phantom print filter (-1=all, 0=exclude, 1=include) |
| `--premarket` | string | 1 | no | Premarket session filter (-1=all, 0=exclude, 1=include) |
| `--rank-snapshot` | int | -1 | no | Trade rank snapshot filter |
| `--relative-size` | int | 5 | no | Relative size threshold |
| `--rth` | string | 1 | no | Regular trading hours filter (-1=all, 0=exclude, 1=include) |
| `--security-type` | int | -1 | no | Security type key |
| `--sig-prints` | string | -1 | no | Signature print filter (-1=all, 0=exclude, 1=only) |
| `--start-date` | string | - | no | Start date YYYY-MM-DD (required unless --days is set) |
| `--sweeps` | string | -1 | no | Sweep filter (-1=all, 0=exclude, 1=only) |
| `--trade-rank` | int | -1 | no | Trade rank filter |
| `--vcd` | int | 97 | no | VCD filter |

**Example:**

```bash
volumeleaders-agent trade sentiment --start-date 2025-04-21 --end-date 2025-04-25
```

#### `volumeleaders-agent update`

Download the latest GitHub release for the current platform, verify it against the release checksum file, and replace the running binary atomically. Automatic update notifications are enabled by default, cached for one day, skipped in CI and non-interactive output, and can be disabled with update config.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--force` | bool | false | no | Install the latest release even when the current version is already latest |

**Example:**

```bash
volumeleaders-agent update
volumeleaders-agent update --force
```

#### `volumeleaders-agent update check`

Check the latest GitHub release for the current platform and report whether it is newer than the running binary. This command only reports status and never modifies the installed binary.

**Example:**

```bash
volumeleaders-agent update check
```

#### `volumeleaders-agent update config`

Show updater notification settings, or persist a new automatic notification preference when --check-notifications is set. This updater-specific settings file only controls update checks and does not enable general CLI config loading.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--check-notifications` | bool | true | no | Set automatic update notification preference; true enables notifications, false disables them, omitted only displays current settings |

**Example:**

```bash
volumeleaders-agent update config
volumeleaders-agent update config --check-notifications=false
```

#### `volumeleaders-agent volume ah-institutional`

Query the after-hours institutional volume leaderboard, ranking tickers by total institutional trade volume during after-hours sessions for a given date. Accepts optional ticker positional arguments; also accepts --tickers flag. Requires --date.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--date` | string | - | yes | Date YYYY-MM-DD |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--length` | int | 100 | no | Number of results |
| `--order-col` | int | 1 | no | Order column index |
| `--order-dir` | string | asc | no | Order direction |
| `--start` | int | 0 | no | DataTables start offset |
| `--tickers` | string | - | no | Comma-separated ticker symbols |

**Example:**

```bash
volumeleaders-agent volume ah-institutional --date 2025-01-15
```

#### `volumeleaders-agent volume institutional`

Query the regular-hours institutional volume leaderboard, ranking tickers by total institutional trade volume for a given date. Accepts optional ticker positional arguments to filter results; also accepts --tickers flag. Requires --date. Outputs compact JSON or CSV/TSV with --format. PREREQUISITES: choose a trading date in YYYY-MM-DD format. RECOVERY: if --date is missing or invalid, retry with an explicit trading day. NEXT STEPS: run trade dashboard for interesting single tickers first, then use trade list, trade levels, or trade clusters only when a dashboard section needs deeper detail.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--date` | string | - | yes | Date YYYY-MM-DD |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--length` | int | 100 | no | Number of results |
| `--order-col` | int | 1 | no | Order column index |
| `--order-dir` | string | asc | no | Order direction |
| `--start` | int | 0 | no | DataTables start offset |
| `--tickers` | string | - | no | Comma-separated ticker symbols |

**Example:**

```bash
volumeleaders-agent volume institutional AAPL MSFT --date 2025-01-15
```

#### `volumeleaders-agent volume total`

Query the total volume leaderboard combining all session types, ranking tickers by total institutional trade volume for a given date. Accepts optional ticker positional arguments; also accepts --tickers flag. Requires --date.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--date` | string | - | yes | Date YYYY-MM-DD |
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--length` | int | 100 | no | Number of results |
| `--order-col` | int | 1 | no | Order column index |
| `--order-dir` | string | asc | no | Order direction |
| `--start` | int | 0 | no | DataTables start offset |
| `--tickers` | string | - | no | Comma-separated ticker symbols |

**Example:**

```bash
volumeleaders-agent volume total XLE --date 2025-01-15 --length 20
```

#### `volumeleaders-agent watchlist add-ticker`

Add a ticker symbol to an existing watchlist. Requires --watchlist-key with the watchlist key and --ticker with the symbol to add. The ticker is appended to the watchlist without affecting existing symbols.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--ticker` | string | - | yes | Ticker symbol to add |
| `--watchlist-key` | int | 0 | yes | Watch list key |

**Example:**

```bash
volumeleaders-agent watchlist add-ticker --watchlist-key 1 --ticker NVDA
```

#### `volumeleaders-agent watchlist configs`

List all saved watchlist configurations with their keys and names. Outputs compact JSON or CSV/TSV with --format. Each row shows the watchlist key and name; use the tickers subcommand to view symbols in a specific watchlist.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--format` | string | json | no | Output format: json, csv, or tsv |

**Example:**

```bash
volumeleaders-agent watchlist configs
```

#### `volumeleaders-agent watchlist create`

Create a new watchlist configuration with a name and optional filter settings such as minimum volume, price range, sector, and trade conditions. Requires --name. Use --tickers to specify an explicit ticker list or leave unset for a filter-based watchlist.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--ah-trades` | bool | true | no | Include after-hours trades |
| `--blocks` | bool | true | no | Include block trades |
| `--closing-trades` | bool | true | no | Include closing trades |
| `--dark-pools` | bool | true | no | Include dark pool trades |
| `--late-prints` | bool | true | no | Include late prints |
| `--lit-exchanges` | bool | true | no | Include lit exchange trades |
| `--max-dollars` | float64 | 3e+10 | no | Maximum dollars filter |
| `--max-price` | float64 | 100000 | no | Maximum price filter |
| `--max-trade-rank` | string | -1 | no | Maximum trade rank (-1=all, 1/3/5/10/25/50/100) |
| `--max-volume` | int | 2000000000 | no | Maximum volume filter |
| `--min-dollars` | float64 | 0 | no | Minimum dollars filter |
| `--min-price` | float64 | 0 | no | Minimum price filter |
| `--min-relative-size` | string | 0 | no | Minimum relative size (0/5/10/25/50/100) |
| `--min-vcd` | float64 | 0 | no | Minimum VCD percentile (0-100) |
| `--min-volume` | int | 0 | no | Minimum volume filter |
| `--name` | string | - | yes | Watch list name |
| `--normal-prints` | bool | true | no | Include normal prints |
| `--offsetting-trades` | bool | true | no | Include offsetting trades |
| `--opening-trades` | bool | true | no | Include opening trades |
| `--phantom-trades` | bool | true | no | Include phantom trades |
| `--premarket-trades` | bool | true | no | Include premarket trades |
| `--rsi-overbought-daily` | string | -1 | no | RSI overbought daily (-1=ignore, 0=no, 1=yes) |
| `--rsi-overbought-hourly` | string | -1 | no | RSI overbought hourly (-1=ignore, 0=no, 1=yes) |
| `--rsi-oversold-daily` | string | -1 | no | RSI oversold daily (-1=ignore, 0=no, 1=yes) |
| `--rsi-oversold-hourly` | string | -1 | no | RSI oversold hourly (-1=ignore, 0=no, 1=yes) |
| `--rth-trades` | bool | true | no | Include regular trading hours trades |
| `--sector-industry` | string | - | no | Sector/industry filter (max 100 chars) |
| `--security-type` | string | -1 | no | Security type (-1=all, 1=stocks, 26=ETFs, 4=REITs) |
| `--signature-prints` | bool | true | no | Include signature prints |
| `--sweeps` | bool | true | no | Include sweep trades |
| `--tickers` | string | - | no | Comma-separated ticker symbols (max 500) |
| `--timely-prints` | bool | true | no | Include timely prints |

**Example:**

```bash
volumeleaders-agent watchlist create --name "Tech stocks" --tickers AAPL,MSFT,GOOGL
volumeleaders-agent watchlist create --name "Large caps" --security-type 1 --min-dollars 10000000
```

#### `volumeleaders-agent watchlist delete`

Remove a saved watchlist configuration by its numeric key. Requires --key with the watchlist key (visible in configs output). The deletion is permanent and removes the watchlist and all its tickers.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--key` | int | 0 | yes | Watch list key to delete |

**Example:**

```bash
volumeleaders-agent watchlist delete --key 1
```

#### `volumeleaders-agent watchlist edit`

Modify an existing watchlist configuration identified by its numeric key. Requires --key with the watchlist key. Specify the fields you want to set; unspecified fields are replaced with their default values.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--ah-trades` | bool | true | no | Include after-hours trades |
| `--blocks` | bool | true | no | Include block trades |
| `--closing-trades` | bool | true | no | Include closing trades |
| `--dark-pools` | bool | true | no | Include dark pool trades |
| `--key` | int | 0 | yes | Watch list key to edit |
| `--late-prints` | bool | true | no | Include late prints |
| `--lit-exchanges` | bool | true | no | Include lit exchange trades |
| `--max-dollars` | float64 | 3e+10 | no | Maximum dollars filter |
| `--max-price` | float64 | 100000 | no | Maximum price filter |
| `--max-trade-rank` | string | -1 | no | Maximum trade rank (-1=all, 1/3/5/10/25/50/100) |
| `--max-volume` | int | 2000000000 | no | Maximum volume filter |
| `--min-dollars` | float64 | 0 | no | Minimum dollars filter |
| `--min-price` | float64 | 0 | no | Minimum price filter |
| `--min-relative-size` | string | 0 | no | Minimum relative size (0/5/10/25/50/100) |
| `--min-vcd` | float64 | 0 | no | Minimum VCD percentile (0-100) |
| `--min-volume` | int | 0 | no | Minimum volume filter |
| `--name` | string | - | no | Watch list name |
| `--normal-prints` | bool | true | no | Include normal prints |
| `--offsetting-trades` | bool | true | no | Include offsetting trades |
| `--opening-trades` | bool | true | no | Include opening trades |
| `--phantom-trades` | bool | true | no | Include phantom trades |
| `--premarket-trades` | bool | true | no | Include premarket trades |
| `--rsi-overbought-daily` | string | -1 | no | RSI overbought daily (-1=ignore, 0=no, 1=yes) |
| `--rsi-overbought-hourly` | string | -1 | no | RSI overbought hourly (-1=ignore, 0=no, 1=yes) |
| `--rsi-oversold-daily` | string | -1 | no | RSI oversold daily (-1=ignore, 0=no, 1=yes) |
| `--rsi-oversold-hourly` | string | -1 | no | RSI oversold hourly (-1=ignore, 0=no, 1=yes) |
| `--rth-trades` | bool | true | no | Include regular trading hours trades |
| `--sector-industry` | string | - | no | Sector/industry filter (max 100 chars) |
| `--security-type` | string | -1 | no | Security type (-1=all, 1=stocks, 26=ETFs, 4=REITs) |
| `--signature-prints` | bool | true | no | Include signature prints |
| `--sweeps` | bool | true | no | Include sweep trades |
| `--tickers` | string | - | no | Comma-separated ticker symbols (max 500) |
| `--timely-prints` | bool | true | no | Include timely prints |

**Example:**

```bash
volumeleaders-agent watchlist edit --key 1 --name "Updated watchlist" --tickers AAPL,MSFT
```

#### `volumeleaders-agent watchlist tickers`

Query the ticker symbols belonging to a specific watchlist identified by --watchlist-key. Returns all tickers in the watchlist with their metadata. Outputs compact JSON or CSV/TSV with --format.

**Flags:**

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--format` | string | json | no | Output format: json, csv, or tsv |
| `--watchlist-key` | int | -1 | no | Watch list key (-1 for all) |

**Example:**

```bash
volumeleaders-agent watchlist tickers --watchlist-key 1
```

### Examples

#### volumeleaders-agent alert configs

```bash
volumeleaders-agent alert configs
```

#### volumeleaders-agent alert create

```bash
volumeleaders-agent alert create --name "Big trades" --tickers AAPL,MSFT --trade-rank-lte 5
volumeleaders-agent alert create --name "Dark pool sweeps" --sweep --dark-pool --trade-volume-gte 1000000
```

#### volumeleaders-agent alert delete

```bash
volumeleaders-agent alert delete --key 42
```

#### volumeleaders-agent alert edit

```bash
volumeleaders-agent alert edit --key 42 --name "Updated alert" --trade-rank-lte 3
```

#### volumeleaders-agent market earnings

```bash
volumeleaders-agent market earnings --days 5
```

#### volumeleaders-agent market exhaustion

```bash
volumeleaders-agent market exhaustion --date 2025-01-15
```

#### volumeleaders-agent outputschema

```bash
volumeleaders-agent outputschema
volumeleaders-agent outputschema trade list
```

#### volumeleaders-agent report dark-pool-20x

```bash
volumeleaders-agent report dark-pool-20x
volumeleaders-agent report dark-pool-20x --tickers SPY,QQQ --days 5
```

#### volumeleaders-agent report dark-pool-sweeps

```bash
volumeleaders-agent report dark-pool-sweeps
volumeleaders-agent report dark-pool-sweeps --tickers AAPL,TSLA --days 5
```

#### volumeleaders-agent report disproportionately-large

```bash
volumeleaders-agent report disproportionately-large
volumeleaders-agent report disproportionately-large --tickers XLE,XLK --days 5
```

#### volumeleaders-agent report leveraged-etfs

```bash
volumeleaders-agent report leveraged-etfs
volumeleaders-agent report leveraged-etfs --tickers TQQQ,SQQQ --days 5
```

#### volumeleaders-agent report list

```bash
volumeleaders-agent report list
```

#### volumeleaders-agent report offsetting-trades

```bash
volumeleaders-agent report offsetting-trades
volumeleaders-agent report offsetting-trades --tickers SPY,QQQ --days 5
```

#### volumeleaders-agent report phantom-trades

```bash
volumeleaders-agent report phantom-trades
volumeleaders-agent report phantom-trades --tickers AAPL,MSFT --days 5
```

#### volumeleaders-agent report rsi-overbought

```bash
volumeleaders-agent report rsi-overbought
volumeleaders-agent report rsi-overbought --tickers NVDA,AMD --days 5
```

#### volumeleaders-agent report rsi-oversold

```bash
volumeleaders-agent report rsi-oversold
volumeleaders-agent report rsi-oversold --tickers IWM,QQQ --days 5
```

#### volumeleaders-agent report top-10-rank

```bash
volumeleaders-agent report top-10-rank
volumeleaders-agent report top-10-rank --tickers SPY,QQQ --days 3
```

#### volumeleaders-agent report top-100-rank

```bash
volumeleaders-agent report top-100-rank
volumeleaders-agent report top-100-rank --tickers NVDA,MSFT --days 5
```

#### volumeleaders-agent report top-30-rank-10x-99th

```bash
volumeleaders-agent report top-30-rank-10x-99th
volumeleaders-agent report top-30-rank-10x-99th --tickers XLK,XLF --days 5
```

#### volumeleaders-agent trade alerts

```bash
volumeleaders-agent trade alerts --date 2025-01-15
```

#### volumeleaders-agent trade cluster-alerts

```bash
volumeleaders-agent trade cluster-alerts --date 2025-01-15
```

#### volumeleaders-agent trade cluster-bombs

```bash
volumeleaders-agent trade cluster-bombs TSLA --days 3
```

#### volumeleaders-agent trade clusters

```bash
volumeleaders-agent trade clusters AAPL --days 7
```

#### volumeleaders-agent trade dashboard

```bash
volumeleaders-agent trade dashboard IGV
```

#### volumeleaders-agent trade level-touches

```bash
volumeleaders-agent trade level-touches AAPL --days 14
```

#### volumeleaders-agent trade levels

```bash
volumeleaders-agent trade levels AAPL
```

#### volumeleaders-agent trade list

```bash
volumeleaders-agent trade list AAPL MSFT
volumeleaders-agent trade list --tickers AAPL,MSFT
volumeleaders-agent trade list --tickers NVDA --dark-pools 1 --min-dollars 1000000
volumeleaders-agent trade list --sector Technology --relative-size 10
volumeleaders-agent trade list --preset "Top-100 Rank" --start-date 2025-04-01 --end-date 2025-04-24
volumeleaders-agent trade list --watchlist "Magnificent 7" --start-date 2025-04-01 --end-date 2025-04-24
```

#### volumeleaders-agent trade sentiment

```bash
volumeleaders-agent trade sentiment --start-date 2025-04-21 --end-date 2025-04-25
```

#### volumeleaders-agent update

```bash
volumeleaders-agent update
volumeleaders-agent update --force
```

#### volumeleaders-agent update check

```bash
volumeleaders-agent update check
```

#### volumeleaders-agent update config

```bash
volumeleaders-agent update config
volumeleaders-agent update config --check-notifications=false
```

#### volumeleaders-agent volume ah-institutional

```bash
volumeleaders-agent volume ah-institutional --date 2025-01-15
```

#### volumeleaders-agent volume institutional

```bash
volumeleaders-agent volume institutional AAPL MSFT --date 2025-01-15
```

#### volumeleaders-agent volume total

```bash
volumeleaders-agent volume total XLE --date 2025-01-15 --length 20
```

#### volumeleaders-agent watchlist add-ticker

```bash
volumeleaders-agent watchlist add-ticker --watchlist-key 1 --ticker NVDA
```

#### volumeleaders-agent watchlist configs

```bash
volumeleaders-agent watchlist configs
```

#### volumeleaders-agent watchlist create

```bash
volumeleaders-agent watchlist create --name "Tech stocks" --tickers AAPL,MSFT,GOOGL
volumeleaders-agent watchlist create --name "Large caps" --security-type 1 --min-dollars 10000000
```

#### volumeleaders-agent watchlist delete

```bash
volumeleaders-agent watchlist delete --key 1
```

#### volumeleaders-agent watchlist edit

```bash
volumeleaders-agent watchlist edit --key 1 --name "Updated watchlist" --tickers AAPL,MSFT
```

#### volumeleaders-agent watchlist tickers

```bash
volumeleaders-agent watchlist tickers --watchlist-key 1
```
