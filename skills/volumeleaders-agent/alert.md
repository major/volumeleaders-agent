# alert

Manage saved alert configurations. Alerts trigger when trades or clusters match thresholds.

| Command | Use when | Required | Output |
|---|---|---|---|
| `alert configs` | List alert configs | optional `--fields` | JSON, CSV, TSV |
| `alert delete` | Delete a config | `--key` | JSON |
| `alert create` | Create a config | `--name` | JSON |
| `alert edit` | Replace/update a config | `--key` | JSON |

```bash
volumeleaders-agent alert configs --format csv
volumeleaders-agent alert configs --fields all --format csv
volumeleaders-agent alert create --name "Top Trades" --tickers "AAPL,MSFT" --trade-rank-lte 5
volumeleaders-agent alert edit --key 12345 --trade-rank-lte 3
volumeleaders-agent alert delete --key 12345
```

`alert configs` defaults to a compact LLM-friendly field set: `AlertConfigKey`, `Name`, `Tickers`, `TradeConditions`, `ClosingTradeConditions`, `DarkPool`, `Sweep`, `OffsettingPrint`, and `PhantomPrint`. Use `--fields FieldA,FieldB` to request specific JSON fields, or `--fields all` for the full alert model with every threshold. CSV/TSV headers follow the same field selection.

Full `alert configs` fields include `AlertConfigKey` for edit/delete, `UserKey`, `Name`, `Tickers`, and threshold fields. Threshold names follow `{Category}{Metric}{LTE|GTE}` where LTE is maximum rank and GTE is minimum value.

Create/edit options: `--name` max 50 chars, `--ticker-group AllTickers|SelectedTickers`, `--tickers` comma-separated and auto-selects `SelectedTickers`.

Threshold flags default to `0` disabled unless noted:

| Category | Flags |
|---|---|
| Trade | `--trade-rank-lte`, `--trade-vcd-gte`, `--trade-mult-gte`, `--trade-volume-gte`, `--trade-dollars-gte`, `--trade-conditions` |
| Cluster | `--cluster-rank-lte`, `--cluster-vcd-gte`, `--cluster-mult-gte`, `--cluster-volume-gte`, `--cluster-dollars-gte` |
| Closing | `--closing-trade-rank-lte`, `--closing-trade-vcd-gte`, `--closing-trade-mult-gte`, `--closing-trade-volume-gte`, `--closing-trade-dollars-gte`, `--closing-trade-conditions` |
| Total | `--total-rank-lte`, `--total-volume-gte`, `--total-dollars-gte` |
| After-hours | `--ah-rank-lte`, `--ah-volume-gte`, `--ah-dollars-gte` |
| Booleans | `--dark-pool`, `--sweep`, `--offsetting-print`, `--phantom-print` |

Edit gotcha: unspecified edit flags reset to defaults, so include every value that must remain set.
