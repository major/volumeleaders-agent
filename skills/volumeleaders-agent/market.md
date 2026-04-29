# market

Market-wide prices, earnings calendar, and exhaustion scores.

| Command | Use when | Required | Output |
|---|---|---|---|
| `market snapshots` | Get current prices for all tracked symbols | none | JSON object |
| `market earnings` | Find earnings with prior institutional activity | date range or `--days` | JSON, CSV, TSV |
| `market exhaustion` | Check reversal/exhaustion signals | none | JSON |

```bash
volumeleaders-agent market snapshots
volumeleaders-agent market earnings --start-date 2026-04-21 --end-date 2026-04-28 --format csv
volumeleaders-agent market earnings --days 7
volumeleaders-agent market exhaustion --date 2026-04-28
```

`market earnings` fields: `Ticker`, `Name`, `Sector`, `Industry`, `EarningsDate`, `AfterMarketClose`, `TradeCount`, `TradeClusterCount`, `TradeClusterBombCount`.

`market exhaustion` optional flag: `--date`, omitted for current day. Fields: `DateKey`, `ExhaustionScoreRank`, `ExhaustionScoreRank30Day`, `ExhaustionScoreRank90Day`, `ExhaustionScoreRank365Day`. Lower rank = stronger exhaustion signal. Multiple low ranks across timeframes reinforce reversal risk.
