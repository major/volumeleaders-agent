# alert

Manage saved alert configurations. Alerts trigger when trades/clusters match threshold criteria.

## alert configs

List all saved alert configurations.

Optional: `--format json|csv|tsv`

```bash
volumeleaders-agent alert configs
volumeleaders-agent alert configs --format csv
```

Output fields: `AlertConfigKey` (use as `--key` for edit/delete), `Name`, `Tickers`, plus threshold fields. Threshold naming pattern: `{Category}{Metric}{LTE|GTE}` where LTE = max rank (lower rank = more significant) and GTE = minimum value (higher = bigger trade).

Output format: `alert configs` defaults to JSON and supports CSV/TSV. Create/edit/delete responses remain JSON-only.

## alert delete

Required: `--key` (AlertConfigKey from `alert configs`)

```bash
volumeleaders-agent alert delete --key 12345
```

## alert create

Required: `--name` (max 50 chars)
Optional: `--ticker-group` (AllTickers or SelectedTickers, default AllTickers), `--tickers` (comma-separated, auto-sets SelectedTickers when provided)

Threshold flags (all default 0 = disabled):

| Category | Flags |
|---|---|
| **Trade** | `--trade-rank-lte` (0/1/3/5/10/25/50/100), `--trade-vcd-gte` (0/99/100), `--trade-mult-gte` (0/5/10/25/50/100), `--trade-volume-gte`, `--trade-dollars-gte`, `--trade-conditions` |
| **Cluster** | `--cluster-rank-lte`, `--cluster-vcd-gte` (0/97/98/99/100), `--cluster-mult-gte`, `--cluster-volume-gte`, `--cluster-dollars-gte` |
| **Closing** | `--closing-trade-rank-lte`, `--closing-trade-vcd-gte` (0/97/98/99/100), `--closing-trade-mult-gte`, `--closing-trade-volume-gte`, `--closing-trade-dollars-gte` |
| **Total** | `--total-rank-lte` (0/1/3/10/25/50/100), `--total-volume-gte`, `--total-dollars-gte` |
| **After-hours** | `--ah-rank-lte`, `--ah-volume-gte`, `--ah-dollars-gte` |
| **Booleans** | `--dark-pool`, `--sweep`, `--offsetting-print`, `--phantom-print` (all default false) |

```bash
volumeleaders-agent alert create --name "Top Trades" --tickers "AAPL,MSFT" --trade-rank-lte 5
```

## alert edit

Required: `--key` (AlertConfigKey)
All flags from `alert create` available. `--name` is optional for edit. Unspecified flags reset to their defaults.

```bash
volumeleaders-agent alert edit --key 12345 --trade-rank-lte 3
```
