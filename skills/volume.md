# volume

Volume leaderboards ranking stocks by trading activity. All three subcommands share the same flags and return Trade model objects.

Required: `--date` (YYYY-MM-DD)
Optional: `--tickers`, pagination (`--length 100 --order-col 1 --order-dir asc`)

Note: default sort direction is `asc` (ascending), unlike trade commands which default to `desc`.

## volume institutional

Stocks ranked by institutional dollar volume for a date. The primary "what are institutions trading today?" command.

```bash
volumeleaders-agent volume institutional --date 2025-04-23
```

## volume ah-institutional

Stocks ranked by after-hours institutional activity. After-hours institutional trades often precede next-day moves.

```bash
volumeleaders-agent volume ah-institutional --date 2025-04-23
```

## volume total

Overall volume leaders across all trade types. Compare institutional vs total volume.

```bash
volumeleaders-agent volume total --date 2025-04-23
```

Output fields (all three): same Trade model as `trade list`. Key fields for volume analysis: `Ticker`, `Name`, `Sector`, `Industry`, `TotalInstitutionalDollars`, `TotalInstitutionalDollarsRank`, `TotalInstitutionalVolume`, `AHInstitutionalDollars`, `AHInstitutionalDollarsRank`, `AHInstitutionalVolume`, `ClosingTradeDollars`, `ClosingTradeDollarsRank`, `ClosingTradeVolume`, `TotalDollars`, `TotalDollarsRank`, `TotalVolume`, `AverageDailyVolume`, `PercentDailyVolume`, `ClosePrice`, `RSIHour`, `RSIDay`
