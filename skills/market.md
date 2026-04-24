# market

Market-wide data: prices, earnings calendar, exhaustion scores.

## market snapshots

Current prices for all tracked symbols. No flags.

```bash
volumeleaders-agent market snapshots
```

Output: JSON object mapping ticker to price data.

## market earnings

Earnings calendar with institutional activity counts. Shows how much institutional positioning preceded each earnings event.

Required: `--start-date`, `--end-date`

```bash
volumeleaders-agent market earnings --start-date 2025-04-21 --end-date 2025-04-25
```

Output fields: `Ticker`, `Name`, `Sector`, `Industry`, `EarningsDate`, `AfterMarketClose` (bool), `TradeCount`, `TradeClusterCount`, `TradeClusterBombCount`

## market exhaustion

Market exhaustion scores indicating potential trend reversals. Measures when institutional buying/selling pressure may be running out of steam.

Optional: `--date` (omit for current day)

```bash
volumeleaders-agent market exhaustion
```

Output fields:

| Field | Meaning |
|---|---|
| `DateKey` | Trading date (YYYYMMDD format) |
| `ExhaustionScoreRank` | Raw exhaustion score (unbounded, higher = more exhausted) |
| `ExhaustionScoreRank30Day` | Rank within last ~21 trading days (1 = most exhausted) |
| `ExhaustionScoreRank90Day` | Rank within last ~63 trading days |
| `ExhaustionScoreRank365Day` | Rank within last ~252 trading days |

Interpreting: **lower rank = stronger exhaustion signal**. When multiple timeframes rank low simultaneously, the reversal signal is reinforced. High ranks across all timeframes means the trend likely has room to continue.
